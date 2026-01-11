package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"portfolio-manager/internal/backup"
	"portfolio-manager/internal/portfolio"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// CLI handles command line interface
type CLI struct {
	commands      map[string]*Command
	backupSvc     backup.BackupService
	defaultDBPath string
	baseURL       string
	httpClient    *http.Client
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{
		commands:      make(map[string]*Command),
		backupSvc:     backup.NewService(),
		defaultDBPath: "./portfolio-manager.db",
		baseURL:       "http://localhost:8080",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	cli.registerCommands()
	return cli
}

// SetBaseURL allows overriding the default API base URL
func (c *CLI) SetBaseURL(url string) {
	if url != "" {
		c.baseURL = url
	}
}

func (c *CLI) isRemoteURL() bool {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host != "localhost" && host != "127.0.0.1" && host != ""
}

// registerCommands registers all available commands
func (c *CLI) registerCommands() {
	// Local commands
	c.commands["backup"] = &Command{
		Name:        "backup",
		Description: "Create a backup of the local database (and optionally the data folder)",
		Handler:     c.handleBackup,
	}

	c.commands["restore-from-backup"] = &Command{
		Name:        "restore-from-backup",
		Description: "Restore local database from a backup",
		Handler:     c.handleRestore,
	}

	c.commands["version"] = &Command{
		Name:        "version",
		Description: "Show version information",
		Handler:     c.handleVersion,
	}

	// Remote (API) commands
	c.commands["position-list"] = &Command{
		Name:        "position-list",
		Description: "List all positions via REST API",
		Handler:     c.handlePositionList,
	}

	c.commands["position-delete"] = &Command{
		Name:        "position-delete",
		Description: "Delete a specific position via REST API",
		Handler:     c.handlePositionDelete,
	}
}

// ParseAndExecute parses command line arguments and executes the appropriate command
func (c *CLI) ParseAndExecute(args []string) error {
	if len(args) < 2 {
		c.ShowHelp()
		return fmt.Errorf("no command specified")
	}

	commandName := args[1]

	// Handle special case for version flags
	if commandName == "-v" || commandName == "--version" {
		return c.handleVersion(args[2:])
	}

	// Handle help flags
	if commandName == "-h" || commandName == "--help" || commandName == "help" {
		c.ShowHelp()
		return nil
	}

	command, exists := c.commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	return command.Handler(args[2:])
}

// RunCLI handles CLI subcommands for compatibility with main.go
func RunCLI(args []string, baseURL string) error {
	if len(args) == 0 {
		PrintUsage()
		return fmt.Errorf("no command provided")
	}

	c := NewCLI()
	c.SetBaseURL(baseURL)

	command, exists := c.commands[args[0]]
	if !exists {
		PrintUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}

	return command.Handler(args[1:])
}

// --- Local Command Handlers ---

func (c *CLI) handleBackup(args []string) error {
	var source, uri, user, password string
	var includeData bool

	if c.isRemoteURL() {
		return fmt.Errorf("backup command only works for a local portfolio-manager server. To backup a remote server, run this command locally on that server")
	}

	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.StringVar(&source, "source", "local", "Backup source (local, gdrive, nextcloud)")
	fs.StringVar(&uri, "uri", "", "File location or URL")
	fs.StringVar(&user, "user", "", "Username for remote sources")
	fs.StringVar(&password, "password", "", "Password for remote sources")
	fs.BoolVar(&includeData, "include-data", true, "Include the data folder (funds flow, reports etc.) in the backup")
	if err := fs.Parse(args); err != nil {
		return err
	}

	running, err := c.backupSvc.IsApplicationRunning(c.baseURL)
	if err != nil {
		fmt.Printf("Warning: Could not check if application is running: %v\n", err)
	} else if running {
		return fmt.Errorf("application appears to be running. Please stop the service before backing up to ensure database consistency")
	}

	if _, err := os.Stat(c.defaultDBPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found at %s. Make sure the application has been run at least once", c.defaultDBPath)
	}

	size, err := c.backupSvc.GetBackupSize(c.defaultDBPath, includeData)
	if err != nil {
		return fmt.Errorf("failed to calculate backup size: %w", err)
	}

	fmt.Printf("Backup size will be approximately: %s\n", formatFileSize(size))
	fmt.Println("WARNING: Backups from older versions might not be compatible with the current version.")
	fmt.Print("Do you want to proceed with the backup? (y/N): ")

	if !c.promptForConfirmation() {
		fmt.Println("Backup cancelled.")
		return nil
	}

	config := backup.BackupConfig{
		Source:      source,
		URI:         uri,
		User:        user,
		Password:    password,
		IncludeData: includeData,
	}

	return c.backupSvc.Backup(context.Background(), c.defaultDBPath, config)
}

func (c *CLI) handleRestore(args []string) error {
	var source, uri, user, password string

	if c.isRemoteURL() {
		return fmt.Errorf("restore-from-backup command only works for a local portfolio-manager server. To restore a remote server, run this command locally on that server")
	}

	fs := flag.NewFlagSet("restore-from-backup", flag.ContinueOnError)
	fs.StringVar(&source, "source", "local", "Backup source (local, gdrive, nextcloud)")
	fs.StringVar(&uri, "uri", "", "File location or URL")
	fs.StringVar(&user, "user", "", "Username for remote sources")
	fs.StringVar(&password, "password", "", "Password for remote sources")
	if err := fs.Parse(args); err != nil {
		return err
	}

	running, err := c.backupSvc.IsApplicationRunning(c.baseURL)
	if err != nil {
		fmt.Printf("Warning: Could not check if application is running: %v\n", err)
	} else if running {
		return fmt.Errorf("application appears to be running. Please stop the service before restoring")
	}

	if _, err := os.Stat(c.defaultDBPath); err == nil {
		fmt.Printf("Existing database found at %s. This will completely replace it. Continue? (y/N): ", c.defaultDBPath)
		if !c.promptForConfirmation() {
			fmt.Println("Restore cancelled.")
			return nil
		}
	}

	config := backup.BackupConfig{
		Source:   source,
		URI:      uri,
		User:     user,
		Password: password,
	}

	return c.backupSvc.Restore(context.Background(), c.defaultDBPath, config)
}

func (c *CLI) handleVersion(args []string) error {
	version, err := getVersion()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	fmt.Printf("Portfolio Manager version: %s\n", strings.TrimSpace(version))
	return nil
}

// --- Remote (API) Command Handlers ---

func (c *CLI) handlePositionList(args []string) error {
	u := fmt.Sprintf("%s/api/v1/portfolio/positions", c.baseURL)

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var positions []*portfolio.Position
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("%-15s %-10s %-15s %-12s %-15s %-15s %-15s\n",
		"Book", "Ticker", "Asset Class", "Quantity", "Avg Price", "Market Value", "P&L")
	fmt.Println("--------------------------------------------------------------------------------------------")

	for _, pos := range positions {
		fmt.Printf("%-15s %-10s %-15s %-12.2f %-15.2f %-15.2f %-15.2f\n",
			pos.Book, pos.Ticker, pos.AssetClass, pos.Qty, pos.AvgPx, pos.Mv, pos.PnL)
	}

	fmt.Printf("\nTotal positions: %d\n", len(positions))
	return nil
}

func (c *CLI) handlePositionDelete(args []string) error {
	fs := flag.NewFlagSet("position-delete", flag.ContinueOnError)
	book := fs.String("book", "", "Book name (required)")
	ticker := fs.String("ticker", "", "Ticker symbol (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if *book == "" && len(remaining) > 0 {
		*book = remaining[0]
	}
	if *ticker == "" && len(remaining) > 1 {
		*ticker = remaining[1]
	}

	if *book == "" || *ticker == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s position-delete --book <book> --ticker <ticker>\n", executable())
		return fmt.Errorf("both book and ticker are required")
	}

	u, err := url.Parse(fmt.Sprintf("%s/api/v1/portfolio/position", c.baseURL))
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("book", *book)
	q.Set("ticker", *ticker)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully deleted position for book: %s, ticker: %s\n", *book, *ticker)
	return nil
}

// --- Utilities ---

func (c *CLI) promptForConfirmation() bool {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "y" || response == "yes"
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getVersion() (string, error) {
	versionPaths := []string{"VERSION", "../../VERSION", "../../../VERSION"}
	for _, path := range versionPaths {
		if content, err := os.ReadFile(path); err == nil {
			return string(content), nil
		}
	}
	return "unknown", nil
}

func executable() string {
	base := filepath.Base(os.Args[0])
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (c *CLI) ShowHelp() {
	PrintUsage()
}

func PrintUsage() {
	exe := executable()
	fmt.Fprintf(os.Stderr, `Portfolio Manager CLI

Usage:
  %[1]s [global options] <command> [flags]

Local Commands:
  backup                  Create a backup of the database and data folder
  restore-from-backup     Restore database from a backup
  version                 Show version information

Remote API Commands:
  position-list           List all positions
  position-delete         Delete a specific position (--book B --ticker T)

Global Flags:
  --url=<base-url>        Base URL of the API (default: http://localhost:8080)
  -v, --version           Show version information

Examples:
  %[1]s backup
  %[1]s position-list
  %[1]s position-delete --book main --ticker AAPL
`, exe)
}
