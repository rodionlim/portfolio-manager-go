package testify

import (
	"encoding/csv"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/mock"
)

// MockMarketDataManager is a mock implementation of the MarketDataManager interface
type MockMarketDataManager struct {
	mock.Mock
}

// GetAssetPrice mocks the GetAssetPrice method
func (m *MockMarketDataManager) GetAssetPrice(ticker string) (*types.AssetData, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.AssetData), args.Error(1)
}

// GetHistoricalData mocks the GetHistoricalData method
func (m *MockMarketDataManager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, error) {
	args := m.Called(ticker, fromDate, toDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.AssetData), args.Error(1)
}

// StoreCustomDividendsMetadata mocks the StoreCustomDividendsMetadata method
func (m *MockMarketDataManager) StoreCustomDividendsMetadata(ticker string, dividends []types.DividendsMetadata) error {
	args := m.Called(ticker, dividends)
	return args.Error(0)
}

// ImportCustomDividendsFromCSVReader mocks the ImportCustomDividendsFromCSVReader method
func (m *MockMarketDataManager) ImportCustomDividendsFromCSVReader(reader *csv.Reader) (int, error) {
	args := m.Called(reader)
	return args.Int(0), args.Error(1)
}

// GetDividendsMetadata mocks the GetDividendsMetadata method
func (m *MockMarketDataManager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.DividendsMetadata), args.Error(1)
}

// GetDividendsMetadataFromTickerRef mocks the GetDividendsMetadataFromTickerRef method
func (m *MockMarketDataManager) GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error) {
	args := m.Called(tickerRef)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.DividendsMetadata), args.Error(1)
}
