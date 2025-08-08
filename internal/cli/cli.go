package cli

import (
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

	"portfolio-manager/internal/portfolio"
)

// CLIClient handles CLI operations
type CLIClient struct {
	BaseURL string
	Client  *http.Client
}

// NewCLIClient creates a new CLI client
func NewCLIClient(baseURL string) *CLIClient {
	return &CLIClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DeletePosition deletes a single position via REST API
func (c *CLIClient) DeletePosition(book, ticker string) error {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/portfolio/position", c.BaseURL))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("book", book)
	q.Set("ticker", ticker)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully deleted position for book: %s, ticker: %s\n", book, ticker)
	return nil
}

// ListPositions lists all positions via REST API
func (c *CLIClient) ListPositions() error {
	url := fmt.Sprintf("%s/api/v1/portfolio/positions", c.BaseURL)

	resp, err := c.Client.Get(url)
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

	// Display positions in a table format
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

// RunCLI handles CLI subcommands
func RunCLI(args []string, baseURL string) error {
	if len(args) == 0 {
		PrintUsage()
		return fmt.Errorf("no command provided")
	}

	cmd := args[0]
	client := NewCLIClient(baseURL)

	switch cmd {
	case "position-delete":
		fs := flag.NewFlagSet("position-delete", flag.ContinueOnError)
		book := fs.String("book", "", "Book name (required)")
		ticker := fs.String("ticker", "", "Ticker symbol (required)")
		// Allow legacy positional args: position-delete <book> <ticker>
		fs.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: %s position-delete --book <book> --ticker <ticker>\n", executable())
			fmt.Fprintln(os.Stderr)
			fs.PrintDefaults()
		}
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		// fallback to positional if flags not provided
		remaining := fs.Args()
		if *book == "" && len(remaining) > 0 {
			*book = remaining[0]
		}
		if *ticker == "" && len(remaining) > 1 {
			*ticker = remaining[1]
		}
		if *book == "" || *ticker == "" {
			fs.Usage()
			return fmt.Errorf("both --book and --ticker are required")
		}
		return client.DeletePosition(*book, *ticker)

	case "position-list":
		// No flags yet, but create a FlagSet for extensibility
		fs := flag.NewFlagSet("position-list", flag.ContinueOnError)
		fs.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: %s position-list\n", executable())
		}
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return client.ListPositions()

	case "help", "--help", "-h":
		PrintUsage()
		return nil

	default:
		PrintUsage()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func executable() string {
	base := filepath.Base(os.Args[0])
	// strip possible extensions on Windows
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// PrintUsage prints CLI usage information
func PrintUsage() {
	exe := executable()
	fmt.Fprintf(os.Stderr, `Portfolio Manager CLI

Usage:
	%[1]s [--url=<base-url>] <command> [flags]

Commands:
	position-list                         List all positions
	position-delete --book B --ticker T   Delete a specific position

Shortcuts / Legacy Positional Forms:
	%[1]s position-delete B T

Global Flags:
	--url=<base-url>        Base URL of the API (default: http://localhost:8080)

Examples:
	%[1]s position-list
	%[1]s position-delete --book main --ticker AAPL
	%[1]s --url=http://localhost:9090 position-list
	%[1]s position-delete main AAPL

`, exe)
}
