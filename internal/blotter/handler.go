package blotter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/logging"
	"time"
)

// TradeRequest represents the request payload for a trade.
type TradeRequest struct {
	TradeDate string  `json:"tradeDate"`
	Ticker    string  `json:"ticker"`
	Side      string  `json:"side"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	Yield     float64 `json:"yield"`
	Trader    string  `json:"trader"`
	Broker    string  `json:"broker"`
	Account   string  `json:"account"`
	SeqNum    int     `json:"seqNum"` // Sequence number
}

// HandleTradePost handles the addition of trades to the blotter service.
// @Summary Add a new trade
// @Description Add a new trade to the blotter
// @Tags trades
// @Accept  json
// @Produce  json
// @Param   trade  body  TradeRequest  true  "Trade Request"
// @Success 201 {object} Trade
// @Failure 400 {string} string "Invalid request payload"
// @Failure 500 {string} string "Failed to add trade"
// @Router /blotter/trade [post]
func HandleTradePost(blotter *TradeBlotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tradeRequest TradeRequest
		err := json.NewDecoder(r.Body).Decode(&tradeRequest)
		if err != nil {
			http.Error(w, "ERROR: Invalid request payload", http.StatusBadRequest)
			return
		}

		tradeDate, err := time.Parse(time.RFC3339, tradeRequest.TradeDate)
		if err != nil {
			http.Error(w, "ERROR: Invalid trade date format", http.StatusBadRequest)
			return
		}

		trade, err := NewTrade(
			tradeRequest.Side,
			tradeRequest.Quantity,
			tradeRequest.Ticker,
			tradeRequest.Trader,
			tradeRequest.Broker,
			tradeRequest.Account,
			tradeRequest.Price,
			tradeRequest.Yield,
			tradeDate)
		if err != nil {
			http.Error(w, fmt.Sprintf("ERROR: %s", err.Error()), http.StatusBadRequest)
			return
		}

		err = blotter.AddTrade(*trade)
		if err != nil {
			logging.GetLogger().Error("Failed to add trade", err)
			http.Error(w, "ERROR: Failed to add trade", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(trade)
	}
}

// HandleTradeGet handles retrieving trades from the blotter service.
// @Summary Get all trades
// @Description Retrieve all trades from the blotter
// @Tags trades
// @Produce  json
// @Success 200 {array} Trade
// @Router /blotter/trade [get]
func HandleTradeGet(blotter *TradeBlotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trades := blotter.GetTrades()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	}
}

// HandleTradeImportCSV handles importing trades from a CSV file
// @Summary Import trades from CSV
// @Description Import trades from a CSV file
// @Tags trades
// @Accept  multipart/form-data
// @Produce  json
// @Param   file  formData  file  true  "CSV file"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Failed to get file from request"
// @Failure 500 {string} string "Failed to import trades"
// @Router /blotter/import [post]
func HandleTradeImportCSV(blotter *TradeBlotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "ERROR: Failed to get file from request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		err = blotter.ImportFromCSVReader(reader)
		if err != nil {
			http.Error(w, fmt.Sprintf("ERROR: %s", err.Error()), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleTradeExportCSV handles exporting trades to a CSV file
// @Summary Export trades to CSV
// @Description Export all trades to a CSV file
// @Tags trades
// @Produce  text/csv
// @Success 200 {file} file "trades.csv"
// @Failure 500 {string} string "Failed to export trades"
// @Router /blotter/export [get]
func HandleTradeExportCSV(blotter *TradeBlotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trades, err := blotter.ExportToCSVBytes()
		if err != nil {
			http.Error(w, fmt.Sprintf("ERROR: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=trades.csv")

		w.Write(trades)
	}
}

// RegisterHandlers registers the handlers for the blotter service.
func RegisterHandlers(mux *http.ServeMux, blotter *TradeBlotter) {
	mux.HandleFunc("/blotter/trade", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandleTradePost(blotter).ServeHTTP(w, r)
		case http.MethodGet:
			HandleTradeGet(blotter).ServeHTTP(w, r)
		default:
			http.Error(w, "ERROR: Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/blotter/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "ERROR: Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleTradeImportCSV(blotter).ServeHTTP(w, r)
	})

	mux.HandleFunc("/blotter/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "ERROR: Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleTradeExportCSV(blotter).ServeHTTP(w, r)
	})
}
