package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"portfolio-manager/pkg/logging"
	"time"
)

// loggingMiddleware logs details about the request.
func loggingMiddleware(next http.Handler, logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := r.RemoteAddr
		method := r.Method
		uri := r.RequestURI
		userAgent := r.UserAgent()
		query := r.URL.Query().Encode()

		// Read and restore the body
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Log request details
		logger.Info(fmt.Sprintf("Received request: method=%s uri=%s client_ip=%s user_agent=%s query_params=%s body=%s",
			method, uri, clientIP, userAgent, query, string(bodyBytes)))

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log response details
		duration := time.Since(start)
		logger.Info(fmt.Sprintf("Completed request: method=%s uri=%s client_ip=%s duration=%s", method, uri, clientIP, duration))
	})
}

// corsMiddleware to add CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
