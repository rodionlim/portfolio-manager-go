package historical

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"portfolio-manager/internal/analytics"
	"portfolio-manager/internal/config"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/pkg/csvutil"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/scheduler"
	"portfolio-manager/pkg/types"
	"strconv"
	"strings"
	"time"
)

// Service implements the HistoricalMetricsManager interface
type Service struct {
	metricsService   metrics.MetricsServicer
	analyticsService analytics.Service
	db               dal.Database
	scheduler        scheduler.Scheduler
	mdataManager     mdata.MarketDataManager
	logger           *logging.Logger
	collectionTasks  []scheduler.TaskID
}

// NewService creates a new historical metrics service
func NewService(
	metricsService metrics.MetricsServicer,
	analyticsService analytics.Service,
	db dal.Database,
	scheduler scheduler.Scheduler,
	mdataManager mdata.MarketDataManager,
) *Service {
	return &Service{
		metricsService:   metricsService,
		analyticsService: analyticsService,
		db:               db,
		scheduler:        scheduler,
		mdataManager:     mdataManager,
		logger:           logging.GetLogger(),
	}
}

func (s *Service) transformBookFilter(book_filter string) string {
	if book_filter == "" {
		return "portfolio" // Default to portfolio if no book filter is provided
	}
	return book_filter
}

// StartSGXReportCollection starts collection of sgx report based on a cron expression
func (s *Service) StartSGXReportCollection(cronExpr string) func() {
	metricsTask := func(ctx context.Context) error {
		analysis, err := s.analyticsService.FetchAndAnalyzeLatestReportByType("fund flow")
		if err != nil {
			return fmt.Errorf("failed to collect SGX fund flow report: %w", err)
		}
		s.logger.Info("Successfully collected SGX fund flow reports for", analysis.ReportTitle)
		return nil
	}

	sched, err := scheduler.NewCronSchedule(cronExpr)
	if err != nil {
		s.logger.Errorf("Invalid cron expression for sgx report collection: %v", err)
		return func() {}
	}

	taskId := s.scheduler.ScheduleTaskFunc(metricsTask, sched)
	s.collectionTasks = append(s.collectionTasks, taskId)

	s.logger.Infof("Started sgx report collection with cron schedule: %s", cronExpr)

	return func() {
		s.scheduler.Unschedule(taskId)
		s.logger.Info("Stopped sgx report collection")
	}
}

// StoreCurrentMetrics stores the current portfolio/book metrics with the current timestamp
func (s *Service) StoreCurrentMetrics(book_filter string) error {
	// Get current metrics
	result, err := s.metricsService.CalculatePortfolioMetrics(book_filter)
	if err != nil {
		return fmt.Errorf("failed to calculate [book_filter: %s] metrics: %w", book_filter, err)
	}

	// Create timestamped metrics (date only)
	now := time.Now()
	// Set time to midnight to ensure only date info is used
	dateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	timestampedMetrics := TimestampedMetrics{
		Timestamp: dateOnly,
		Metrics:   result.Metrics,
	}

	label := "portfolio"
	if book_filter != "" {
		label = book_filter
	}

	// Generate key for LevelDB
	// Format: metrics:book:YYYY-MM-DD
	key := fmt.Sprintf("%s:%s:%s",
		types.HistoricalMetricsKeyPrefix,
		label,
		dateOnly.Format("2006-01-02"),
	)

	// Store in LevelDB
	err = s.db.Put(key, timestampedMetrics)
	if err != nil {
		return fmt.Errorf("failed to store metrics [book_filter: %s]: %w", label, err)
	}

	s.logger.Infof("Stored portfolio metrics [book_filter: %s] for timestamp %s", label, now.Format(time.RFC3339))
	return nil
}

// GetMetrics retrieves all historical metrics
func (s *Service) GetMetrics(book_filter string) ([]TimestampedMetrics, error) {
	label := s.transformBookFilter(book_filter) // Ensure we use the correct book filter
	prefix := fmt.Sprintf("%s:%s:", types.HistoricalMetricsKeyPrefix, label)
	keys, err := s.db.GetAllKeysWithPrefix(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics keys: %w", err)
	}

	// Initialize with an empty slice instead of nil to ensure JSON encodes as [] not null
	results := []TimestampedMetrics{}
	for _, key := range keys {
		var metrics TimestampedMetrics
		err := s.db.Get(key, &metrics)
		if err != nil {
			s.logger.Warnf("Failed to get metrics for key %s: %v", key, err)
			continue
		}
		results = append(results, metrics)
	}
	return results, nil
}

// GetMetricsByDateRange retrieves historical metrics for a given time range
func (s *Service) GetMetricsByDateRange(book_filter string, start, end time.Time) ([]TimestampedMetrics, error) {
	label := s.transformBookFilter(book_filter)

	prefix := fmt.Sprintf("%s:%s:", types.HistoricalMetricsKeyPrefix, label)
	keys, err := s.db.GetAllKeysWithPrefix(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics keys: %w", err)
	}

	// Initialize with an empty slice instead of nil to ensure JSON encodes as [] not null
	results := []TimestampedMetrics{}
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			s.logger.Warnf("Malformed metrics key: %s", key)
			continue
		}
		dateStr := parts[2]
		timestamp, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			s.logger.Warnf("Failed to parse date from key %s: %v", key, err)
			continue
		}
		if (timestamp.Equal(start) || timestamp.After(start)) &&
			(timestamp.Equal(end) || timestamp.Before(end)) {
			var metrics TimestampedMetrics
			err := s.db.Get(key, &metrics)
			if err != nil {
				s.logger.Warnf("Failed to get metrics for key %s: %v", key, err)
				continue
			}
			results = append(results, metrics)
		}
	}
	return results, nil
}

// StartMetricsCollection starts collection of metrics based on a cron expression
func (s *Service) StartMetricsCollection(cronExpr string, book_filter string) func() {
	metricsTask := func(ctx context.Context) error {
		return s.StoreCurrentMetrics(book_filter)
	}

	sched, err := scheduler.NewCronSchedule(cronExpr)
	if err != nil {
		s.logger.Errorf("Invalid cron expression for metrics collection: %v", err)
		return func() {}
	}

	taskId := s.scheduler.ScheduleTaskFunc(metricsTask, sched)
	s.collectionTasks = append(s.collectionTasks, taskId)

	s.logger.Infof("Started metrics collection [book_filter: %s] with cron schedule: %s", book_filter, cronExpr)

	return func() {
		s.scheduler.Unschedule(taskId)
		s.logger.Infof("Stopped metrics collection [book filter: %s]", book_filter)
	}
}

// StopMetricsCollection stops the periodic collection of metrics
func (s *Service) StopMetricsCollection() {
	s.logger.Info("Stopping metrics collection")
	// Unschedule all collection tasks
	for _, taskId := range s.collectionTasks {
		if ok := s.scheduler.Unschedule(taskId); !ok {
			s.logger.Warnf("Failed to unschedule task %s", taskId)
		}
	}
	s.collectionTasks = []scheduler.TaskID{}
}

// ExportMetricsToCSV exports all historical metrics to a CSV file in memory and returns it as a byte slice.
func (s *Service) ExportMetricsToCSV(book_filter string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{"Date", "IRR", "PricePaid", "MV", "TotalDividends", "BookFilter"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	metrics, err := s.GetMetrics(book_filter)
	if err != nil {
		return nil, err
	}
	for _, m := range metrics {
		record := []string{
			m.Timestamp.Format("2006-01-02"),
			csvutil.FormatFloat(m.Metrics.IRR, 6),
			csvutil.FormatFloat(m.Metrics.PricePaid, 2),
			csvutil.FormatFloat(m.Metrics.MV, 2),
			csvutil.FormatFloat(m.Metrics.TotalDividends, 2),
			book_filter,
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ImportMetricsFromCSVFile imports historical metrics from a CSV file and adds them to the database.
func (s *Service) ImportMetricsFromCSVFile(file multipart.File) (int, error) {
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return 0, err
	}
	expectedHeaders := []string{"Date", "IRR", "PricePaid", "MV", "TotalDividends", "BookFilter"}
	if len(header) < len(expectedHeaders) {
		return 0, fmt.Errorf("invalid CSV header length: expected at least %d columns, got %d", len(expectedHeaders), len(header))
	}

	for i, h := range expectedHeaders {
		if header[i] != h {
			return 0, fmt.Errorf("invalid CSV header: expected %s at position %d, got %s", h, i, header[i])
		}
	}
	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		if len(row) < 6 {
			return count, fmt.Errorf("invalid row length: %v", row)
		}
		ts, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			return count, fmt.Errorf("invalid date: %w", err)
		}
		irr, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return count, fmt.Errorf("invalid IRR: %w", err)
		}
		pricePaid, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return count, fmt.Errorf("invalid PricePaid: %w", err)
		}
		mv, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return count, fmt.Errorf("invalid MV: %w", err)
		}
		totalDiv, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return count, fmt.Errorf("invalid TotalDividends: %w", err)
		}
		label := s.transformBookFilter(row[5])
		metrics := metrics.MetricsResult{
			IRR:            irr,
			PricePaid:      pricePaid,
			MV:             mv,
			TotalDividends: totalDiv,
		}
		tm := TimestampedMetrics{
			Timestamp: ts,
			Metrics:   metrics,
		}
		// Store in DB (overwrite if exists)
		key := fmt.Sprintf("%s:%s:%s", types.HistoricalMetricsKeyPrefix, label, ts.Format("2006-01-02"))
		if err := s.db.Put(key, tm); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// UpsertMetric inserts or updates a single historical metric in the database
func (s *Service) UpsertMetric(metric TimestampedMetrics, book_filter string) error {
	book_filter = s.transformBookFilter(book_filter)
	key := fmt.Sprintf("%s:%s:%s", types.HistoricalMetricsKeyPrefix, book_filter, metric.Timestamp.Format("2006-01-02"))
	return s.db.Put(key, metric)
}

// DeleteMetric deletes a historical metric by timestamp
func (s *Service) DeleteMetric(timestamp string, book_filter string) error {
	label := s.transformBookFilter(book_filter)
	s.logger.Info(fmt.Sprintf("Deleting historical metric [book_filter: %s] for timestamp: %s", book_filter, timestamp))

	// Parse the timestamp to validate it
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Invalid timestamp format: %s", timestamp))
		return fmt.Errorf("invalid timestamp format: %v", err)
	}

	// Create the key for the database - use the date part only
	dateOnly := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, parsedTime.Location())
	key := fmt.Sprintf("%s:%s:%s", types.HistoricalMetricsKeyPrefix, label, dateOnly.Format("2006-01-02"))

	// Check if the metric exists by trying to get it
	var metric TimestampedMetrics
	err = s.db.Get(key, &metric)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Metric not found for timestamp: %s [book_filter: %s] - %v", timestamp, book_filter, err))
		return fmt.Errorf("metric not found")
	}

	// Delete the metric
	err = s.db.Delete(key)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete metric [book_filter: %s]: %v", book_filter, err))
		return err
	}

	s.logger.Info(fmt.Sprintf("Successfully deleted metric [book_filter: %s] for timestamp: %s", book_filter, timestamp))
	return nil
}

// DeleteMetrics deletes multiple historical metrics by their timestamps
func (s *Service) DeleteMetrics(timestamps []string, book_filter string) (DeleteMetricsResponse, error) {
	s.logger.Info(fmt.Sprintf("Batch deleting %d historical metrics [book_filter: %s]", len(timestamps), book_filter))

	result := DeleteMetricsResponse{
		Deleted:  0,
		Failed:   0,
		Failures: []string{},
	}

	for _, timestamp := range timestamps {
		err := s.DeleteMetric(timestamp, book_filter)
		if err != nil {
			result.Failed++
			result.Failures = append(result.Failures, fmt.Sprintf("Failed to delete metric with timestamp %s [book_filter: %s]: %v", timestamp, book_filter, err))
			s.logger.Warn(fmt.Sprintf("Failed to delete metric with timestamp %s [book_filter: %s]: %v", timestamp, book_filter, err))
		} else {
			result.Deleted++
		}
	}

	s.logger.Info(fmt.Sprintf("Batch delete completed: %d deleted, %d failed [book_filter: %s]", result.Deleted, result.Failed, book_filter))
	return result, nil
}

// CreateMetricsJob creates a new custom metrics job with a cron expression and book filter
func (s *Service) CreateMetricsJob(cronExpr string, book_filter string) (*MetricsJob, error) {
	if book_filter == "" {
		return nil, fmt.Errorf("book_filter cannot be empty for custom metrics job")
	}

	if book_filter == "portfolio" {
		return nil, fmt.Errorf("portfolio cannot be used as book_filter. It is reserved for entire portfolio across all books")
	}

	// Use default cron expression from config if not provided
	if cronExpr == "" {
		if config.DefaultMetricsSchedule == "" {
			return nil, fmt.Errorf("no cron expression provided and no default schedule configured")
		}
		cronExpr = config.DefaultMetricsSchedule
	}

	// Validate cron expression
	_, err := scheduler.NewCronSchedule(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Check if job already exists for this book filter
	key := fmt.Sprintf("%s:%s:%s", types.ScheduledJobKeyPrefix, types.CustomMetricsJobKeyPrefix, book_filter)
	var existingJob MetricsJob
	if err := s.db.Get(key, &existingJob); err == nil {
		return nil, fmt.Errorf("metrics job already exists for book_filter: %s", book_filter)
	}

	// Start metrics collection and get the cancellation function
	cancelFunc := s.StartMetricsCollection(cronExpr, book_filter)

	// Get the task ID from the last added task
	var taskId scheduler.TaskID
	if len(s.collectionTasks) > 0 {
		taskId = s.collectionTasks[len(s.collectionTasks)-1]
	}

	// Create and store the metrics job
	job := MetricsJob{
		BookFilter: book_filter,
		CronExpr:   cronExpr,
		TaskId:     taskId,
	}

	if err := s.db.Put(key, job); err != nil {
		// If storage fails, cancel the started job
		cancelFunc()
		return nil, fmt.Errorf("failed to store metrics job: %w", err)
	}

	s.logger.Infof("Created metrics job for book_filter: %s with cron: %s", book_filter, cronExpr)
	return &job, nil
}

// DeleteMetricsJob deletes a custom metrics job by book filter
func (s *Service) DeleteMetricsJob(book_filter string) error {
	if book_filter == "" {
		return fmt.Errorf("book_filter cannot be empty")
	}

	key := fmt.Sprintf("%s:%s:%s", types.ScheduledJobKeyPrefix, types.CustomMetricsJobKeyPrefix, book_filter)

	// Get the job to retrieve the task ID
	var job MetricsJob
	if err := s.db.Get(key, &job); err != nil {
		return fmt.Errorf("metrics job not found for book_filter: %s", book_filter)
	}

	// Stop the scheduled task
	if !s.scheduler.Unschedule(job.TaskId) {
		s.logger.Warnf("Failed to unschedule task %s for book_filter: %s", job.TaskId, book_filter)
	}

	// Remove the task ID from our collection tasks list
	for i, taskId := range s.collectionTasks {
		if taskId == job.TaskId {
			s.collectionTasks = append(s.collectionTasks[:i], s.collectionTasks[i+1:]...)
			break
		}
	}

	// Delete from database
	if err := s.db.Delete(key); err != nil {
		return fmt.Errorf("failed to delete metrics job: %w", err)
	}

	s.logger.Infof("Deleted metrics job for book_filter: %s", book_filter)
	return nil
}

// ListMetricsJobs lists all custom metrics jobs (excluding the default portfolio job)
func (s *Service) ListMetricsJobs() ([]MetricsJob, error) {
	prefix := fmt.Sprintf("%s:%s", types.ScheduledJobKeyPrefix, types.CustomMetricsJobKeyPrefix)
	keys, err := s.db.GetAllKeysWithPrefix(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics job keys: %w", err)
	}

	// Initialize with an empty slice instead of nil to ensure JSON encodes as [] not null
	jobs := []MetricsJob{}
	for _, key := range keys {
		var job MetricsJob
		if err := s.db.Get(key, &job); err != nil {
			s.logger.Warnf("Failed to get metrics job for key %s: %v", key, err)
			continue
		}

		// Exclude the default portfolio job (book_filter would be "portfolio" or empty)
		if job.BookFilter != "" && job.BookFilter != "portfolio" {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

// ListAllMetricsJobsIncludingPortfolio lists all custom metrics jobs and includes a dummy portfolio job
func (s *Service) ListAllMetricsJobsIncludingPortfolio() ([]MetricsJob, error) {
	// Get all custom jobs first
	customJobs, err := s.ListMetricsJobs()
	if err != nil {
		return nil, err
	}

	// Initialize with the portfolio job at the top
	jobs := []MetricsJob{
		{
			BookFilter: "",
			CronExpr:   config.DefaultMetricsSchedule,
			TaskId:     "", // Empty since this is a dummy job for UI purposes
		},
	}

	// Append custom jobs
	jobs = append(jobs, customJobs...)
	return jobs, nil
}
