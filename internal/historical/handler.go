package historical

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
)

// HandleGetMetrics handles the GET /api/v1/historical/metrics endpoint
// @Summary Get historical portfolio metrics
// @Description Get all historical portfolio metrics (date-stamped portfolio metrics), optionally filtered by book
// @Tags historical
// @Produce json
// @Param book_filter query string false "Filter metrics by book (optional)"
// @Success 200 {array} TimestampedMetrics "List of historical portfolio metrics by date"
// @Failure 500 {object} common.ErrorResponse "Failed to get historical metrics"
// @Router /api/v1/historical/metrics [get]
func HandleGetMetrics(service HistoricalMetricsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookFilter := r.URL.Query().Get("book_filter")
		metrics, err := service.GetMetrics(bookFilter)
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
// @Param book_filter query string false "Filter metrics by book (optional)"
// @Success 200 {string} string "CSV file with historical metrics"
// @Failure 500 {object} common.ErrorResponse "Failed to export historical metrics"
// @Router /api/v1/historical/metrics/export [get]
func HandleExportMetricsCSV(service HistoricalMetricsCsvManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookFilter := r.URL.Query().Get("book_filter")
		csvBytes, err := service.ExportMetricsToCSV(bookFilter)
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
func HandleImportMetricsCSV(service HistoricalMetricsCsvManager) http.HandlerFunc {
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
// @Description Insert or update a single historical portfolio metric (date-stamped portfolio metric), optionally filtered by book
// @Tags historical
// @Accept json
// @Produce json
// @Param book_filter query string false "Filter metric by book (optional)"
// @Param metric body TimestampedMetrics true "Historical metric"
// @Success 201 {object} TimestampedMetrics
// @Failure 400 {object} common.ErrorResponse "Invalid request payload"
// @Failure 500 {object} common.ErrorResponse "Failed to upsert historical metric"
// @Router /api/v1/historical/metrics [post]
// @Router /api/v1/historical/metrics [put]
func HandleUpsertMetric(service HistoricalMetricsSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookFilter := r.URL.Query().Get("book_filter")
		var metric TimestampedMetrics
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			common.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err := service.UpsertMetric(metric, bookFilter)
		if err != nil {
			common.WriteJSONError(w, "Failed to upsert historical metric: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(metric)
	}
}

// HandleDeleteMetrics handles deleting one or more historical metrics
// @Summary Delete historical portfolio metrics
// @Description Delete one or more historical portfolio metrics by their timestamps, optionally filtered by book
// @Tags historical
// @Accept json
// @Produce json
// @Param book_filter query string false "Filter metrics by book (optional)"
// @Param request body DeleteMetricsRequest true "List of timestamps to delete"
// @Success 200 {object} DeleteMetricsResponse "Result of the deletion operation"
// @Failure 400 {object} common.ErrorResponse "Invalid request payload"
// @Failure 500 {object} common.ErrorResponse "Failed to delete historical metrics"
// @Router /api/v1/historical/metrics/delete [post]
func HandleDeleteMetrics(service HistoricalMetricsSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookFilter := r.URL.Query().Get("book_filter")
		var request DeleteMetricsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			common.WriteJSONError(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(request.Timestamps) == 0 {
			common.WriteJSONError(w, "No timestamps provided for deletion", http.StatusBadRequest)
			return
		}

		response, err := service.DeleteMetrics(request.Timestamps, bookFilter)
		if err != nil {
			common.WriteJSONError(w, "Failed to delete historical metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// CreateMetricsJobRequest represents the request to create a metrics job
type CreateMetricsJobRequest struct {
	CronExpr   string `json:"cronExpr"`   // Optional, uses default if empty
	BookFilter string `json:"bookFilter"` // Required
}

// HandleCreateMetricsJob handles the POST /api/v1/historical/metrics/jobs endpoint
// @Summary Create a custom metrics job
// @Description Create a new custom metrics job with a cron expression and book filter
// @Tags historical
// @Accept json
// @Produce json
// @Param request body CreateMetricsJobRequest true "Metrics job request"
// @Success 201 {object} MetricsJob "Created metrics job"
// @Failure 400 {object} common.ErrorResponse "Invalid request payload"
// @Failure 500 {object} common.ErrorResponse "Failed to create metrics job"
// @Router /api/v1/historical/metrics/jobs [post]
func HandleCreateMetricsJob(service HistoricalMetricsScheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateMetricsJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.BookFilter == "" {
			common.WriteJSONError(w, "bookFilter is required", http.StatusBadRequest)
			return
		}

		job, err := service.CreateMetricsJob(req.CronExpr, req.BookFilter)
		if err != nil {
			common.WriteJSONError(w, "Failed to create metrics job: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(job)
	}
}

// HandleDeleteMetricsJob handles the DELETE /api/v1/historical/metrics/jobs/{bookFilter} endpoint
// @Summary Delete a custom metrics job
// @Description Delete a custom metrics job by book filter
// @Tags historical
// @Param bookFilter path string true "Book filter"
// @Success 204 "Metrics job deleted successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid book filter"
// @Failure 404 {object} common.ErrorResponse "Metrics job not found"
// @Failure 500 {object} common.ErrorResponse "Failed to delete metrics job"
// @Router /api/v1/historical/metrics/jobs/{bookFilter} [delete]
func HandleDeleteMetricsJob(service HistoricalMetricsScheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract book filter from URL path
		bookFilter := r.URL.Path[len("/api/v1/historical/metrics/jobs/"):]
		if bookFilter == "" {
			common.WriteJSONError(w, "Book filter is required", http.StatusBadRequest)
			return
		}

		err := service.DeleteMetricsJob(bookFilter)
		if err != nil {
			if err.Error() == "metrics job not found for book_filter: "+bookFilter {
				common.WriteJSONError(w, "Metrics job not found", http.StatusNotFound)
				return
			}
			common.WriteJSONError(w, "Failed to delete metrics job: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleListMetricsJobs handles the GET /api/v1/historical/metrics/jobs endpoint
// @Summary List all custom metrics jobs
// @Description List all custom metrics jobs (excluding the default portfolio job)
// @Tags historical
// @Produce json
// @Success 200 {array} MetricsJob "List of custom metrics jobs"
// @Failure 500 {object} common.ErrorResponse "Failed to list metrics jobs"
// @Router /api/v1/historical/metrics/jobs [get]
func HandleListMetricsJobs(service HistoricalMetricsScheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := service.ListMetricsJobs()
		if err != nil {
			common.WriteJSONError(w, "Failed to list metrics jobs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	}
}

// HandleListAllMetricsJobs handles the GET /api/v1/historical/metrics/jobs/all endpoint
// @Summary List all metrics jobs including portfolio job
// @Description List all custom metrics jobs and include a dummy portfolio job for UI purposes
// @Tags historical
// @Produce json
// @Success 200 {array} MetricsJob "List of all metrics jobs including portfolio"
// @Failure 500 {object} common.ErrorResponse "Failed to list metrics jobs"
// @Router /api/v1/historical/metrics/jobs/all [get]
func HandleListAllMetricsJobs(service HistoricalMetricsScheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := service.ListAllMetricsJobsIncludingPortfolio()
		if err != nil {
			common.WriteJSONError(w, "Failed to list metrics jobs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	}
}

// TriggerMetricsCollectionRequest represents the request to manually trigger metrics collection
type TriggerMetricsCollectionRequest struct {
	BookFilter string `json:"bookFilter"` // Optional, empty for entire portfolio
}

// HandleTriggerMetricsCollection handles the POST /api/v1/historical/metrics/trigger endpoint
// @Summary Manually trigger metrics collection
// @Description Manually trigger metrics collection for a specific book or entire portfolio
// @Tags historical
// @Accept json
// @Produce json
// @Param request body TriggerMetricsCollectionRequest true "Trigger metrics collection request"
// @Success 200 {object} map[string]interface{} "Metrics collection triggered successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request payload"
// @Failure 500 {object} common.ErrorResponse "Failed to trigger metrics collection"
// @Router /api/v1/historical/metrics/trigger [post]
func HandleTriggerMetricsCollection(service HistoricalMetricsSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TriggerMetricsCollectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteJSONError(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Trigger metrics collection
		err := service.StoreCurrentMetrics(req.BookFilter)
		if err != nil {
			common.WriteJSONError(w, "Failed to trigger metrics collection: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"message":    "Metrics collection triggered successfully",
			"bookFilter": req.BookFilter,
		})
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

	// Register metrics deletion endpoint
	mux.HandleFunc("/api/v1/historical/metrics/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleDeleteMetrics(service).ServeHTTP(w, r)
	})

	// Register metrics jobs endpoints
	mux.HandleFunc("/api/v1/historical/metrics/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandleCreateMetricsJob(service).ServeHTTP(w, r)
		case http.MethodGet:
			HandleListMetricsJobs(service).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/historical/metrics/jobs/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleListAllMetricsJobs(service).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/historical/metrics/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			HandleDeleteMetricsJob(service).ServeHTTP(w, r)
		} else {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Register manual trigger endpoint
	mux.HandleFunc("/api/v1/historical/metrics/trigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		HandleTriggerMetricsCollection(service).ServeHTTP(w, r)
	})

	// Register historical market data endpoints
	mux.HandleFunc("GET /api/v1/historical/config", HandleGetAssetConfigs(service))
	mux.HandleFunc("POST /api/v1/historical/config", HandleUpdateAssetConfig(service))
	mux.HandleFunc("DELETE /api/v1/historical/config/{ticker}", HandleRemoveAssetConfig(service))
	mux.HandleFunc("POST /api/v1/historical/sync", HandleSyncAssetData(service))
	mux.HandleFunc("GET /api/v1/historical/data/{ticker}", HandleGetHistoricalAssetData(service))
}
