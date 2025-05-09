package historical

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
)

// HandleGetMetrics handles the GET /api/v1/historical/metrics endpoint
// @Summary Get historical portfolio metrics
// @Description Get all historical portfolio metrics (date-stamped portfolio metrics)
// @Tags historical
// @Produce json
// @Success 200 {array} TimestampedMetrics "List of historical portfolio metrics by date"
// @Failure 500 {object} common.ErrorResponse "Failed to get historical metrics"
// @Router /api/v1/historical/metrics [get]
func HandleGetMetrics(service interface {
	GetMetrics() ([]TimestampedMetrics, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := service.GetMetrics()
		if err != nil {
			common.WriteJSONError(w, "Failed to get historical metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			common.WriteJSONError(w, "Failed to encode metrics as JSON", http.StatusInternalServerError)
		}
	}
}

// RegisterHandlers registers the historical metrics handlers
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("/api/v1/historical/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleGetMetrics(service).ServeHTTP(w, r)
	})
}
