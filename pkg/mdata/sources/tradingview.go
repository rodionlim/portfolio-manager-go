package sources

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"

	"portfolio-manager/pkg/types"

	"github.com/patrickmn/go-cache"
)

const (
	tradingViewFacadeURL                 = "https://screener-facade.tradingview.com/screener-facade/api/v1"
	tradingViewIndustryTableID           = "sector_and_industry.industry"
	tradingViewStocksTableID             = "sector_and_industry.industries_companies"
	tradingViewStockUnusualVolumeTableID = "stocks_market_movers.unusual_volume"
	tradingViewStockPreMarketActiveID    = "stocks_market_movers.active_pre_market_stocks"
	tradingViewETFLargestInflowsTableID  = "etfs_funds.largest_inflows"
	tradingViewETFLargestOutflowsTableID = "etfs_funds.largest_outflows"
	tradingViewETFSectorTableID          = "etfs_funds.sector_etfs"
	tradingViewVersion                   = "57"
	tradingViewCacheTTL                  = time.Hour
	tradingViewETFResultLimit            = 100

	tradingViewOverviewCacheKey    = "usa-industry-overview"
	tradingViewPerformanceCacheKey = "usa-industry-performance"
)

type TradingViewSource struct {
	client  *http.Client
	baseURL string
	cache   *cache.Cache

	overviewMu          sync.Mutex
	performanceMu       sync.Mutex
	stocksOverviewMu    sync.Mutex
	stocksPerformanceMu sync.Mutex
	stockMoversMu       sync.Mutex
	etfMu               sync.Mutex
}

var _ types.ScreenerSource = (*TradingViewSource)(nil)

func NewTradingView() *TradingViewSource {
	return &TradingViewSource{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: tradingViewFacadeURL,
		cache:   cache.New(tradingViewCacheTTL, time.Hour),
	}
}

type tradingViewScanResponse struct {
	TotalCount int                 `json:"totalCount"`
	Symbols    []string            `json:"symbols"`
	Data       []tradingViewColumn `json:"data"`
}

type tradingViewColumn struct {
	ID            string              `json:"id"`
	RawValues     []json.RawMessage   `json:"rawValues"`
	ViewPropsArgs [][]json.RawMessage `json:"viewPropsArgs"`
}

type tradingViewIndustry struct {
	Description   string `json:"description"`
	DescriptionEN string `json:"description_en"`
}

type tradingViewStock struct {
	Description string `json:"description"`
	Exchange    string `json:"exchange"`
	Name        string `json:"name"`
}

func (src *TradingViewSource) FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error) {
	if cached, found := src.cache.Get(tradingViewOverviewCacheKey); found {
		if result, ok := cached.([]types.USAIndustryOverview); ok {
			return result, nil
		}
	}
	src.overviewMu.Lock()
	defer src.overviewMu.Unlock()
	if cached, found := src.cache.Get(tradingViewOverviewCacheKey); found {
		if result, ok := cached.([]types.USAIndustryOverview); ok {
			return result, nil
		}
	}

	response, err := src.scan(tradingViewIndustryTableID, "overview", map[string]string{"market": "america"}, 500)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 7 {
		return nil, fmt.Errorf("unexpected TradingView overview column count: %d", len(response.Data))
	}

	result := make([]types.USAIndustryOverview, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		industry, err := industryAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.USAIndustryOverview{
			ID:            id,
			Industry:      industry,
			MarketCap:     numberAt(response.Data[1], i),
			Currency:      currencyAt(response.Data[1], i),
			DividendYield: numberAt(response.Data[2], i),
			Change:        numberAt(response.Data[3], i),
			Volume:        numberAt(response.Data[4], i),
			Sector:        stringAt(response.Data[5], i),
			Stocks:        intAt(response.Data[6], i),
		})
	}
	src.cache.Set(tradingViewOverviewCacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error) {
	if cached, found := src.cache.Get(tradingViewPerformanceCacheKey); found {
		if result, ok := cached.([]types.USAIndustryPerformance); ok {
			return result, nil
		}
	}
	src.performanceMu.Lock()
	defer src.performanceMu.Unlock()
	if cached, found := src.cache.Get(tradingViewPerformanceCacheKey); found {
		if result, ok := cached.([]types.USAIndustryPerformance); ok {
			return result, nil
		}
	}

	response, err := src.scan(tradingViewIndustryTableID, "performance", map[string]string{"market": "america"}, 500)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 11 {
		return nil, fmt.Errorf("unexpected TradingView performance column count: %d", len(response.Data))
	}

	result := make([]types.USAIndustryPerformance, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		industry, err := industryAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.USAIndustryPerformance{
			ID: id, Industry: industry,
			Change: numberAt(response.Data[1], i), OneWeek: numberAt(response.Data[2], i),
			OneMonth: numberAt(response.Data[3], i), ThreeMonths: numberAt(response.Data[4], i),
			SixMonths: numberAt(response.Data[5], i), YTD: numberAt(response.Data[6], i),
			OneYear: numberAt(response.Data[7], i), FiveYears: numberAt(response.Data[8], i),
			TenYears: numberAt(response.Data[9], i), AllTime: numberAt(response.Data[10], i),
		})
	}
	src.cache.Set(tradingViewPerformanceCacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchUSAIndustryStocksOverview(industry string) ([]types.USAIndustryStockOverview, error) {
	resolvedIndustry, err := src.resolveUSAIndustry(industry)
	if err != nil {
		return nil, err
	}
	cacheKey := "usa-industry-stocks-overview:" + industrySlug(resolvedIndustry)
	if cached, found := src.cache.Get(cacheKey); found {
		if result, ok := cached.([]types.USAIndustryStockOverview); ok {
			return result, nil
		}
	}
	src.stocksOverviewMu.Lock()
	defer src.stocksOverviewMu.Unlock()
	if cached, found := src.cache.Get(cacheKey); found {
		if result, ok := cached.([]types.USAIndustryStockOverview); ok {
			return result, nil
		}
	}

	response, err := src.scan(tradingViewStocksTableID, "overview", map[string]string{"market": "america", "division_type": resolvedIndustry}, 500)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 12 {
		return nil, fmt.Errorf("unexpected TradingView stock overview column count: %d", len(response.Data))
	}
	result := make([]types.USAIndustryStockOverview, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		stock, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.USAIndustryStockOverview{
			ID: id, Ticker: stock.Name, Company: stock.Description, Exchange: stock.Exchange,
			MarketCap: numberAt(response.Data[1], i), Currency: currencyAt(response.Data[2], i),
			Price: numberAt(response.Data[2], i), Change: numberAt(response.Data[3], i),
			Volume: numberAt(response.Data[4], i), RelativeVolume: numberAt(response.Data[5], i),
			PriceToEarnings: numberAt(response.Data[6], i), EPSDilutedTTM: numberAt(response.Data[7], i),
			EPSDilutedGrowthYoYTTM: numberAt(response.Data[8], i), DividendYieldTTM: numberAt(response.Data[9], i),
			Sector: stringAt(response.Data[10], i), AnalystRating: stringAt(response.Data[11], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchUSAIndustryStocksPerformance(industry string) ([]types.USAIndustryStockPerformance, error) {
	resolvedIndustry, err := src.resolveUSAIndustry(industry)
	if err != nil {
		return nil, err
	}
	cacheKey := "usa-industry-stocks-performance:" + industrySlug(resolvedIndustry)
	if cached, found := src.cache.Get(cacheKey); found {
		if result, ok := cached.([]types.USAIndustryStockPerformance); ok {
			return result, nil
		}
	}
	src.stocksPerformanceMu.Lock()
	defer src.stocksPerformanceMu.Unlock()
	if cached, found := src.cache.Get(cacheKey); found {
		if result, ok := cached.([]types.USAIndustryStockPerformance); ok {
			return result, nil
		}
	}

	response, err := src.scan(tradingViewStocksTableID, "performance", map[string]string{"market": "america", "division_type": resolvedIndustry}, 500)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 14 {
		return nil, fmt.Errorf("unexpected TradingView stock performance column count: %d", len(response.Data))
	}
	result := make([]types.USAIndustryStockPerformance, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		stock, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.USAIndustryStockPerformance{
			ID: id, Ticker: stock.Name, Company: stock.Description, Exchange: stock.Exchange,
			Currency: currencyAt(response.Data[1], i), Price: numberAt(response.Data[1], i),
			Change: numberAt(response.Data[2], i), OneWeek: numberAt(response.Data[3], i),
			OneMonth: numberAt(response.Data[4], i), ThreeMonths: numberAt(response.Data[5], i),
			SixMonths: numberAt(response.Data[6], i), YTD: numberAt(response.Data[7], i),
			OneYear: numberAt(response.Data[8], i), FiveYears: numberAt(response.Data[9], i),
			TenYears: numberAt(response.Data[10], i), AllTime: numberAt(response.Data[11], i),
			VolatilityOneWeek: numberAt(response.Data[12], i), VolatilityOneMonth: numberAt(response.Data[13], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchUSAStockUnusualVolumeOverview() ([]types.USAStockUnusualVolumeOverview, error) {
	cacheKey := "usa-stock-market-movers:unusual-volume:overview"
	if result, ok := cachedValue[[]types.USAStockUnusualVolumeOverview](src, cacheKey); ok {
		return result, nil
	}
	src.stockMoversMu.Lock()
	defer src.stockMoversMu.Unlock()
	if result, ok := cachedValue[[]types.USAStockUnusualVolumeOverview](src, cacheKey); ok {
		return result, nil
	}

	response, err := src.scan(tradingViewStockUnusualVolumeTableID, "overview", map[string]string{"market": "america"}, tradingViewETFResultLimit)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 12 {
		return nil, fmt.Errorf("unexpected TradingView unusual-volume overview column count: %d", len(response.Data))
	}
	result := make([]types.USAStockUnusualVolumeOverview, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		stock, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.USAStockUnusualVolumeOverview{
			ID: id, Ticker: stock.Name, Company: stock.Description, Exchange: stock.Exchange,
			RelativeVolume: numberAt(response.Data[1], i),
			Currency:       currencyAt(response.Data[2], i), Price: numberAt(response.Data[2], i),
			Change: numberAt(response.Data[3], i), Volume: numberAt(response.Data[4], i),
			MarketCap: numberAt(response.Data[5], i), MarketCapCurrency: currencyAt(response.Data[5], i),
			PriceToEarnings: numberAt(response.Data[6], i),
			EPSDilutedTTM:   numberAt(response.Data[7], i), EPSDilutedCurrency: currencyAt(response.Data[7], i),
			EPSDilutedGrowthYoYTTM: numberAt(response.Data[8], i), DividendYieldTTM: numberAt(response.Data[9], i),
			Sector: stringAt(response.Data[10], i), AnalystRating: stringAt(response.Data[11], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchUSAStockPreMarketMostActiveOverview() ([]types.USAStockPreMarketMostActiveOverview, error) {
	cacheKey := "usa-stock-market-movers:pre-market-most-active:overview"
	if result, ok := cachedValue[[]types.USAStockPreMarketMostActiveOverview](src, cacheKey); ok {
		return result, nil
	}
	src.stockMoversMu.Lock()
	defer src.stockMoversMu.Unlock()
	if result, ok := cachedValue[[]types.USAStockPreMarketMostActiveOverview](src, cacheKey); ok {
		return result, nil
	}

	response, err := src.scan(tradingViewStockPreMarketActiveID, "overview", map[string]string{"market": "america"}, tradingViewETFResultLimit)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 11 {
		return nil, fmt.Errorf("unexpected TradingView pre-market most-active overview column count: %d", len(response.Data))
	}
	result := make([]types.USAStockPreMarketMostActiveOverview, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		stock, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.USAStockPreMarketMostActiveOverview{
			ID: id, Ticker: stock.Name, Company: stock.Description, Exchange: stock.Exchange,
			PreMarketVolume: numberAt(response.Data[1], i),
			PreMarketClose:  numberAt(response.Data[2], i), PreMarketCurrency: currencyAt(response.Data[2], i),
			PreMarketChangeAbs: numberAt(response.Data[3], i), PreMarketChange: numberAt(response.Data[4], i),
			PreMarketGap: numberAt(response.Data[5], i),
			Currency:     currencyAt(response.Data[6], i), Price: numberAt(response.Data[6], i),
			Change: numberAt(response.Data[7], i), Volume: numberAt(response.Data[8], i),
			MarketCap: numberAt(response.Data[9], i), MarketCapCurrency: currencyAt(response.Data[9], i),
			MarketCapPerformance: numberAt(response.Data[10], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchETFLargestInflowsOverview() ([]types.ETFFundFlowOverview, error) {
	return src.fetchETFFundFlowOverview("largest-inflows", tradingViewETFLargestInflowsTableID)
}

func (src *TradingViewSource) FetchETFLargestInflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	return src.fetchETFFundFlowPerformance("largest-inflows", tradingViewETFLargestInflowsTableID)
}

func (src *TradingViewSource) FetchETFLargestInflowsFundFlows() ([]types.ETFFundFlows, error) {
	return src.fetchETFFundFlows("largest-inflows", tradingViewETFLargestInflowsTableID)
}

func (src *TradingViewSource) FetchETFLargestOutflowsOverview() ([]types.ETFFundFlowOverview, error) {
	return src.fetchETFFundFlowOverview("largest-outflows", tradingViewETFLargestOutflowsTableID)
}

func (src *TradingViewSource) FetchETFLargestOutflowsPerformance() ([]types.ETFFundFlowPerformance, error) {
	return src.fetchETFFundFlowPerformance("largest-outflows", tradingViewETFLargestOutflowsTableID)
}

func (src *TradingViewSource) FetchETFLargestOutflowsFundFlows() ([]types.ETFFundFlows, error) {
	return src.fetchETFFundFlows("largest-outflows", tradingViewETFLargestOutflowsTableID)
}

func (src *TradingViewSource) FetchETFSectorOverview() ([]types.ETFSectorOverview, error) {
	cacheKey := "etf-fund-flow:sector-etfs:overview"
	if result, ok := cachedValue[[]types.ETFSectorOverview](src, cacheKey); ok {
		return result, nil
	}
	src.etfMu.Lock()
	defer src.etfMu.Unlock()
	if result, ok := cachedValue[[]types.ETFSectorOverview](src, cacheKey); ok {
		return result, nil
	}

	response, err := src.scan(tradingViewETFSectorTableID, "overview", nil, tradingViewETFResultLimit)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 10 {
		return nil, fmt.Errorf("unexpected TradingView sector ETF overview column count: %d", len(response.Data))
	}
	result := make([]types.ETFSectorOverview, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		fund, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.ETFSectorOverview{
			ID: id, Ticker: fund.Name, Fund: fund.Description, Exchange: fund.Exchange,
			AssetsUnderManagement: numberAt(response.Data[1], i), AUMCurrency: currencyAt(response.Data[1], i),
			Price: numberAt(response.Data[2], i), Currency: currencyAt(response.Data[2], i),
			Change: numberAt(response.Data[3], i), Volume: numberAt(response.Data[4], i),
			VolumeCurrency: currencyAt(response.Data[4], i), RelativeVolume: numberAt(response.Data[5], i),
			NAVTotalReturn3Y: numberAt(response.Data[6], i), ExpenseRatio: numberAt(response.Data[7], i),
			AssetClass: stringAt(response.Data[8], i), Focus: stringAt(response.Data[9], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) FetchETFSectorPerformance() ([]types.ETFFundFlowPerformance, error) {
	return src.fetchETFFundFlowPerformance("sector-etfs", tradingViewETFSectorTableID)
}

func (src *TradingViewSource) FetchETFSectorFundFlows() ([]types.ETFFundFlows, error) {
	return src.fetchETFFundFlows("sector-etfs", tradingViewETFSectorTableID)
}

func (src *TradingViewSource) fetchETFFundFlowOverview(screen, tableID string) ([]types.ETFFundFlowOverview, error) {
	cacheKey := "etf-fund-flow:" + screen + ":overview"
	if result, ok := cachedValue[[]types.ETFFundFlowOverview](src, cacheKey); ok {
		return result, nil
	}
	src.etfMu.Lock()
	defer src.etfMu.Unlock()
	if result, ok := cachedValue[[]types.ETFFundFlowOverview](src, cacheKey); ok {
		return result, nil
	}

	response, err := src.scan(tableID, "overview", nil, tradingViewETFResultLimit)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 11 {
		return nil, fmt.Errorf("unexpected TradingView ETF overview column count: %d", len(response.Data))
	}
	result := make([]types.ETFFundFlowOverview, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		fund, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.ETFFundFlowOverview{
			ID: id, Ticker: fund.Name, Fund: fund.Description, Exchange: fund.Exchange,
			FundFlowsOneYear: numberAt(response.Data[1], i), FundFlowsCurrency: currencyAt(response.Data[1], i),
			Price: numberAt(response.Data[2], i), Currency: currencyAt(response.Data[2], i),
			Change: numberAt(response.Data[3], i), Volume: numberAt(response.Data[4], i),
			VolumeCurrency: currencyAt(response.Data[4], i), RelativeVolume: numberAt(response.Data[5], i),
			AssetsUnderManagement: numberAt(response.Data[6], i), AUMCurrency: currencyAt(response.Data[6], i),
			NAVTotalReturn3Y: numberAt(response.Data[7], i), ExpenseRatio: numberAt(response.Data[8], i),
			AssetClass: stringAt(response.Data[9], i), Focus: stringAt(response.Data[10], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) fetchETFFundFlowPerformance(screen, tableID string) ([]types.ETFFundFlowPerformance, error) {
	cacheKey := "etf-fund-flow:" + screen + ":performance"
	if result, ok := cachedValue[[]types.ETFFundFlowPerformance](src, cacheKey); ok {
		return result, nil
	}
	src.etfMu.Lock()
	defer src.etfMu.Unlock()
	if result, ok := cachedValue[[]types.ETFFundFlowPerformance](src, cacheKey); ok {
		return result, nil
	}

	response, err := src.scan(tableID, "performance", nil, tradingViewETFResultLimit)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 12 {
		return nil, fmt.Errorf("unexpected TradingView ETF performance column count: %d", len(response.Data))
	}
	result := make([]types.ETFFundFlowPerformance, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		fund, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.ETFFundFlowPerformance{
			ID: id, Ticker: fund.Name, Fund: fund.Description, Exchange: fund.Exchange,
			Currency: currencyAt(response.Data[1], i), Price: numberAt(response.Data[1], i),
			Change: numberAt(response.Data[2], i), OneWeek: numberAt(response.Data[3], i),
			OneMonth: numberAt(response.Data[4], i), ThreeMonths: numberAt(response.Data[5], i),
			SixMonths: numberAt(response.Data[6], i), YTD: numberAt(response.Data[7], i),
			OneYear: numberAt(response.Data[8], i), FiveYears: numberAt(response.Data[9], i),
			TenYears: numberAt(response.Data[10], i), AllTime: numberAt(response.Data[11], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func (src *TradingViewSource) fetchETFFundFlows(screen, tableID string) ([]types.ETFFundFlows, error) {
	cacheKey := "etf-fund-flow:" + screen + ":fund-flows"
	if result, ok := cachedValue[[]types.ETFFundFlows](src, cacheKey); ok {
		return result, nil
	}
	src.etfMu.Lock()
	defer src.etfMu.Unlock()
	if result, ok := cachedValue[[]types.ETFFundFlows](src, cacheKey); ok {
		return result, nil
	}

	response, err := src.scan(tableID, "fundFlows", nil, tradingViewETFResultLimit)
	if err != nil {
		return nil, err
	}
	if len(response.Data) != 6 {
		return nil, fmt.Errorf("unexpected TradingView ETF fund-flow column count: %d", len(response.Data))
	}
	result := make([]types.ETFFundFlows, 0, len(response.Symbols))
	for i, id := range response.Symbols {
		fund, err := stockAt(response.Data[0], i)
		if err != nil {
			return nil, err
		}
		result = append(result, types.ETFFundFlows{
			ID: id, Ticker: fund.Name, Fund: fund.Description, Exchange: fund.Exchange,
			Currency: currencyAt(response.Data[1], i), OneMonth: numberAt(response.Data[1], i),
			ThreeMonths: numberAt(response.Data[2], i), OneYear: numberAt(response.Data[3], i),
			ThreeYears: numberAt(response.Data[4], i), YTD: numberAt(response.Data[5], i),
		})
	}
	src.cache.Set(cacheKey, result, tradingViewCacheTTL)
	return result, nil
}

func cachedValue[T any](src *TradingViewSource, key string) (T, bool) {
	var zero T
	value, found := src.cache.Get(key)
	if !found {
		return zero, false
	}
	result, ok := value.(T)
	return result, ok
}

func (src *TradingViewSource) scan(tableID, columnSet string, scanParams map[string]string, limit int) (*tradingViewScanResponse, error) {
	endpoint, err := url.Parse(src.baseURL + "/screener-table/scan")
	if err != nil {
		return nil, fmt.Errorf("failed to build TradingView URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("table_id", tableID)
	query.Set("version", tradingViewVersion)
	query.Set("columnset_id", columnSet)
	for key, value := range scanParams {
		query.Set(key, value)
	}
	endpoint.RawQuery = query.Encode()

	body, err := json.Marshal(map[string]any{
		"lang": "en", "range": []int{0, limit}, "scanner_product_label": "markets-screener",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode TradingView request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create TradingView request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.tradingview.com")
	req.Header.Set("Referer", "https://www.tradingview.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := src.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TradingView %s data: %w", columnSet, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TradingView returned status %d for %s data", resp.StatusCode, columnSet)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read TradingView response: %w", err)
	}
	var result tradingViewScanResponse
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("failed to decode TradingView response: %w", err)
	}
	if len(result.Symbols) == 0 || len(result.Data) == 0 {
		return nil, fmt.Errorf("TradingView returned no %s data", columnSet)
	}
	return &result, nil
}

func (src *TradingViewSource) resolveUSAIndustry(industry string) (string, error) {
	industry = strings.TrimSpace(industry)
	if industry == "" {
		return "", errors.New("industry is required")
	}
	industries, err := src.FetchUSAIndustryOverview()
	if err != nil {
		return "", fmt.Errorf("failed to resolve USA industry %q: %w", industry, err)
	}
	for _, candidate := range industries {
		if strings.EqualFold(candidate.Industry, industry) ||
			strings.EqualFold(candidate.ID, industry) ||
			strings.EqualFold(industrySlug(candidate.Industry), industrySlug(industry)) {
			return candidate.Industry, nil
		}
	}
	return "", fmt.Errorf("unknown USA industry: %s", industry)
}

func industrySlug(value string) string {
	var result strings.Builder
	separator := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(r)
			separator = false
		} else {
			separator = true
		}
	}
	return result.String()
}

func industryAt(column tradingViewColumn, index int) (string, error) {
	if index >= len(column.RawValues) {
		return "", fmt.Errorf("TradingView industry column is missing row %d", index)
	}
	var value tradingViewIndustry
	if err := json.Unmarshal(column.RawValues[index], &value); err != nil {
		return "", fmt.Errorf("failed to decode TradingView industry row %d: %w", index, err)
	}
	if value.DescriptionEN != "" {
		return value.DescriptionEN, nil
	}
	return value.Description, nil
}

func stockAt(column tradingViewColumn, index int) (tradingViewStock, error) {
	if index >= len(column.RawValues) {
		return tradingViewStock{}, fmt.Errorf("TradingView stock column is missing row %d", index)
	}
	var value tradingViewStock
	if err := json.Unmarshal(column.RawValues[index], &value); err != nil {
		return tradingViewStock{}, fmt.Errorf("failed to decode TradingView stock row %d: %w", index, err)
	}
	return value, nil
}

func numberAt(column tradingViewColumn, index int) *float64 {
	if index >= len(column.RawValues) || string(column.RawValues[index]) == "null" {
		return nil
	}
	var value float64
	if json.Unmarshal(column.RawValues[index], &value) != nil {
		return nil
	}
	return &value
}

func stringAt(column tradingViewColumn, index int) string {
	if index >= len(column.RawValues) {
		return ""
	}
	var value string
	_ = json.Unmarshal(column.RawValues[index], &value)
	return value
}

func intAt(column tradingViewColumn, index int) *int {
	value := numberAt(column, index)
	if value == nil {
		return nil
	}
	result := int(*value)
	return &result
}

func currencyAt(column tradingViewColumn, index int) string {
	if index >= len(column.ViewPropsArgs) || len(column.ViewPropsArgs[index]) < 3 {
		return ""
	}
	var currency string
	_ = json.Unmarshal(column.ViewPropsArgs[index][2], &currency)
	return currency
}
