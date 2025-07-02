package historical

import (
	"portfolio-manager/internal/metrics"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/internal/mocks/testify"
	"portfolio-manager/pkg/scheduler"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestStoreCurrentMetrics tests the StoreCurrentMetrics function
func TestStoreCurrentMetrics(t *testing.T) {
	// Create mock dependencies
	mockMetrics := new(testify.MockMetricsService)
	mockDB := new(mocks.MockDatabase)
	mockScheduler := scheduler.NewScheduler() // Using real scheduler for simplicity

	// Sample metrics result
	sampleMetrics := metrics.MetricResultsWithCashFlows{
		Metrics: metrics.MetricsResult{
			IRR:            0.05,
			PricePaid:      1000.0,
			MV:             1200.0,
			TotalDividends: 50.0,
		},
		CashFlows: []metrics.CashFlow{
			{
				Date:        "2023-01-01T00:00:00Z",
				Cash:        -1000.0,
				Ticker:      "AAPL",
				Description: metrics.CashFlowTypeBuy,
			},
			{
				Date:        "2023-06-01T00:00:00Z",
				Cash:        25.0,
				Ticker:      "AAPL",
				Description: metrics.CashFlowTypeDividend,
			},
		},
	}

	// Configure mock behavior
	mockMetrics.On("CalculatePortfolioMetrics").Return(sampleMetrics, nil)
	mockDB.On("Put", mock.AnythingOfType("string"), mock.AnythingOfType("historical.TimestampedMetrics")).Run(func(args mock.Arguments) {
		// Verify that the timestamp is midnight (date only)
		metrics := args.Get(1).(TimestampedMetrics)
		// Check that the time component is zeroed out (midnight)
		assert.Equal(t, 0, metrics.Timestamp.Hour())
		assert.Equal(t, 0, metrics.Timestamp.Minute())
		assert.Equal(t, 0, metrics.Timestamp.Second())

		// Also check that the key format is just YYYY-MM-DD (no time component)
		key := args.Get(0).(string)
		parts := strings.Split(key, ":")
		dateStr := parts[2]
		_, err := time.Parse("2006-01-02", dateStr)
		assert.NoError(t, err, "Date format should be YYYY-MM-DD")
	}).Return(nil)

	// Create the service with our mock
	service := &Service{
		metricsService: mockMetrics,
		db:             mockDB,
		scheduler:      mockScheduler,
		logger:         nil, // Not needed for test
	}

	// Call the method under test
	err := service.StoreCurrentMetrics("")

	// Assertions
	assert.NoError(t, err)
	mockMetrics.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// TestGetMetrics tests fetching all historical metrics
func TestGetMetrics(t *testing.T) {
	mockMetrics := new(testify.MockMetricsService)
	mockDB := new(mocks.MockDatabase)
	mockScheduler := scheduler.NewScheduler()

	keys := []string{
		"METRICS:portfolio:2023-05-01",
		"METRICS:portfolio:2023-05-15",
		"METRICS:portfolio:2023-06-01",
	}

	mockDB.On("GetAllKeysWithPrefix", mock.AnythingOfType("string")).Return(keys, nil)
	mockDB.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("*historical.TimestampedMetrics")).Run(func(args mock.Arguments) {
		metricsPtr := args.Get(1).(*TimestampedMetrics)
		key := args.Get(0).(string)
		parts := strings.Split(key, ":")
		dateStr := parts[2]
		date, _ := time.Parse("2006-01-02", dateStr)
		*metricsPtr = TimestampedMetrics{
			Timestamp: date,
			Metrics: metrics.MetricsResult{
				IRR:       0.05,
				PricePaid: 1000.0,
				MV:        1200.0,
			},
		}
	}).Return(nil)

	service := &Service{
		metricsService: mockMetrics,
		db:             mockDB,
		scheduler:      mockScheduler,
		logger:         nil,
	}

	results, err := service.GetMetrics("")
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	mockDB.AssertExpectations(t)
}

// TestGetMetricsByDateRange tests fetching metrics within a date range
func TestGetMetricsByDateRange(t *testing.T) {
	mockMetrics := new(testify.MockMetricsService)
	mockDB := new(mocks.MockDatabase)
	mockScheduler := scheduler.NewScheduler()

	startDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	keys := []string{
		"METRICS:portfolio:2023-05-01",
		"METRICS:portfolio:2023-05-15",
		"METRICS:portfolio:2023-06-01",
	}

	mockDB.On("GetAllKeysWithPrefix", mock.AnythingOfType("string")).Return(keys, nil)
	mockDB.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("*historical.TimestampedMetrics")).Run(func(args mock.Arguments) {
		metricsPtr := args.Get(1).(*TimestampedMetrics)
		key := args.Get(0).(string)
		parts := strings.Split(key, ":")
		dateStr := parts[2]
		date, _ := time.Parse("2006-01-02", dateStr)
		*metricsPtr = TimestampedMetrics{
			Timestamp: date,
			Metrics: metrics.MetricsResult{
				IRR:       0.05,
				PricePaid: 1000.0,
				MV:        1200.0,
			},
		}
	}).Return(nil)

	service := &Service{
		metricsService: mockMetrics,
		db:             mockDB,
		scheduler:      mockScheduler,
		logger:         nil,
	}

	results, err := service.GetMetricsByDateRange("", startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	mockDB.AssertExpectations(t)
}
