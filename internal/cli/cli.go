package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"portfolio-manager/internal/backup"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// CLI handles command line interface
type CLI struct {
	commands     map[string]*Command
	backupSvc    backup.BackupService
	defaultDBPath string
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{
		commands:      make(map[string]*Command),
		backupSvc:     backup.NewService(),
		defaultDBPath: "./portfolio-manager.db",
	}
	
	cli.registerCommands()
	return cli
}

// registerCommands registers all available commands
func (c *CLI) registerCommands() {
	c.commands["backup"] = &Command{
		Name:        "backup",
		Description: "Create a backup of the database",
		Handler:     c.handleBackup,
	}
	
	c.commands["restore-from-backup"] = &Command{
		Name:        "restore-from-backup",
		Description: "Restore database from a backup",
		Handler:     c.handleRestore,
	}
	
	c.commands["-v"] = &Command{
		Name:        "-v",
		Description: "Show version information",
		Handler:     c.handleVersion,
	}
	
	c.commands["--version"] = &Command{
		Name:        "--version",
		Description: "Show version information",
		Handler:     c.handleVersion,
	}
}

// ParseAndExecute parses command line arguments and executes the appropriate command
func (c *CLI) ParseAndExecute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("no command specified")
	}
	
	commandName := args[1]
	
	// Handle special case for version flags
	if commandName == "-v" || commandName == "--version" {
		return c.handleVersion(args[1:])
	}
	
	command, exists := c.commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}
	
	return command.Handler(args[2:])
}

// handleBackup handles the backup command
func (c *CLI) handleBackup(args []string) error {
	var source, uri, user, password string
	
	// Parse flags
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	fs.StringVar(&source, "source", "local", "Backup source (local, gdrive, nextcloud)")
	fs.StringVar(&uri, "uri", "", "File location or URL")
	fs.StringVar(&user, "user", "", "Username for remote sources")
	fs.StringVar(&password, "password", "", "Password for remote sources")
	fs.Parse(args)
	
	// Check if database exists
	if _, err := os.Stat(c.defaultDBPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found at %s. Make sure the application has been run at least once", c.defaultDBPath)
	}
	
	// Calculate backup size
	size, err := c.backupSvc.GetBackupSize(c.defaultDBPath)
	if err != nil {
		return fmt.Errorf("failed to calculate backup size: %w", err)
	}
	
	// Convert size to human readable format
	sizeStr := formatFileSize(size)
	
	// Prompt user for confirmation
	fmt.Printf("Backup size will be approximately: %s\n", sizeStr)
	fmt.Println("WARNING: Backups from older versions might not be compatible with the current version.")
	fmt.Println("It is recommended to use the CLI tool of the same version as when the backup was taken.")
	fmt.Print("Do you want to proceed with the backup? (y/N): ")
	
	if !c.promptForConfirmation() {
		fmt.Println("Backup cancelled.")
		return nil
	}
	
	// Create backup configuration
	config := backup.BackupConfig{
		Source:   source,
		URI:      uri,
		User:     user,
		Password: password,
	}
	
	// Create backup
	ctx := context.Background()
	err = c.backupSvc.Backup(ctx, c.defaultDBPath, config)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	
	fmt.Println("Backup completed successfully!")
	return nil
}

// handleRestore handles the restore command
func (c *CLI) handleRestore(args []string) error {
	var source, uri, user, password string
	
	// Parse flags
	fs := flag.NewFlagSet("restore-from-backup", flag.ExitOnError)
	fs.StringVar(&source, "source", "local", "Backup source (local, gdrive, nextcloud)")
	fs.StringVar(&uri, "uri", "", "File location or URL")
	fs.StringVar(&user, "user", "", "Username for remote sources")
	fs.StringVar(&password, "password", "", "Password for remote sources")
	fs.Parse(args)
	
	// Check if application is running
	running, err := c.backupSvc.IsApplicationRunning()
	if err != nil {
		fmt.Printf("Warning: Could not check if application is running: %v\n", err)
	} else if running {
		return fmt.Errorf("application appears to be running. Please stop the portfolio-manager service before restoring from backup")
	}
	
	// Check if database already exists
	if _, err := os.Stat(c.defaultDBPath); err == nil {
		fmt.Printf("Existing database found at %s\n", c.defaultDBPath)
		fmt.Print("This will completely replace the existing database. Are you sure you want to continue? (y/N): ")
		
		if !c.promptForConfirmation() {
			fmt.Println("Restore cancelled.")
			return nil
		}
	}
	
	// Create backup configuration
	config := backup.BackupConfig{
		Source:   source,
		URI:      uri,
		User:     user,
		Password: password,
	}
	
	// Restore from backup
	ctx := context.Background()
	err = c.backupSvc.Restore(ctx, c.defaultDBPath, config)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	
	fmt.Println("Restore completed successfully!")
	return nil
}

// handleVersion handles the version command
func (c *CLI) handleVersion(args []string) error {
	version, err := getVersion()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	
	fmt.Printf("Portfolio Manager version: %s\n", strings.TrimSpace(version))
	return nil
}

// promptForConfirmation prompts user for yes/no confirmation
func (c *CLI) promptForConfirmation() bool {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "y" || response == "yes"
}

// getVersion reads version from VERSION file
func getVersion() (string, error) {
	// Try different paths for VERSION file
	versionPaths := []string{
		"VERSION",
		"../../VERSION", // For tests running from subdirectories
		"../../../VERSION", // For tests running from deeper subdirectories
	}
	
	for _, path := range versionPaths {
		if content, err := ioutil.ReadFile(path); err == nil {
			return string(content), nil
		}
	}
	
	return "", fmt.Errorf("VERSION file not found")
}

// formatFileSize formats file size in human readable format
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

// ShowHelp displays help information
func (c *CLI) ShowHelp() {
	fmt.Println("Portfolio Manager CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  ./portfolio-manager [command] [options]")
	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Println("  backup                  Create a backup of the database")
	fmt.Println("  restore-from-backup     Restore database from a backup")
	fmt.Println("  -v, --version          Show version information")
	fmt.Println("")
	fmt.Println("For command-specific help, run:")
	fmt.Println("  ./portfolio-manager [command] -h")
}