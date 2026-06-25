package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"portfolio-manager/internal/backup"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/rdata"

	"gopkg.in/yaml.v2"
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

	c.commands["refdata-list"] = &Command{
		Name:        "refdata-list",
		Description: "List reference data via REST API",
		Handler:     c.handleReferenceDataList,
	}

	c.commands["refdata-get"] = &Command{
		Name:        "refdata-get",
		Description: "Get one reference data record via REST API",
		Handler:     c.handleReferenceDataGet,
	}

	c.commands["refdata-add"] = &Command{
		Name:        "refdata-add",
		Description: "Add reference data via REST API",
		Handler:     c.handleReferenceDataAdd,
	}

	c.commands["refdata-update"] = &Command{
		Name:        "refdata-update",
		Description: "Update reference data via REST API",
		Handler:     c.handleReferenceDataUpdate,
	}

	c.commands["refdata-delete"] = &Command{
		Name:        "refdata-delete",
		Description: "Delete reference data via REST API",
		Handler:     c.handleReferenceDataDelete,
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

func (c *CLI) handleReferenceDataList(args []string) error {
	fs := flag.NewFlagSet("refdata-list", flag.ContinueOnError)
	idsArg := fs.String("id", "", "Comma-separated reference data ids to include")
	assetClass := fs.String("asset-class", "", "Filter by asset class")
	assetSubClass := fs.String("asset-sub-class", "", "Filter by asset sub-class")
	category := fs.String("category", "", "Filter by category")
	asJSON := fs.Bool("json", false, "Print filtered reference data as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ids := normalizeCLIIDs(splitCSVAndArgs(*idsArg, fs.Args()))
	refData, err := c.fetchReferenceData()
	if err != nil {
		return err
	}

	filtered := filterReferenceData(refData, ids, *assetClass, *assetSubClass, *category)
	if *asJSON {
		return printJSON(filtered)
	}

	printReferenceDataTable(filtered)
	return nil
}

func (c *CLI) handleReferenceDataGet(args []string) error {
	fs := flag.NewFlagSet("refdata-get", flag.ContinueOnError)
	id := fs.String("id", "", "Reference data id")
	asJSON := fs.Bool("json", false, "Print reference data as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if *id == "" && len(remaining) > 0 {
		*id = remaining[0]
	}
	if *id == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s refdata-get --id <id>\n", executable())
		return fmt.Errorf("id is required")
	}

	refData, err := c.fetchReferenceData()
	if err != nil {
		return err
	}
	ref, ok := refData[strings.ToUpper(*id)]
	if !ok {
		return fmt.Errorf("reference data not found for id %s", *id)
	}

	if *asJSON {
		return printJSON(ref)
	}
	printReferenceDataTable(map[string]rdata.TickerReferenceWithSGXMapped{ref.ID: ref})
	return nil
}

func (c *CLI) handleReferenceDataAdd(args []string) error {
	ticker, err := parseTickerReferenceCommand("refdata-add", args)
	if err != nil {
		return err
	}

	respBody, err := c.doJSONRequest(http.MethodPost, "/api/v1/refdata", ticker)
	if err != nil {
		return err
	}

	var id string
	if err := json.Unmarshal(respBody, &id); err == nil && id != "" {
		fmt.Printf("Successfully added reference data for id: %s\n", id)
		return nil
	}

	fmt.Printf("Successfully added reference data for id: %s\n", ticker.ID)
	return nil
}

func (c *CLI) handleReferenceDataUpdate(args []string) error {
	ticker, err := parseTickerReferenceCommand("refdata-update", args)
	if err != nil {
		return err
	}

	if _, err := c.doJSONRequest(http.MethodPut, "/api/v1/refdata", ticker); err != nil {
		return err
	}

	fmt.Printf("Successfully updated reference data for id: %s\n", ticker.ID)
	return nil
}

func (c *CLI) handleReferenceDataDelete(args []string) error {
	fs := flag.NewFlagSet("refdata-delete", flag.ContinueOnError)
	idsArg := fs.String("id", "", "Comma-separated reference data ids to delete")
	yes := fs.Bool("yes", false, "Delete without interactive confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ids := normalizeCLIIDs(splitCSVAndArgs(*idsArg, fs.Args()))
	if len(ids) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s refdata-delete --id <id>[,<id>...] [--yes]\n", executable())
		return fmt.Errorf("at least one id is required")
	}

	if !*yes {
		fmt.Printf("Delete reference data for ids %v? This cannot be undone. (y/N): ", ids)
		if !c.promptForConfirmation() {
			fmt.Println("Reference data delete cancelled.")
			return nil
		}
	}

	if _, err := c.doJSONRequest(http.MethodDelete, "/api/v1/refdata", ids); err != nil {
		return err
	}

	fmt.Printf("Successfully deleted reference data for ids: %s\n", strings.Join(ids, ", "))
	return nil
}

// --- Utilities ---

func (c *CLI) fetchReferenceData() (map[string]rdata.TickerReferenceWithSGXMapped, error) {
	respBody, err := c.doJSONRequest(http.MethodGet, "/api/v1/refdata", nil)
	if err != nil {
		return nil, err
	}

	var refData map[string]rdata.TickerReferenceWithSGXMapped
	if err := json.Unmarshal(respBody, &refData); err != nil {
		return nil, fmt.Errorf("failed to decode reference data response: %w", err)
	}

	return refData, nil
}

func (c *CLI) doJSONRequest(method, path string, body any) ([]byte, error) {
	endpoint := strings.TrimRight(c.baseURL, "/") + path

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func parseTickerReferenceCommand(command string, args []string) (rdata.TickerReference, error) {
	var ticker rdata.TickerReference
	var filePath string

	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.StringVar(&filePath, "file", "", "JSON or YAML file containing one reference data record")
	fs.StringVar(&ticker.ID, "id", "", "Reference data id")
	fs.StringVar(&ticker.Name, "name", "", "Display name")
	fs.StringVar(&ticker.UnderlyingTicker, "underlying-ticker", "", "Underlying ticker")
	fs.StringVar(&ticker.YahooTicker, "yahoo-ticker", "", "Yahoo Finance ticker")
	fs.StringVar(&ticker.GoogleTicker, "google-ticker", "", "Google Finance ticker")
	fs.StringVar(&ticker.DividendsSgTicker, "dividends-sg-ticker", "", "Dividends.sg ticker")
	fs.StringVar(&ticker.NasdaqTicker, "nasdaq-ticker", "", "Nasdaq ticker")
	fs.StringVar(&ticker.BarchartTicker, "barchart-ticker", "", "Barchart ticker")
	fs.StringVar(&ticker.AssetClass, "asset-class", "", "Asset class")
	fs.StringVar(&ticker.AssetSubClass, "asset-sub-class", "", "Asset sub-class")
	fs.StringVar(&ticker.Category, "category", "", "Category")
	fs.StringVar(&ticker.SubCategory, "sub-category", "", "Sub-category")
	fs.StringVar(&ticker.Ccy, "ccy", "", "Currency")
	fs.StringVar(&ticker.Domicile, "domicile", "", "Domicile")
	fs.Float64Var(&ticker.CouponRate, "coupon-rate", 0, "Coupon rate")
	fs.StringVar(&ticker.MaturityDate, "maturity-date", "", "Maturity date in YYYY-MM-DD format")
	fs.Float64Var(&ticker.StrikePrice, "strike-price", 0, "Strike price")
	fs.StringVar(&ticker.CallPut, "call-put", "", "Option side: call or put")
	if err := fs.Parse(args); err != nil {
		return ticker, err
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return ticker, fmt.Errorf("failed to read reference data file: %w", err)
		}
		if err := yaml.Unmarshal(data, &ticker); err != nil {
			return ticker, fmt.Errorf("failed to parse reference data file: %w", err)
		}
	}

	ticker.ID = strings.ToUpper(strings.TrimSpace(ticker.ID))
	ticker.UnderlyingTicker = strings.ToUpper(strings.TrimSpace(ticker.UnderlyingTicker))
	ticker.YahooTicker = strings.ToUpper(strings.TrimSpace(ticker.YahooTicker))
	ticker.GoogleTicker = strings.ToUpper(strings.TrimSpace(ticker.GoogleTicker))
	ticker.DividendsSgTicker = strings.ToUpper(strings.TrimSpace(ticker.DividendsSgTicker))
	ticker.NasdaqTicker = strings.ToUpper(strings.TrimSpace(ticker.NasdaqTicker))
	ticker.BarchartTicker = strings.ToUpper(strings.TrimSpace(ticker.BarchartTicker))
	ticker.Ccy = strings.ToUpper(strings.TrimSpace(ticker.Ccy))
	ticker.Domicile = strings.ToUpper(strings.TrimSpace(ticker.Domicile))
	ticker.CallPut = strings.ToLower(strings.TrimSpace(ticker.CallPut))

	if ticker.ID == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s %s --id <id> [flags] or --file <json-or-yaml-file>\n", executable(), command)
		return ticker, fmt.Errorf("id is required")
	}

	return ticker, nil
}

func splitCSVAndArgs(csv string, args []string) []string {
	values := make([]string, 0, len(args)+1)
	for _, value := range strings.Split(csv, ",") {
		if strings.TrimSpace(value) != "" {
			values = append(values, value)
		}
	}
	values = append(values, args...)
	return values
}

func normalizeCLIIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		normalized := strings.ToUpper(strings.TrimSpace(id))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func filterReferenceData(refData map[string]rdata.TickerReferenceWithSGXMapped, ids []string, assetClass, assetSubClass, category string) map[string]rdata.TickerReferenceWithSGXMapped {
	idFilter := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idFilter[id] = struct{}{}
	}

	filtered := make(map[string]rdata.TickerReferenceWithSGXMapped)
	for id, ref := range refData {
		normalizedID := strings.ToUpper(id)
		if len(idFilter) > 0 {
			if _, ok := idFilter[normalizedID]; !ok {
				continue
			}
		}
		if assetClass != "" && ref.AssetClass != assetClass {
			continue
		}
		if assetSubClass != "" && ref.AssetSubClass != assetSubClass {
			continue
		}
		if category != "" && ref.Category != category {
			continue
		}
		filtered[id] = ref
	}
	return filtered
}

func printReferenceDataTable(refData map[string]rdata.TickerReferenceWithSGXMapped) {
	ids := make([]string, 0, len(refData))
	for id := range refData {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Printf("%-15s %-28s %-12s %-15s %-20s %-6s %-8s\n",
		"ID", "Name", "Asset Class", "Asset Subclass", "Category", "Ccy", "Domicile")
	fmt.Println("------------------------------------------------------------------------------------------------------------")

	for _, id := range ids {
		ref := refData[id]
		fmt.Printf("%-15s %-28s %-12s %-15s %-20s %-6s %-8s\n",
			ref.ID, truncate(ref.Name, 28), ref.AssetClass, ref.AssetSubClass, ref.Category, ref.Ccy, ref.Domicile)
	}

	fmt.Printf("\nTotal reference data: %d\n", len(ids))
}

func printJSON(value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

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
  refdata-list            List reference data
  refdata-get             Get reference data (--id ID)
  refdata-add             Add reference data (--id ID [flags] or --file FILE)
  refdata-update          Update reference data (--id ID [flags] or --file FILE)
  refdata-delete          Delete reference data (--id ID[,ID...] [--yes])

Global Flags:
  --url=<base-url>        Base URL of the API (default: http://localhost:8080)
  -v, --version           Show version information

Examples:
  %[1]s backup
  %[1]s position-list
  %[1]s position-delete --book main --ticker AAPL
  %[1]s refdata-list --asset-class eq
  %[1]s refdata-get --id AAPL --json
  %[1]s refdata-add --id AAPL --name Apple --underlying-ticker AAPL --asset-class eq --asset-sub-class stock --category technology --ccy USD --domicile US
  %[1]s refdata-update --file ./refdata-aapl.yaml
  %[1]s refdata-delete --id AAPL --yes
`, exe)
}
