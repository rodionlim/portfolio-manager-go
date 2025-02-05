package common

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse represents the error response payload.
type ErrorResponse struct {
	Message string `json:"message"`
}

// SuccessResponse represents the success response payload.
type SuccessResponse struct {
	Message string `json:"message"`
}

func NewHttpRequestWithUserAgent(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	return req, nil
}

// WriteJSONError writes an error message in JSON format to the response.
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}
