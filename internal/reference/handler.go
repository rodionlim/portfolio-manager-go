package reference

import (
	"encoding/json"
	"net/http"
)

// HandleReferenceDataGet handles retrieving all reference data
func HandleReferenceDataGet(refSvc *ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := refSvc.GetRefData()
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
