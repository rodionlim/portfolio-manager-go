package mdata

import (
	"net/http"
	"net/http/httptest"
	"portfolio-manager/internal/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleInterestRatesGet(t *testing.T) {
	// Create a mock market data manager
	mockManager := mocks.NewMockMarketDataManager()

	// Create a test request
	req := httptest.NewRequest("GET", "/api/v1/mdata/interest-rates/SG?points=10", nil)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create the handler
	handler := HandleInterestRatesGet(mockManager)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "SORA")
}

func TestHandleInterestRatesGet_MissingCountry(t *testing.T) {
	// Create a mock market data manager
	mockManager := mocks.NewMockMarketDataManager()

	// Create a test request without country
	req := httptest.NewRequest("GET", "/api/v1/mdata/interest-rates/", nil)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create the handler
	handler := HandleInterestRatesGet(mockManager)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Country is required")
}

func TestHandleInterestRatesGet_InvalidPoints(t *testing.T) {
	// Create a mock market data manager
	mockManager := mocks.NewMockMarketDataManager()

	// Create a test request with invalid points
	req := httptest.NewRequest("GET", "/api/v1/mdata/interest-rates/SG?points=invalid", nil)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create the handler
	handler := HandleInterestRatesGet(mockManager)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid points parameter")
}

func TestHandleInterestRatesGet_DefaultPoints(t *testing.T) {
	// Create a mock market data manager
	mockManager := mocks.NewMockMarketDataManager()

	// Create a test request without points (should default to 250)
	req := httptest.NewRequest("GET", "/api/v1/mdata/interest-rates/SG", nil)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create the handler
	handler := HandleInterestRatesGet(mockManager)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "SORA")
}
