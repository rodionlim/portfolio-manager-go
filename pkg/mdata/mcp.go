package mdata

import (
	"context"
	"encoding/json"
	"fmt"
	"portfolio-manager/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	toolFetchStockIndustryOverview          = "fetch_stock_industry_overview"
	toolFetchStockIndustryPerformance       = "fetch_stock_industry_performance"
	toolFetchStockIndustryStocksOverview    = "fetch_stock_industry_stocks_overview"
	toolFetchStockIndustryStocksPerformance = "fetch_stock_industry_stocks_performance"
	toolFetchStockUnusualVolumeOverview     = "fetch_stock_unusual_volume_overview"
	toolFetchStockPreMarketActiveOverview   = "fetch_stock_premarket_most_active_overview"
	toolFetchSectorETFOverview              = "fetch_sector_etf_overview"
	toolFetchSectorETFPerformance           = "fetch_sector_etf_performance"
	toolFetchSectorETFFundFlows             = "fetch_sector_etf_fund_flows"
	toolScreenDailyMarketRotation           = "screen_daily_market_rotation"
)

// RegisterMCPTools registers market data MCP tools.
func RegisterMCPTools(mcpServer *server.MCPServer, manager MarketDataManager) {
	// Fetch benchmark interest rates tool
	benchmarkRatesTool := mcp.NewTool("fetch_benchmark_interest_rates",
		mcp.WithDescription("Fetch benchmark interest rates for a specific country"),
		mcp.WithString("country",
			mcp.Description("Country code (e.g., 'SG' for Singapore)"),
		),
		mcp.WithNumber("points",
			mcp.Description("Number of data points to fetch (default: 100)"),
		),
	)

	mcpServer.AddTool(benchmarkRatesTool, createHandleFetchBenchmarkInterestRates(manager))
}

// RegisterScreenerMCPTools registers aggregate market screening MCP tools.
func RegisterScreenerMCPTools(mcpServer *server.MCPServer, screener MarketDataScreener) {
	industryOverviewTool := mcp.NewTool(toolFetchStockIndustryOverview,
		mcp.WithDescription("Fetch overview metrics for USA stock-market industries. This is not an ETF dataset and does not contain fund flows. Do not use it for sector ETF or capital-flow requests."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	mcpServer.AddTool(industryOverviewTool, createHandleFetchUSAIndustryOverview(screener))

	industryPerformanceTool := mcp.NewTool(toolFetchStockIndustryPerformance,
		mcp.WithDescription("Fetch performance for USA stock-market industries across daily, weekly, monthly, YTD, and multi-year periods. This is not an ETF dataset and has no fund-flow fields. Do not use it for sector ETF fund flows."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	mcpServer.AddTool(industryPerformanceTool, createHandleFetchUSAIndustryPerformance(screener))

	industryStocksOverviewTool := mcp.NewTool(toolFetchStockIndustryStocksOverview,
		mcp.WithDescription("Fetch overview metrics for individual stocks within a USA stock-market industry. This returns underlying companies, not ETFs, and does not provide fund flows."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("industry",
			mcp.Required(),
			mcp.Description("Industry name or URL slug, for example 'Semiconductors' or 'semiconductors'"),
		),
	)
	mcpServer.AddTool(industryStocksOverviewTool, createHandleFetchUSAIndustryStocksOverview(screener))

	industryStocksPerformanceTool := mcp.NewTool(toolFetchStockIndustryStocksPerformance,
		mcp.WithDescription("Fetch performance for individual stocks within a USA stock-market industry. This returns underlying companies, not ETFs, and does not provide fund flows."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("industry",
			mcp.Required(),
			mcp.Description("Industry name or URL slug, for example 'Semiconductors' or 'semiconductors'"),
		),
	)
	mcpServer.AddTool(industryStocksPerformanceTool, createHandleFetchUSAIndustryStocksPerformance(screener))

	stockUnusualVolumeTool := mcp.NewTool(toolFetchStockUnusualVolumeOverview,
		mcp.WithDescription("Fetch the TradingView USA unusual-volume stock market-mover overview screen, sorted by relative volume. This returns stocks, not ETFs, and does not provide fund flows."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	mcpServer.AddTool(stockUnusualVolumeTool, createHandleFetchUSAStockUnusualVolumeOverview(screener))

	stockPreMarketActiveTool := mcp.NewTool(toolFetchStockPreMarketActiveOverview,
		mcp.WithDescription("Fetch the TradingView USA pre-market most-active stock market-mover overview screen, sorted by pre-market volume. This returns stocks, not ETFs, and does not provide fund flows."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	mcpServer.AddTool(stockPreMarketActiveTool, createHandleFetchUSAStockPreMarketMostActiveOverview(screener))

	etfTools := []struct {
		name        string
		description string
		handler     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{"fetch_etf_largest_inflows_overview", "Fetch the top 100 global ETFs ranked by largest inflows with overview metrics", createHandleFetchETFFundFlowOverview("largest-inflows", screener.FetchETFLargestInflowsOverview)},
		{"fetch_etf_largest_inflows_performance", "Fetch performance metrics for the top 100 global ETFs ranked by largest inflows. Percentage values use percent units.", createHandleFetchETFFundFlowPerformance("largest-inflows", screener.FetchETFLargestInflowsPerformance)},
		{"fetch_etf_largest_inflows_fund_flows", "Fetch 1M, 3M, 1Y, 3Y, and YTD fund flows for the top 100 global ETFs ranked by largest inflows", createHandleFetchETFFundFlows("largest-inflows", screener.FetchETFLargestInflowsFundFlows)},
		{"fetch_etf_largest_outflows_overview", "Fetch the top 100 global ETFs ranked by largest outflows with overview metrics", createHandleFetchETFFundFlowOverview("largest-outflows", screener.FetchETFLargestOutflowsOverview)},
		{"fetch_etf_largest_outflows_performance", "Fetch performance metrics for the top 100 global ETFs ranked by largest outflows. Percentage values use percent units.", createHandleFetchETFFundFlowPerformance("largest-outflows", screener.FetchETFLargestOutflowsPerformance)},
		{"fetch_etf_largest_outflows_fund_flows", "Fetch 1M, 3M, 1Y, 3Y, and YTD fund flows for the top 100 global ETFs ranked by largest outflows", createHandleFetchETFFundFlows("largest-outflows", screener.FetchETFLargestOutflowsFundFlows)},
		{toolFetchSectorETFPerformance, "Fetch performance for the top 100 global sector ETFs. This returns ETFs, not underlying stock industries. Use for sector ETF performance, not individual-stock performance.", createHandleFetchETFFundFlowPerformance("sector-etfs", screener.FetchETFSectorPerformance)},
		{toolFetchSectorETFFundFlows, "Fetch 1M, 3M, 1Y, 3Y, and YTD monetary fund flows for global sector ETFs. Use this when a request mentions sector fund flows, capital flows, inflows, or outflows; stock industries do not provide fund flows.", createHandleFetchETFFundFlows("sector-etfs", screener.FetchETFSectorFundFlows)},
	}
	for _, tool := range etfTools {
		mcpServer.AddTool(mcp.NewTool(tool.name, mcp.WithDescription(tool.description), mcp.WithReadOnlyHintAnnotation(true)), tool.handler)
	}
	mcpServer.AddTool(
		mcp.NewTool(toolFetchSectorETFOverview, mcp.WithDescription("Fetch overview metrics for the top 100 global sector ETFs. This returns ETFs, not stock-market industries or their underlying companies."), mcp.WithReadOnlyHintAnnotation(true)),
		createHandleFetchETFSectorOverview(screener.FetchETFSectorOverview),
	)
	if rotationScreener, ok := screener.(MarketRotationScreener); ok {
		rotationTool := mcp.NewTool(toolScreenDailyMarketRotation,
			mcp.WithDescription("Build a deterministic daily US market-rotation brief from TradingView sector ETF fund flows, sector/industry performance, and stock performance. All joins, calculations, rankings, exclusions, and history transitions are precomputed; narrate the returned order without recalculating or inventing stock-level fund flows."),
			mcp.WithBoolean("persist_history", mcp.Description("Persist this US session for transition detection. Defaults to true; set false for validation and dry runs.")),
			mcp.WithNumber("max_stock_candidates", mcp.Description("Maximum stock candidates to return, from 1 to 5. Defaults to 5.")),
		)
		mcpServer.AddTool(rotationTool, createHandleScreenDailyMarketRotation(rotationScreener))
	}
}

func createHandleFetchUSAIndustryOverview(manager MarketDataScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		industries, err := manager.FetchUSAIndustryOverview()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch USA industry overview: %v", err)), nil
		}
		return industryToolResult(newUSAIndustryOverviewResponse(industries))
	}
}

func createHandleFetchUSAIndustryPerformance(manager MarketDataScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		industries, err := manager.FetchUSAIndustryPerformance()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch USA industry performance: %v", err)), nil
		}
		return industryToolResult(newUSAIndustryPerformanceResponse(industries))
	}
}

func createHandleFetchUSAIndustryStocksOverview(manager MarketDataScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		industry, err := request.RequireString("industry")
		if err != nil || industry == "" {
			return mcp.NewToolResultError("Industry parameter is required"), nil
		}
		stocks, err := manager.FetchUSAIndustryStocksOverview(industry)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch USA industry stock overview: %v", err)), nil
		}
		return industryToolResult(newUSAIndustryStocksOverviewResponse(industry, stocks))
	}
}

func createHandleFetchUSAIndustryStocksPerformance(manager MarketDataScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		industry, err := request.RequireString("industry")
		if err != nil || industry == "" {
			return mcp.NewToolResultError("Industry parameter is required"), nil
		}
		stocks, err := manager.FetchUSAIndustryStocksPerformance(industry)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch USA industry stock performance: %v", err)), nil
		}
		return industryToolResult(newUSAIndustryStocksPerformanceResponse(industry, stocks))
	}
}

func createHandleFetchUSAStockUnusualVolumeOverview(manager MarketDataScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stocks, err := manager.FetchUSAStockUnusualVolumeOverview()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch USA unusual-volume stock overview: %v", err)), nil
		}
		return industryToolResult(newUSAStockUnusualVolumeOverviewResponse(stocks))
	}
}

func createHandleFetchUSAStockPreMarketMostActiveOverview(manager MarketDataScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stocks, err := manager.FetchUSAStockPreMarketMostActiveOverview()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch USA pre-market most-active stock overview: %v", err)), nil
		}
		return industryToolResult(newUSAStockPreMarketMostActiveOverviewResponse(stocks))
	}
}

func createHandleFetchETFFundFlowOverview(screen string, fetch func() ([]types.ETFFundFlowOverview, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		etfs, err := fetch()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch ETF %s overview: %v", screen, err)), nil
		}
		return industryToolResult(newETFFundFlowOverviewResponse(screen, etfs))
	}
}

func createHandleFetchETFFundFlowPerformance(screen string, fetch func() ([]types.ETFFundFlowPerformance, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		etfs, err := fetch()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch ETF %s performance: %v", screen, err)), nil
		}
		return industryToolResult(newETFFundFlowPerformanceResponse(screen, etfs))
	}
}

func createHandleFetchETFFundFlows(screen string, fetch func() ([]types.ETFFundFlows, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		etfs, err := fetch()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch ETF %s fund flows: %v", screen, err)), nil
		}
		return industryToolResult(newETFFundFlowsResponse(screen, etfs))
	}
}

func createHandleFetchETFSectorOverview(fetch func() ([]types.ETFSectorOverview, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		etfs, err := fetch()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch sector ETF overview: %v", err)), nil
		}
		return industryToolResult(newETFSectorOverviewResponse(etfs))
	}
}

func createHandleScreenDailyMarketRotation(screener MarketRotationScreener) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		maxCandidates, err := request.RequireInt("max_stock_candidates")
		if err != nil {
			maxCandidates = defaultMaxStockCandidates
		}
		if maxCandidates < 1 || maxCandidates > defaultMaxStockCandidates {
			return mcp.NewToolResultError("max_stock_candidates must be between 1 and 5"), nil
		}
		brief, err := screener.ScreenDailyMarketRotation(MarketRotationOptions{
			PersistHistory:     request.GetBool("persist_history", true),
			MaxStockCandidates: maxCandidates,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to screen daily market rotation: %v", err)), nil
		}
		jsonData, err := json.Marshal(brief)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal market rotation brief: %v", err)), nil
		}
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func industryToolResult(industries any) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(industries, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal industry data: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// createHandleFetchBenchmarkInterestRates creates a handler for fetching benchmark interest rates
func createHandleFetchBenchmarkInterestRates(manager MarketDataManager) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		country, _ := request.RequireString("country")
		if country == "" {
			return mcp.NewToolResultError("Country parameter is required"), nil
		}

		points, err := request.RequireInt("points")
		if err != nil {
			points = 100 // default points
		}

		// Fetch benchmark interest rates
		rates, err := manager.FetchBenchmarkInterestRates(country, points)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch benchmark interest rates: %v", err)), nil
		}

		// Prepare response
		response := map[string]interface{}{
			"total_rates": len(rates),
			"rates":       rates,
			"filter": map[string]interface{}{
				"country": country,
				"points":  points,
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
