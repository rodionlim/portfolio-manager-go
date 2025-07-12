package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer represents the MCP server instance
type MCPServer struct {
	server    *server.MCPServer
	blotter   *blotter.TradeBlotter
	portfolio *portfolio.Portfolio
	addr      string
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(addr string, blotterSvc *blotter.TradeBlotter, portfolioSvc *portfolio.Portfolio) *MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Portfolio Manager MCP 📊",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	mcpServer := &MCPServer{
		server:    s,
		blotter:   blotterSvc,
		portfolio: portfolioSvc,
		addr:      addr,
	}

	// Register tools
	mcpServer.registerTools()

	return mcpServer
}

// registerTools registers all available MCP tools
func (m *MCPServer) registerTools() {
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

	m.server.AddTool(queryBlotterTool, m.handleQueryBlotterTrades)

	// Get portfolio positions tool
	portfolioTool := mcp.NewTool("get_portfolio_positions",
		mcp.WithDescription("Get current portfolio positions with market values and P&L, if user does not specify book or asks for all books, pass in empty string as the book parameter"),
		mcp.WithString("book",
			mcp.Description("Filter by specific book (optional, default: all books)"),
		),
	)

	m.server.AddTool(portfolioTool, m.handleGetPortfolioPositions)
}

// handleQueryBlotterTrades handles the query blotter trades tool
func (m *MCPServer) handleQueryBlotterTrades(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	trades := m.blotter.GetTrades()

	// Apply filters
	var filteredTrades []blotter.Trade
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

// handleGetPortfolioPositions handles the get portfolio positions tool
func (m *MCPServer) handleGetPortfolioPositions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	book, _ := request.RequireString("book")

	// Get positions for the specified book (or all books if empty)
	positions, err := m.portfolio.GetAllPositions()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get positions: %v", err)), nil
	}

	// Prepare response
	response := map[string]interface{}{
		"total_positions": len(positions),
		"positions":       positions,
		"filter": map[string]interface{}{
			"book": book,
		},
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// Start starts the MCP server
func (m *MCPServer) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)

	httpServer := server.NewStreamableHTTPServer(m.server)
	logger.Infof("MCP server listening on http://%s/mcp", m.addr)

	return httpServer.Start(m.addr)
}
