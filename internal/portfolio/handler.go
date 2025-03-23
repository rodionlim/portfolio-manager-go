// Package portfolio provides handlers for managing investment portfolio operations
// @title Portfolio Manager API
// @version 1.0
// @description API for managing investment portfolio positions
package portfolio

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// HandlePositionsGet handles retrieving all positions from the portfolio service.
// @Summary Get all portfolio positions
// @Description Retrieves all positions currently in the portfolio
// @Tags portfolio
// @Produce json
// @Success 200 {array} Position
// @Failure 500 {object} common.ErrorResponse "Failed to get positions"
// @Router /api/v1/portfolio/positions [get]
func HandlePositionsGet(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions, err := portfolio.GetAllPositions()
		if err != nil {
			logging.GetLogger().Errorf("Failed to get positions: %v", err)
			common.WriteJSONError(w, "Failed to get positions", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(positions)
	}
}

// HandlePositionsDelete handles deleting all positions from the portfolio service.
// @Summary Delete all portfolio positions
// @Description Deletes all positions currently in the portfolio
// @Tags portfolio
// @Produce json
// @Success 200 {string} string "Positions deleted successfully"
// @Failure 500 {object} common.ErrorResponse "Failed to delete positions"
// @Router /api/v1/portfolio/positions [delete]
func HandlePositionsDelete(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := portfolio.DeletePositions()
		if err != nil {
			logging.GetLogger().Errorf("Failed to delete positions: %v", err)
			common.WriteJSONError(w, "Failed to delete positions", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Positions deleted successfully")
	}
}

// HandlePositionsCleanup handles closing positions that have expired
// @Summary Close positions that have expired
// @Description Closes positions that have expired without a corresponding closure trade
// @Tags portfolio
// @Produce json
// @Success 200 {array} []string "Trade Ids of closed trades"
// @Failure 500 {object} common.ErrorResponse "Failed to auto close positions"
// @Router /api/v1/portfolio/cleanup [post]
func HandlePositionsCleanup(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		closedTrades, err := portfolio.AutoCloseTrades()
		if err != nil {
			logging.GetLogger().Errorf("Failed to auto close positions: %v", err)
			common.WriteJSONError(w, "Failed to auto close positions", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(closedTrades)
	}
}

// RegisterHandlers registers the handlers for the portfolio service.
func RegisterHandlers(mux *http.ServeMux, portfolio *Portfolio) {
	mux.HandleFunc("/api/v1/portfolio/positions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandlePositionsGet(portfolio).ServeHTTP(w, r)
		case http.MethodDelete:
			HandlePositionsDelete(portfolio).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/portfolio/cleanup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandlePositionsCleanup(portfolio).ServeHTTP(w, r)
	})
}
