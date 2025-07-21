package mdata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterMCPTools registers all market data related MCP tools
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
