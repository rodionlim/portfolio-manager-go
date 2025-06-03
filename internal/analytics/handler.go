package analytics

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// AnalyzeFileRequest represents a request to analyze a file
type AnalyzeFileRequest struct {
	FilePath string `json:"filePath" binding:"required"`
}

// HandleGetLatestReport handles getting the latest SGX report analysis
// @Summary Get latest SGX report analysis
// @Description Fetches the latest SGX report, downloads it, and provides AI analysis
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} ReportAnalysis
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/latest [get]
func HandleGetLatestReport(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		analysis, err := service.FetchLatestReport(ctx)
		if err != nil {
			logging.GetLogger().Error("Failed to fetch latest report:", err)
			common.WriteJSONError(w, "Failed to fetch latest report: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(analysis); err != nil {
			logging.GetLogger().Error("Failed to write analysis response as JSON:", err)
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
// @Param type query string true "Report type (e.g., 'fund_flow', 'daily_momentum')"
// @Success 200 {object} ReportAnalysis
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/analytics/latest [get]
func HandleGetLatestReportByType(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reportType := r.URL.Query().Get("type")
		if reportType == "" {
			common.WriteJSONError(w, "report type is required", http.StatusBadRequest)
			return
		}

		analysis, err := service.FetchLatestReportByType(ctx, reportType)
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
		ctx := r.Context()

		var req AnalyzeFileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.FilePath == "" {
			common.WriteJSONError(w, "filePath is required", http.StatusBadRequest)
			return
		}

		analysis, err := service.AnalyzeExistingFile(ctx, req.FilePath)
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
			HandleGetLatestReport(service).ServeHTTP(w, r)
		}
	})

	mux.HandleFunc("/api/v1/analytics/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleAnalyzeExistingFile(service).ServeHTTP(w, r)
	})
}
