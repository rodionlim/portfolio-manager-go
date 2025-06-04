package analytics

import (
	"context"
	"portfolio-manager/internal/dal"
)

// SGXReport represents a single SGX report from the API
type SGXReport struct {
	Data struct {
		Title      string `json:"title"`
		ReportDate int64  `json:"reportDate"`
		Report     struct {
			Data struct {
				MediaType string `json:"mediaType"`
				Name      string `json:"name"`
				Date      int64  `json:"date"`
				File      struct {
					Data struct {
						URL      string `json:"url"`
						FileMime string `json:"filemime"`
					} `json:"data"`
				} `json:"file"`
			} `json:"data"`
		} `json:"report"`
		FundsFlowType []struct {
			Data struct {
				Data struct {
					ID           string  `json:"id"`
					Name         string  `json:"name"`
					Order        string  `json:"order"`
					ParentCode   *string `json:"parentCode"`
					EntityBundle string  `json:"entityBundle"`
				} `json:"data"`
			} `json:"data"`
		} `json:"fundsFlowType"`
	} `json:"data"`
}

// SGXReportsResponse represents the API response structure
type SGXReportsResponse struct {
	Data struct {
		ReportTypes struct {
			Count   int `json:"count"`
			Results []struct {
				Data struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"data"`
			} `json:"results"`
		} `json:"reportTypes"`
		List struct {
			Count   int         `json:"count"`
			Results []SGXReport `json:"results"`
		} `json:"list"`
	} `json:"data"`
}

// ReportAnalysis represents the AI analysis of a report
type ReportAnalysis struct {
	Summary      string            `json:"summary"`
	KeyInsights  []string          `json:"keyInsights"`
	ReportDate   int64             `json:"reportDate"`
	ReportTitle  string            `json:"reportTitle"`
	ReportType   string            `json:"reportType"`
	DownloadURL  string            `json:"downloadUrl"`
	FilePath     string            `json:"filePath"`
	AnalysisDate int64             `json:"analysisDate"`
	Metadata     map[string]string `json:"metadata"`
}

// SGXClient interface for fetching SGX reports
type SGXClient interface {
	// FetchReports fetches the latest SGX reports
	FetchReports(ctx context.Context) (*SGXReportsResponse, error)

	// DownloadFile downloads a file from the given URL
	DownloadFile(ctx context.Context, url, filePath string) error
}

// AIAnalyzer interface for analyzing documents with AI
type AIAnalyzer interface {
	// AnalyzeDocument analyzes a document and returns insights
	AnalyzeDocument(ctx context.Context, filePath string, fileType string) (*ReportAnalysis, error)

	// SetDatabase sets the database instance for storing analysis results
	SetDatabase(db dal.Database)

	// FetchAnalysisByFileName fetches analysis results by file name
	FetchAnalysisByFileName(fileName string) (*ReportAnalysis, error)

	// GetAllAnalysisKeys gets all analysis keys from the database
	GetAllAnalysisKeys() ([]string, error)
}

// Service interface for the analytics service
type Service interface {
	// FetchLatestReportByType fetches the latest report of a specific type and analyzes it
	FetchLatestReportByType(ctx context.Context, reportType string) (*ReportAnalysis, error)

	// ListReportsInDataDir lists all available SGX reports in the data directory
	ListReportsInDataDir() ([]string, error)

	// ListAllAnalysis lists all available analysis reports that was previously stored in database
	ListAllAnalysis() ([]*ReportAnalysis, error)

	// AnalyzeExistingFile analyzes an existing file
	AnalyzeExistingFile(ctx context.Context, filePath string) (*ReportAnalysis, error)
}
