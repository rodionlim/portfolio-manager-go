package server

import (
	"context"
	"fmt"
	"net/http"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/portfolio"
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
	Addr      string
	blotter   *blotter.TradeBlotter
	portfolio *portfolio.Portfolio
}

// NewServer creates a new Server instance.
func NewServer(addr string, blotterSvc *blotter.TradeBlotter, portfolioSvc *portfolio.Portfolio) *Server {
	return &Server{
		Addr:      addr,
		blotter:   blotterSvc,
		portfolio: portfolioSvc,
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

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)

	mux := http.NewServeMux()

	// Serve embed assets. If the build tag builtinassets is set,
	// ui.AssetsHandler() will serve files; otherwise it will return a dummy handler.
	mux.Handle("/", ui.AssetsHandler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		upcheckHandler(w, r.WithContext(ctx))
	})

	// Application handlers registration
	blotter.RegisterHandlers(mux, s.blotter)
	portfolio.RegisterHandlers(mux, s.portfolio)
	if s.portfolio != nil {
		// Register market data service handlers
		mdata.RegisterHandlers(mux, s.portfolio.GetMdataManager())
		rdata.RegisterHandlers(mux, s.portfolio.GetRdataManager())
		dividends.RegisterHandlers(mux, s.portfolio.GetDividendsManager())
	}

	// Swagger registration
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap mux with loggingMiddleware and corsMiddleware
	loggedCorsMux := loggingMiddleware(corsMiddleware(mux), logger)

	logger.Info("Starting server on", fmt.Sprintf("http://%s", s.Addr))
	logger.Info("Swagger UI available at", fmt.Sprintf("http://%s/swagger/index.html", s.Addr))
	return http.ListenAndServe(s.Addr, loggedCorsMux)
}
