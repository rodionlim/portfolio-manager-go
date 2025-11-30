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

// HandlePositionsGetLite handles retrieving all positions without enrichment (no FX, dividends, market values).
// @Summary Get all portfolio positions without enrichment
// @Description Retrieves all positions without costly enrichment operations (FX rates, dividends, market values). Useful for UI components that only need ticker information.
// @Tags portfolio
// @Produce json
// @Success 200 {array} Position
// @Router /api/v1/portfolio/positions/lite [get]
func HandlePositionsGetLite(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions := portfolio.GetAllPositionsWithoutEnrichment()
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

// HandlePositionDelete handles deleting a single position from the portfolio service.
// @Summary Delete a single portfolio position
// @Description Deletes a single position by book and ticker
// @Tags portfolio
// @Produce json
// @Param book query string true "Book name"
// @Param ticker query string true "Ticker symbol"
// @Success 200 {string} string "Position deleted successfully"
// @Failure 400 {object} common.ErrorResponse "Missing required parameters"
// @Failure 404 {object} common.ErrorResponse "Position not found"
// @Failure 500 {object} common.ErrorResponse "Failed to delete position"
// @Router /api/v1/portfolio/position [delete]
func HandlePositionDelete(portfolio *Portfolio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		book := r.URL.Query().Get("book")
		ticker := r.URL.Query().Get("ticker")

		if book == "" || ticker == "" {
			common.WriteJSONError(w, "Missing required parameters: book and ticker", http.StatusBadRequest)
			return
		}

		err := portfolio.DeletePosition(book, ticker)
		if err != nil {
			logging.GetLogger().Errorf("Failed to delete position for book %s, ticker %s: %v", book, ticker, err)
			common.WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Position deleted successfully")
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

	mux.HandleFunc("/api/v1/portfolio/positions/lite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandlePositionsGetLite(portfolio).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/portfolio/position", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			HandlePositionDelete(portfolio).ServeHTTP(w, r)
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
