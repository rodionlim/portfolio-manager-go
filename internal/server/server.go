package server

import (
	"context"
	"fmt"
	"net/http"

	"portfolio-manager/pkg/logging"
)

// Server represents the HTTP server.
type Server struct {
	Addr string
}

// NewServer creates a new Server instance.
func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

// health check handler
func upcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "I'm up!")
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	logger := ctx.Value("logger").(*logging.Logger)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upcheckHandler(w, r.WithContext(ctx))
	})

	// Wrap mux with loggingMiddleware
	loggedMux := loggingMiddleware(mux, logger)

	logger.Info("Starting server on", fmt.Sprintf("http://%s", s.Addr))
	return http.ListenAndServe(s.Addr, loggedMux)
}
