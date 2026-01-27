// Package historical provides functionality for storing and retrieving historical portfolio metrics
package historical

import (
	"mime/multipart"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/pkg/scheduler"
	"time"
)

// TimeSeriesKey represents a key for time series data
type TimeSeriesKey struct {
	Type string    // Type of data (e.g., "portfolio-metrics")
	Date time.Time // Date of the snapshot
}

type MetricsJob struct {
	BookFilter string // Filter for specific book
	CronExpr   string // Cron expression for scheduling
	TaskId     scheduler.TaskID
}

// Historical* interfaces for managing historical metrics
// - GetMetrics() fetches all metrics
// - GetMetricsByDateRange(start, end) fetches by date range
// - CreateMetricsJob creates a new custom metrics job with a cron expression and book filter. Set cron expression to empty to use the default schedule in config
// - DeleteMetricsJob deletes a custom metrics job by book filter
// - ListMetricsJobs lists all custom metrics jobs
// - StartMetricsCollection uses a cron expression to start collection of metrics
// - StopMetricsCollection stops the collection of metrics
// - StoreCurrentMetrics stores the current metrics
// - ExportMetricsToCSV exports metrics to CSV

// - StartSGXReportCollection uses a cron expression to start collection of SGX reports

type HistoricalMetricsScheduler interface {
	CreateMetricsJob(cronExpr string, book_filter string) (*MetricsJob, error)
	DeleteMetricsJob(book_filter string) error
	ListMetricsJobs() ([]MetricsJob, error)
	ListAllMetricsJobsIncludingPortfolio() ([]MetricsJob, error)
	StartMetricsCollection(cronExpr string, book_filter string) func()
	StopMetricsCollection()
}

type HistoricalReportsScheduler interface {
	StartSGXReportCollection(cronExpr string) func()
}

type HistoricalMetricsGetter interface {
	GetMetrics(book_filter string) ([]TimestampedMetrics, error)
	GetMetricsByDateRange(book_filter string, start, end time.Time) ([]TimestampedMetrics, error)
}

type HistoricalMetricsSetter interface {
	StoreCurrentMetrics(book_filter string) error
	UpsertMetric(metric TimestampedMetrics, book_filter string) error
	DeleteMetric(timestamp string, book_filter string) error
	DeleteMetrics(timestamps []string, book_filter string) (DeleteMetricsResponse, error)
}

type HistoricalMetricsCsvManager interface {
	ExportMetricsToCSV(book_filter string) ([]byte, error)
	ImportMetricsFromCSVFile(file multipart.File) (int, error)
}

type HistoricalCachedPricesGetter interface {
	GetCachedPricesWithLatestMetrics(tickers []string) (CachedPricesResponse, error)
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

// CachedPricesRequest represents the request for cached prices by ticker.
type CachedPricesRequest struct {
	Tickers []string `json:"tickers"`
}

// CachedPrice represents a cached daily price snapshot for a ticker.
type CachedPrice struct {
	Ticker    string    `json:"ticker"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// CachedPricesResponse returns cached prices with the latest metrics snapshot.
type CachedPricesResponse struct {
	Metrics     *TimestampedMetrics `json:"metrics,omitempty"`
	Prices      []CachedPrice       `json:"prices"`
	PricesPrev2 []CachedPrice       `json:"pricesPrev2"`
	Missing     []string            `json:"missing,omitempty"`
}
