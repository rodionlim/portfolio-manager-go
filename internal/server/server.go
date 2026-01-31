package server

import (
	"context"
	"fmt"
	"net/http"

	"portfolio-manager/internal/analytics"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/fxinfer"
	"portfolio-manager/internal/historical"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/internal/user"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"portfolio-manager/web/ui"

	_ "portfolio-manager/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Server represents the HTTP server.
type Server struct {
	Addr         string
	mux          *http.ServeMux
	blotter      *blotter.TradeBlotter
	confirmation *blotter.ConfirmationService
	portfolio    *portfolio.Portfolio
	fxinfer      *fxinfer.Service
	metrics      *metrics.MetricsService
	historical   *historical.Service
	analytics    analytics.Service
	user         *user.Service
	mcpServer    *MCPServer
}

// NewServer creates a new Server instance.
func NewServer(addr string, blotterSvc *blotter.TradeBlotter, confirmationSvc *blotter.ConfirmationService, portfolioSvc *portfolio.Portfolio, fxinferSvc *fxinfer.Service, metricsSvc *metrics.MetricsService, historicalSvc *historical.Service, analyticsSvc analytics.Service, userSvc *user.Service, mcpSvc *MCPServer) *Server {
	return &Server{
		Addr:         addr,
		mux:          http.NewServeMux(),
		blotter:      blotterSvc,
		confirmation: confirmationSvc,
		portfolio:    portfolioSvc,
		fxinfer:      fxinferSvc,
		metrics:      metricsSvc,
		historical:   historicalSvc,
		analytics:    analyticsSvc,
		user:         userSvc,
		mcpServer:    mcpSvc,
	}
}

// @Summary Health check
// @Description Returns a simple message to indicate that the server is up and running
// @Tags health
// @Produce plain
// @Success 200 {string} string "I'm up!"
// @Router /healthz [get]
func upcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I'm up!")
}

// Start starts the HTTP server and optionally the MCP server.
func (s *Server) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)

	// Serve embed assets. If the build tag builtinassets is set,
	// ui.AssetsHandler() will serve files; otherwise it will return a dummy handler.
	s.mux.Handle("/", ui.AssetsHandler())

	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		upcheckHandler(w, r.WithContext(ctx))
	})

	// Application handlers registration
	blotter.RegisterHandlers(s.mux, s.blotter)
	blotter.RegisterConfirmationHandlers(s.mux, s.confirmation)
	portfolio.RegisterHandlers(s.mux, s.portfolio)
	if s.portfolio != nil {
		// Register market data service handlers
		mdata.RegisterHandlers(s.mux, s.portfolio.GetMdataManager())
		rdata.RegisterHandlers(s.mux, s.portfolio.GetRdataManager())
		dividends.RegisterHandlers(s.mux, s.portfolio.GetDividendsManager())
	}
	fxinfer.RegisterHandlers(s.mux, s.fxinfer)
	metrics.RegisterHandlers(s.mux, s.metrics)
	if s.historical != nil {
		historical.RegisterHandlers(s.mux, s.historical)
	}
	if s.analytics != nil {
		analytics.RegisterHandlers(s.mux, s.analytics)
	}
	if s.user != nil {
		user.RegisterHandlers(s.mux, s.user)
	}

	// Swagger registration
	s.mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap mux with loggingMiddleware and corsMiddleware
	loggedCorsMux := loggingMiddleware(corsMiddleware(s.mux), logger)

	logger.Info("Starting server on", fmt.Sprintf("http://%s", s.Addr))
	logger.Info("Swagger UI available at", fmt.Sprintf("http://%s/swagger/index.html", s.Addr))

	// Start MCP server in a goroutine if it's provided
	if s.mcpServer != nil {
		go func() {
			if err := s.mcpServer.Start(ctx); err != nil {
				logger.Error("MCP server error:", err)
			}
		}()
	}

	// Start HTTP server (blocking call)
	return http.ListenAndServe(s.Addr, loggedCorsMux)
}
