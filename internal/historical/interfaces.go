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
// Updated to match the implementation in service.go
// - GetMetrics() fetches all metrics
// - GetMetricsByDateRange(start, end) fetches by date range
// - StartMetricsCollection uses a cron expression
// - StopMetricsCollection stops the collection
type HistoricalMetricsManager interface {
	StoreCurrentMetrics() error
	GetMetrics() ([]TimestampedMetrics, error)
	GetMetricsByDateRange(start, end time.Time) ([]TimestampedMetrics, error)
	StartMetricsCollection(cronExpr string) func()
	StopMetricsCollection()
	StartSGXReportCollection(cronExpr string) func()
}

// TimestampedMetrics represents portfolio metrics with a timestamp (date only)
type TimestampedMetrics struct {
	Timestamp time.Time             `json:"timestamp"` // Only the date portion of this field will be used
	Metrics   metrics.MetricsResult `json:"metrics"`
}

// DeleteMetricsRequest represents a request to delete multiple metrics by timestamps
type DeleteMetricsRequest struct {
	Timestamps []string `json:"timestamps"`
}

// DeleteMetricsResponse represents the response from a batch delete operation
type DeleteMetricsResponse struct {
	Deleted  int      `json:"deleted"`
	Failed   int      `json:"failed"`
	Failures []string `json:"failures,omitempty"`
}
