package rdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-manager/pkg/logging"
	"strings"

	"gopkg.in/yaml.v2"
)

// ErrorResponse represents the error response payload.
type ErrorResponse struct {
	Message string `json:"message"`
}

// SuccessResponse represents the success response payload.
type SuccessResponse struct {
	Message string `json:"message"`
}

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
			logging.GetLogger().Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// @Summary Add reference data
// @Description Adds reference data
// @Tags Reference
// @Accept json
// @Produce json
// @Param body body TickerReference true "Reference data"
// @Success 201 {object} string
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /api/v1/refdata [post]
func HandleReferenceDataPost(refSvc ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refDataRequest TickerReference
		err := json.NewDecoder(r.Body).Decode(&refDataRequest)
		if err != nil {
			logging.GetLogger().Error(err)
			writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		id, err := refSvc.AddTicker(refDataRequest)
		if err != nil {
			msg := fmt.Sprintf("Failed to add reference data [%s]\n", id)
			logging.GetLogger().Error(msg, err)
			writeJSONError(w, msg, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id)
	}
}

// @Summary Delete reference data
// @Description Deletes reference data by id
// @Tags Reference
// @Accept json
// @Produce json
// @Param body body []string true "Reference data ids (underlying tickers)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /api/v1/refdata [delete]
func HandleReferenceDataDelete(refSvc ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ids []string
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			logging.GetLogger().Error(err)
			writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		for _, id := range ids {
			err = refSvc.DeleteTicker(strings.ToUpper(id))
			if err != nil {
				logging.GetLogger().Error("Failed to delete ref data", id, err)
				writeJSONError(w, fmt.Sprintf("Failed to delete ref data for id %s", id), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse{Message: fmt.Sprintf("Reference data deleted successfully for ids %v", ids)})
	}
}

// @Summary Update reference data
// @Description Updates reference data
// @Tags Reference
// @Accept json
// @Produce json
// @Param body body TickerReference true "Reference data"
// @Success 201 {object} TickerReference
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /api/v1/refdata [put]
func HandleReferenceDataUpdate(refSvc ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refDataRequest TickerReference
		err := json.NewDecoder(r.Body).Decode(&refDataRequest)
		if err != nil {
			logging.GetLogger().Error(err)
			writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		err = refSvc.UpdateTicker(&refDataRequest)
		if err != nil {
			logging.GetLogger().Error("Failed to update reference data", err)
			writeJSONError(w, "Failed to update reference data", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(refDataRequest)
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
			logging.GetLogger().Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", "attachment; filename=refdata.yaml")

		w.Write(refData)
	}
}

// @Summary Import reference data
// @Description Imports reference data from a YAML file
// @Tags Reference
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "YAML file"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /api/v1/refdata/import [post]
func HandleReferenceDataImport(refSvc ReferenceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10MB max
		if err != nil {
			writeJSONError(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "Failed to get uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		var tickers []TickerReference
		if err := yaml.NewDecoder(file).Decode(&tickers); err != nil {
			writeJSONError(w, "Failed to parse YAML file", http.StatusBadRequest)
			return
		}

		inserted, updated := 0, 0
		for _, ticker := range tickers {
			_, err := refSvc.GetTicker(ticker.ID)
			if err == nil {
				// Exists, update
				err = refSvc.UpdateTicker(&ticker)
				if err == nil {
					updated++
				}
			} else {
				// Not found, add
				_, err = refSvc.AddTicker(ticker)
				if err == nil {
					inserted++
				}
			}
		}

		msg := fmt.Sprintf("Reference data import complete: %d inserted, %d updated", inserted, updated)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse{Message: msg})
	}
}

// writeJSONError writes an error message in JSON format to the response.
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// RegisterHandlers registers the handlers for the reference data service
func RegisterHandlers(mux *http.ServeMux, refSvc ReferenceManager) {
	mux.HandleFunc("/api/v1/refdata", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleReferenceDataGet(refSvc).ServeHTTP(w, r)
		case http.MethodPost:
			HandleReferenceDataPost(refSvc).ServeHTTP(w, r)
		case http.MethodPut:
			HandleReferenceDataUpdate(refSvc).ServeHTTP(w, r)
		case http.MethodDelete:
			HandleReferenceDataDelete(refSvc).ServeHTTP(w, r)
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

	mux.HandleFunc("/api/v1/refdata/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleReferenceDataImport(refSvc).ServeHTTP(w, r)
	})
}
