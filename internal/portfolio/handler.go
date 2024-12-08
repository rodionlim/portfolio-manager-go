package portfolio

import (
	"encoding/json"
	"net/http"
)

// HandlePositionsGet handles retrieving all positions from the portfolio service.
func HandlePositionsGet(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions := portfolio.GetAllPositions()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(positions)
	}
}

// RegisterHandlers registers the handlers for the portfolio service.
func RegisterHandlers(mux *http.ServeMux, portfolio *Portfolio) {
	mux.HandleFunc("/portfolio/positions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandlePositionsGet(portfolio).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
