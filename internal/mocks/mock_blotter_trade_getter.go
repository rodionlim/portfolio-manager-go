package mocks

import (
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
)

type MockTradeGetterBlotter struct{}

func NewMockTradeGetterBlotter(db dal.Database) *MockTradeGetterBlotter {
	return &MockTradeGetterBlotter{}
}

func (m *MockTradeGetterBlotter) GetTrades() []blotter.Trade {
	// Mock implementation
	return []blotter.Trade{
		{TradeDate: "2022-12-31", Quantity: 100},
		{TradeDate: "2023-01-15", Quantity: 200},
	}
}

func (m *MockTradeGetterBlotter) GetTradesByTicker(ticker string) ([]blotter.Trade, error) {
	return []blotter.Trade{
		{TradeDate: "2022-12-31", Quantity: 100},
		{TradeDate: "2023-01-15", Quantity: 200},
	}, nil
}

func (m *MockTradeGetterBlotter) GetTradeByID(tradeID string) (*blotter.Trade, error) {
	return &blotter.Trade{
		TradeDate: "2022-12-31", Quantity: 100,
	}, nil
}
