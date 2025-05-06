package testify

import (
	"encoding/csv"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/pkg/event"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockTradeBlotter is a mock implementation of the TradeBlotter struct
type MockTradeBlotter struct {
	mock.Mock
}

// GetTrades mocks the GetTrades method of TradeBlotter
func (m *MockTradeBlotter) GetTrades() []blotter.Trade {
	args := m.Called()
	return args.Get(0).([]blotter.Trade)
}

// GetTradeByID mocks the GetTradeByID method of TradeBlotter
func (m *MockTradeBlotter) GetTradeByID(tradeID string) (*blotter.Trade, error) {
	args := m.Called(tradeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blotter.Trade), args.Error(1)
}

// GetTradesByTicker mocks the GetTradesByTicker method of TradeBlotter
func (m *MockTradeBlotter) GetTradesByTicker(ticker string) ([]blotter.Trade, error) {
	args := m.Called(ticker)
	return args.Get(0).([]blotter.Trade), args.Error(1)
}

// GetAllTickers mocks the GetAllTickers method of TradeBlotter
func (m *MockTradeBlotter) GetAllTickers() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// AddTrade mocks the AddTrade method of TradeBlotter
func (m *MockTradeBlotter) AddTrade(trade blotter.Trade) error {
	args := m.Called(trade)
	return args.Error(0)
}

// UpdateTrade mocks the UpdateTrade method of TradeBlotter
func (m *MockTradeBlotter) UpdateTrade(trade blotter.Trade) error {
	args := m.Called(trade)
	return args.Error(0)
}

// RemoveTrade mocks the RemoveTrade method of TradeBlotter
func (m *MockTradeBlotter) RemoveTrade(tradeID string) error {
	args := m.Called(tradeID)
	return args.Error(0)
}

// RemoveTrades mocks the RemoveTrades method of TradeBlotter
func (m *MockTradeBlotter) RemoveTrades(tradeIDs []string) error {
	args := m.Called(tradeIDs)
	return args.Error(0)
}

// RemoveAllTrades mocks the RemoveAllTrades method of TradeBlotter
func (m *MockTradeBlotter) RemoveAllTrades() error {
	args := m.Called()
	return args.Error(0)
}

// Subscribe mocks the Subscribe method of TradeBlotter
func (m *MockTradeBlotter) Subscribe(eventName string, handler event.EventHandler) {
	m.Called(eventName, handler)
}

// Unsubscribe mocks the Unsubscribe method of TradeBlotter
func (m *MockTradeBlotter) Unsubscribe(eventName string, corrID uuid.UUID) {
	m.Called(eventName, corrID)
}

// GetCurrentSeqNum mocks the GetCurrentSeqNum method of TradeBlotter
func (m *MockTradeBlotter) GetCurrentSeqNum() int {
	args := m.Called()
	return args.Int(0)
}

// GetTradesBySeqNumRange mocks the GetTradesBySeqNumRange method of TradeBlotter
func (m *MockTradeBlotter) GetTradesBySeqNumRange(startSeqNum, endSeqNum int) []blotter.Trade {
	args := m.Called(startSeqNum, endSeqNum)
	return args.Get(0).([]blotter.Trade)
}

// GetTradesBySeqNumRangeWithCallback mocks the GetTradesBySeqNumRangeWithCallback method of TradeBlotter
func (m *MockTradeBlotter) GetTradesBySeqNumRangeWithCallback(startSeqNum, endSeqNum int, callback func(blotter.Trade)) {
	m.Called(startSeqNum, endSeqNum, callback)
}

// ExportToCSVBytes mocks the ExportToCSVBytes method of TradeBlotter
func (m *MockTradeBlotter) ExportToCSVBytes() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// ImportFromCSVFile mocks the ImportFromCSVFile method of TradeBlotter
func (m *MockTradeBlotter) ImportFromCSVFile(filepath string) (int, error) {
	args := m.Called(filepath)
	return args.Int(0), args.Error(1)
}

// ImportFromCSVReader mocks the ImportFromCSVReader method of TradeBlotter
func (m *MockTradeBlotter) ImportFromCSVReader(reader *csv.Reader) (int, error) {
	args := m.Called(reader)
	return args.Int(0), args.Error(1)
}

// MockBlotterTradeGetter is a mock implementation of blotter.TradeGetter
type MockBlotterTradeGetter struct {
	mock.Mock
}

// GetTrades mocks the GetTrades method of TradeGetter
func (m *MockBlotterTradeGetter) GetTrades() []blotter.Trade {
	args := m.Called()
	return args.Get(0).([]blotter.Trade)
}

// GetTradeByID mocks the GetTradeByID method of TradeGetter
func (m *MockBlotterTradeGetter) GetTradeByID(tradeID string) (*blotter.Trade, error) {
	args := m.Called(tradeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blotter.Trade), args.Error(1)
}

// GetTradesByTicker mocks the GetTradesByTicker method of TradeGetter
func (m *MockBlotterTradeGetter) GetTradesByTicker(ticker string) ([]blotter.Trade, error) {
	args := m.Called(ticker)
	return args.Get(0).([]blotter.Trade), args.Error(1)
}

// GetAllTickers mocks the GetAllTickers method of TradeGetter
func (m *MockBlotterTradeGetter) GetAllTickers() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}
