package mdata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/require"
)

func TestUSAIndustryPerformanceResponseDeclaresPercentageUnits(t *testing.T) {
	change := 5.26
	manager := &mcpTestScreener{performance: []types.USAIndustryPerformance{
		{Industry: "Semiconductors", Change: &change},
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/mdata/screener/usa/industries/performance", nil)

	HandleUSAIndustryPerformanceGet(manager).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response types.USAIndustryPerformanceResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "percent", response.PercentageValues.Unit)
	require.Contains(t, response.PercentageValues.Fields, "change")
	require.Contains(t, response.PercentageValues.Note, "5.26 means 5.26%")
	require.Equal(t, 5.26, *response.Industries[0].Change)
}

func TestUSAIndustryStocksOverviewHandler(t *testing.T) {
	price := 210.69
	manager := &mcpTestScreener{stocksOverview: []types.USAIndustryStockOverview{
		{Ticker: "NVDA", Company: "NVIDIA Corporation", Price: &price},
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/mdata/screener/usa/industries/semiconductors/stocks/overview", nil)
	request.SetPathValue("industry", "semiconductors")

	HandleUSAIndustryStocksOverviewGet(manager).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response types.USAIndustryStocksOverviewResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "semiconductors", response.Industry)
	require.Equal(t, "NVDA", response.Stocks[0].Ticker)
	require.Equal(t, "semiconductors", manager.requestedIndustry)
}

func TestETFFundFlowOverviewHandler(t *testing.T) {
	flows := 390148137674.63
	fetch := func() ([]types.ETFFundFlowOverview, error) {
		return []types.ETFFundFlowOverview{{Ticker: "CSP1", FundFlowsOneYear: &flows, FundFlowsCurrency: "USD"}}, nil
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/mdata/screener/etfs/largest-inflows/overview", nil)

	handleETFFundFlowOverviewGet("largest-inflows", fetch).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response types.ETFFundFlowOverviewResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "largest-inflows", response.Screen)
	require.Equal(t, "percent", response.PercentageValues.Unit)
	require.Contains(t, response.MonetaryValues.Fields, "fund_flows_one_year")
	require.Equal(t, "CSP1", response.ETFs[0].Ticker)
}

func TestETFSectorOverviewHandler(t *testing.T) {
	aum := 143073007656.36
	fetch := func() ([]types.ETFSectorOverview, error) {
		return []types.ETFSectorOverview{{Ticker: "VGT", AssetsUnderManagement: &aum, AUMCurrency: "USD"}}, nil
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/mdata/screener/etfs/sector-etfs/overview", nil)

	handleETFSectorOverviewGet(fetch).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response types.ETFSectorOverviewResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "sector-etfs", response.Screen)
	require.Equal(t, "aum_currency", response.MonetaryValues.CurrencyFields["assets_under_management"])
	require.Equal(t, "VGT", response.ETFs[0].Ticker)
}
