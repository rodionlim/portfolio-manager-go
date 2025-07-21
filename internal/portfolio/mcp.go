package portfolio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterMCPTools registers all portfolio related MCP tools
func RegisterMCPTools(mcpServer *server.MCPServer, portfolio PortfolioGetter) {
	// Get portfolio positions tool
	portfolioTool := mcp.NewTool("get_portfolio_positions",
		mcp.WithDescription("Get current portfolio positions with market values and P&L, if user does not specify book or asks for all books, pass in empty string as the book parameter"),
		mcp.WithString("book",
			mcp.Description("Filter by specific book (optional, default: all books)"),
		),
	)

	mcpServer.AddTool(portfolioTool, createHandleGetPortfolioPositions(portfolio))
}

// createHandleGetPortfolioPositions creates a handler for getting portfolio positions
func createHandleGetPortfolioPositions(portfolio PortfolioGetter) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		book, _ := request.RequireString("book")

		// Get positions for the specified book (or all books if empty)
		positions, err := portfolio.GetAllPositions()
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
}
