package sources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTradingViewFetchETFLargestInflowsOverview(t *testing.T) {
	columns := []map[string]any{
		etfColumn(),
		currencyColumn("FundFlows", 390148137674.63, "USD"),
		currencyColumn("Price", 61001.0, "GBX"),
		{"id": "Change", "rawValues": []any{0.25}},
		currencyColumn("VolumePrice", 4155668.42, "USD"),
		{"id": "RelativeVolume", "rawValues": []any{0.52}},
		currencyColumn("AssetsUnderManagement", 147987682708.09, "USD"),
		{"id": "NavTotalReturn", "rawValues": []any{70.68}},
		{"id": "ExpenseRatio", "rawValues": []any{0.07}},
		{"id": "AssetClass", "rawValues": []any{"Equity"}},
		{"id": "Focus", "rawValues": []any{"Large cap"}},
	}
	server, requests := newTradingViewETFTestServer(t, tradingViewETFLargestInflowsTableID, "overview", columns)
	defer server.Close()
	source := NewTradingView()
	source.baseURL = server.URL

	rows, err := source.FetchETFLargestInflowsOverview()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "CSP1", rows[0].Ticker)
	assert.Equal(t, "GBX", rows[0].Currency)
	assert.Equal(t, "USD", rows[0].FundFlowsCurrency)
	assert.Equal(t, 0.07, *rows[0].ExpenseRatio)

	_, err = source.FetchETFLargestInflowsOverview()
	require.NoError(t, err)
	assert.Equal(t, int32(1), requests.Load())
}

func TestTradingViewFetchETFLargestOutflowsPerformance(t *testing.T) {
	columns := []map[string]any{
		etfColumn(), currencyColumn("Price", 440.0, "IDR"), {"id": "Change", "rawValues": []any{-1.25}},
	}
	for _, value := range []float64{-2, -3, -4, -5, -6, -7, -8, -9, -10} {
		columns = append(columns, map[string]any{"id": "Performance", "rawValues": []any{value}})
	}
	server, requests := newTradingViewETFTestServer(t, tradingViewETFLargestOutflowsTableID, "performance", columns)
	defer server.Close()
	source := NewTradingView()
	source.baseURL = server.URL

	rows, err := source.FetchETFLargestOutflowsPerformance()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, -2.0, *rows[0].OneWeek)
	assert.Equal(t, -10.0, *rows[0].AllTime)
	assert.Equal(t, "IDR", rows[0].Currency)

	_, err = source.FetchETFLargestOutflowsPerformance()
	require.NoError(t, err)
	assert.Equal(t, int32(1), requests.Load())
}

func TestTradingViewFetchETFLargestInflowsFundFlows(t *testing.T) {
	columns := []map[string]any{etfColumn()}
	for _, value := range []float64{10, 30, 100, 300, 80} {
		columns = append(columns, currencyColumn("FundFlows", value, "USD"))
	}
	server, requests := newTradingViewETFTestServer(t, tradingViewETFLargestInflowsTableID, "fundFlows", columns)
	defer server.Close()
	source := NewTradingView()
	source.baseURL = server.URL

	rows, err := source.FetchETFLargestInflowsFundFlows()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 10.0, *rows[0].OneMonth)
	assert.Equal(t, 300.0, *rows[0].ThreeYears)
	assert.Equal(t, "USD", rows[0].Currency)

	_, err = source.FetchETFLargestInflowsFundFlows()
	require.NoError(t, err)
	assert.Equal(t, int32(1), requests.Load())
}

func TestTradingViewFetchETFSectorOverview(t *testing.T) {
	columns := []map[string]any{
		etfColumn(),
		currencyColumn("AssetsUnderManagement", 143073007656.36, "USD"),
		currencyColumn("Price", 120.04, "USD"),
		{"id": "Change", "rawValues": []any{2.66}},
		currencyColumn("VolumePrice", 440176596.64, "USD"),
		{"id": "RelativeVolume", "rawValues": []any{0.60}},
		{"id": "NavTotalReturn", "rawValues": []any{123.78}},
		{"id": "ExpenseRatio", "rawValues": []any{0.09}},
		{"id": "AssetClass", "rawValues": []any{"Equity"}},
		{"id": "Focus", "rawValues": []any{"Information technology"}},
	}
	server, requests := newTradingViewETFTestServer(t, tradingViewETFSectorTableID, "overview", columns)
	defer server.Close()
	source := NewTradingView()
	source.baseURL = server.URL

	rows, err := source.FetchETFSectorOverview()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "CSP1", rows[0].Ticker)
	assert.Equal(t, 143073007656.36, *rows[0].AssetsUnderManagement)
	assert.Equal(t, "Information technology", rows[0].Focus)

	_, err = source.FetchETFSectorOverview()
	require.NoError(t, err)
	assert.Equal(t, int32(1), requests.Load())
}

func etfColumn() map[string]any {
	return map[string]any{
		"id": "TickerUniversal",
		"rawValues": []any{map[string]any{
			"description": "iShares Core S&P 500 UCITS ETF", "exchange": "LSE", "name": "CSP1",
		}},
	}
}

func currencyColumn(id string, value float64, currency string) map[string]any {
	return map[string]any{
		"id": id, "rawValues": []any{value},
		"viewPropsArgs": []any{[]any{"fund", []any{"etf"}, currency}},
	}
}

func newTradingViewETFTestServer(t *testing.T, tableID, columnSet string, columns []map[string]any) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requests := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		assert.Equal(t, tableID, r.URL.Query().Get("table_id"))
		assert.Equal(t, columnSet, r.URL.Query().Get("columnset_id"))
		assert.Empty(t, r.URL.Query().Get("market"))
		var body struct {
			Range []int `json:"range"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []int{0, 100}, body.Range)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"totalCount": 100,
			"symbols":    []string{"LSE:CSP1"},
			"data":       columns,
		}))
	}))
	return server, requests
}
