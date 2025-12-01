package testify

import (
	"portfolio-manager/pkg/rdata"

	"github.com/stretchr/testify/mock"
)

// MockReferenceManager is a mock implementation of the ReferenceManager interface
type MockReferenceManager struct {
	mock.Mock
}

// AddTicker mocks the AddTicker method
func (m *MockReferenceManager) AddTicker(ticker rdata.TickerReference) (string, error) {
	args := m.Called(ticker)
	return args.String(0), args.Error(1)
}

// UpdateTicker mocks the UpdateTicker method
func (m *MockReferenceManager) UpdateTicker(ticker *rdata.TickerReference) error {
	args := m.Called(ticker)
	return args.Error(0)
}

// DeleteTicker mocks the DeleteTicker method
func (m *MockReferenceManager) DeleteTicker(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetTicker mocks the GetTicker method
func (m *MockReferenceManager) GetTicker(id string) (rdata.TickerReferenceWithSGXMapped, error) {
	args := m.Called(id)
	return args.Get(0).(rdata.TickerReferenceWithSGXMapped), args.Error(1)
}

// GetAllTickers mocks the GetAllTickers method
func (m *MockReferenceManager) GetAllTickers() (map[string]rdata.TickerReferenceWithSGXMapped, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]rdata.TickerReferenceWithSGXMapped), args.Error(1)
}

// ExportToYamlBytes mocks the ExportToYamlBytes method
func (m *MockReferenceManager) ExportToYamlBytes() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
