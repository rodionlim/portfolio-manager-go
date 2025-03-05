package dividends

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/logging"
	"strings"
)

// HandleGetDividends handles retrieving dividends for a single ticker.
// @Summary Get dividends for a single ticker
// @Description Get dividends for a single ticker
// @Tags dividends
// @Accept  json
// @Produce  json
// @Param ticker path string true "Asset ticker symbol"
// @Success 200 {array} Dividends
// @Failure 400 {string} string "ticker is required"
// @Failure 500 {string} string "failed to calculate dividends"
// @Router /api/v1/dividends/{ticker} [get]
func HandleGetDividends(manager *DividendsManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/api/v1/dividends/"))
		if ticker == "" {
			http.Error(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		dividends, err := manager.CalculateDividendsForSingleTicker(ticker)
		if err != nil {
			logging.GetLogger().Error("Failed to calculate dividends", err)
			http.Error(w, fmt.Sprintf("failed to calculate dividends for %s", ticker), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dividends)
	}
}

// RegisterHandlers registers the handlers for the dividends service.
func RegisterHandlers(mux *http.ServeMux, manager *DividendsManager) {
	mux.HandleFunc("/api/v1/dividends/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetDividends(manager).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
