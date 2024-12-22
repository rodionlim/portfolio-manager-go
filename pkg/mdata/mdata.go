package mdata

import (
	"errors"
	"fmt"
	"strings"

	"portfolio-manager/internal/config"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata/sources"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
)

// MarketDataManager defines the interface for market data management
type MarketDataManager interface {
	GetAssetPrice(ticker string) (*types.AssetData, error)
	GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, error)
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
	iLoveSsb, err := NewDataSource(sources.SSB, db)
	if err != nil {
		return nil, err
	}
	mas, err := NewDataSource(sources.MAS, db)
	if err != nil {
		return nil, err
	}

	m.sources[sources.GoogleFinance] = google
	m.sources[sources.YahooFinance] = yahoo
	m.sources[sources.DividendsSingapore] = dividendsSg
	m.sources[sources.SSB] = iLoveSsb
	m.sources[sources.MAS] = mas

	logging.GetLogger().Info("Market data manager initialized with Yahoo/Google finance, Dividends.sg, ILoveSsb and MAS data sources")

	return m, nil
}

// GetAssetPrice attempts to fetch asset price from available sources
func (m *Manager) GetAssetPrice(ticker string) (*types.AssetData, error) {
	logging.GetLogger().Info("Fetching asset price for ticker", ticker)

	// for SSB, tickers are standardized against the following convention, e.g. SBJAN25
	if common.IsSSB(ticker) {
		if iLoveSsb, ok := m.sources[sources.SSB]; ok {
			data, err := iLoveSsb.GetAssetPrice(ticker)
			return data, err
		}
	}

	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return nil, err
	}

	// Try Yahoo Finance first if ticker is available in ref data
	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			if data, err := yahoo.GetAssetPrice(tickerRef.YahooTicker); err == nil {
				return data, nil
			}
		}
	}

	// Fallback to Google Finance
	if tickerRef.GoogleTicker != "" {
		if google, ok := m.sources[sources.GoogleFinance]; ok {
			if data, err := google.GetAssetPrice(tickerRef.GoogleTicker); err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to fetch asset price %s from any market data sources", ticker)
}

// GetHistoricalData attempts to fetch historical data from available sources
func (m *Manager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, error) {
	logging.GetLogger().Info("Fetching historical data for ticker", ticker)

	// for SSB, tickers are standardized against the following convention, e.g. SBJAN25
	if common.IsSSB(ticker) {
		if iLoveSsb, ok := m.sources[sources.SSB]; ok {
			data, err := iLoveSsb.GetHistoricalData(ticker, fromDate, toDate)
			return data, err
		}
	}

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
	ticker = strings.ToUpper(ticker)

	// All other tickers go through normalization via reference data
	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return nil, err
	}

	return m.GetDividendsMetadataFromTickerRef(tickerRef)
}

// GetDividendsMetadataFromTickerRef attempts to fetch dividends metadata from available sources
func (m *Manager) GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error) {
	witholdingTax := m.MapDomicileToWitholdingTax(tickerRef.Domicile)

	// for SSB, tickers are standardized against the following convention, e.g. SBJAN25
	if common.IsSSB(tickerRef.ID) {
		if iLoveSsb, ok := m.sources[sources.SSB]; ok {
			return iLoveSsb.GetDividendsMetadata(tickerRef.ID, witholdingTax)
		}
	}

	// for SG MAS Bills, tickers are standardized against the following convention, e.g. BS24124Z
	if common.IsSgTBill(tickerRef.ID) {
		if mas, ok := m.sources[sources.MAS]; ok {
			return mas.GetDividendsMetadata(tickerRef.ID, witholdingTax)
		}
	}

	// Try Dividends.sg first
	if tickerRef.DividendsSgTicker != "" {
		if dividendsSg, ok := m.sources[sources.DividendsSingapore]; ok {
			if data, err := dividendsSg.GetDividendsMetadata(tickerRef.DividendsSgTicker, witholdingTax); err == nil {
				return data, nil
			}
		}
	}

	// Fallback to Yahoo Finance
	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			if data, err := yahoo.GetDividendsMetadata(tickerRef.YahooTicker, witholdingTax); err == nil {
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
		return sources.NewYahooFinance(db), nil
	case sources.DividendsSingapore:
		return sources.NewDividendsSg(db), nil
	case sources.SSB:
		return sources.NewILoveSsb(db), nil
	case sources.MAS:
		return sources.NewMas(db), nil
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

func (m *Manager) MapDomicileToWitholdingTax(domicile string) float64 {
	cfg, err := config.GetOrCreateConfig("")
	if err != nil {
		return 0.0
	}

	switch domicile {
	case "SG":
		return cfg.DivWitholdingTaxSG
	case "US":
		return cfg.DivWitholdingTaxUS
	case "HK":
		return cfg.DivWitholdingTaxHK
	case "IE":
		return cfg.DivWitholdingTaxIE
	default:
		return 0.0
	}
}
