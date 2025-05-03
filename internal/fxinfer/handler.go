package fxinfer

import (
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// HandleTradeExportWithInferredFX handles exporting trades with inferred FX rates
// @Summary Export trades with inferred FX rates
// @Description Export all trades as CSV with FX rates inferred where missing
// @Tags trades
// @Produce text/csv
// @Success 200 {file} file "trades_with_fx.csv"
// @Failure 500 {object} common.ErrorResponse "Failed to export trades"
// @Router /api/v1/blotter/export-with-fx [get]
func HandleTradeExportWithInferredFX(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trades, err := service.ExportTradesWithInferredFX()
		if err != nil {
			logging.GetLogger().Error("Failed to export trades with inferred FX rates:", err)
			common.WriteJSONError(w, "Failed to export trades: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=trades_with_fx.csv")

		w.Write(trades)
	}
}

// RegisterHandlers registers the handlers for the FX inference service
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("/api/v1/blotter/export-with-fx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleTradeExportWithInferredFX(service).ServeHTTP(w, r)
	})
}
