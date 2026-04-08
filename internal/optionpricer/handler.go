package optionpricer

import (
	"encoding/json"
	"net/http"

	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// HandlePriceOption handles option pricing requests.
// @Summary Price an equity option
// @Description Prices an American equity option using a Black-Scholes approximation. If rate is omitted, the backend interpolates the latest Fed H15 Treasury constant maturity curve for the option tenor and falls back to 3.75 percent if the fetch fails. If volatility is omitted, the backend first tries to imply it from the supplied premium and otherwise estimates annualized volatility from recent underlying history using a selectable 30, 60, 180, or 360 day lookback.
// @Tags options
// @Accept json
// @Produce json
// @Param request body PriceRequest true "Option pricing request"
// @Success 200 {object} PriceResponse "Option valuation and Greeks"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to price option"
// @Router /api/v1/options/price [post]
func HandlePriceOption(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PriceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := service.Price(req)
		if err != nil {
			logging.GetLogger().Error("Failed to price option:", err)
			common.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logging.GetLogger().Error("Failed to encode option pricing response:", err)
			common.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
		}
	}
}

func RegisterHandlers(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("/api/v1/options/price", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandlePriceOption(service).ServeHTTP(w, r)
	})
}
