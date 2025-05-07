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
	metricsService *metrics.MetricsService
	db             dal.Database
	scheduler      scheduler.Scheduler
	logger         *logging.Logger
	collectionTask scheduler.TaskID
}

// NewService creates a new historical metrics service
func NewService(
	metricsService *metrics.MetricsService,
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

	// Create timestamped metrics
	now := time.Now()
	timestampedMetrics := TimestampedMetrics{
		Timestamp: now,
		Metrics:   result.Metrics,
	}

	// Generate key for LevelDB
	// Format: metrics:portfolio:YYYY-MM-DD:HH:MM:SS
	key := fmt.Sprintf("%s:%s:%s",
		types.KeyPrefixHistoricalMetrics,
		"portfolio",
		now.Format("2006-01-02:15:04:05"),
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
		if len(parts) < 4 {
			s.logger.Warnf("Malformed metrics key: %s", key)
			continue
		}

		// Format should be metrics:portfolio:YYYY-MM-DD:HH:MM:SS
		dateStr := parts[2] + ":" + parts[3]
		timestamp, err := time.Parse("2006-01-02:15:04:05", dateStr)
		if err != nil {
			s.logger.Warnf("Failed to parse timestamp from key %s: %v", key, err)
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

// StartMetricsCollection starts periodic collection of metrics
func (s *Service) StartMetricsCollection(interval time.Duration) func() {
	// Create a task that will store metrics
	metricsTask := func(ctx context.Context) error {
		return s.StoreCurrentMetrics()
	}

	// Schedule the task with the given interval
	s.collectionTask = s.scheduler.ScheduleTaskFunc(metricsTask, scheduler.NewPeriodicSchedule(interval))

	s.logger.Infof("Started metrics collection with interval of %v", interval)

	// Return a function that can be used to stop collection
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
