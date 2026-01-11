package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GDriveBackupSource implements BackupSource for Google Drive
type GDriveBackupSource struct {
	credentialsFile string
	targetPath      string
}

// NewGDriveBackupSource creates a new Google Drive backup source
func NewGDriveBackupSource(credentialsFile string, targetPath string) *GDriveBackupSource {
	if targetPath == "" {
		targetPath = "portfolio-manager-go/backups"
	}
	return &GDriveBackupSource{
		credentialsFile: credentialsFile,
		targetPath:      targetPath,
	}
}

// GetName returns the name of the backup source
func (g *GDriveBackupSource) GetName() string {
	return "gdrive"
}

// SetTargetPath updates the target folder path on Google Drive
func (g *GDriveBackupSource) SetTargetPath(path string) {
	if path == "." || path == "" {
		g.targetPath = "root"
	} else {
		g.targetPath = path
	}
}

// Upload uploads data to Google Drive
func (g *GDriveBackupSource) Upload(ctx context.Context, reader io.Reader, filename string) error {
	srv, err := g.getService(ctx)
	if err != nil {
		return err
	}

	folderID, err := g.getOrCreatePath(ctx, srv, g.targetPath)
	if err != nil {
		return fmt.Errorf("failed to get/create target path: %w", err)
	}

	fmt.Printf("GDrive: Uploading '%s' to folder ID: %s\n", filename, folderID)

	// Check if file already exists in that folder to update it or create new
	existingFile, err := g.findFile(ctx, srv, folderID, filename)
	if err != nil {
		return err
	}

	f := &drive.File{
		Name:    filename,
		Parents: []string{folderID},
	}

	if existingFile != "" {
		fmt.Printf("GDrive: Updating existing file with ID: %s\n", existingFile)
		_, err = srv.Files.Update(existingFile, f).
			Context(ctx).
			SupportsAllDrives(true). // Use SupportsAllDrives for Shared Drive compatibility
			Media(reader).Do()
		if err != nil {
			return fmt.Errorf("failed to update file on GDrive: %w", err)
		}
	} else {
		fmt.Printf("GDrive: Creating new file '%s'\n", filename)
		_, err = srv.Files.Create(f).
			Context(ctx).
			SupportsAllDrives(true). // Use SupportsAllDrives for Shared Drive compatibility
			Media(reader).Do()
		if err != nil {
			return fmt.Errorf("failed to create file on GDrive: %w", err)
		}
	}

	return nil
}

// Download downloads data from Google Drive
func (g *GDriveBackupSource) Download(ctx context.Context, filename string) (io.Reader, error) {
	srv, err := g.getService(ctx)
	if err != nil {
		return nil, err
	}

	folderID, err := g.getOrCreatePath(ctx, srv, g.targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get target path: %w", err)
	}

	fileID, err := g.findFile(ctx, srv, folderID, filename)
	if err != nil {
		return nil, err
	}
	if fileID == "" {
		// If the specific file is not found, try to find the latest file in the folder
		fmt.Printf("GDrive: File '%s' not found, searching for the latest backup in '%s'...\n", filename, g.targetPath)
		latestID, latestName, err := g.findLatestFile(ctx, srv, folderID)
		if err != nil {
			return nil, fmt.Errorf("file not found and failed to find latest: %w", err)
		}
		if latestID == "" {
			return nil, fmt.Errorf("file not found and no other backups found in: %s", g.targetPath)
		}
		fmt.Printf("GDrive: Found latest backup: %s (ID: %s)\n", latestName, latestID)
		fileID = latestID
	}

	resp, err := srv.Files.Get(fileID).
		Context(ctx).
		SupportsAllDrives(true). // Support finding files in Shared Drives
		Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return resp.Body, nil
}

func (g *GDriveBackupSource) getService(ctx context.Context) (*drive.Service, error) {
	if g.credentialsFile == "" {
		return nil, fmt.Errorf("GDrive credentials file path is required")
	}

	credentials, err := os.ReadFile(g.credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	// Detect credential type manually
	var creds struct {
		Type string `json:"type"`
	}
	json.Unmarshal(credentials, &creds)

	if creds.Type == "service_account" {
		config, err := google.JWTConfigFromJSON(credentials, drive.DriveScope)
		if err != nil {
			return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
		}
		client := config.Client(ctx)
		srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve Drive client: %w", err)
		}
		fmt.Printf("GDrive: Authenticated as service account: %s\n", config.Email)
		return srv, nil
	}

	// Handle as OAuth2 User credentials
	config, err := google.ConfigFromJSON(credentials, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials file as OAuth2: %w", err)
	}

	client := g.getOAuthClient(ctx, config)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %w", err)
	}

	fmt.Println("GDrive: Authenticated as user via OAuth2")

	return srv, nil
}

// getOAuthClient retrieves a token, saves the token, then returns the generated client.
func (g *GDriveBackupSource) getOAuthClient(ctx context.Context, config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := g.tokenFromFile(tokFile)
	if err != nil {
		tok = g.getTokenFromWeb(ctx, config)
		g.saveToken(tokFile, tok)
	} else {
		// If token is expired or invalid, refresh it now while we have the chance to save it
		if !tok.Valid() {
			ts := config.TokenSource(ctx, tok)
			newTok, err := ts.Token()
			if err == nil {
				if newTok.AccessToken != tok.AccessToken {
					fmt.Println("GDrive: Token refreshed, updating token.json")
					g.saveToken(tokFile, newTok)
					tok = newTok
				}
			} else {
				fmt.Printf("GDrive: Failed to refresh token: %v. Requesting new token...\n", err)
				tok = g.getTokenFromWeb(ctx, config)
				g.saveToken(tokFile, tok)
			}
		}
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func (g *GDriveBackupSource) getTokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	// Start a local server to receive the authorization code.
	// Google has deprecated the out-of-band (OOB) flow, so loopback is now required.
	port := 8888
	if config.RedirectURL != "" {
		// Use the redirect URL from credentials.json if available
		u, err := url.Parse(config.RedirectURL)
		if err == nil && u.Port() != "" {
			fmt.Sscanf(u.Port(), "%d", &port)
		}
	} else {
		config.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
	}

	// Create a channel to receive the code
	codeCh := make(chan string)

	// Use a custom ServeMux to avoid global state
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Authorization successful! You can close this tab and return to the terminal.")
			codeCh <- code
		} else {
			fmt.Fprintf(w, "Authorization failed! No code found in response.")
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("GDrive: Warning - failed to start redirect server on port %d: %v\n", port, err)
		}
	}()

	fmt.Printf("\nGDrive: Google has blocked the out-of-band (OOB) flow for desktop apps.\n")
	fmt.Printf("Please follow these steps to authenticate:\n\n")

	fmt.Printf("1. Ensure the redirect URL is reachable: %s\n", config.RedirectURL)
	if strings.Contains(config.RedirectURL, "localhost") {
		fmt.Println("   Note: If you are on a REMOTE server, you must tunnel traffic from your local machine:")
		fmt.Printf("   A) Using SSH:   ssh -L %d:localhost:%d [user]@[remote-host]\n", port, port)
		fmt.Printf("   B) Using socat: socat TCP4-LISTEN:%d,fork TCP4:[remote-host]:%d\n", port, port)
	} else {
		fmt.Println("   Note: You are using a custom redirect domain. Ensure your firewall allows incoming traffic on the specified port.")
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("\n2. Open the following link in your browser:")
	fmt.Printf("%v\n\n", authURL)

	fmt.Printf("3. After you authorize, Google will redirect to %s and this app will receive the code automatically.\n", config.RedirectURL)
	fmt.Println("   Waiting for authorization (timeout in 5 minutes)...")

	var authCode string
	select {
	case authCode = <-codeCh:
		// Shutdown the server
		_ = server.Shutdown(ctx)
		fmt.Println("GDrive: Authorization code received successfully.")
	case <-time.After(5 * time.Minute):
		fmt.Println("GDrive: Timed out waiting for authorization code.")
		_ = server.Shutdown(ctx)
		os.Exit(1)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		fmt.Printf("Unable to retrieve token from web: %v", err)
		os.Exit(1)
	}
	return tok
}

// tokenFromFile retrieves a token from a local file.
func (g *GDriveBackupSource) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
func (g *GDriveBackupSource) saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Unable to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (g *GDriveBackupSource) getOrCreatePath(ctx context.Context, srv *drive.Service, path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	parentID := "root"

	fmt.Printf("GDrive: Resolving path: %s\n", path)

	for _, name := range parts {
		if name == "" {
			continue
		}

		// Search for folders including those in Shared Drives
		query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", name)
		if parentID != "root" {
			query = fmt.Sprintf("%s and '%s' in parents", query, parentID)
		}

		res, err := srv.Files.List().
			Context(ctx).
			Q(query).
			SupportsAllDrives(true).         // Required for Shared Drives
			IncludeItemsFromAllDrives(true). // Find folders shared with service account
			Fields("files(id, name)").Do()
		if err != nil {
			return "", fmt.Errorf("failed to list folders: %w", err)
		}
		var id string
		if len(res.Files) > 0 {
			id = res.Files[0].Id
			fmt.Printf("GDrive: Found folder '%s' with ID: %s (Parent: %s)\n", name, id, parentID)
		} else {
			// Create folder
			f := &drive.File{
				Name:     name,
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{parentID},
			}
			newRes, err := srv.Files.Create(f).
				Context(ctx).
				SupportsAllDrives(true). // Create inside Shared Drive if parent is there
				Do()
			if err != nil {
				return "", fmt.Errorf("failed to create folder %s: %w", name, err)
			}
			id = newRes.Id
			fmt.Printf("GDrive: Created folder '%s' with ID: %s (Parent: %s)\n", name, id, parentID)
		}
		parentID = id
	}

	return parentID, nil
}

func (g *GDriveBackupSource) findFile(ctx context.Context, srv *drive.Service, parentID, name string) (string, error) {
	query := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", name, parentID)
	res, err := srv.Files.List().
		Context(ctx).
		Q(query).
		SupportsAllDrives(true).         // Search inside Shared Drives
		IncludeItemsFromAllDrives(true). // Include items from Shared Drives
		Fields("files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}
	if len(res.Files) > 0 {
		return res.Files[0].Id, nil
	}
	return "", nil
}

func (g *GDriveBackupSource) findLatestFile(ctx context.Context, srv *drive.Service, parentID string) (string, string, error) {
	// Query for files in the parent folder, sorted by name descending
	// Assuming backups have timestamped names like portfolio-manager-backup-YYYYMMDD-HHMMSS.tar.gz
	query := fmt.Sprintf("'%s' in parents and mimeType != 'application/vnd.google-apps.folder' and trashed = false", parentID)
	res, err := srv.Files.List().
		Context(ctx).
		Q(query).
		OrderBy("name desc").
		PageSize(1).
		SupportsAllDrives(true).         // Search inside Shared Drives
		IncludeItemsFromAllDrives(true). // Include items from Shared Drives
		Fields("files(id, name)").Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to list files for latest search: %w", err)
	}

	if len(res.Files) > 0 {
		return res.Files[0].Id, res.Files[0].Name, nil
	}
	return "", "", nil
}
