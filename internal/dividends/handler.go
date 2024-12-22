package dividends

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/logging"
)

// HandlePostDividends handles retrieving dividends for a single ticker.
// @Summary Get dividends for a single ticker
// @Description Get dividends for a single ticker
// @Tags dividends
// @Accept  json
// @Produce  json
// @Param   ticker  body  string  true  "Ticker symbol"
// @Success 200 {array} Dividends
// @Failure 400 {string} string "ticker is required"
// @Failure 500 {string} string "failed to calculate dividends"
// @Router /api/v1/dividends [post]
func HandlePostDividends(manager *DividendsManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Ticker string `json:"ticker"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}

		if request.Ticker == "" {
			http.Error(w, "ticker is required", http.StatusBadRequest)
			return
		}

		dividends, err := manager.CalculateDividendsForSingleTicker(request.Ticker)
		if err != nil {
			logging.GetLogger().Error("Failed to calculate dividends", err)
			http.Error(w, "failed to calculate dividends", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dividends)
	}
}

// RegisterHandlers registers the handlers for the dividends service.
func RegisterHandlers(mux *http.ServeMux, manager *DividendsManager) {
	mux.HandleFunc("/api/v1/dividends", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandlePostDividends(manager).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
