package common

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ErrorResponse represents the error response payload.
type ErrorResponse struct {
	Message string `json:"message"`
}

// SuccessResponse represents the success response payload.
type SuccessResponse struct {
	Message string `json:"message"`
}

// userAgents contains a list of different browser user agents to rotate through
var userAgents = []string{
	// Chrome
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",

	// Firefox
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (X11; Linux i686; rv:135.0) Gecko/20100101 Firefox/135.0",

	// Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15",

	// Edge
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/131.0.2903.86",
}

// getRandomUserAgent returns a random user agent from the list
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// RateLimiter implements a simple rate limiter for HTTP requests
type RateLimiter struct {
	lastRequest time.Time
	minInterval time.Duration
	mutex       sync.Mutex
}

// Global rate limiter for Yahoo Finance requests
var yahooRateLimiter = &RateLimiter{
	minInterval: 1200 * time.Millisecond, // Minimum min interval rate limit between requests
}

// GetYahooRateLimiter returns the global Yahoo Finance rate limiter
func GetYahooRateLimiter() *RateLimiter {
	return yahooRateLimiter
}

// SetRateLimitInterval allows updating the rate limit interval
func SetRateLimitInterval(intervalMs int) {
	yahooRateLimiter.mutex.Lock()
	defer yahooRateLimiter.mutex.Unlock()
	yahooRateLimiter.minInterval = time.Duration(intervalMs) * time.Millisecond
}

// Wait blocks until it's safe to make the next request
func (rl *RateLimiter) Wait() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	elapsed := time.Since(rl.lastRequest)
	if elapsed < rl.minInterval {
		time.Sleep(rl.minInterval - elapsed)
	}
	rl.lastRequest = time.Now()
}

func NewHttpRequestWithUserAgent(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", getRandomUserAgent())
	return req, nil
}

// NewBrowserLikeRequest creates an HTTP request with browser-like headers to avoid rate limiting
func NewBrowserLikeRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set comprehensive browser-like headers with randomized user agent
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	// Add referer for Yahoo Finance requests
	if url != "" && (req.URL.Host == "finance.yahoo.com" || req.URL.Host == "query1.finance.yahoo.com") {
		req.Header.Set("Referer", "https://finance.yahoo.com/")
	}

	return req, nil
}

// WriteJSONError writes an error message in JSON format to the response.
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func DoWithRetry(client *http.Client, req *http.Request, maxRetries int, backoff time.Duration, ensureHttpStatusOK bool) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil {
			if ensureHttpStatusOK && resp.StatusCode != http.StatusOK {
				return resp, fmt.Errorf("returned http status code: %d, url: %s", resp.StatusCode, req.URL.String())
			}
			return resp, nil
		}
		time.Sleep(backoff * time.Duration(attempt+1))
	}

	return nil, err
}
