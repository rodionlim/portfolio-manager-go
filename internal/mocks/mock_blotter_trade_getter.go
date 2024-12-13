package mocks

import (
	"errors"
	"fmt"
	"portfolio-manager/internal/blotter"
)

type MockTradeGetterBlotter struct {
	trades map[string][]blotter.Trade
}

func NewMockTradeGetterBlotter() *MockTradeGetterBlotter {
	blotterTradeGetter := MockTradeGetterBlotter{
		trades: make(map[string][]blotter.Trade),
	}
	blotterTradeGetter.SetTrades("AAPL", []blotter.Trade{
		{Ticker: "AAPL", TradeDate: "2022-12-31", Quantity: 100, TradeID: "1"},
		{Ticker: "AAPL", TradeDate: "2023-01-15", Quantity: 200, TradeID: "2"},
	})

	return &blotterTradeGetter
}

func (m *MockTradeGetterBlotter) SetTrades(ticker string, trades []blotter.Trade) {
	m.trades[ticker] = trades
}

func (m *MockTradeGetterBlotter) GetTrades() []blotter.Trade {
	var allTrades []blotter.Trade
	for _, trades := range m.trades {
		allTrades = append(allTrades, trades...)
	}
	return allTrades
}

func (m *MockTradeGetterBlotter) GetTradesByTicker(ticker string) ([]blotter.Trade, error) {
	if trades, ok := m.trades[ticker]; !ok {
		return nil, fmt.Errorf("no trades found for the given ticker %s", ticker)
	} else {
		return trades, nil
	}
}

func (m *MockTradeGetterBlotter) GetTradeByID(tradeID string) (*blotter.Trade, error) {
	return nil, errors.New("not implemented")
}
