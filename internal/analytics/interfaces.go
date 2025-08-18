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

// SectorFundsFlowReport represents the institutional funds flow by sector for a specific report
type SectorFundsFlowReport struct {
	ReportDate        string       `json:"reportDate"`
	ReportTitle       string       `json:"reportTitle"`
	FilePath          string       `json:"filePath"`
	WeekEndingDate    string       `json:"weekEndingDate"`
	OverallNetBuySell float64      `json:"overallNetBuySell"`
	SectorFlows       []SectorFlow `json:"sectorFlows"`
	ExtractedAt       int64        `json:"extractedAt"`
}

// SectorFlow represents the institutional net buy/sell amount for a specific sector
type SectorFlow struct {
	SectorName     string  `json:"sectorName"`
	NetBuySellSGDM float64 `json:"netBuySellSGDM"`
}

// Top10Stock represents a single stock entry in the Weekly Top 10 data
type Top10Stock struct {
	StockName      string  `json:"stockName"`
	StockCode      string  `json:"stockCode"`
	NetBuySellSGDM float64 `json:"netBuySellSGDM"`
	IsNetBuy       bool    `json:"isNetBuy"`     // true for net buy, false for net sell
	InvestorType   string  `json:"investorType"` // "institutional" or "retail"
}

// Top10WeeklyReport represents the complete Weekly Top 10 data from a report
type Top10WeeklyReport struct {
	ReportDate                       string       `json:"reportDate"`
	ReportTitle                      string       `json:"reportTitle"`
	FilePath                         string       `json:"filePath"`
	WeekEndingDate                   string       `json:"weekEndingDate"`
	InstitutionalNetSellTotalSGDM    float64      `json:"institutionalNetSellTotalSGDM"`
	InstitutionalNetSellPreviousSGDM float64      `json:"institutionalNetSellPreviousSGDM"`
	RetailNetBuyTotalSGDM            float64      `json:"retailNetBuyTotalSGDM"`
	RetailNetBuyPreviousSGDM         float64      `json:"retailNetBuyPreviousSGDM"`
	Top10Stocks                      []Top10Stock `json:"top10Stocks"`
	ExtractedAt                      int64        `json:"extractedAt"`
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

	// ListAndExtractSectorFundsFlow filters for SGX Fund Flow reports and extracts the "Institutional" worksheet
	// n - limit results to the latest n reports (0 means no limit)
	ListAndExtractSectorFundsFlow(n int) ([]*SectorFundsFlowReport, error)

	// ListAndExtractTop10Stocks filters for SGX Fund Flow reports and extracts the "Weekly Top 10" worksheet
	// n - limit results to the latest n reports (0 means no limit)
	ListAndExtractTop10Stocks(n int) ([]*Top10WeeklyReport, error)
}
