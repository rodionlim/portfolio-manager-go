package server

import (
	"context"
	"fmt"
	"net/http"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
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
	blotter.RegisterHandlers(mux, s.blotter)
	portfolio.RegisterHandlers(mux, s.portfolio)

	// Wrap mux with loggingMiddleware
	loggedMux := loggingMiddleware(mux, logger)

	logger.Info("Starting server on", fmt.Sprintf("http://%s", s.Addr))
	return http.ListenAndServe(s.Addr, loggedMux)
}
