package server

import (
	"context"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/types"

	"github.com/mark3labs/mcp-go/server"
)

// MCPServer represents the MCP server instance
type MCPServer struct {
	server    *server.MCPServer
	blotter   *blotter.TradeBlotter
	portfolio *portfolio.Portfolio
	mdata     mdata.MarketDataManager
	addr      string
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(addr string, blotterSvc *blotter.TradeBlotter, portfolioSvc *portfolio.Portfolio, mdataSvc mdata.MarketDataManager) *MCPServer {
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
		mdata:     mdataSvc,
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
}

// Start starts the MCP server
func (m *MCPServer) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)

	httpServer := server.NewStreamableHTTPServer(m.server)
	logger.Infof("MCP server listening on http://%s/mcp", m.addr)

	return httpServer.Start(m.addr)
}
