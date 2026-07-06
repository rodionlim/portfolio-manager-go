package mdata

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// MarketDataScreener defines aggregate market screening operations.
type MarketDataScreener interface {
	FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error)
	FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error)
	FetchUSAIndustryStocksOverview(industry string) ([]types.USAIndustryStockOverview, error)
	FetchUSAIndustryStocksPerformance(industry string) ([]types.USAIndustryStockPerformance, error)
	FetchUSAStockUnusualVolumeOverview() ([]types.USAStockUnusualVolumeOverview, error)
	FetchUSAStockPreMarketMostActiveOverview() ([]types.USAStockPreMarketMostActiveOverview, error)
	FetchETFLargestInflowsOverview() ([]types.ETFFundFlowOverview, error)
	FetchETFLargestInflowsPerformance() ([]types.ETFFundFlowPerformance, error)
	FetchETFLargestInflowsFundFlows() ([]types.ETFFundFlows, error)
	FetchETFLargestOutflowsOverview() ([]types.ETFFundFlowOverview, error)
	FetchETFLargestOutflowsPerformance() ([]types.ETFFundFlowPerformance, error)
	FetchETFLargestOutflowsFundFlows() ([]types.ETFFundFlows, error)
	FetchETFSectorOverview() ([]types.ETFSectorOverview, error)
	FetchETFSectorPerformance() ([]types.ETFFundFlowPerformance, error)
	FetchETFSectorFundFlows() ([]types.ETFFundFlows, error)
}

// MarketRotationScreener provides deterministic, compact daily rotation analysis.
type MarketRotationScreener interface {
	ScreenDailyMarketRotation(options MarketRotationOptions) (*types.MarketRotationBrief, error)
}

// MarketRotationOptions controls persistence and response size.
type MarketRotationOptions struct {
	PersistHistory     bool
	MaxStockCandidates int
}

// Manager handles multiple data sources with fallback capability
type Manager struct {
	db              dal.Database
	sources         map[string]types.DataSource
	futuresSources  map[string]*sources.BarchartsSource
	screenerSources map[string]types.ScreenerSource
	rdata           rdata.ReferenceManager
}

var _ MarketDataManager = (*Manager)(nil)
var _ MarketDataScreener = (*Manager)(nil)

func newUSAIndustryOverviewResponse(industries []types.USAIndustryOverview) types.USAIndustryOverviewResponse {
	return types.USAIndustryOverviewResponse{
		PercentageValues: types.PercentageMetadata{
			Unit:   "percent",
			Fields: []string{"dividend_yield", "change"},
			Note:   "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		Industries: industries,
	}
}

func newUSAIndustryPerformanceResponse(industries []types.USAIndustryPerformance) types.USAIndustryPerformanceResponse {
	return types.USAIndustryPerformanceResponse{
		PercentageValues: types.PercentageMetadata{
			Unit: "percent",
			Fields: []string{
				"change", "one_week", "one_month", "three_months", "six_months",
				"ytd", "one_year", "five_years", "ten_years", "all_time",
			},
			Note: "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		Industries: industries,
	}
}

func newUSAIndustryStocksOverviewResponse(industry string, stocks []types.USAIndustryStockOverview) types.USAIndustryStocksOverviewResponse {
	return types.USAIndustryStocksOverviewResponse{
		Industry: industry,
		PercentageValues: types.PercentageMetadata{
			Unit:   "percent",
			Fields: []string{"change", "eps_diluted_growth_yoy_ttm", "dividend_yield_ttm"},
			Note:   "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		Stocks: stocks,
	}
}

func newUSAIndustryStocksPerformanceResponse(industry string, stocks []types.USAIndustryStockPerformance) types.USAIndustryStocksPerformanceResponse {
	return types.USAIndustryStocksPerformanceResponse{
		Industry: industry,
		PercentageValues: types.PercentageMetadata{
			Unit: "percent",
			Fields: []string{
				"change", "one_week", "one_month", "three_months", "six_months", "ytd",
				"one_year", "five_years", "ten_years", "all_time",
				"volatility_one_week", "volatility_one_month",
			},
			Note: "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		Stocks: stocks,
	}
}

func newUSAStockUnusualVolumeOverviewResponse(stocks []types.USAStockUnusualVolumeOverview) types.USAStockUnusualVolumeOverviewResponse {
	return types.USAStockUnusualVolumeOverviewResponse{
		Screen: "unusual-volume",
		PercentageValues: types.PercentageMetadata{
			Unit:   "percent",
			Fields: []string{"change", "eps_diluted_growth_yoy_ttm", "dividend_yield_ttm"},
			Note:   "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		MonetaryValues: types.MonetaryMetadata{
			CurrencyFields: map[string]string{
				"price":           "currency",
				"market_cap":      "market_cap_currency",
				"eps_diluted_ttm": "eps_diluted_currency",
			},
			Fields: []string{"price", "market_cap", "eps_diluted_ttm"},
			Note:   "Monetary values use the corresponding currency field on each stock row.",
		},
		Stocks: stocks,
	}
}

func newUSAStockPreMarketMostActiveOverviewResponse(stocks []types.USAStockPreMarketMostActiveOverview) types.USAStockPreMarketMostActiveOverviewResponse {
	return types.USAStockPreMarketMostActiveOverviewResponse{
		Screen: "pre-market-most-active",
		PercentageValues: types.PercentageMetadata{
			Unit:   "percent",
			Fields: []string{"pre_market_change", "pre_market_gap", "change", "market_cap_performance"},
			Note:   "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		MonetaryValues: types.MonetaryMetadata{
			CurrencyFields: map[string]string{
				"pre_market_close":      "pre_market_currency",
				"pre_market_change_abs": "pre_market_currency",
				"price":                 "currency",
				"market_cap":            "market_cap_currency",
			},
			Fields: []string{"pre_market_close", "pre_market_change_abs", "price", "market_cap"},
			Note:   "Monetary values use the corresponding currency field on each stock row.",
		},
		Stocks: stocks,
	}
}

func newETFFundFlowOverviewResponse(screen string, etfs []types.ETFFundFlowOverview) types.ETFFundFlowOverviewResponse {
	return types.ETFFundFlowOverviewResponse{
		Screen: screen,
		PercentageValues: types.PercentageMetadata{
			Unit:   "percent",
			Fields: []string{"change", "nav_total_return_3y", "expense_ratio"},
			Note:   "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		MonetaryValues: types.MonetaryMetadata{
			CurrencyFields: map[string]string{
				"fund_flows_one_year":     "fund_flows_currency",
				"volume":                  "volume_currency",
				"assets_under_management": "aum_currency",
			},
			Fields: []string{"fund_flows_one_year", "volume", "assets_under_management"},
			Note:   "Monetary values use the corresponding currency field on each ETF row.",
		},
		ETFs: etfs,
	}
}

func newETFFundFlowPerformanceResponse(screen string, etfs []types.ETFFundFlowPerformance) types.ETFFundFlowPerformanceResponse {
	return types.ETFFundFlowPerformanceResponse{
		Screen: screen,
		PercentageValues: types.PercentageMetadata{
			Unit: "percent",
			Fields: []string{
				"change", "one_week", "one_month", "three_months", "six_months",
				"ytd", "one_year", "five_years", "ten_years", "all_time",
			},
			Note: "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		ETFs: etfs,
	}
}

func newETFFundFlowsResponse(screen string, etfs []types.ETFFundFlows) types.ETFFundFlowsResponse {
	return types.ETFFundFlowsResponse{
		Screen: screen,
		MonetaryValues: types.MonetaryMetadata{
			CurrencyFields: map[string]string{
				"one_month": "currency", "three_months": "currency", "one_year": "currency",
				"three_years": "currency", "ytd": "currency",
			},
			Fields: []string{"one_month", "three_months", "one_year", "three_years", "ytd"},
			Note:   "Fund flows are signed monetary values; positive values are inflows and negative values are outflows.",
		},
		ETFs: etfs,
	}
}

func newETFSectorOverviewResponse(etfs []types.ETFSectorOverview) types.ETFSectorOverviewResponse {
	return types.ETFSectorOverviewResponse{
		Screen: "sector-etfs",
		PercentageValues: types.PercentageMetadata{
			Unit:   "percent",
			Fields: []string{"change", "nav_total_return_3y", "expense_ratio"},
			Note:   "Values are percentages: 5.26 means 5.26%, not 526%.",
		},
		MonetaryValues: types.MonetaryMetadata{
			CurrencyFields: map[string]string{
				"volume": "volume_currency", "assets_under_management": "aum_currency",
			},
			Fields: []string{"volume", "assets_under_management"},
			Note:   "Monetary values use the corresponding currency field on each ETF row.",
		},
		ETFs: etfs,
	}
}

// NewManager creates a new data manager with initialized data sources
func NewManager(db dal.Database, rdata rdata.ReferenceManager) (*Manager, error) {
	m := &Manager{
		db:              db,
		sources:         make(map[string]types.DataSource),
		futuresSources:  make(map[string]*sources.BarchartsSource),
		screenerSources: make(map[string]types.ScreenerSource),
		rdata:           rdata,
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
	barcharts := sources.NewBarcharts()
	tradingView := sources.NewTradingView()

	m.sources[sources.GoogleFinance] = google
	m.sources[sources.YahooFinance] = yahoo
	m.sources[sources.DividendsSingapore] = dividendsSg
	m.sources[sources.SSB] = iLoveSsb
	m.sources[sources.MAS] = mas
	m.sources[sources.Nasdaq] = nasdaq
	m.futuresSources[sources.Barcharts] = barcharts
	m.screenerSources[sources.TradingView] = tradingView

	logging.GetLogger().Info("Market data manager initialized with Yahoo/Google finance, Dividends.sg, ILoveSsb, MAS, Nasdaq, Barcharts and TradingView data sources")

	return m, nil
}

func (m *Manager) FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error) {
	source, ok := m.screenerSources[sources.TradingView]
	if !ok {
		return nil, errors.New("TradingView screener source is not configured")
	}
	return source.FetchUSAIndustryPerformance()
}

func (m *Manager) FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error) {
	source, ok := m.screenerSources[sources.TradingView]
	if !ok {
		return nil, errors.New("TradingView screener source is not configured")
	}
	return source.FetchUSAIndustryOverview()
}

func (m *Manager) FetchUSAIndustryStocksOverview(industry string) ([]types.USAIndustryStockOverview, error) {
	source, ok := m.screenerSources[sources.TradingView]
	if !ok {
		return nil, errors.New("TradingView screener source is not configured")
	}
	return source.FetchUSAIndustryStocksOverview(industry)
}

func (m *Manager) FetchUSAIndustryStocksPerformance(industry string) ([]types.USAIndustryStockPerformance, error) {
	source, ok := m.screenerSources[sources.TradingView]
	if !ok {
		return nil, errors.New("TradingView screener source is not configured")
	}
	return source.FetchUSAIndustryStocksPerformance(industry)
}

func (m *Manager) FetchUSAStockUnusualVolumeOverview() ([]types.USAStockUnusualVolumeOverview, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchUSAStockUnusualVolumeOverview()
}

func (m *Manager) FetchUSAStockPreMarketMostActiveOverview() ([]types.USAStockPreMarketMostActiveOverview, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchUSAStockPreMarketMostActiveOverview()
}

func (m *Manager) FetchETFLargestInflowsOverview() ([]types.ETFFundFlowOverview, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFLargestInflowsOverview()
}

func (m *Manager) FetchETFLargestInflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFLargestInflowsPerformance()
}

func (m *Manager) FetchETFLargestInflowsFundFlows() ([]types.ETFFundFlows, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFLargestInflowsFundFlows()
}

func (m *Manager) FetchETFLargestOutflowsOverview() ([]types.ETFFundFlowOverview, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFLargestOutflowsOverview()
}

func (m *Manager) FetchETFLargestOutflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFLargestOutflowsPerformance()
}

func (m *Manager) FetchETFLargestOutflowsFundFlows() ([]types.ETFFundFlows, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFLargestOutflowsFundFlows()
}

func (m *Manager) FetchETFSectorOverview() ([]types.ETFSectorOverview, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFSectorOverview()
}

func (m *Manager) FetchETFSectorPerformance() ([]types.ETFFundFlowPerformance, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFSectorPerformance()
}

func (m *Manager) FetchETFSectorFundFlows() ([]types.ETFFundFlows, error) {
	source, err := m.tradingViewScreener()
	if err != nil {
		return nil, err
	}
	return source.FetchETFSectorFundFlows()
}

func (m *Manager) tradingViewScreener() (types.ScreenerSource, error) {
	source, ok := m.screenerSources[sources.TradingView]
	if !ok {
		return nil, errors.New("TradingView screener source is not configured")
	}
	return source, nil
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

	if common.IsFutures(ticker) {
		normalized := strings.ToUpper(strings.TrimSpace(ticker))
		if len(normalized) >= 4 {
			baseTicker := normalized[:len(normalized)-3]
			if refData, err := m.getReferenceData(baseTicker); err == nil {
				if fromDate > 0 && toDate > 0 {
					duration := time.Unix(toDate, 0).Sub(time.Unix(fromDate, 0))
					if duration > 2*365*24*time.Hour {
						logging.GetLogger().Warnf("Barcharts request for %s spans more than 2 years; UI should cap to 2 years", normalized)
					}
				}

				if barcharts, ok := m.futuresSources[sources.Barcharts]; ok {
					futuresData, err := barcharts.GetHistoricalData(normalized, fromDate, toDate)
					if err != nil {
						return nil, false, err
					}
					assetData := make([]*types.AssetData, 0, len(futuresData))
					for _, entry := range futuresData {
						assetData = append(assetData, &types.AssetData{
							Ticker:    normalized,
							Price:     entry.LastPrice,
							AdjClose:  0,
							Currency:  refData.Ccy,
							Timestamp: entry.Timestamp,
						})
					}
					return assetData, false, nil
				}
			}
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
