// Package historical provides functionality for storing and retrieving historical portfolio metrics
package historical

import (
	"mime/multipart"
	"portfolio-manager/internal/metrics"
	"time"
)

// TimeSeriesKey represents a key for time series data
type TimeSeriesKey struct {
	Type string    // Type of data (e.g., "portfolio-metrics")
	Date time.Time // Date of the snapshot
}

// Historical* interfaces for managing historical metrics
// - GetMetrics() fetches all metrics
// - GetMetricsByDateRange(start, end) fetches by date range
// - StartMetricsCollection uses a cron expression to start collection of metrics
// - StopMetricsCollection stops the collection of metrics
// - StartSGXReportCollection uses a cron expression to start collection of SGX reports
// - StoreCurrentMetrics stores the current metrics
// - ExportMetricsToCSV exports metrics to CSV

type HistoricalMetricsScheduler interface {
	StartMetricsCollection(cronExpr string) func()
	StopMetricsCollection()
}

type HistoricalReportsScheduler interface {
	StartSGXReportCollection(cronExpr string) func()
}
type HistoricalMetricsGetter interface {
	GetMetrics() ([]TimestampedMetrics, error)
	GetMetricsByDateRange(start, end time.Time) ([]TimestampedMetrics, error)
}

type HistoricalMetricsSetter interface {
	StoreCurrentMetrics(book_filter string) error // set book_filter to "" to get metrics for all books
}

type HistoricalMetricsCsvManager interface {
	ExportMetricsToCSV() ([]byte, error)
	ImportMetricsFromCSVFile(file multipart.File) (int, error)
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
