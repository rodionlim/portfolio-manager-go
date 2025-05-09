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
func (m *MockMetricsService) CalculatePortfolioMetrics() (metrics.MetricResultsWithCashFlows, error) {
	args := m.Called()
	return args.Get(0).(metrics.MetricResultsWithCashFlows), args.Error(1)
}
