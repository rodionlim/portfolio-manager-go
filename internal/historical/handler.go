package historical

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"portfolio-manager/pkg/common"
)

// HandleGetMetrics handles the GET /api/v1/historical/metrics endpoint
// @Summary Get historical portfolio metrics
// @Description Get all historical portfolio metrics (date-stamped portfolio metrics)
// @Tags historical
// @Produce json
// @Success 200 {array} TimestampedMetrics "List of historical portfolio metrics by date"
// @Failure 500 {object} common.ErrorResponse "Failed to get historical metrics"
// @Router /api/v1/historical/metrics [get]
func HandleGetMetrics(service interface {
	GetMetrics() ([]TimestampedMetrics, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := service.GetMetrics()
		if err != nil {
			common.WriteJSONError(w, "Failed to get historical metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			common.WriteJSONError(w, "Failed to encode metrics as JSON", http.StatusInternalServerError)
		}
	}
}

// HandleExportMetricsCSV handles exporting all historical metrics as CSV
// @Summary Export historical portfolio metrics as CSV
// @Description Export all historical portfolio metrics (date-stamped portfolio metrics) as a CSV file
// @Tags historical
// @Produce text/csv
// @Success 200 {string} string "CSV file with historical metrics"
// @Failure 500 {object} common.ErrorResponse "Failed to export historical metrics"
// @Router /api/v1/historical/metrics/export [get]
func HandleExportMetricsCSV(service interface {
	ExportMetricsToCSV() ([]byte, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csvBytes, err := service.ExportMetricsToCSV()
		if err != nil {
			common.WriteJSONError(w, "Failed to export historical metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=historical_metrics_export.csv")
		w.Write(csvBytes)
	}
}

// HandleImportMetricsCSV handles importing historical metrics from a CSV file
// @Summary Import historical portfolio metrics from CSV
// @Description Import historical portfolio metrics (date-stamped portfolio metrics) from a CSV file
// @Tags historical
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file to import"
// @Success 200 {object} map[string]interface{} "Import result"
// @Failure 400 {object} common.ErrorResponse "Invalid file or format"
// @Failure 500 {object} common.ErrorResponse "Failed to import historical metrics"
// @Router /api/v1/historical/metrics/import [post]
func HandleImportMetricsCSV(service interface {
	ImportMetricsFromCSVFile(file multipart.File) (int, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10MB max
		if err != nil {
			common.WriteJSONError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			common.WriteJSONError(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		count, err := service.ImportMetricsFromCSVFile(file)
		if err != nil {
			common.WriteJSONError(w, "Failed to import historical metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"imported": count})
	}
}

// HandleUpsertMetric handles inserting or updating a single historical metric
// @Summary Upsert a historical portfolio metric
// @Description Insert or update a single historical portfolio metric (date-stamped portfolio metric)
// @Tags historical
// @Accept json
// @Produce json
// @Param metric body TimestampedMetrics true "Historical metric"
// @Success 201 {object} TimestampedMetrics
// @Failure 400 {object} common.ErrorResponse "Invalid request payload"
// @Failure 500 {object} common.ErrorResponse "Failed to upsert historical metric"
// @Router /api/v1/historical/metrics [post]
// @Router /api/v1/historical/metrics [put]
func HandleUpsertMetric(service interface {
	UpsertMetric(metric TimestampedMetrics) error
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric TimestampedMetrics
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			common.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err := service.UpsertMetric(metric)
		if err != nil {
			common.WriteJSONError(w, "Failed to upsert historical metric: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(metric)
	}
}

// HandleDeleteMetric handles deleting a single historical metric by timestamp
// @Summary Delete a historical portfolio metric
// @Description Delete a single historical portfolio metric by its timestamp
// @Tags historical
// @Produce json
// @Param timestamp path string true "Timestamp of the metric to delete (URL encoded)"
// @Success 200 {object} map[string]string "Message confirming deletion"
// @Failure 400 {object} common.ErrorResponse "Invalid timestamp format"
// @Failure 404 {object} common.ErrorResponse "Metric not found"
// @Failure 500 {object} common.ErrorResponse "Failed to delete historical metric"
// @Router /api/v1/historical/metrics/{timestamp} [delete]
func HandleDeleteMetric(service interface {
	DeleteMetric(timestamp string) error
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timestamp := r.PathValue("timestamp")
		if timestamp == "" {
			common.WriteJSONError(w, "Timestamp is required", http.StatusBadRequest)
			return
		}

		err := service.DeleteMetric(timestamp)
		if err != nil {
			if err.Error() == "metric not found" {
				common.WriteJSONError(w, "Metric not found: "+err.Error(), http.StatusNotFound)
				return
			}
			common.WriteJSONError(w, "Failed to delete historical metric: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Metric deleted successfully"})
	}
}

// RegisterHandlers registers the historical metrics handlers
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("/api/v1/historical/metrics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetMetrics(service).ServeHTTP(w, r)
		case http.MethodPost, http.MethodPut:
			HandleUpsertMetric(service).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/historical/metrics/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleExportMetricsCSV(service).ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/v1/historical/metrics/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleImportMetricsCSV(service).ServeHTTP(w, r)
	})
	
	// Register DELETE endpoint with path parameter
	mux.HandleFunc("/api/v1/historical/metrics/{timestamp}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleDeleteMetric(service).ServeHTTP(w, r)
	})
}
