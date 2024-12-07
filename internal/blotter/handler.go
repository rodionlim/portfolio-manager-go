package blotter

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/logging"
	"time"
)

// TradeRequest represents the request payload for a trade.
type TradeRequest struct {
	TradeDate     string  `json:"tradeDate"`
	Ticker        string  `json:"ticker"`
	Side          string  `json:"side"`
	Quantity      float64 `json:"quantity"`
	AssetClass    string  `json:"assetClass"`
	AssetSubClass string  `json:"assetSubClass"`
	Price         float64 `json:"price"`
	Yield         float64 `json:"yield"`
	Trader        string  `json:"trader"`
	SeqNum        int     `json:"seqNum"` // Sequence number
}

// HandleTradePost handles the addition of trades to the blotter service.
func HandleTradePost(blotter *TradeBlotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tradeRequest TradeRequest
		err := json.NewDecoder(r.Body).Decode(&tradeRequest)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		tradeDate, err := time.Parse(time.RFC3339, tradeRequest.TradeDate)
		if err != nil {
			http.Error(w, "Invalid trade date format", http.StatusBadRequest)
			return
		}

		trade, err := NewTrade(tradeRequest.Side, tradeRequest.Quantity, tradeRequest.AssetClass, tradeRequest.AssetSubClass, tradeRequest.Ticker, tradeRequest.Trader, tradeRequest.Price, tradeRequest.Yield, tradeDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = blotter.AddTrade(*trade)
		if err != nil {
			logging.GetLogger().Error("Failed to add trade", err)
			http.Error(w, "Failed to add trade", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(trade)
	}
}

// HandleTradeGet handles retrieving trades from the blotter service.
func HandleTradeGet(blotter *TradeBlotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trades := blotter.GetTrades()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
