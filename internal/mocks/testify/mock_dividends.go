package testify

import (
	"portfolio-manager/internal/dividends"

	"github.com/stretchr/testify/mock"
)

// MockDividendsManager is a mock implementation of the dividends.Manager interface
type MockDividendsManager struct {
	mock.Mock
}

// CalculateDividendsForAllTickers mocks the CalculateDividendsForAllTickers method
func (m *MockDividendsManager) CalculateDividendsForAllTickers() (map[string][]dividends.Dividends, error) {
	args := m.Called()
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(map[string][]dividends.Dividends), args.Error(1)
}

// CalculateDividendsForSingleTicker mocks the CalculateDividendsForSingleTicker method
func (m *MockDividendsManager) CalculateDividendsForSingleTicker(ticker string) ([]dividends.Dividends, error) {
	args := m.Called(ticker)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]dividends.Dividends), args.Error(1)
}

// CalculateDividendsForSingleBook mocks the CalculateDividendsForSingleBook method
func (m *MockDividendsManager) CalculateDividendsForSingleBook(book string) (map[string][]dividends.Dividends, error) {
	args := m.Called(book)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(map[string][]dividends.Dividends), args.Error(1)
}
