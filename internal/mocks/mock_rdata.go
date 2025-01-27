package mocks

import (
	"errors"
	"portfolio-manager/pkg/rdata"
)

type MockReferenceManager struct {
	Tickers map[string]rdata.TickerReference
}

func NewMockReferenceManager() *MockReferenceManager {
	refMgr := &MockReferenceManager{
		Tickers: make(map[string]rdata.TickerReference),
	}

	refMgr.AddTicker(rdata.TickerReference{
		ID:                "AAPL",
		DividendsSgTicker: "AAPL",
	})

	return refMgr
}

func (m *MockReferenceManager) AddTicker(ticker rdata.TickerReference) (string, error) {
	if ticker.ID == "" {
		return "", errors.New("ticker ID is required")
	}
	m.Tickers[ticker.ID] = ticker
	return ticker.ID, nil
}

func (m *MockReferenceManager) UpdateTicker(ticker *rdata.TickerReference) error {
	if ticker.ID == "" {
		return errors.New("ticker ID is required")
	}
	m.Tickers[ticker.ID] = *ticker
	return nil
}

func (m *MockReferenceManager) DeleteTicker(id string) error {
	delete(m.Tickers, id)
	return nil
}

func (m *MockReferenceManager) GetTicker(id string) (rdata.TickerReference, error) {
	ticker, exists := m.Tickers[id]
	if !exists {
		return rdata.TickerReference{}, errors.New("ticker not found")
	}
	return ticker, nil
}

func (m *MockReferenceManager) GetAllTickers() (map[string]rdata.TickerReference, error) {
	return m.Tickers, nil
}

func (m *MockReferenceManager) ExportToYamlBytes() ([]byte, error) {
	return []byte{}, nil
}
