package server

import (
	"context"
	"fmt"
	"net/http"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
)

// Server represents the HTTP server.
type Server struct {
	Addr    string
	Blotter *blotter.TradeBlotter
}

// NewServer creates a new Server instance.
func NewServer(addr string, blotterSvc *blotter.TradeBlotter) *Server {
	return &Server{
		Addr:    addr,
		Blotter: blotterSvc,
	}
}

// health check handler
func upcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "I'm up!")
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	logger := ctx.Value(types.LoggerKey).(*logging.Logger)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upcheckHandler(w, r.WithContext(ctx))
	})

	// Wrap mux with loggingMiddleware
	loggedMux := loggingMiddleware(mux, logger)

	logger.Info("Starting server on", fmt.Sprintf("http://%s", s.Addr))
	return http.ListenAndServe(s.Addr, loggedMux)
}
