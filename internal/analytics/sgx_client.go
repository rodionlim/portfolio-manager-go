package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SGXClientImpl implements the SGXClient interface
type SGXClientImpl struct {
	httpClient *http.Client
	baseURL    string
}

// NewSGXClient creates a new SGX client
func NewSGXClient() SGXClient {
	return &SGXClientImpl{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api2.sgx.com",
	}
}

// SGXQueryParams represents the variables for SGX API queries
type SGXQueryParams struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Lang   string `json:"lang"`
}

// FetchReports fetches the latest SGX reports
func (c *SGXClientImpl) FetchReports() (*SGXReportsResponse, error) {
	// Query parameters in human-readable format
	queryID := "54d7880bed915819b82da8c0cf77d10e299ea9cc:funds_flow_reports_list"
	params := SGXQueryParams{
		Limit:  50,
		Offset: 0,
		Lang:   "EN",
	}

	// Convert params to JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query parameters: %w", err)
	}

	// Build URL with proper encoding
	baseURL := "https://api2.sgx.com/content-api"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	query := u.Query()
	query.Set("queryId", queryID)
	query.Set("variables", string(paramsJSON))
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("origin", "https://www.sgx.com")
	req.Header.Set("referer", "https://www.sgx.com/")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var reportsResponse SGXReportsResponse
	if err := json.NewDecoder(resp.Body).Decode(&reportsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &reportsResponse, nil
}

// DownloadFile downloads a file from the given URL to the specified path
func (c *SGXClientImpl) DownloadFile(url, filePath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code when downloading: %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// generateSafeFileName creates a safe filename from the report title
func generateSafeFileName(title string, extension string) string {
	// Replace unsafe characters with underscores
	safe := strings.ReplaceAll(title, " ", "_")
	safe = strings.ReplaceAll(safe, "(", "")
	safe = strings.ReplaceAll(safe, ")", "")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	safe = strings.ReplaceAll(safe, "*", "_")
	safe = strings.ReplaceAll(safe, "?", "_")
	safe = strings.ReplaceAll(safe, "\"", "_")
	safe = strings.ReplaceAll(safe, "<", "_")
	safe = strings.ReplaceAll(safe, ">", "_")
	safe = strings.ReplaceAll(safe, "|", "_")

	return fmt.Sprintf("%s%s", safe, extension)
}
