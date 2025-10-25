package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Wait(t *testing.T) {
	// Create a rate limiter with 100ms interval
	rl := &RateLimiter{
		minInterval: 100 * time.Millisecond,
	}

	// First request should not block
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 50*time.Millisecond, "First request should not be delayed")

	// Second request should be delayed by at least the interval
	start = time.Now()
	rl.Wait()
	elapsed = time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond, "Second request should be delayed by at least the interval")
	assert.Less(t, elapsed, 150*time.Millisecond, "Delay should not be excessive")
}

func TestSetRateLimitInterval(t *testing.T) {
	// Test setting a new rate limit interval
	originalInterval := yahooRateLimiter.minInterval
	defer func() {
		yahooRateLimiter.minInterval = originalInterval
	}()

	SetRateLimitInterval(200)
	assert.Equal(t, 200*time.Millisecond, yahooRateLimiter.minInterval)

	SetRateLimitInterval(1000)
	assert.Equal(t, 1000*time.Millisecond, yahooRateLimiter.minInterval)
}

func TestNewBrowserLikeRequest(t *testing.T) {
	req, err := NewBrowserLikeRequest("GET", "https://query1.finance.yahoo.com/v8/finance/chart/AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, req)

	// Check that browser-like headers are set
	assert.Equal(t, "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7", req.Header.Get("Accept"))
	assert.Equal(t, "en-US,en;q=0.9", req.Header.Get("Accept-Language"))
	assert.Equal(t, "gzip, deflate, br", req.Header.Get("Accept-Encoding"))
	assert.Equal(t, "1", req.Header.Get("DNT"))
	assert.Equal(t, "keep-alive", req.Header.Get("Connection"))
	assert.Equal(t, "https://finance.yahoo.com/", req.Header.Get("Referer"))
}

func TestNewBrowserLikeRequest_NonYahoo(t *testing.T) {
	req, err := NewBrowserLikeRequest("GET", "https://example.com/api")
	assert.NoError(t, err)
	assert.NotNil(t, req)

	// Should not have Yahoo-specific referer for non-Yahoo URLs
	assert.Empty(t, req.Header.Get("Referer"))
}
