package mdata

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"strconv"
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
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
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
			logging.GetLogger().Error("Tickers query parameter is required")
			common.WriteJSONError(w, "Tickers query parameter is required", http.StatusBadRequest)
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
// @Router /api/v1/mdata/dividends/{ticker} [get]
func HandleDividendsGet(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/dividends/")
		if ticker == "" {
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		data, err := mdataSvc.GetDividendsMetadata(ticker)
		if err != nil {
			logging.GetLogger().Error("Failed to get dividends metadata", err)
			common.WriteJSONError(w, "Failed to get dividends metadata", http.StatusInternalServerError)
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
// @Router /api/v1/mdata/dividends/{ticker} [post]
func HandleDividendsStore(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/dividends/")
		if ticker == "" {
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		// fetch the body which should cotain a list of dividends metadata
		var dividends []types.DividendsMetadata
		err := json.NewDecoder(r.Body).Decode(&dividends)
		if err != nil {
			common.WriteJSONError(w, "Failed to decode request body", http.StatusBadRequest)
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

// @Summary Import dividend data from CSV stream
// @Description Handles the import of dividend data from an uploaded CSV file for a multiple tickers
// @Tags market-data
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file containing dividend data"
// @Success 200 {object} common.SuccessResponse "Successfully imported dividends data"
// @Failure 400 {string} string "Bad request - Invalid form data"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/mdata/dividends/import [post]
func HandleDividendsImportCSVStream(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form data with a reasonable max memory
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			logging.GetLogger().Error("Failed to parse form", err)
			common.WriteJSONError(w, "Failed to parse upload form", http.StatusBadRequest)
			return
		}

		// Get the file from the form data
		file, handler, err := r.FormFile("file")
		if err != nil {
			logging.GetLogger().Error("Failed to get file from form", err)
			common.WriteJSONError(w, "Failed to get uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		logging.GetLogger().Info("Received file: ", handler.Filename)

		// Process the CSV file
		reader := csv.NewReader(file)
		count, err := mdataSvc.ImportCustomDividendsFromCSVReader(reader)
		if err != nil {
			logging.GetLogger().Error("Failed to import CSV data", err)
			common.WriteJSONError(w, "Error processing CSV data: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.SuccessResponse{
			Message: "Successfully imported " + strconv.Itoa(count) + " custom dividends data",
		})
	}
}

// @Summary Get historical price data for a ticker
// @Description Retrieves historical price data for a specified ticker between start and end dates
// @Tags market-data
// @Accept json
// @Produce json
// @Param ticker path string true "Ticker symbol (see reference data)"
// @Param start query string true "Start date in YYYYMMDD format"
// @Param end query string false "End date in YYYYMMDD format (defaults to today)"
// @Success 200 {array} types.AssetData "Historical price data for the ticker"
// @Failure 400 {string} string "Bad request - Invalid parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/mdata/price/historical/{ticker} [get]
func HandleHistoricalDataGet(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ticker from URL path
		ticker := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/price/historical/")
		if ticker == "" {
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		// Get start date from query params (required)
		startDateStr := r.URL.Query().Get("start")
		if startDateStr == "" {
			common.WriteJSONError(w, "Start date is required (format: YYYYMMDD)", http.StatusBadRequest)
			return
		}

		// Parse start date
		startDate, err := common.ParseDateToEpoch(startDateStr)
		if err != nil {
			common.WriteJSONError(w, "Invalid start date format, use YYYYMMDD", http.StatusBadRequest)
			return
		}

		// Get end date from query params (optional, defaults to today)
		endDate := common.GetCurrentEpochTime()
		endDateStr := r.URL.Query().Get("end")
		if endDateStr != "" {
			endDate, err = common.ParseDateToEpoch(endDateStr)
			if err != nil {
				common.WriteJSONError(w, "Invalid end date format, use YYYYMMDD", http.StatusBadRequest)
				return
			}
		}

		data, err := mdataSvc.GetHistoricalData(ticker, startDate, endDate)
		if err != nil {
			logging.GetLogger().Error("Failed to get historical data", err)
			common.WriteJSONError(w, "Failed to get historical data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// @Summary Get benchmark interest rates for a country
// @Description Retrieves benchmark interest rates for a specified country
// @Tags market-data
// @Accept json
// @Produce json
// @Param country path string true "Country code (e.g., SG for Singapore)"
// @Param points query int false "Number of recent records to return (default: 250)"
// @Success 200 {array} types.InterestRates "Benchmark interest rates for the country"
// @Failure 400 {string} string "Bad request - Invalid parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/mdata/interest-rates/{country} [get]
func HandleInterestRatesGet(mdataSvc MarketDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract country from URL path
		country := strings.TrimPrefix(r.URL.Path, "/api/v1/mdata/interest-rates/")
		if country == "" {
			common.WriteJSONError(w, "Country is required", http.StatusBadRequest)
			return
		}

		// Get points from query params (optional, defaults to 250)
		points := 250
		pointsStr := r.URL.Query().Get("points")
		if pointsStr != "" {
			var err error
			points, err = strconv.Atoi(pointsStr)
			if err != nil {
				common.WriteJSONError(w, "Invalid points parameter, must be an integer", http.StatusBadRequest)
				return
			}
			if points <= 0 {
				common.WriteJSONError(w, "Points must be greater than 0", http.StatusBadRequest)
				return
			}
		}

		data, err := mdataSvc.FetchBenchmarkInterestRates(country, points)
		if err != nil {
			logging.GetLogger().Error("Failed to get benchmark interest rates", err)
			common.WriteJSONError(w, "Failed to get benchmark interest rates: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// RegisterHandlers registers the handlers for the market data service
func RegisterHandlers(mux *http.ServeMux, mdataSvc MarketDataManager) {
	mux.HandleFunc("/api/v1/mdata/price/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickerGet(mdataSvc).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/tickers/price", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleTickersGet(mdataSvc).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/dividends/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleDividendsGet(mdataSvc).ServeHTTP(w, r)
		case http.MethodPost:
			HandleDividendsStore(mdataSvc).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/dividends/upload", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			logging.GetLogger().Info("Received request to import dividends data")
			HandleDividendsImportCSVStream(mdataSvc).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/price/historical/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleHistoricalDataGet(mdataSvc).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/mdata/interest-rates/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleInterestRatesGet(mdataSvc).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
