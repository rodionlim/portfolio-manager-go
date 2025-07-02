package historical

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"portfolio-manager/internal/analytics"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/pkg/csvutil"
	"portfolio-manager/pkg/logging"
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
	logger           *logging.Logger
	collectionTask   scheduler.TaskID
}

// NewService creates a new historical metrics service
func NewService(
	metricsService metrics.MetricsServicer,
	analyticsService analytics.Service,
	db dal.Database,
	scheduler scheduler.Scheduler,
) *Service {
	return &Service{
		metricsService:   metricsService,
		analyticsService: analyticsService,
		db:               db,
		scheduler:        scheduler,
		logger:           logging.GetLogger(),
	}
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

	s.collectionTask = s.scheduler.ScheduleTaskFunc(metricsTask, sched)

	s.logger.Infof("Started sgx report collection with cron schedule: %s", cronExpr)

	return func() {
		if s.collectionTask != "" {
			s.scheduler.Unschedule(s.collectionTask)
			s.logger.Info("Stopped sgx report collection")
		}
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

	// Generate key for LevelDB
	// Format: metrics:book:YYYY-MM-DD
	key := fmt.Sprintf("%s:%s:%s",
		types.HistoricalMetricsKeyPrefix,
		"portfolio",
		dateOnly.Format("2006-01-02"),
	)

	// Store in LevelDB
	err = s.db.Put(key, timestampedMetrics)
	if err != nil {
		return fmt.Errorf("failed to store metrics [book_filter: %s]: %w", book_filter, err)
	}

	s.logger.Infof("Stored portfolio metrics [book_filter: %s] for timestamp %s", book_filter, now.Format(time.RFC3339))
	return nil
}

// GetMetrics retrieves all historical metrics
func (s *Service) GetMetrics() ([]TimestampedMetrics, error) {
	prefix := fmt.Sprintf("%s:%s:", types.HistoricalMetricsKeyPrefix, "portfolio")
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
func (s *Service) GetMetricsByDateRange(start, end time.Time) ([]TimestampedMetrics, error) {
	prefix := fmt.Sprintf("%s:%s:", types.HistoricalMetricsKeyPrefix, "portfolio")
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

	s.collectionTask = s.scheduler.ScheduleTaskFunc(metricsTask, sched)

	s.logger.Infof("Started metrics collection with cron schedule: %s", cronExpr)

	return func() {
		if s.collectionTask != "" {
			s.scheduler.Unschedule(s.collectionTask)
			s.logger.Info("Stopped metrics collection")
		}
	}
}

// StopMetricsCollection stops the periodic collection of metrics
func (s *Service) StopMetricsCollection() {
	if s.collectionTask != "" {
		s.scheduler.Unschedule(s.collectionTask)
		s.collectionTask = ""
		s.logger.Info("Stopped metrics collection")
	}
}

// ExportMetricsToCSV exports all historical metrics to a CSV file in memory and returns it as a byte slice.
func (s *Service) ExportMetricsToCSV() ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{"Date", "IRR", "PricePaid", "MV", "TotalDividends"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	metrics, err := s.GetMetrics()
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
	expectedHeaders := []string{"Date", "IRR", "PricePaid", "MV", "TotalDividends"}
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
		if len(row) < 5 {
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
		key := fmt.Sprintf("%s:%s:%s", types.HistoricalMetricsKeyPrefix, "portfolio", ts.Format("2006-01-02"))
		if err := s.db.Put(key, tm); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// UpsertMetric inserts or updates a single historical metric in the database
func (s *Service) UpsertMetric(metric TimestampedMetrics) error {
	key := fmt.Sprintf("%s:%s:%s", types.HistoricalMetricsKeyPrefix, "portfolio", metric.Timestamp.Format("2006-01-02"))
	return s.db.Put(key, metric)
}

// DeleteMetric deletes a historical metric by timestamp
func (s *Service) DeleteMetric(timestamp string) error {
	s.logger.Info(fmt.Sprintf("Deleting historical metric for timestamp: %s", timestamp))

	// Parse the timestamp to validate it
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Invalid timestamp format: %s", timestamp))
		return fmt.Errorf("invalid timestamp format: %v", err)
	}

	// Create the key for the database - use the date part only
	dateOnly := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, parsedTime.Location())
	key := fmt.Sprintf("%s:%s:%s", types.HistoricalMetricsKeyPrefix, "portfolio", dateOnly.Format("2006-01-02"))

	// Check if the metric exists by trying to get it
	var metric TimestampedMetrics
	err = s.db.Get(key, &metric)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Metric not found for timestamp: %s - %v", timestamp, err))
		return fmt.Errorf("metric not found")
	}

	// Delete the metric
	err = s.db.Delete(key)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete metric: %v", err))
		return err
	}

	s.logger.Info(fmt.Sprintf("Successfully deleted metric for timestamp: %s", timestamp))
	return nil
}

// DeleteMetrics deletes multiple historical metrics by their timestamps
func (s *Service) DeleteMetrics(timestamps []string) (DeleteMetricsResponse, error) {
	s.logger.Info(fmt.Sprintf("Batch deleting %d historical metrics", len(timestamps)))

	result := DeleteMetricsResponse{
		Deleted:  0,
		Failed:   0,
		Failures: []string{},
	}

	for _, timestamp := range timestamps {
		err := s.DeleteMetric(timestamp)
		if err != nil {
			result.Failed++
			result.Failures = append(result.Failures, fmt.Sprintf("Failed to delete metric with timestamp %s: %v", timestamp, err))
			s.logger.Warn(fmt.Sprintf("Failed to delete metric with timestamp %s: %v", timestamp, err))
		} else {
			result.Deleted++
		}
	}

	s.logger.Info(fmt.Sprintf("Batch delete completed: %d deleted, %d failed", result.Deleted, result.Failed))
	return result, nil
}
