package rdata

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
// @Router /api/v1/refdata [get]
func HandleReferenceDataGet(refSvc ReferenceManager) http.HandlerFunc {
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

// @Summary Export reference data
// @Description Exports reference data in yaml format
// @Tags Reference
// @Produce application/x-yaml
// @Success 200 {file} file "refdata.yaml"
// @Failure 500 {object} error
// @Router /api/v1/refdata/export [get]
func HandleReferenceDataExport(refSvc ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refData, err := refSvc.ExportToYamlBytes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", "attachment; filename=refdata.yaml")

		w.Write(refData)
	}
}

// RegisterHandlers registers the handlers for the reference data service
func RegisterHandlers(mux *http.ServeMux, refSvc ReferenceManager) {
	mux.HandleFunc("/api/v1/refdata", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleReferenceDataGet(refSvc).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/refdata/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleReferenceDataExport(refSvc).ServeHTTP(w, r)
	})
}
