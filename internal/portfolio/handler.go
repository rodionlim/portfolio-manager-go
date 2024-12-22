// Package portfolio provides handlers for managing investment portfolio operations
// @title Portfolio Manager API
// @version 1.0
// @description API for managing investment portfolio positions
package portfolio

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/logging"
)

// HandlePositionsGet handles retrieving all positions from the portfolio service.
// @Summary Get all portfolio positions
// @Description Retrieves all positions currently in the portfolio
// @Tags portfolio
// @Produce json
// @Success 200 {array} Position
// @Failure 500 {object} error
// @Router /api/v1/portfolio/positions [get]
func HandlePositionsGet(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions, err := portfolio.GetAllPositions()
		if err != nil {
			logging.GetLogger().Errorf("Failed to get positions: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(positions)
	}
}

// RegisterHandlers registers the handlers for the portfolio service.
func RegisterHandlers(mux *http.ServeMux, portfolio *Portfolio) {
	mux.HandleFunc("/api/v1/portfolio/positions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandlePositionsGet(portfolio).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
