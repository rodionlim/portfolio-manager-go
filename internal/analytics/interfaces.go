package analytics

import (
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
	FilePath     string            `json:"filePath"`
	AnalysisDate int64             `json:"analysisDate"`
	Metadata     map[string]string `json:"metadata"`
}

// MostTradedStock represents a single entry in the 100 Most Traded Stocks worksheet
type MostTradedStock struct {
	StockName                       string   `json:"stockName"`
	StockCode                       string   `json:"stockCode"`
	YTDAvgDailyTurnoverSGDM         float64  `json:"ytdAvgDailyTurnoverSGDM"`
	YTDInstitutionNetBuySellSGDM    float64  `json:"ytdInstitutionNetBuySellSGDM"`
	Past5SessionsInstitutionNetSGDM float64  `json:"past5SessionsInstitutionNetSGDM"`
	Sector                          string   `json:"sector"`
	InstitutionNetBuySellChange     *float64 `json:"institutionNetBuySellChange,omitempty"` // Change from previous report
}

// MostTradedStocksReport represents the complete 100 Most Traded Stocks data from a report
type MostTradedStocksReport struct {
	ReportDate  string            `json:"reportDate"`
	ReportTitle string            `json:"reportTitle"`
	FilePath    string            `json:"filePath"`
	Stocks      []MostTradedStock `json:"stocks"`
	ExtractedAt int64             `json:"extractedAt"`
}

// SGXClient interface for fetching SGX reports
type SGXClient interface {
	// FetchReports fetches the latest SGX reports
	FetchReports() (*SGXReportsResponse, error)

	// DownloadFile downloads a file from the given URL
	DownloadFile(url, filePath string) error
}

// AIAnalyzer interface for analyzing documents with AI
type AIAnalyzer interface {
	// AnalyzeDocument analyzes a document and returns insights
	AnalyzeDocument(filePath string, fileType string) (*ReportAnalysis, error)

	// SetDatabase sets the database instance for storing analysis results
	SetDatabase(db dal.Database)

	// FetchAnalysisByFileName fetches analysis results by file name
	FetchAnalysisByFileName(fileName string) (*ReportAnalysis, error)

	// GetAllAnalysisKeys gets all analysis keys from the database
	GetAllAnalysisKeys() ([]string, error)
}

// Service interface for the analytics service
type Service interface {
	// DownloadLatestNReports downloads the latest N SGX reports and returns their file paths
	DownloadLatestNReports(n int, reportType string) ([]string, error)

	// AnalyzeLatestNReports analyzes the latest N SGX reports and returns their analysis results
	AnalyzeLatestNReports(n int, reportType string, forceReanalysis bool) ([]*ReportAnalysis, error)

	// FetchAndAnalyzeLatestReportByType downloads the latest report of a specific type and analyzes it
	FetchAndAnalyzeLatestReportByType(reportType string) (*ReportAnalysis, error)

	// ListReportsInDataDir lists all available SGX reports in the data directory
	ListReportsInDataDir() ([]string, error)

	// ListAllAnalysis lists all available analysis reports that was previously stored in database
	ListAllAnalysis() ([]*ReportAnalysis, error)

	// AnalyzeExistingFile analyzes an existing file
	AnalyzeExistingFile(filePath string) (*ReportAnalysis, error)

	// ListAndExtractMostTradedStocks filters for SGX Fund Flow reports and extracts the "100 Most Traded Stocks" worksheet
	// n - limit results to the latest n reports (0 means no limit)
	ListAndExtractMostTradedStocks(n int) ([]*MostTradedStocksReport, error)
}
