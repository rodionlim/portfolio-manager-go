package blotter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"strings"
	"time"
)

// HandleConfirmationUpload handles uploading a trade confirmation
// @Summary Upload trade confirmation
// @Description Upload a trade confirmation file for a specific trade
// @Tags confirmations
// @Accept  multipart/form-data
// @Produce  json
// @Param   tradeId  path  string  true  "Trade ID"
// @Param   file  formData  file  true  "Confirmation file"
// @Success 200 {object} common.SuccessResponse "Confirmation uploaded successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to upload confirmation"
// @Router /api/v1/blotter/confirmation/{tradeId} [post]
func HandleConfirmationUpload(confirmationService *ConfirmationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract trade ID from URL path using helper
		tradeID, err := extractTradeIDFromPath(r.URL.Path)
		if err != nil {
			common.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Parse multipart form
		err = r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			logging.GetLogger().Error("Failed to parse form", err)
			common.WriteJSONError(w, "Failed to parse upload form", http.StatusBadRequest)
			return
		}

		// Get file from form
		file, header, err := r.FormFile("file")
		if err != nil {
			logging.GetLogger().Error("Failed to get file from form", err)
			common.WriteJSONError(w, "Failed to get uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate content type from header
		contentType := header.Header.Get("Content-Type")
		validContentTypes := map[string]bool{
			"application/pdf": true,
			"image/png":       true,
			"image/jpeg":      true,
		}
		if !validContentTypes[contentType] {
			common.WriteJSONError(w, "Invalid file type. Only PDF, PNG, and JPEG files are allowed", http.StatusBadRequest)
			return
		}

		// Read file data
		data, err := io.ReadAll(file)
		if err != nil {
			logging.GetLogger().Error("Failed to read file data", err)
			common.WriteJSONError(w, "Failed to read file data", http.StatusInternalServerError)
			return
		}

		// Save confirmation
		uploadedDate := time.Now().Format(time.RFC3339)
		err = confirmationService.SaveConfirmation(tradeID, header.Filename, contentType, data, uploadedDate)
		if err != nil {
			logging.GetLogger().Error("Failed to save confirmation", err)
			common.WriteJSONError(w, "Failed to save confirmation: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.SuccessResponse{
			Message: "Confirmation uploaded successfully",
		})
	}
}

// HandleConfirmationDelete handles deleting a trade confirmation
// @Summary Delete trade confirmation
// @Description Delete a trade confirmation for a specific trade
// @Tags confirmations
// @Produce  json
// @Param   tradeId  path  string  true  "Trade ID"
// @Success 200 {object} common.SuccessResponse "Confirmation deleted successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to delete confirmation"
// @Router /api/v1/blotter/confirmation/{tradeId} [delete]
func HandleConfirmationDelete(confirmationService *ConfirmationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract trade ID from URL path using helper
		tradeID, err := extractTradeIDFromPath(r.URL.Path)
		if err != nil {
			common.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = confirmationService.DeleteConfirmation(tradeID)
		if err != nil {
			logging.GetLogger().Error("Failed to delete confirmation", err)
			common.WriteJSONError(w, "Failed to delete confirmation", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.SuccessResponse{
			Message: "Confirmation deleted successfully",
		})
	}
}

// HandleConfirmationGet handles retrieving a trade confirmation
// @Summary Get trade confirmation
// @Description Get a trade confirmation file for a specific trade
// @Tags confirmations
// @Produce  application/octet-stream
// @Param   tradeId  path  string  true  "Trade ID"
// @Success 200 {file} file "Confirmation file"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Confirmation not found"
// @Router /api/v1/blotter/confirmation/{tradeId} [get]
func HandleConfirmationGet(confirmationService *ConfirmationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract trade ID from URL path using helper
		tradeID, err := extractTradeIDFromPath(r.URL.Path)
		if err != nil {
			common.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		confirmation, err := confirmationService.GetConfirmation(tradeID)
		if err != nil {
			common.WriteJSONError(w, "Confirmation not found", http.StatusNotFound)
			return
		}

		// Set appropriate headers with properly encoded filename
		w.Header().Set("Content-Type", confirmation.Metadata.ContentType)
		w.Header().Set("Content-Disposition", "attachment; "+sanitizeContentDispositionFilename(confirmation.Metadata.FileName))

		w.Write(confirmation.Data)
	}
}

// HandleConfirmationMetadataGet handles retrieving confirmation metadata
// @Summary Get trade confirmation metadata
// @Description Get metadata for a trade confirmation
// @Tags confirmations
// @Produce  json
// @Param   tradeId  path  string  true  "Trade ID"
// @Success 200 {object} ConfirmationMetadata
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 404 {object} common.ErrorResponse "Confirmation not found"
// @Router /api/v1/blotter/confirmation/{tradeId}/metadata [get]
func HandleConfirmationMetadataGet(confirmationService *ConfirmationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract trade ID from URL path using helper
		tradeID, err := extractTradeIDFromPath(r.URL.Path)
		if err != nil {
			common.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		metadata, err := confirmationService.GetConfirmationMetadata(tradeID)
		if err != nil {
			common.WriteJSONError(w, "Confirmation not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	}
}

// HandleConfirmationsExport handles exporting confirmations as a zip file
// @Summary Export confirmations
// @Description Export trade confirmations as a zip archive
// @Tags confirmations
// @Accept  json
// @Produce  application/zip
// @Param   tradeIds  body  []string  true  "Trade IDs to export"
// @Success 200 {file} file "confirmations_YYYYMMDD.zip"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Failed to export confirmations"
// @Router /api/v1/blotter/confirmations/export [post]
func HandleConfirmationsExport(confirmationService *ConfirmationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tradeIDs []string
		err := json.NewDecoder(r.Body).Decode(&tradeIDs)
		if err != nil {
			logging.GetLogger().Error(err)
			common.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if len(tradeIDs) == 0 {
			common.WriteJSONError(w, "No trade IDs provided", http.StatusBadRequest)
			return
		}

		zipData, err := confirmationService.ExportConfirmationsAsZip(tradeIDs)
		if err != nil {
			logging.GetLogger().Error("Failed to export confirmations", err)
			common.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Generate filename with current date
		now := time.Now()
		dateString := now.Format("20060102")
		filename := "confirmations_" + dateString + ".zip"

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; "+sanitizeContentDispositionFilename(filename))

		w.Write(zipData)
	}
}

// HandleConfirmationsMetadataGet handles retrieving all confirmation metadata
// @Summary Get all confirmations metadata
// @Description Get metadata for all trade confirmations
// @Tags confirmations
// @Produce  json
// @Success 200 {array} ConfirmationMetadata
// @Failure 500 {object} common.ErrorResponse "Failed to retrieve confirmations"
// @Router /api/v1/blotter/confirmations/metadata [get]
func HandleConfirmationsMetadataGet(confirmationService *ConfirmationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metadata, err := confirmationService.GetAllConfirmationMetadata()
		if err != nil {
			logging.GetLogger().Error("Failed to get confirmations metadata", err)
			common.WriteJSONError(w, "Failed to retrieve confirmations", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	}
}

// RegisterConfirmationHandlers registers the handlers for confirmation endpoints
func RegisterConfirmationHandlers(mux *http.ServeMux, confirmationService *ConfirmationService) {
	// Single confirmation operations
	mux.HandleFunc("/api/v1/blotter/confirmation/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a metadata request
		if strings.HasSuffix(r.URL.Path, "/metadata") {
			if r.Method != http.MethodGet {
				common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			HandleConfirmationMetadataGet(confirmationService).ServeHTTP(w, r)
			return
		}

		// Regular confirmation operations
		switch r.Method {
		case http.MethodPost:
			HandleConfirmationUpload(confirmationService).ServeHTTP(w, r)
		case http.MethodGet:
			HandleConfirmationGet(confirmationService).ServeHTTP(w, r)
		case http.MethodDelete:
			HandleConfirmationDelete(confirmationService).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Bulk operations
	mux.HandleFunc("/api/v1/blotter/confirmations/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleConfirmationsExport(confirmationService).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/blotter/confirmations/metadata", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleConfirmationsMetadataGet(confirmationService).ServeHTTP(w, r)
	})
}

// extractTradeIDFromPath extracts the trade ID from the URL path
func extractTradeIDFromPath(path string) (string, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		return "", fmt.Errorf("invalid URL path format")
	}

	tradeID := parts[len(parts)-1]
	if strings.HasSuffix(path, "/metadata") && len(parts) >= 7 {
		tradeID = parts[len(parts)-2]
	}

	if tradeID == "" {
		return "", fmt.Errorf("trade ID cannot be empty")
	}

	return tradeID, nil
}

// sanitizeContentDispositionFilename properly encodes filename for Content-Disposition header
func sanitizeContentDispositionFilename(filename string) string {
	// For simplicity, URL-encode the filename for RFC 6266 compatibility
	return fmt.Sprintf("filename*=UTF-8''%s", url.QueryEscape(filename))
}
