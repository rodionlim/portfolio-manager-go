package testify

import (
	"portfolio-manager/internal/metrics"

	"github.com/stretchr/testify/mock"
)

// MockMetricsService is a mock of the metrics.MetricsService
type MockMetricsService struct {
	mock.Mock
}

// CalculatePortfolioMetrics is a mock implementation of the CalculatePortfolioMetrics method
func (m *MockMetricsService) CalculatePortfolioMetrics(book_filter string) (metrics.MetricResultsWithCashFlows, error) {
	args := m.Called()
	return args.Get(0).(metrics.MetricResultsWithCashFlows), args.Error(1)
}

// BenchmarkPortfolioPerformance is a mock implementation of the BenchmarkPortfolioPerformance method
func (m *MockMetricsService) BenchmarkPortfolioPerformance(req metrics.BenchmarkRequest) (metrics.BenchmarkComparisonResult, error) {
	args := m.Called(req)
	return args.Get(0).(metrics.BenchmarkComparisonResult), args.Error(1)
}
