package fxinfer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// HandleTradeExportWithInferredFX handles exporting trades with inferred FX rates
// @Summary Export trades with inferred FX rates
// @Description Export all trades as CSV with FX rates inferred where missing
// @Tags trades
// @Produce text/csv
// @Success 200 {file} file "trades_with_fx.csv"
// @Failure 500 {object} common.ErrorResponse "Failed to export trades"
// @Router /api/v1/blotter/export-with-fx [get]
func HandleTradeExportWithInferredFX(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trades, err := service.ExportTradesWithInferredFX()
		if err != nil {
			logging.GetLogger().Error("Failed to export trades with inferred FX rates:", err)
			common.WriteJSONError(w, "Failed to export trades: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=trades_with_fx.csv")

		w.Write(trades)
	}
}

// HandleGetCurrentFXRates handles getting current FX rates for all currencies in the blotter
// @Summary Get current FX rates
// @Description Get current FX rates for all currencies in blotter trades
// @Tags fx
// @Produce json
// @Success 200 {object} map[string]float64 "Map of currencies to their current FX rates"
// @Failure 500 {object} common.ErrorResponse "Failed to fetch FX rates"
// @Router /api/v1/blotter/fx [get]
func HandleGetCurrentFXRates(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all trades from blotter
		trades := service.blotterSvc.GetTrades()

		// Create a cache for ticker reference data to avoid duplicate lookups
		tickerRefCache := make(map[string]string) // map[ticker]currency

		// Extract unique currencies from trades
		currencies := make(map[string]bool)
		for _, trade := range trades {
			// Check if we already have reference data for this ticker in cache
			currency, exists := tickerRefCache[trade.Ticker]
			if !exists {
				// Get reference data for the trade's ticker
				refData, err := service.rdataSvc.GetTicker(trade.Ticker)
				if err != nil {
					logging.GetLogger().Debugf("Failed to get reference data for ticker %s: %v", trade.Ticker, err)
					continue
				}
				// Cache the currency for this ticker
				tickerRefCache[trade.Ticker] = refData.Ccy
				currency = refData.Ccy
			}

			// Add currency to the set if not already present
			currencies[currency] = true
		}

		// Always include the base currency
		currencies[service.baseCcy] = true

		// Fetch current FX rates for each currency
		fxRates := make(map[string]float64)
		for currency := range currencies {
			// Base currency always has FX rate of 1.0
			if currency == service.baseCcy {
				fxRates[currency] = 1.0
				continue
			}

			// Construct FX ticker - e.g., USD-SGD
			fxTicker := fmt.Sprintf("%s-%s", currency, service.baseCcy)

			// Get current FX rate
			assetData, err := service.mdataSvc.GetAssetPrice(fxTicker)
			if err != nil {
				logging.GetLogger().Errorf("Failed to get current FX rate for %s: %v", fxTicker, err)
				continue
			}

			// Store the rate (in terms of base currency)
			fxRates[currency] = assetData.Price
		}

		// Return the FX rates as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fxRates); err != nil {
			logging.GetLogger().Error("Failed to write FX rates as JSON:", err)
			common.WriteJSONError(w, "Failed to write FX rates", http.StatusInternalServerError)
		}
	}
}

// RegisterHandlers registers the handlers for the FX inference service
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("/api/v1/blotter/export-with-fx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleTradeExportWithInferredFX(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/blotter/fx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleGetCurrentFXRates(service).ServeHTTP(w, r)
	})
}
