// Package historical provides functionality for storing and retrieving historical portfolio metrics
package historical

import (
	"portfolio-manager/internal/metrics"
	"time"
)

// TimeSeriesKey represents a key for time series data
type TimeSeriesKey struct {
	Type string    // Type of data (e.g., "portfolio-metrics")
	Date time.Time // Date of the snapshot
}

// HistoricalMetricsManager defines the interface for managing historical metrics
type HistoricalMetricsManager interface {
	// StoreCurrentMetrics stores the current portfolio metrics with the current timestamp
	StoreCurrentMetrics() error

	// GetMetrics retrieves historical metrics for a given time range
	GetMetrics(start, end time.Time) ([]TimestampedMetrics, error)

	// StartMetricsCollection starts periodic collection of metrics
	// interval specifies how often to collect metrics
	// Returns a function that can be called to stop collection
	StartMetricsCollection(interval time.Duration) func()
}

// TimestampedMetrics represents portfolio metrics with a timestamp
type TimestampedMetrics struct {
	Timestamp time.Time             `json:"timestamp"`
	Metrics   metrics.MetricsResult `json:"metrics"`
}
