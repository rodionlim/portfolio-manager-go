package server

import (
	"context"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/mark3labs/mcp-go/server"
)

const mcpInstructions = "Routing rule: distinguish stock-market industries from sector ETFs. Stock industries and their underlying stocks provide overview and performance only; they do not provide fund flows. For raw requests mentioning sector fund flows, capital flows, inflows, or outflows, use the sector ETF fund-flow tool unless the user explicitly rejects ETFs. For a daily market brief, rotation screen, or ranked sector-to-stock drill-down, use screen_daily_market_rotation and preserve its precomputed order and classifications without recalculating. Describe its flows as ETF product flows, never direct stock-level institutional flows. If the user asks for underlying-stock fund flows, explain that this dataset is unavailable. If raw dataset intent remains unclear, ask whether they mean stock-market industries or sector ETFs. Fund-flow periods are 1M, 3M, 1Y, 3Y, and YTD."

// MCPServer represents the MCP server instance
type MCPServer struct {
	server    *server.MCPServer
	blotter   *blotter.TradeBlotter
	portfolio *portfolio.Portfolio
	mdata     mdata.MarketDataManager
	screener  mdata.MarketDataScreener
	rdata     rdata.ReferenceManager
	addr      string
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(addr string, blotterSvc *blotter.TradeBlotter, portfolioSvc *portfolio.Portfolio, mdataSvc mdata.MarketDataManager, screenerSvc mdata.MarketDataScreener, rdataSvc rdata.ReferenceManager) *MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Portfolio Manager MCP 📊",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithInstructions(mcpInstructions),
		server.WithRecovery(),
	)

	mcpServer := &MCPServer{
		server:    s,
		blotter:   blotterSvc,
		portfolio: portfolioSvc,
		mdata:     mdataSvc,
		screener:  screenerSvc,
		rdata:     rdataSvc,
		addr:      addr,
	}

	// Register tools
	mcpServer.registerTools()

	return mcpServer
}

// registerTools registers all available MCP tools
func (m *MCPServer) registerTools() {
	// Register blotter tools
	blotter.RegisterMCPTools(m.server, m.blotter)

	// Register portfolio tools
	portfolio.RegisterMCPTools(m.server, m.portfolio)

	// Register market data tools
	mdata.RegisterMCPTools(m.server, m.mdata)
	if m.screener != nil {
		mdata.RegisterScreenerMCPTools(m.server, m.screener)
	}

	// Register reference data tools
	if m.rdata != nil {
		rdata.RegisterMCPTools(m.server, m.rdata)
	}
}

// Start starts the MCP server
func (m *MCPServer) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)

	httpServer := server.NewStreamableHTTPServer(m.server)
	logger.Infof("MCP server listening on http://%s/mcp", m.addr)

	return httpServer.Start(m.addr)
}
