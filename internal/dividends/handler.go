package dividends

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/common"
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
// @Failure 500 {string} string "failed to calculate dividends"
// @Router /api/v1/dividends/{ticker} [get]
func HandleGetDividends(manager *DividendsManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/api/v1/dividends/"))

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

// HandleGetAllDividends handles retrieving dividends for all tickers.
// @Summary Get dividends for all tickers
// @Description Get dividends for all tickers
// @Tags dividends
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string][]Dividends "Mapping of ticker to dividends"
// @Failure 500 {object} common.ErrorResponse "failed to calculate dividends"
// @Router /api/v1/dividends [get]
func HandleGetAllDividends(manager *DividendsManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dividends, err := manager.CalculateDividendsForAllTickers()
		if err != nil {
			logging.GetLogger().Error("Failed to calculate dividends", err)
			common.WriteJSONError(w, fmt.Sprintf("Failed to calculate dividends for all tickers [%v]", err), http.StatusInternalServerError)
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

	mux.HandleFunc("/api/v1/dividends", HandleGetAllDividends(manager))
}
