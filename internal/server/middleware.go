package server

import (
	"fmt"
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

		// Log request details
		logger.Info(fmt.Sprintf("Received request: method=%s uri=%s client_ip=%s user_agent=%s", method, uri, clientIP, userAgent))

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log response details
		duration := time.Since(start)
		logger.Info(fmt.Sprintf("Completed request: method=%s uri=%s client_ip=%s duration=%s", method, uri, clientIP, duration))
	})
}
