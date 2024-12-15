package mocks

import (
	"errors"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
)

// MockMarketDataManager is a mock implementation of the MarketDataManager interface
type MockMarketDataManager struct {
	AssetPriceData    map[string]*types.AssetData
	HistoricalData    map[string][]*types.AssetData
	DividendsMetadata map[string][]types.DividendsMetadata
}

// NewMockMarketDataManager creates a new instance of MockMarketDataManager
func NewMockMarketDataManager() *MockMarketDataManager {
	mgr := &MockMarketDataManager{
		AssetPriceData:    make(map[string]*types.AssetData),
		HistoricalData:    make(map[string][]*types.AssetData),
		DividendsMetadata: make(map[string][]types.DividendsMetadata),
	}

	// some sensible defaults, though tests should set this themselves
	mgr.SetDividendMetadata("AAPL", []types.DividendsMetadata{
		{Ticker: "AAPL", ExDate: "2023-01-01", Amount: 1.0, WithholdingTax: 0.3},
		{Ticker: "AAPL", ExDate: "2023-02-01", Amount: 2.0, WithholdingTax: 0.3},
	})

	return mgr
}

// GetAssetPrice returns mock asset price data
func (m *MockMarketDataManager) GetAssetPrice(ticker string) (*types.AssetData, error) {
	if data, ok := m.AssetPriceData[ticker]; ok {
		return data, nil
	}
	return nil, errors.New("mock: unable to fetch stock price")
}

// GetHistoricalData returns mock historical data
func (m *MockMarketDataManager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, error) {
	if data, ok := m.HistoricalData[ticker]; ok {
		return data, nil
	}
	return nil, errors.New("mock: unable to fetch historical data")
}

// GetDividendsMetadata returns mock dividends metadata
func (m *MockMarketDataManager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	if data, ok := m.DividendsMetadata[ticker]; ok {
		return data, nil
	}
	return nil, errors.New("mock: unable to fetch dividends metadata")
}

func (m *MockMarketDataManager) GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error) {
	if data, ok := m.DividendsMetadata[tickerRef.ID]; ok {
		return data, nil
	}
	return nil, errors.New("mock: unable to fetch dividends metadata")
}

// SetDividendMetadata sets mock dividends metadata
func (m *MockMarketDataManager) SetDividendMetadata(ticker string, data []types.DividendsMetadata) {
	m.DividendsMetadata[ticker] = data
}

// SetAssetkPrice sets mock asset price data
func (m *MockMarketDataManager) SetAssetPrice(ticker string, data *types.AssetData) {
	m.AssetPriceData[ticker] = data
}
