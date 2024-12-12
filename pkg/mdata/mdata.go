package mdata

import (
	"errors"

	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata/sources"
	"portfolio-manager/pkg/types"
)

// Manager handles multiple data sources with fallback capability
type Manager struct {
	sources map[string]types.DataSource
}

// NewManager creates a new data manager with initialized data sources
func NewManager() (*Manager, error) {
	m := &Manager{
		sources: make(map[string]types.DataSource),
	}

	// Initialize default data sources
	google, err := NewDataSource(sources.GoogleFinance)
	if err != nil {
		return nil, err
	}
	yahoo, err := NewDataSource(sources.YahooFinance)
	if err != nil {
		return nil, err
	}
	dividendsSg, err := NewDataSource(sources.DividendsSingapore)
	if err != nil {
		return nil, err
	}

	m.sources[sources.GoogleFinance] = google
	m.sources[sources.YahooFinance] = yahoo
	m.sources[sources.DividendsSingapore] = dividendsSg

	logging.GetLogger().Info("Market data manager initialized with Yahoo/Google finance and Dividends.sg data sources")

	return m, nil
}

// GetStockPrice attempts to fetch stock price from available sources
func (m *Manager) GetStockPrice(ticker string) (*types.StockData, error) {
	logging.GetLogger().Info("Fetching stock price for ticker", ticker)

	// Try Yahoo Finance first
	if yahoo, ok := m.sources[sources.YahooFinance]; ok {
		if data, err := yahoo.GetStockPrice(ticker); err == nil {
			return data, nil
		}
	}

	// Fallback to Google Finance
	if google, ok := m.sources[sources.GoogleFinance]; ok {
		if data, err := google.GetStockPrice(ticker); err == nil {
			return data, nil
		}
	}

	return nil, errors.New("unable to fetch stock price from any source")
}

// GetHistoricalData attempts to fetch historical data from available sources
func (m *Manager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.StockData, error) {
	// Try Yahoo Finance first
	if yahoo, ok := m.sources[sources.YahooFinance]; ok {
		if data, err := yahoo.GetHistoricalData(ticker, fromDate, toDate); err == nil {
			return data, nil
		}
	}

	// Fallback to Google Finance
	if google, ok := m.sources[sources.GoogleFinance]; ok {
		if data, err := google.GetHistoricalData(ticker, fromDate, toDate); err == nil {
			return data, nil
		}
	}

	return nil, errors.New("unable to fetch historical data from any source")
}

// GetDividends attempts to fetch dividends metadata from available sources
func (m *Manager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	// Try Dividends.sg first
	if dividendsSg, ok := m.sources[sources.DividendsSingapore]; ok {
		if data, err := dividendsSg.GetDividendsMetadata(ticker); err == nil {
			return data, nil
		}
	}

	return nil, errors.New("unable to fetch dividends from any source")
}

// NewDataSource creates a new data source engine based on the source type
func NewDataSource(sourceType string) (types.DataSource, error) {
	switch sourceType {
	case sources.GoogleFinance:
		return sources.NewGoogleFinance(), nil
	case sources.YahooFinance:
		return sources.NewYahooFinance(), nil
	case sources.DividendsSingapore:
		return sources.NewDividendsSg(), nil
	default:
		return nil, errors.New("unsupported data source")
	}
}
