package mdata

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"portfolio-manager/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

func TestScreenerMCPToolNamesDisambiguateStockIndustriesAndSectorETFs(t *testing.T) {
	server := mcpserver.NewMCPServer("test", "1.0.0")
	RegisterScreenerMCPTools(server, &Manager{})
	tools := server.ListTools()

	for _, name := range []string{
		toolFetchStockIndustryOverview,
		toolFetchStockIndustryPerformance,
		toolFetchStockIndustryStocksOverview,
		toolFetchStockIndustryStocksPerformance,
		toolFetchStockUnusualVolumeOverview,
		toolFetchStockPreMarketActiveOverview,
		toolFetchSectorETFOverview,
		toolFetchSectorETFPerformance,
		toolFetchSectorETFFundFlows,
		toolScreenDailyMarketRotation,
	} {
		require.Contains(t, tools, name)
	}
	for _, oldName := range []string{
		"fetch_usa_industry_overview", "fetch_usa_industry_performance",
		"fetch_usa_industry_stocks_overview", "fetch_usa_industry_stocks_performance",
		"fetch_etf_sector_overview", "fetch_etf_sector_performance", "fetch_etf_sector_fund_flows",
	} {
		require.NotContains(t, tools, oldName)
	}
	require.Contains(t, tools[toolFetchStockIndustryPerformance].Tool.Description, "no fund-flow fields")
	require.Contains(t, tools[toolFetchStockUnusualVolumeOverview].Tool.Description, "does not provide fund flows")
	require.NotNil(t, tools[toolFetchStockPreMarketActiveOverview].Tool.Annotations.ReadOnlyHint)
	require.True(t, *tools[toolFetchStockPreMarketActiveOverview].Tool.Annotations.ReadOnlyHint)
	require.Contains(t, tools[toolFetchSectorETFFundFlows].Tool.Description, "stock industries do not provide fund flows")
}

func TestScreenDailyMarketRotationMCPHandler(t *testing.T) {
	runner := &rotationMCPTestRunner{brief: &types.MarketRotationBrief{MethodologyVersion: types.MarketRotationMethodologyVersion}}
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"persist_history": false, "max_stock_candidates": float64(3),
	}}}

	result, err := createHandleScreenDailyMarketRotation(runner)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.False(t, runner.options.PersistHistory)
	require.Equal(t, 3, runner.options.MaxStockCandidates)
}

type rotationMCPTestRunner struct {
	brief   *types.MarketRotationBrief
	options MarketRotationOptions
}

func (r *rotationMCPTestRunner) ScreenDailyMarketRotation(options MarketRotationOptions) (*types.MarketRotationBrief, error) {
	r.options = options
	return r.brief, nil
}

func TestFetchUSAIndustryOverviewMCPHandler(t *testing.T) {
	marketCap := 15.5e12
	manager := &mcpTestScreener{overview: []types.USAIndustryOverview{
		{ID: "INDUSTRY_US:SEMICONDUCTORS", Industry: "Semiconductors", MarketCap: &marketCap, Currency: "USD"},
	}}

	result, err := createHandleFetchUSAIndustryOverview(manager)(context.Background(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	var response types.USAIndustryOverviewResponse
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, "percent", response.PercentageValues.Unit)
	require.Contains(t, response.PercentageValues.Note, "5.26 means 5.26%")
	require.Equal(t, "Semiconductors", response.Industries[0].Industry)
}

func TestFetchUSAIndustryPerformanceMCPHandlerError(t *testing.T) {
	manager := &mcpTestScreener{performanceErr: errors.New("upstream unavailable")}

	result, err := createHandleFetchUSAIndustryPerformance(manager)(context.Background(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, content.Text, "upstream unavailable")
}

func TestFetchUSAIndustryStocksPerformanceMCPHandler(t *testing.T) {
	oneWeek := 4.56
	manager := &mcpTestScreener{stocksPerformance: []types.USAIndustryStockPerformance{
		{Ticker: "NVDA", OneWeek: &oneWeek},
	}}
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]any{"industry": "semiconductors"},
	}}

	result, err := createHandleFetchUSAIndustryStocksPerformance(manager)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Equal(t, "semiconductors", manager.requestedIndustry)
	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	var response types.USAIndustryStocksPerformanceResponse
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, "percent", response.PercentageValues.Unit)
	require.Equal(t, "NVDA", response.Stocks[0].Ticker)
}

func TestFetchUSAStockUnusualVolumeOverviewMCPHandler(t *testing.T) {
	relativeVolume := 916.75
	manager := &mcpTestScreener{unusualVolume: []types.USAStockUnusualVolumeOverview{
		{Ticker: "LUCY", RelativeVolume: &relativeVolume},
	}}

	result, err := createHandleFetchUSAStockUnusualVolumeOverview(manager)(context.Background(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	var response types.USAStockUnusualVolumeOverviewResponse
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, "unusual-volume", response.Screen)
	require.Equal(t, "percent", response.PercentageValues.Unit)
	require.Equal(t, "LUCY", response.Stocks[0].Ticker)
	require.Equal(t, 916.75, *response.Stocks[0].RelativeVolume)
}

func TestFetchUSAStockPreMarketMostActiveOverviewMCPHandler(t *testing.T) {
	volume := 303349565.0
	manager := &mcpTestScreener{preMarketMostActive: []types.USAStockPreMarketMostActiveOverview{
		{Ticker: "YHC", PreMarketVolume: &volume},
	}}

	result, err := createHandleFetchUSAStockPreMarketMostActiveOverview(manager)(context.Background(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	var response types.USAStockPreMarketMostActiveOverviewResponse
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, "pre-market-most-active", response.Screen)
	require.Contains(t, response.PercentageValues.Fields, "pre_market_change")
	require.Equal(t, "YHC", response.Stocks[0].Ticker)
}

func TestFetchETFFundFlowsMCPHandler(t *testing.T) {
	oneMonth := -125000000.0
	fetch := func() ([]types.ETFFundFlows, error) {
		return []types.ETFFundFlows{{Ticker: "SPY", Currency: "USD", OneMonth: &oneMonth}}, nil
	}

	result, err := createHandleFetchETFFundFlows("largest-outflows", fetch)(context.Background(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	content, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	var response types.ETFFundFlowsResponse
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, "largest-outflows", response.Screen)
	require.Equal(t, "currency", response.MonetaryValues.CurrencyFields["one_month"])
	require.Contains(t, response.MonetaryValues.Note, "negative values are outflows")
}

type mcpTestScreener struct {
	MarketDataScreener
	overview             []types.USAIndustryOverview
	overviewErr          error
	performance          []types.USAIndustryPerformance
	performanceErr       error
	stocksOverview       []types.USAIndustryStockOverview
	stocksOverviewErr    error
	stocksPerformance    []types.USAIndustryStockPerformance
	stocksPerformanceErr error
	unusualVolume        []types.USAStockUnusualVolumeOverview
	unusualVolumeErr     error
	preMarketMostActive  []types.USAStockPreMarketMostActiveOverview
	preMarketErr         error
	requestedIndustry    string
}

func (m *mcpTestScreener) FetchUSAIndustryOverview() ([]types.USAIndustryOverview, error) {
	return m.overview, m.overviewErr
}

func (m *mcpTestScreener) FetchUSAIndustryPerformance() ([]types.USAIndustryPerformance, error) {
	return m.performance, m.performanceErr
}

func (m *mcpTestScreener) FetchUSAIndustryStocksOverview(industry string) ([]types.USAIndustryStockOverview, error) {
	m.requestedIndustry = industry
	return m.stocksOverview, m.stocksOverviewErr
}

func (m *mcpTestScreener) FetchUSAIndustryStocksPerformance(industry string) ([]types.USAIndustryStockPerformance, error) {
	m.requestedIndustry = industry
	return m.stocksPerformance, m.stocksPerformanceErr
}

func (m *mcpTestScreener) FetchUSAStockUnusualVolumeOverview() ([]types.USAStockUnusualVolumeOverview, error) {
	return m.unusualVolume, m.unusualVolumeErr
}

func (m *mcpTestScreener) FetchUSAStockPreMarketMostActiveOverview() ([]types.USAStockPreMarketMostActiveOverview, error) {
	return m.preMarketMostActive, m.preMarketErr
}
