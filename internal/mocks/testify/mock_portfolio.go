package testify

import (
	"portfolio-manager/internal/portfolio"

	"github.com/stretchr/testify/mock"
)

// MockPortfolioGetter is a mock implementation of the portfolio.PortfolioGetter interface
type MockPortfolioGetter struct {
	mock.Mock
}

// GetAllPositions mocks the GetAllPositions method
func (m *MockPortfolioGetter) GetAllPositions() ([]*portfolio.Position, error) {
	args := m.Called()
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*portfolio.Position), args.Error(1)
}
