package mdata

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
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
	GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, bool, error) // bool indicates if resync is needed
	GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error)
	GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error)
	ImportCustomDividendsFromCSVReader(*csv.Reader) (int, error)
	StoreCustomDividendsMetadata(ticker string, dividends []types.DividendsMetadata) error
	DeleteDividendsMetadata(ticker string, isCustom bool) error
	FetchBenchmarkInterestRates(country string, points int) ([]types.InterestRates, error)
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
	nasdaq, err := NewDataSource(sources.Nasdaq, db)
	if err != nil {
		return nil, err
	}

	m.sources[sources.GoogleFinance] = google
	m.sources[sources.YahooFinance] = yahoo
	m.sources[sources.DividendsSingapore] = dividendsSg
	m.sources[sources.SSB] = iLoveSsb
	m.sources[sources.MAS] = mas
	m.sources[sources.Nasdaq] = nasdaq

	logging.GetLogger().Info("Market data manager initialized with Yahoo/Google finance, Dividends.sg, ILoveSsb, MAS and Nasdaq data sources")

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

	// for SG MAS Bills, tickers are standardized against the following convention, e.g. BS24124Z, BY24124Z etc.
	if common.IsSgTBill(ticker) {
		if mas, ok := m.sources[sources.MAS]; ok {
			data, err := mas.GetAssetPrice(ticker)
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

	// Also allow Dividends Sg for exotic SG tickers like bonds etc.
	if tickerRef.DividendsSgTicker != "" {
		if dividendsSg, ok := m.sources[sources.DividendsSingapore]; ok {
			if data, err := dividendsSg.GetAssetPrice(tickerRef.DividendsSgTicker); err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to fetch asset price %s from any market data sources", ticker)
}

// GetHistoricalData attempts to fetch historical data from available sources
func (m *Manager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, bool, error) {
	logging.GetLogger().Info("Fetching historical data for ticker", ticker)

	// for SSB, tickers are standardized against the following convention, e.g. SBJAN25
	if common.IsSSB(ticker) {
		if iLoveSsb, ok := m.sources[sources.SSB]; ok {
			return iLoveSsb.GetHistoricalData(ticker, fromDate, toDate)
		}
	}

	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return nil, false, err
	}

	// Try Yahoo Finance first if ticker is available in ref data
	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			if data, resync, err := yahoo.GetHistoricalData(tickerRef.YahooTicker, fromDate, toDate); err == nil {
				return data, resync, nil
			} else {
				logging.GetLogger().Errorf("Failed to fetch historical data from Yahoo Finance for %s: %v", tickerRef.YahooTicker, err)
			}
		}
	}

	// Fallback to Google Finance
	if tickerRef.GoogleTicker != "" {
		if google, ok := m.sources[sources.GoogleFinance]; ok {
			if data, resync, err := google.GetHistoricalData(tickerRef.GoogleTicker, fromDate, toDate); err == nil {
				return data, resync, nil
			}
		}
	}

	return nil, false, errors.New("unable to fetch historical data from any market data sources")
}

// GetDividendsMetadata attempts to fetch dividends metadata from available sources
func (m *Manager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	ticker = strings.ToUpper(ticker)

	// All other tickers go through normalization via reference data
	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference data for ticker %s, %v", ticker, err)
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

	// Try Nasdaq next
	if tickerRef.NasdaqTicker != "" {
		if nasdaq, ok := m.sources[sources.Nasdaq]; ok {
			if data, err := nasdaq.GetDividendsMetadata(tickerRef.NasdaqTicker, witholdingTax); err == nil {
				return data, nil
			} else {
				logging.GetLogger().Errorf("Failed to fetch dividends metadata from Nasdaq for %s: %v", tickerRef.NasdaqTicker, err)
			}
		}
	}

	// Fallback to Yahoo Finance as last resort
	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			if data, err := yahoo.GetDividendsMetadata(tickerRef.YahooTicker, witholdingTax); err == nil {
				return data, nil
			} else {
				logging.GetLogger().Errorf("Failed to fetch dividends metadata from Yahoo Finance for %s: %v", tickerRef.YahooTicker, err)
			}
		}
	}

	return nil, fmt.Errorf("unable to fetch dividends [%s] from any source", tickerRef.ID)
}

// ImportCustomDividendsFromCSVReader imports custom dividends metadata from a CSV reader
func (m *Manager) ImportCustomDividendsFromCSVReader(reader *csv.Reader) (int, error) {
	logging.GetLogger().Info("Importing dividends metadata from CSV")

	// Read and validate header
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("error reading CSV header: %w", err)
	}

	expectedHeaders := []string{"Ticker", "ExDate", "Amount", "Interest", "AvgInterest", "WithholdingTax"}
	if len(header) != len(expectedHeaders) {
		return 0, fmt.Errorf("invalid CSV format: expected %d columns, got %d", len(expectedHeaders), len(header))
	}

	for i, h := range expectedHeaders {
		if header[i] != h {
			return 0, fmt.Errorf("invalid CSV header: expected %s at position %d, got %s", h, i, header[i])
		}
	}

	// Read all rows and create dividends metadata
	var dividends []*types.DividendsMetadata
	lineNum := 1
	for {
		cnt := lineNum - 1
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return cnt, fmt.Errorf("error reading CSV line %d: %w", lineNum, err)
		}

		amount, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return cnt, fmt.Errorf("invalid amount at line %d: %w", lineNum, err)
		}

		var interest float64
		if row[3] != "" {
			interest, err = strconv.ParseFloat(row[3], 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid interest at line %d: %w", lineNum, err)
			}
		}

		var avgInterest float64
		if row[4] != "" {
			avgInterest, err = strconv.ParseFloat(row[4], 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid average interest at line %d: %w", lineNum, err)
			}
		}

		var withholdingTax float64
		if row[5] != "" {
			withholdingTax, err = strconv.ParseFloat(row[5], 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid withholding tax at line %d: %w", lineNum, err)
			}
		}

		dividendsMetadata := &types.DividendsMetadata{
			Ticker:         row[0],
			ExDate:         row[1],
			Amount:         amount,
			Interest:       interest,
			AvgInterest:    avgInterest,
			WithholdingTax: withholdingTax,
		}

		dividends = append(dividends, dividendsMetadata)
		lineNum++
	}

	// Add all dividends by ticker after validation
	// First group by ticker
	dividendsByTicker := make(map[string][]types.DividendsMetadata)
	for _, dividend := range dividends {
		dividendsByTicker[dividend.Ticker] = append(dividendsByTicker[dividend.Ticker], *dividend)
	}

	// Store all dividends by ticker
	for ticker, dividends := range dividendsByTicker {
		if err := m.StoreCustomDividendsMetadata(ticker, dividends); err != nil {
			return lineNum, fmt.Errorf("error adding dividends metadata: %w", err)
		}
	}

	return len(dividends), nil
}

// StoreCustomDividendsMetadata stores custom dividends metadata for a single ticker
func (m *Manager) StoreCustomDividendsMetadata(ticker string, dividends []types.DividendsMetadata) error {
	ticker = strings.ToUpper(ticker)

	// All other tickers go through normalization via reference data
	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return err
	}

	if tickerRef.DividendsSgTicker != "" {
		if dividendsSg, ok := m.sources[sources.DividendsSingapore]; ok {
			_, err = dividendsSg.StoreDividendsMetadata(tickerRef.DividendsSgTicker, dividends, true)
			return err
		}
	}

	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			_, err = yahoo.StoreDividendsMetadata(tickerRef.YahooTicker, dividends, true)
			return err
		}
	}

	return errors.New("unable to store custom dividends metadata for this data source")
}

// DeleteDividendsMetadata deletes custom or official dividends metadata for a single ticker
func (m *Manager) DeleteDividendsMetadata(ticker string, isCustom bool) error {
	ticker = strings.ToUpper(ticker)

	// All other tickers go through normalization via reference data
	tickerRef, err := m.getReferenceData(ticker)
	if err != nil {
		return err
	}

	if tickerRef.DividendsSgTicker != "" {
		if dividendsSg, ok := m.sources[sources.DividendsSingapore]; ok {
			return dividendsSg.DeleteDividendsMetadata(tickerRef.DividendsSgTicker, isCustom)
		}
	}

	if tickerRef.YahooTicker != "" {
		if yahoo, ok := m.sources[sources.YahooFinance]; ok {
			return yahoo.DeleteDividendsMetadata(tickerRef.YahooTicker, isCustom)
		}
	}

	return errors.New("unable to delete dividends metadata for this data source")
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
	case sources.Nasdaq:
		return sources.NewNasdaq(db), nil
	default:
		return nil, errors.New("unsupported data source")
	}
}

func (m *Manager) getReferenceData(ticker string) (rdata.TickerReference, error) {
	refData, err := m.rdata.GetTicker(strings.ToUpper(ticker))
	if err != nil {
		return rdata.TickerReference{}, err
	}

	return refData.TickerReference, nil
}

func (m *Manager) MapDomicileToWitholdingTax(domicile string) float64 {
	cfg, err := config.GetOrCreateConfig("config.yaml")
	if err != nil {
		return 0.0
	}

	switch domicile {
	case "SG":
		return cfg.Dividends.WithholdingTaxSG
	case "US":
		return cfg.Dividends.WithholdingTaxUS
	case "HK":
		return cfg.Dividends.WithholdingTaxHK
	case "IE":
		return cfg.Dividends.WithholdingTaxIE
	default:
		return 0.0
	}
}

// FetchBenchmarkInterestRates fetches benchmark interest rates from available sources
func (m *Manager) FetchBenchmarkInterestRates(country string, points int) ([]types.InterestRates, error) {
	logging.GetLogger().Infof("Fetching benchmark interest rates for country %s with %d points", country, points)

	// For now, default to MAS data source for Singapore
	if country == "SG" {
		if mas, ok := m.sources[sources.MAS]; ok {
			return mas.FetchBenchmarkInterestRates(country, points)
		}
	}

	return nil, fmt.Errorf("unsupported country: %s", country)
}
