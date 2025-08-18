package analytics

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"strconv"
)

// AnalyzeFileRequest represents a request to analyze a file
type AnalyzeFileRequest struct {
	FilePath string `json:"filePath" binding:"required"`
}

// HandleListReports handles listing all available SGX reports
// @Summary List all available SGX reports
// @Description Lists all available SGX reports in the data directory
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {array} string "List of report file paths"
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/list [get]
func HandleListReports(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reports, err := service.ListReportsInDataDir()
		if err != nil {
			logging.GetLogger().Error("Failed to list reports:", err)
			common.WriteJSONError(w, "Failed to list reports: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reports); err != nil {
			logging.GetLogger().Error("Failed to write reports response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleGetLatestReportByType handles getting the latest SGX report analysis by type
// @Summary Get latest SGX report analysis by type
// @Description Fetches the latest SGX report of a specific type, downloads it, and provides AI analysis
// @Tags analytics
// @Accept json
// @Produce json
// @Param type query string true "Report type (e.g., 'fund%20flow', 'daily%20momentum')"
// @Success 200 {object} ReportAnalysis
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/latest [get]
func HandleGetLatestReportByType(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportType := r.URL.Query().Get("type")
		if reportType == "" {
			common.WriteJSONError(w, "report type is required", http.StatusBadRequest)
			return
		}

		analysis, err := service.FetchAndAnalyzeLatestReportByType(reportType)
		if err != nil {
			logging.GetLogger().Error("Failed to fetch latest report by type:", err)
			common.WriteJSONError(w, "Failed to fetch latest report by type: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(analysis); err != nil {
			logging.GetLogger().Error("Failed to write analysis response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleAnalyzeExistingFile handles analyzing an existing file
// @Summary Analyze existing file
// @Description Analyzes an existing file in the data directory
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body AnalyzeFileRequest true "File analysis request"
// @Success 200 {object} ReportAnalysis
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/analyze [post]
func HandleAnalyzeExistingFile(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AnalyzeFileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.FilePath == "" {
			common.WriteJSONError(w, "filePath is required", http.StatusBadRequest)
			return
		}

		analysis, err := service.AnalyzeExistingFile(req.FilePath)
		if err != nil {
			logging.GetLogger().Error("Failed to analyze existing file:", err)
			common.WriteJSONError(w, "Failed to analyze existing file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(analysis); err != nil {
			logging.GetLogger().Error("Failed to write analysis response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleListAnalysis handles listing all available analysis reports
// @Summary List all available analysis reports
// @Description Lists all analysis reports that were previously stored in the database
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {array} ReportAnalysis "List of analysis reports"
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/list_analysis [get]
func HandleListAnalysis(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		analyses, err := service.ListAllAnalysis()
		if err != nil {
			logging.GetLogger().Error("Failed to list analyses:", err)
			common.WriteJSONError(w, "Failed to list analyses: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(analyses); err != nil {
			logging.GetLogger().Error("Failed to write analyses response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleDownloadLatestNReports handles downloading the latest N SGX reports
// @Summary Download latest N SGX reports
// @Description Downloads the latest N SGX reports from SGX and stores them in the data directory. Optionally filter by report type.
// @Tags analytics
// @Accept json
// @Produce json
// @Param n query int true "Number of latest reports to download"
// @Param type query string false "Report type filter (e.g., 'fund%20flow', 'daily%20momentum'). If not provided, downloads all types."
// @Success 200 {array} string "List of downloaded report file paths"
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/download [get]
func HandleDownloadLatestNReports(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nStr := r.URL.Query().Get("n")
		if nStr == "" {
			common.WriteJSONError(w, "parameter 'n' is required", http.StatusBadRequest)
			return
		}

		n, err := strconv.Atoi(nStr)
		if err != nil {
			common.WriteJSONError(w, "parameter 'n' must be a valid integer", http.StatusBadRequest)
			return
		}

		if n <= 0 {
			common.WriteJSONError(w, "parameter 'n' must be greater than 0", http.StatusBadRequest)
			return
		}

		// Check if type parameter is provided for filtering by type
		reportType := r.URL.Query().Get("type")

		downloadedFiles, err := service.DownloadLatestNReports(n, reportType)
		if err != nil {
			logging.GetLogger().Error("Failed to download latest N reports:", err)
			common.WriteJSONError(w, "Failed to download latest N reports: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(downloadedFiles); err != nil {
			logging.GetLogger().Error("Failed to write download response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleAnalyzeLatestNReports handles analyzing the latest N SGX reports
// @Summary Analyze latest N SGX reports
// @Description Downloads and analyzes the latest N SGX reports from SGX. Optionally filter by report type and force reanalysis.
// @Tags analytics
// @Accept json
// @Produce json
// @Param n query int true "Number of latest reports to analyze"
// @Param type query string false "Report type filter (e.g., 'fund%20flow', 'daily%20momentum'). If not provided, analyzes all types."
// @Param force query bool false "Force reanalysis even if analysis exists in database (default: false)"
// @Success 200 {array} ReportAnalysis "List of analysis results"
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/analyze_latest [get]
func HandleAnalyzeLatestNReports(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nStr := r.URL.Query().Get("n")
		if nStr == "" {
			common.WriteJSONError(w, "parameter 'n' is required", http.StatusBadRequest)
			return
		}

		n, err := strconv.Atoi(nStr)
		if err != nil {
			common.WriteJSONError(w, "parameter 'n' must be a valid integer", http.StatusBadRequest)
			return
		}

		if n <= 0 {
			common.WriteJSONError(w, "parameter 'n' must be greater than 0", http.StatusBadRequest)
			return
		}

		// Check if type parameter is provided for filtering by type
		reportType := r.URL.Query().Get("type")

		// Check if force reanalysis is requested
		forceReanalysis := false
		if forceStr := r.URL.Query().Get("force"); forceStr != "" {
			forceReanalysis, err = strconv.ParseBool(forceStr)
			if err != nil {
				common.WriteJSONError(w, "parameter 'force' must be a valid boolean", http.StatusBadRequest)
				return
			}
		}

		analyses, err := service.AnalyzeLatestNReports(n, reportType, forceReanalysis)
		if err != nil {
			logging.GetLogger().Error("Failed to analyze latest N reports:", err)
			common.WriteJSONError(w, "Failed to analyze latest N reports: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(analyses); err != nil {
			logging.GetLogger().Error("Failed to write analysis response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleListAndExtractMostTradedStocks (DEPRECATED) handles extracting the "100 Most Traded Stocks" data from SGX Fund Flow reports
// @Summary Extract 100 Most Traded Stocks data from SGX Fund Flow reports
// @Description Filters for SGX Fund Flow Weekly Tracker reports and extracts the "100 Most Traded Stocks" worksheet data
// @Tags analytics
// @Accept json
// @Produce json
// @Param n query int false "Limit results to latest n reports (0 or not provided means no limit)"
// @Success 200 {array} MostTradedStocksReport "List of 100 Most Traded Stocks reports"
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/most_traded_stocks [get]
func HandleListAndExtractMostTradedStocks(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the optional 'n' query parameter
		n := 0
		if nParam := r.URL.Query().Get("n"); nParam != "" {
			var err error
			n, err = strconv.Atoi(nParam)
			if err != nil || n < 0 {
				logging.GetLogger().Error("Invalid 'n' parameter:", err)
				common.WriteJSONError(w, "Invalid 'n' parameter: must be a non-negative integer", http.StatusBadRequest)
				return
			}
		}

		reports, err := service.ListAndExtractMostTradedStocks(n)
		if err != nil {
			logging.GetLogger().Error("Failed to extract most traded stocks:", err)
			common.WriteJSONError(w, "Failed to extract most traded stocks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reports); err != nil {
			logging.GetLogger().Error("Failed to write most traded stocks response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleListAndExtractSectorFundsFlow handles extracting the "Institutional" sector funds flow data from SGX Fund Flow reports
// @Summary Extract Institutional sector funds flow data from SGX Fund Flow reports
// @Description Filters for SGX Fund Flow Weekly Tracker reports and extracts the "Institutional" worksheet data showing weekly institutional flow by sector
// @Tags analytics
// @Accept json
// @Produce json
// @Param n query int false "Limit results to latest n reports (0 or not provided means no limit)"
// @Success 200 {array} SectorFundsFlowReport "List of sector funds flow reports"
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/sector_funds_flow [get]
func HandleListAndExtractSectorFundsFlow(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the optional 'n' query parameter
		n := 0
		if nParam := r.URL.Query().Get("n"); nParam != "" {
			var err error
			n, err = strconv.Atoi(nParam)
			if err != nil || n < 0 {
				logging.GetLogger().Error("Invalid 'n' parameter:", err)
				common.WriteJSONError(w, "Invalid 'n' parameter: must be a non-negative integer", http.StatusBadRequest)
				return
			}
		}

		reports, err := service.ListAndExtractSectorFundsFlow(n)
		if err != nil {
			logging.GetLogger().Error("Failed to extract sector funds flow:", err)
			common.WriteJSONError(w, "Failed to extract sector funds flow: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reports); err != nil {
			logging.GetLogger().Error("Failed to write sector funds flow response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleListAndExtractTop10Stocks handles extracting the "Weekly Top 10" data from SGX Fund Flow reports
// @Summary Extract Weekly Top 10 stocks data from SGX Fund Flow reports
// @Description Filters for SGX Fund Flow Weekly Tracker reports and extracts the "Weekly Top 10" worksheet data showing institutional and retail top 10 net buy/sell stocks
// @Tags analytics
// @Accept json
// @Produce json
// @Param n query int false "Limit results to latest n reports (0 or not provided means no limit)"
// @Success 200 {array} Top10WeeklyReport "List of Weekly Top 10 stocks reports"
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/top10_stocks [get]
func HandleListAndExtractTop10Stocks(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the optional 'n' query parameter
		n := 0
		if nParam := r.URL.Query().Get("n"); nParam != "" {
			var err error
			n, err = strconv.Atoi(nParam)
			if err != nil || n < 0 {
				logging.GetLogger().Error("Invalid 'n' parameter:", err)
				common.WriteJSONError(w, "Invalid 'n' parameter: must be a non-negative integer", http.StatusBadRequest)
				return
			}
		}

		reports, err := service.ListAndExtractTop10Stocks(n)
		if err != nil {
			logging.GetLogger().Error("Failed to extract top 10 stocks:", err)
			common.WriteJSONError(w, "Failed to extract top 10 stocks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reports); err != nil {
			logging.GetLogger().Error("Failed to write top 10 stocks response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// RegisterHandlers registers the analytics handlers
func RegisterHandlers(mux *http.ServeMux, service Service) {
	mux.HandleFunc("/api/v1/analytics/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if type parameter is provided for filtering by type
		reportType := r.URL.Query().Get("type")
		if reportType != "" {
			HandleGetLatestReportByType(service).ServeHTTP(w, r)
		} else {
			common.WriteJSONError(w, "report type is required", http.StatusBadRequest)
			return
		}
	})

	mux.HandleFunc("/api/v1/analytics/list_files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleListReports(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleAnalyzeExistingFile(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/list_analysis", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleListAnalysis(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/download_latest_n", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleDownloadLatestNReports(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/analyze_latest_n", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleAnalyzeLatestNReports(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/most_traded_stocks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleListAndExtractMostTradedStocks(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/sector_funds_flow", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleListAndExtractSectorFundsFlow(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/analytics/top10_stocks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleListAndExtractTop10Stocks(service).ServeHTTP(w, r)
	})
}
