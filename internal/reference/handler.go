package reference

import (
	"encoding/json"
	"net/http"
)

// @Summary Get reference data
// @Description Retrieves all reference data
// @Tags Reference
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} error
// @Router /refdata [get]
func HandleReferenceDataGet(refSvc *ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := refSvc.GetAllTickers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// RegisterHandlers registers the handlers for the reference data service
func RegisterHandlers(mux *http.ServeMux, refSvc *ReferenceManager) {
	mux.HandleFunc("/refdata", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleReferenceDataGet(refSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
