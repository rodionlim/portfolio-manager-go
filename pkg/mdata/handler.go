package mdata

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/types"
	"strings"
)

// @Summary Get market data for a single ticker
// @Description Retrieves current market data for a specified ticker
// @Tags market-data
// @Accept json
// @Produce json
// @Param ticker path string true "Ticker symbol (see reference data)"
// @Success 200 {object} interface{} "Market data for the ticker"
// @Failure 400 {string} string "Bad request - Ticker is required"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/mdata/price/{ticker} [get]
func HandleTickerGet(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/price/")
		if ticker == "" {
			http.Error(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		data, err := mdataSvc.GetAssetPrice(ticker)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// @Summary Get market data for multiple tickers
// @Description Retrieves current market data for multiple asset tickers
// @Tags market-data
// @Accept json
// @Produce json
// @Param tickers query string true "Comma-separated list of asset ticker symbols"
// @Success 200 {object} map[string]interface{} "Market data for all requested tickers"
// @Failure 400 {string} string "Bad request - Tickers query parameter is required"
// @Router /api/v1/mdata/tickers/price [get]
func HandleTickersGet(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tickers := r.URL.Query().Get("tickers")
		if tickers == "" {
			http.Error(w, "Tickers query parameter is required", http.StatusBadRequest)
			return
		}

		tickerList := strings.Split(tickers, ",")
		data := make(map[string]interface{})

		for _, ticker := range tickerList {
			mdata, err := mdataSvc.GetAssetPrice(ticker)
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
// @Param ticker path string true "Asset ticker symbol"
// @Success 200 {object} interface{} "Dividend data for the ticker"
// @Failure 400 {string} string "Bad request - Ticker is required"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/mdata/dividend/{ticker} [get]
func HandleDividendsGet(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/dividend/")
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

// @Summary Store custom dividend metadata for a ticker
// @Description Stores user-provided dividend history data for a specified stock ticker
// @Tags market-data
// @Accept json
// @Produce json
// @Param ticker path string true "Asset ticker symbol"
// @Param dividend_data body []types.DividendsMetadata true "Array of dividend metadata to store"
// @Success 200 {object} common.SuccessResponse "Dividends metadata stored successfully"
// @Failure 400 {string} string "Bad request - Ticker is required or invalid request body"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/mdata/dividend/{ticker} [post]
func HandleDividendsStore(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/dividend/")
		if ticker == "" {
			http.Error(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		// fetch the body which should cotain a list of dividends metadata
		var dividends []types.DividendsMetadata
		err := json.NewDecoder(r.Body).Decode(&dividends)
		if err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		err = mdataSvc.StoreCustomDividendsMetadata(ticker, dividends)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.SuccessResponse{Message: "Dividends metadata stored successfully"})
	}
}

// RegisterHandlers registers the handlers for the market data service
func RegisterHandlers(mux *http.ServeMux, mdataSvc MarketDataManager) {
	mux.HandleFunc("/api/v1/mdata/price/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickerGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/tickers/price", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickersGet(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/dividend/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleDividendsGet(mdataSvc).ServeHTTP(w, r)
		case http.MethodPost:
			HandleDividendsStore(mdataSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
