package historical

import (
	"encoding/json"
	"net/http"
	"strconv"

	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/types"
)

type HistoricalDataManager interface {
	GetAssetConfigs() ([]AssetConfig, error)
	UpdateAssetConfig(config AssetConfig) error
	EnableAssetConfig(ticker string, enabled bool) error
	RemoveAssetConfig(ticker string) error
	SyncAssetData(ticker string) (string, error)
	GetHistoricalAssetData(ticker string, from, to int64, page, limit int) ([]types.AssetData, int, error)
}

// HandleGetAssetConfigs handles GET /api/v1/historical/config
// @Summary Get historical asset configurations
// @Description Get configurations for historical data collection
// @Tags historical
// @Produce json
// @Success 200 {array} AssetConfig
// @Failure 500 {object} common.ErrorResponse "Failed to get asset configs"
// @Router /api/v1/historical/config [get]
func HandleGetAssetConfigs(service HistoricalDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configs, err := service.GetAssetConfigs()
		if err != nil {
			common.WriteJSONError(w, "Failed to get asset configs: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	}
}

// HandleUpdateAssetConfig handles POST /api/v1/historical/config
// @Summary Update or add historical asset configuration
// @Description Update or add configuration for historical data collection
// @Tags historical
// @Accept json
// @Produce json
// @Param config body AssetConfig true "Asset Configuration"
// @Success 200 {object} AssetConfig
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to update asset config"
// @Router /api/v1/historical/config [post]
func HandleUpdateAssetConfig(service HistoricalDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config AssetConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			common.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := service.UpdateAssetConfig(config); err != nil {
			common.WriteJSONError(w, "Failed to update asset config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	}
}

// HandleRemoveAssetConfig handles DELETE /api/v1/historical/config/{ticker}
// @Summary Remove historical asset configuration
// @Description Remove configuration for historical data collection
// @Tags historical
// @Param ticker path string true "Asset Ticker"
// @Success 200 {object} map[string]string
// @Failure 500 {object} common.ErrorResponse "Failed to remove asset config"
// @Router /api/v1/historical/config/{ticker} [delete]
func HandleRemoveAssetConfig(service HistoricalDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := r.PathValue("ticker")
		if ticker == "" {
			// Fallback for older mux or manual parsing if PathValue not populated
			// (Only works if registered with {ticker})
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		if err := service.RemoveAssetConfig(ticker); err != nil {
			common.WriteJSONError(w, "Failed to remove asset config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	}
}

// HandleSyncAssetData handles POST /api/v1/historical/sync
// @Summary Trigger on-demand sync for an asset
// @Description Trigger immediate historical data sync for a specific asset
// @Tags historical
// @Accept json
// @Produce json
// @Param request body map[string]string true "Request with ticker"
// @Success 200 {object} map[string]string
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to sync data"
// @Router /api/v1/historical/sync [post]
func HandleSyncAssetData(service HistoricalDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Ticker string `json:"ticker"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Ticker == "" {
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		msg, err := service.SyncAssetData(req.Ticker)
		if err != nil {
			common.WriteJSONError(w, "Failed to sync data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "synced", "message": msg})
	}
}

// HandleGetHistoricalAssetData handles GET /api/v1/historical/data/{ticker}
// @Summary Get historical data for an asset
// @Description Get paginated historical data for a specific asset
// @Tags historical
// @Produce json
// @Param ticker path string true "Asset Ticker"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page limit (default 20)"
// @Param from query int false "From timestamp (unix)"
// @Param to query int false "To timestamp (unix)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} common.ErrorResponse "Failed to get historical data"
// @Router /api/v1/historical/data/{ticker} [get]
func HandleGetHistoricalAssetData(service HistoricalDataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := r.PathValue("ticker")
		if ticker == "" {
			ticker = r.URL.Query().Get("ticker")
		}
		if ticker == "" {
			common.WriteJSONError(w, "Ticker is required", http.StatusBadRequest)
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 {
			limit = 20
		}
		from, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
		to, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)

		data, total, err := service.GetHistoricalAssetData(ticker, from, to, page, limit)
		if err != nil {
			common.WriteJSONError(w, "Failed to get historical data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"data":  data,
			"total": total,
			"page":  page,
			"limit": limit,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
