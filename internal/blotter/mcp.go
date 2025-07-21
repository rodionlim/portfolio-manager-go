package blotter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterMCPTools registers all blotter related MCP tools
func RegisterMCPTools(mcpServer *server.MCPServer, blotter *TradeBlotter) {
	// Query blotter trades tool
	queryBlotterTool := mcp.NewTool("query_blotter_trades",
		mcp.WithDescription("Query blotter trades based on criteria like ticker, date range, or trade type"),
		mcp.WithString("ticker",
			mcp.Description("Filter by ticker symbol (optional)"),
		),
		mcp.WithString("start_date",
			mcp.Description("Start date for filtering trades in YYYY-MM-DD format (optional)"),
		),
		mcp.WithString("end_date",
			mcp.Description("End date for filtering trades in YYYY-MM-DD format (optional)"),
		),
		mcp.WithString("trade_type",
			mcp.Description("Filter by trade type: BUY or SELL (optional)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Limit the number of results returned (default: 100)"),
		),
	)

	mcpServer.AddTool(queryBlotterTool, createHandleQueryBlotterTrades(blotter))
}

// createHandleQueryBlotterTrades creates a handler for querying blotter trades
func createHandleQueryBlotterTrades(blotter *TradeBlotter) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ticker, _ := request.RequireString("ticker")
		startDateStr, _ := request.RequireString("start_date")
		endDateStr, _ := request.RequireString("end_date")
		tradeTypeStr, _ := request.RequireString("trade_type")
		limitStr, _ := request.RequireString("limit")

		limit := 100 // default limit
		if limitStr != "" {
			if parsedLimit, err := json.Number(limitStr).Int64(); err == nil {
				limit = int(parsedLimit)
			}
		}

		// Get all trades from blotter
		trades := blotter.GetTrades()

		// Apply filters
		var filteredTrades []Trade
		for _, trade := range trades {
			// Filter by ticker
			if ticker != "" && trade.Ticker != ticker {
				continue
			}

			// Filter by trade type (Side field in blotter.Trade)
			if tradeTypeStr != "" && trade.Side != tradeTypeStr {
				continue
			}

			// Filter by date range
			if startDateStr != "" {
				startDate, err := time.Parse("2006-01-02", startDateStr)
				if err == nil {
					tradeDate, parseErr := time.Parse("2006-01-02", trade.TradeDate)
					if parseErr == nil && tradeDate.Before(startDate) {
						continue
					}
				}
			}

			if endDateStr != "" {
				endDate, err := time.Parse("2006-01-02", endDateStr)
				if err == nil {
					tradeDate, parseErr := time.Parse("2006-01-02", trade.TradeDate)
					if parseErr == nil && tradeDate.After(endDate) {
						continue
					}
				}
			}

			filteredTrades = append(filteredTrades, trade)

			// Apply limit
			if len(filteredTrades) >= limit {
				break
			}
		}

		// Prepare response
		response := map[string]interface{}{
			"total_trades": len(filteredTrades),
			"trades":       filteredTrades,
			"filters": map[string]interface{}{
				"ticker":     ticker,
				"start_date": startDateStr,
				"end_date":   endDateStr,
				"trade_type": tradeTypeStr,
				"limit":      limit,
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
