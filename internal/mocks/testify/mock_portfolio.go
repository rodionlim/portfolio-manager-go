package testify

import (
	"portfolio-manager/internal/portfolio"

	"github.com/stretchr/testify/mock"
)

// MockPortfolioReader is a mock implementation of the portfolio.PortfolioReader interface
type MockPortfolioReader struct {
	mock.Mock
}

// GetAllPositions mocks the GetAllPositions method
func (m *MockPortfolioReader) GetAllPositions() ([]*portfolio.Position, error) {
	args := m.Called()
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*portfolio.Position), args.Error(1)
}
