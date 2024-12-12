package server

import (
	"context"
	"fmt"
	"net/http"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/internal/reference"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/types"

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

// health check handler
func upcheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/actuator/health" {
		fmt.Fprint(w, "I'm up!")
		return
	}
	http.NotFound(w, r)
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upcheckHandler(w, r.WithContext(ctx))
	})

	// Application handlers registration
	blotter.RegisterHandlers(mux, s.blotter)
	portfolio.RegisterHandlers(mux, s.portfolio)
	if s.portfolio != nil {
		// Register market data service handlers
		mdata.RegisterHandlers(mux, s.portfolio.GetMdataManager())
		reference.RegisterHandlers(mux, s.portfolio.GetRdataManager())
	}

	// Swagger registration
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap mux with loggingMiddleware
	loggedMux := loggingMiddleware(mux, logger)

	logger.Info("Starting server on", fmt.Sprintf("http://%s", s.Addr))
	logger.Info("Swagger UI available at", fmt.Sprintf("http://%s/swagger/index.html", s.Addr))
	return http.ListenAndServe(s.Addr, loggedMux)
}
