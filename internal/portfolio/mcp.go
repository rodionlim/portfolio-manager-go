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

	// Delete single position tool (requires explicit confirmation from user)
	deleteTool := mcp.NewTool("delete_portfolio_position",
		mcp.WithDescription("Delete a single position by book and ticker. This is destructive and will wipe the position from the database. ALWAYS ask the user to confirm they want to proceed before calling this tool."),
		mcp.WithString("book", mcp.Description("Book name of the position to delete (required)"), mcp.Required()),
		mcp.WithString("ticker", mcp.Description("Ticker symbol of the position to delete (required)"), mcp.Required()),
		mcp.WithString("confirm", mcp.Description("Must be set to 'yes' to actually perform the deletion. If not 'yes', the tool will return a prompt requesting confirmation.")),
	)

	mcpServer.AddTool(deleteTool, createHandleDeletePortfolioPosition(portfolio))
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
		response := map[string]any{
			"total_positions": len(positions),
			"positions":       positions,
			"filter": map[string]any{
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

// createHandleDeletePortfolioPosition creates a handler for deleting a portfolio position with confirmation
func createHandleDeletePortfolioPosition(portfolio PortfolioGetter) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		book, err := request.RequireString("book")
		if err != nil || book == "" {
			return mcp.NewToolResultError("missing required parameter: book"), nil
		}
		ticker, err := request.RequireString("ticker")
		if err != nil || ticker == "" {
			return mcp.NewToolResultError("missing required parameter: ticker"), nil
		}

		confirm, _ := request.RequireString("confirm")
		if confirm != "yes" {
			// Return an instructional message asking user to confirm
			msg := fmt.Sprintf("You are requesting to delete position (book=%s, ticker=%s). This is DESTRUCTIVE and cannot be undone. If you wish to proceed, call the tool again with confirm='yes'.", book, ticker)
			return mcp.NewToolResultText(msg), nil
		}

		if err := portfolio.DeletePosition(book, ticker); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to delete position: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted position (book=%s, ticker=%s)", book, ticker)), nil
	}
}
