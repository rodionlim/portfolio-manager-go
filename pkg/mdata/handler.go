package mdata

import (
	"encoding/json"
	"net/http"
	"strings"
)

// @Summary Get market data for a single ticker
// @Description Retrieves current market data for a specified stock ticker
// @Tags market-data
// @Accept json
// @Produce json
// @Param ticker path string true "Stock ticker symbol"
// @Success 200 {object} interface{} "Market data for the ticker"
// @Failure 400 {string} string "Bad request - Ticker is required"
// @Failure 500 {string} string "Internal server error"
// @Router /mdata/price/{ticker} [get]
func HandleTickerGet(mdataSvc *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/mdata/price/")
		if ticker == "" {
			http.Error(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		// TODO: make generic across different asset class
		data, err := mdataSvc.GetStockPrice(ticker)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// @Summary Get market data for multiple tickers
// @Description Retrieves current market data for multiple stock tickers
// @Tags market-data
// @Accept json
// @Produce json
// @Param tickers query string true "Comma-separated list of stock ticker symbols"
// @Success 200 {object} map[string]interface{} "Market data for all requested tickers"
// @Failure 400 {string} string "Bad request - Tickers query parameter is required"
// @Router /mdata/tickers/price [get]
func HandleTickersGet(mdataSvc *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tickers := r.URL.Query().Get("tickers")
		if tickers == "" {
			http.Error(w, "Tickers query parameter is required", http.StatusBadRequest)
			return
		}

		tickerList := strings.Split(tickers, ",")
		data := make(map[string]interface{})

		for _, ticker := range tickerList {
			mdata, err := mdataSvc.GetStockPrice(ticker)
			if err != nil {
				data[ticker] = err.Error()
				continue
			}
			data[ticker] = mdata
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// @Summary Get dividend metadata for a ticker
// @Description Retrieves dividend history data for a specified stock ticker
// @Tags market-data
// @Accept json
// @Produce json
// @Param ticker path string true "Stock ticker symbol"
// @Success 200 {object} interface{} "Dividend data for the ticker"
// @Failure 400 {string} string "Bad request - Ticker is required"
// @Failure 500 {string} string "Internal server error"
// @Router /mdata/dividend/{ticker} [get]
func HandleDividendsGet(mdataSvc *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/mdata/dividend/")
		if ticker == "" {
			http.Error(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		data, err := mdataSvc.GetDividendsMetadata(ticker)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// RegisterHandlers registers the handlers for the market data service
func RegisterHandlers(mux *http.ServeMux, mdataSvc *Manager) {
	mux.HandleFunc("/mdata/price/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickerGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mdata/tickers/price", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickersGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mdata/dividend/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleDividendsGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
