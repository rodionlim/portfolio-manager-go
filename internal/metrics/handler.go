package metrics

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// HandleGetPortfolioMetrics handles getting the portfolio metrics including IRR
// @Summary Get portfolio IRR
// @Description Get the Internal Rate of Return (IRR), MV, Price Paid for the entire portfolio or a specific book
// @Tags metrics
// @Produce json
// @Param book_filter query string false "Filter metrics by book (optional)"
// @Success 200 {object} MetricsResult "The portfolio metrics, including IRR, cash flows and others"
// @Failure 500 {object} common.ErrorResponse "Failed to calculate portoflio metrics"
// @Router /api/v1/metrics [get]
func HandleGetPortfolioMetrics(service *MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get book_filter query parameter
		bookFilter := r.URL.Query().Get("book_filter")

		result, err := service.CalculatePortfolioMetrics(bookFilter)
		if err != nil {
			logging.GetLogger().Error("Failed to calculate IRR:", err)
			common.WriteJSONError(w, "Failed to calculate portfolio IRR: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logging.GetLogger().Error("Failed to write IRR response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// HandleBenchmarkMetrics handles benchmarking portfolio performance against a user-specified benchmark
// @Summary Benchmark portfolio performance
// @Description Compare portfolio IRR against a benchmark using buy_at_start or match_trades mode
// @Tags metrics
// @Accept json
// @Produce json
// @Param request body BenchmarkRequest true "Benchmark request"
// @Success 200 {object} BenchmarkComparisonResult "Benchmark comparison result"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to benchmark portfolio"
// @Router /api/v1/metrics/benchmark [post]
func HandleBenchmarkMetrics(service *MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BenchmarkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := service.BenchmarkPortfolioPerformance(req)
		if err != nil {
			logging.GetLogger().Error("Failed to benchmark portfolio:", err)
			common.WriteJSONError(w, "Failed to benchmark portfolio: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logging.GetLogger().Error("Failed to write benchmark response as JSON:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

// RegisterHandlers registers the handlers for the metrics service
func RegisterHandlers(mux *http.ServeMux, service *MetricsService) {
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleGetPortfolioMetrics(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/metrics/benchmark", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleBenchmarkMetrics(service).ServeHTTP(w, r)
	})
}
