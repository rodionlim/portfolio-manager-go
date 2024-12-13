package mdata

import (
	"errors"
	"fmt"
	"strings"

	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata/sources"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
)

// MarketDataManager defines the interface for market data management
type MarketDataManager interface {
	GetStockPrice(ticker string) (*types.StockData, error)
	GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.StockData, error)
	GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error)
	GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error)
}

// Manager handles multiple data sources with fallback capability
type Manager struct {
	sources map[string]types.DataSource
	rdata   rdata.ReferenceManager
}

// NewManager creates a new data manager with initialized data sources
func NewManager(db dal.Database, rdata rdata.ReferenceManager) (*Manager, error) {
	m := &Manager{
		sources: make(map[string]types.DataSource),
		rdata:   rdata,
	}

	// Initialize default data sources
	google, err := NewDataSource(sources.GoogleFinance, db)
	if err != nil {
		return nil, err
	}
	yahoo, err := NewDataSource(sources.YahooFinance, db)
	if err != nil {
		return nil, err
	}
	dividendsSg, err := NewDataSource(sources.DividendsSingapore, db)
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

	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return nil, err
	}

	// Try Yahoo Finance first if ticker is available in ref data
	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			if data, err := yahoo.GetStockPrice(tickerRef.YahooTicker); err == nil {
				return data, nil
			}
		}
	}

	// Fallback to Google Finance
	if tickerRef.GoogleTicker != "" {
		if google, ok := m.sources[sources.GoogleFinance]; ok {
			if data, err := google.GetStockPrice(tickerRef.GoogleTicker); err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to fetch stock price %s from any market data sources", ticker)
}

// GetHistoricalData attempts to fetch historical data from available sources
func (m *Manager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.StockData, error) {
	logging.GetLogger().Info("Fetching historical data for ticker", ticker)

	tickerRef, err := m.getReferenceData(ticker)
	if err == nil {
		return nil, err
	}

	// Try Yahoo Finance first if ticker is available in ref data
	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			if data, err := yahoo.GetHistoricalData(tickerRef.YahooTicker, fromDate, toDate); err == nil {
				return data, nil
			}
		}
	}

	// Fallback to Google Finance
	if tickerRef.GoogleTicker != "" {
		if google, ok := m.sources[sources.GoogleFinance]; ok {
			if data, err := google.GetHistoricalData(tickerRef.GoogleTicker, fromDate, toDate); err == nil {
				return data, nil
			}
		}
	}

	return nil, errors.New("unable to fetch historical data from any market data sources")
}

// GetDividendsMetadata attempts to fetch dividends metadata from available sources
func (m *Manager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	tickerRef, err := m.getReferenceData(ticker)
	if err == nil {
		return nil, err
	}

	return m.GetDividendsMetadataFromTickerRef(tickerRef)
}

// GetDividendsMetadataFromTickerRef attempts to fetch dividends metadata from available sources
func (m *Manager) GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error) {
	if tickerRef.DividendsSgTicker != "" {
		// Try Dividends.sg first
		if dividendsSg, ok := m.sources[sources.DividendsSingapore]; ok {
			if data, err := dividendsSg.GetDividendsMetadata(tickerRef.DividendsSgTicker); err == nil {
				return data, nil
			}
		}
	}
	return nil, errors.New("unable to fetch dividends from any source")
}

// NewDataSource creates a new data source engine based on the source type
func NewDataSource(sourceType string, db dal.Database) (types.DataSource, error) {
	switch sourceType {
	case sources.GoogleFinance:
		return sources.NewGoogleFinance(), nil
	case sources.YahooFinance:
		return sources.NewYahooFinance(), nil
	case sources.DividendsSingapore:
		return sources.NewDividendsSg(db), nil
	default:
		return nil, errors.New("unsupported data source")
	}
}

func (m *Manager) getReferenceData(ticker string) (rdata.TickerReference, error) {
	refData, err := m.rdata.GetTicker(strings.ToUpper(ticker))
	if err != nil {
		return rdata.TickerReference{}, err
	}

	return refData, nil
}
