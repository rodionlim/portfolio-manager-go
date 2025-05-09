package historical

import (
	"context"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/scheduler"
	"portfolio-manager/pkg/types"
	"strings"
	"time"
)

// Service implements the HistoricalMetricsManager interface
type Service struct {
	metricsService metrics.MetricsServicer
	db             dal.Database
	scheduler      scheduler.Scheduler
	logger         *logging.Logger
	collectionTask scheduler.TaskID
}

// NewService creates a new historical metrics service
func NewService(
	metricsService metrics.MetricsServicer,
	db dal.Database,
	scheduler scheduler.Scheduler,
) *Service {
	return &Service{
		metricsService: metricsService,
		db:             db,
		scheduler:      scheduler,
		logger:         logging.GetLogger(),
	}
}

// StoreCurrentMetrics stores the current portfolio metrics with the current timestamp
func (s *Service) StoreCurrentMetrics() error {
	// Get current metrics
	result, err := s.metricsService.CalculatePortfolioMetrics()
	if err != nil {
		return fmt.Errorf("failed to calculate metrics: %w", err)
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
	// Format: metrics:portfolio:YYYY-MM-DD
	key := fmt.Sprintf("%s:%s:%s",
		types.KeyPrefixHistoricalMetrics,
		"portfolio",
		dateOnly.Format("2006-01-02"),
	)

	// Store in LevelDB
	err = s.db.Put(key, timestampedMetrics)
	if err != nil {
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	s.logger.Infof("Stored portfolio metrics for timestamp %s", now.Format(time.RFC3339))
	return nil
}

// GetMetrics retrieves historical metrics for a given time range
func (s *Service) GetMetrics(start, end time.Time) ([]TimestampedMetrics, error) {
	// Get all keys with prefix
	prefix := fmt.Sprintf("%s:%s:", types.KeyPrefixHistoricalMetrics, "portfolio")
	keys, err := s.db.GetAllKeysWithPrefix(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics keys: %w", err)
	}

	var results []TimestampedMetrics

	// Process each key
	for _, key := range keys {
		// Extract the timestamp from the key
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			s.logger.Warnf("Malformed metrics key: %s", key)
			continue
		}

		// Format should be metrics:portfolio:YYYY-MM-DD
		dateStr := parts[2]
		timestamp, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			s.logger.Warnf("Failed to parse date from key %s: %v", key, err)
			continue
		}

		// Check if the timestamp is in the requested range
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
func (s *Service) StartMetricsCollection(cronExpr string) func() {
	metricsTask := func(ctx context.Context) error {
		return s.StoreCurrentMetrics()
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
