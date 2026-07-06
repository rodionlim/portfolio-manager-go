package sources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTradingViewFetchUSAIndustryOverview(t *testing.T) {
	server, requestCount := newTradingViewTestServer(t, "overview", []map[string]any{
		industryColumn(),
		{"id": "MarketCap", "rawValues": []any{15500000000000.0, 100.0}, "viewPropsArgs": []any{[]any{"industry", nil, "USD"}, []any{"industry", nil, "USD"}}},
		{"id": "DividendsYieldForward", "rawValues": []any{0.46, nil}},
		{"id": "Change", "rawValues": []any{5.27, -1.0}},
		{"id": "Volume", "rawValues": []any{114420000.0, 20.0}},
		{"id": "Sector", "rawValues": []any{"Electronic technology", "Finance"}},
		{"id": "BasicElements", "rawValues": []any{100.0, 2.0}},
	})
	defer server.Close()

	source := NewTradingView()
	source.baseURL = server.URL
	rows, err := source.FetchUSAIndustryOverview()
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "Semiconductors", rows[0].Industry)
	assert.Equal(t, "USD", rows[0].Currency)
	assert.Equal(t, 15500000000000.0, *rows[0].MarketCap)
	assert.Nil(t, rows[1].DividendYield)

	cachedRows, err := source.FetchUSAIndustryOverview()
	require.NoError(t, err)
	assert.Equal(t, rows, cachedRows)
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestTradingViewFetchUSAIndustryPerformance(t *testing.T) {
	columns := []map[string]any{industryColumn(), {"id": "Change", "rawValues": []any{5.27, nil}}}
	for _, values := range [][]any{
		{11.79, 1.0}, {19.42, 2.0}, {68.27, 3.0}, {97.65, 4.0},
		{83.98, 5.0}, {190.30, 6.0}, {711.11, 7.0}, {7971.03, 8.0}, {205814.69, 9.0},
	} {
		columns = append(columns, map[string]any{"id": "Performance", "rawValues": values})
	}
	server, requestCount := newTradingViewTestServer(t, "performance", columns)
	defer server.Close()

	source := NewTradingView()
	source.baseURL = server.URL
	rows, err := source.FetchUSAIndustryPerformance()
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, 11.79, *rows[0].OneWeek)
	assert.Equal(t, 205814.69, *rows[0].AllTime)
	assert.Nil(t, rows[1].Change)

	cachedRows, err := source.FetchUSAIndustryPerformance()
	require.NoError(t, err)
	assert.Equal(t, rows, cachedRows)
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestTradingViewFetchUSAIndustryStocksOverview(t *testing.T) {
	columns := []map[string]any{
		stockColumn(),
		{"id": "MarketCap", "rawValues": []any{5.1e12}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "Price", "rawValues": []any{210.69}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "Change", "rawValues": []any{2.95}},
		{"id": "Volume", "rawValues": []any{241271170.0}},
		{"id": "RelativeVolume", "rawValues": []any{1.56}},
		{"id": "PriceToEarnings", "rawValues": []any{32.27}},
		{"id": "EpsDiluted", "rawValues": []any{6.53}},
		{"id": "EpsDilutedGrowth", "rawValues": []any{110.33}},
		{"id": "DividendsYield", "rawValues": []any{0.019}},
		{"id": "Sector", "rawValues": []any{"Electronic technology"}},
		{"id": "AnalystRating", "rawValues": []any{"StrongBuy"}},
	}
	server, requestCount := newTradingViewStocksTestServer(t, "overview", columns)
	defer server.Close()
	source := newTradingViewTestSource(server.URL)

	rows, err := source.FetchUSAIndustryStocksOverview("semiconductors")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "NVDA", rows[0].Ticker)
	assert.Equal(t, "NVIDIA Corporation", rows[0].Company)
	assert.Equal(t, "USD", rows[0].Currency)
	assert.Equal(t, 210.69, *rows[0].Price)

	cachedRows, err := source.FetchUSAIndustryStocksOverview("Semiconductors")
	require.NoError(t, err)
	assert.Equal(t, rows, cachedRows)
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestTradingViewFetchUSAIndustryStocksPerformance(t *testing.T) {
	columns := []map[string]any{
		stockColumn(),
		{"id": "Price", "rawValues": []any{210.69}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "Change", "rawValues": []any{2.95}},
	}
	for _, value := range []float64{4.56, -4.07, 18.37, 19.26, 10.99, 46.31, 1042.86, 17653.53, 481474.94, 2.26, 3.33} {
		columns = append(columns, map[string]any{"id": "Performance", "rawValues": []any{value}})
	}
	server, requestCount := newTradingViewStocksTestServer(t, "performance", columns)
	defer server.Close()
	source := newTradingViewTestSource(server.URL)

	rows, err := source.FetchUSAIndustryStocksPerformance("semiconductors")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 4.56, *rows[0].OneWeek)
	assert.Equal(t, 3.33, *rows[0].VolatilityOneMonth)

	_, err = source.FetchUSAIndustryStocksPerformance("semiconductors")
	require.NoError(t, err)
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestTradingViewFetchUSAStockUnusualVolumeOverview(t *testing.T) {
	columns := []map[string]any{
		stockColumn(),
		{"id": "RelativeVolume", "rawValues": []any{916.75}},
		{"id": "Price", "rawValues": []any{1.115}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "Change", "rawValues": []any{46.71}},
		{"id": "Volume", "rawValues": []any{58890000.0}},
		{"id": "MarketCap", "rawValues": []any{7140000.0}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "PriceToEarnings", "rawValues": []any{nil}},
		{"id": "EpsDiluted", "rawValues": []any{-1.74}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "EpsDilutedGrowth", "rawValues": []any{60.81}},
		{"id": "DividendsYield", "rawValues": []any{0.0}},
		{"id": "Sector", "rawValues": []any{"Retail trade"}},
		{"id": "AnalystRating", "rawValues": []any{"Neutral"}},
	}
	server, requestCount := newTradingViewStockMoversTestServer(t, tradingViewStockUnusualVolumeTableID, columns)
	defer server.Close()
	source := NewTradingView()
	source.baseURL = server.URL

	rows, err := source.FetchUSAStockUnusualVolumeOverview()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "NVDA", rows[0].Ticker)
	assert.Equal(t, 916.75, *rows[0].RelativeVolume)
	assert.Equal(t, "USD", rows[0].MarketCapCurrency)
	assert.Nil(t, rows[0].PriceToEarnings)

	cachedRows, err := source.FetchUSAStockUnusualVolumeOverview()
	require.NoError(t, err)
	assert.Equal(t, rows, cachedRows)
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestTradingViewFetchUSAStockPreMarketMostActiveOverview(t *testing.T) {
	columns := []map[string]any{
		stockColumn(),
		{"id": "PreMarketVolume", "rawValues": []any{303349565.0}},
		{"id": "PreMarketClose", "rawValues": []any{0.042}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "PreMarketChangeAbs", "rawValues": []any{-0.0031}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "PreMarketChange", "rawValues": []any{-6.87}},
		{"id": "PreMarketGap", "rawValues": []any{9.76}},
		{"id": "Price", "rawValues": []any{0.04175}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "Change", "rawValues": []any{-7.43}},
		{"id": "Volume", "rawValues": []any{337901538.0}},
		{"id": "MarketCap", "rawValues": []any{892266.93}, "viewPropsArgs": []any{[]any{"stock", []any{"common"}, "USD"}}},
		{"id": "MarketCapPerf", "rawValues": []any{-34.34}},
	}
	server, requestCount := newTradingViewStockMoversTestServer(t, tradingViewStockPreMarketActiveID, columns)
	defer server.Close()
	source := NewTradingView()
	source.baseURL = server.URL

	rows, err := source.FetchUSAStockPreMarketMostActiveOverview()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "NVDA", rows[0].Ticker)
	assert.Equal(t, 303349565.0, *rows[0].PreMarketVolume)
	assert.Equal(t, "USD", rows[0].PreMarketCurrency)
	assert.Equal(t, -34.34, *rows[0].MarketCapPerformance)

	_, err = source.FetchUSAStockPreMarketMostActiveOverview()
	require.NoError(t, err)
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestIndustrySlug(t *testing.T) {
	assert.Equal(t, "internet-software-services", industrySlug("Internet Software/Services"))
	assert.Equal(t, "property-casualty-insurance", industrySlug("Property/Casualty Insurance"))
}

func industryColumn() map[string]any {
	return map[string]any{
		"id": "TickerLinkIndustry",
		"rawValues": []any{
			map[string]any{"description": "Semiconductors", "description_en": "Semiconductors"},
			map[string]any{"description": "Major banks", "description_en": "Major Banks"},
		},
	}
}

func stockColumn() map[string]any {
	return map[string]any{
		"id": "TickerUniversal",
		"rawValues": []any{map[string]any{
			"description": "NVIDIA Corporation", "exchange": "NASDAQ", "name": "NVDA",
		}},
	}
}

func newTradingViewTestSource(baseURL string) *TradingViewSource {
	source := NewTradingView()
	source.baseURL = baseURL
	source.cache.Set(tradingViewOverviewCacheKey, []types.USAIndustryOverview{
		{ID: "INDUSTRY_US:ELECTRONIC.TECHNOLOGY.SEMICONDUCTORS", Industry: "Semiconductors"},
	}, tradingViewCacheTTL)
	return source
}

func newTradingViewStocksTestServer(t *testing.T, columnSet string, columns []map[string]any) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requestCount := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		assert.Equal(t, tradingViewStocksTableID, r.URL.Query().Get("table_id"))
		assert.Equal(t, columnSet, r.URL.Query().Get("columnset_id"))
		assert.Equal(t, "Semiconductors", r.URL.Query().Get("division_type"))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"totalCount": 1,
			"symbols":    []string{"NASDAQ:NVDA"},
			"data":       columns,
		}))
	}))
	return server, requestCount
}

func newTradingViewStockMoversTestServer(t *testing.T, tableID string, columns []map[string]any) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requestCount := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, tableID, r.URL.Query().Get("table_id"))
		assert.Equal(t, "overview", r.URL.Query().Get("columnset_id"))
		assert.Equal(t, "america", r.URL.Query().Get("market"))
		var body struct {
			Range []int `json:"range"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []int{0, tradingViewETFResultLimit}, body.Range)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"totalCount": 1,
			"symbols":    []string{"NASDAQ:NVDA"},
			"data":       columns,
		}))
	}))
	return server, requestCount
}

func newTradingViewTestServer(t *testing.T, columnSet string, columns []map[string]any) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requestCount := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, tradingViewIndustryTableID, r.URL.Query().Get("table_id"))
		assert.Equal(t, columnSet, r.URL.Query().Get("columnset_id"))
		var body struct {
			Range []int `json:"range"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []int{0, 500}, body.Range)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"totalCount": 2,
			"symbols":    []string{"INDUSTRY_US:SEMICONDUCTORS", "INDUSTRY_US:MAJOR.BANKS"},
			"data":       columns,
		}))
	}))
	return server, requestCount
}
