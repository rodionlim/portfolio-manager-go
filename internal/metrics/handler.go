package metrics

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// HandleGetPortfolioMetrics handles getting the portfolio metrics including IRR
// @Summary Get portfolio IRR
// @Description Get the Internal Rate of Return (IRR), MV, Price Paid for the entire portfolio
// @Tags metrics
// @Produce json
// @Success 200 {object} MetricsResult "The portfolio metrics, including IRR, cash flows and others"
// @Failure 500 {object} common.ErrorResponse "Failed to calculate portoflio metrics"
// @Router /api/v1/metrics [get]
func HandleGetPortfolioMetrics(service *MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := service.CalculatePortfolioMetrics()
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

// RegisterHandlers registers the handlers for the metrics service
func RegisterHandlers(mux *http.ServeMux, service *MetricsService) {
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleGetPortfolioMetrics(service).ServeHTTP(w, r)
	})
}
