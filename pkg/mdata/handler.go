package mdata

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HandleTickerGet handles retrieving market data for a single ticker
func HandleTickerGet(mdataSvc *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/mdata/ticker/")
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

// HandleTickersGet handles retrieving market data for multiple tickers
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

// RegisterHandlers registers the handlers for the market data service
func RegisterHandlers(mux *http.ServeMux, mdataSvc *Manager) {
	mux.HandleFunc("/mdata/ticker/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickerGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mdata/tickers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickersGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
