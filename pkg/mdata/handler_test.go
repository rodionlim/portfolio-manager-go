package mdata

import (
	"net/http"
	"net/http/httptest"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/types"
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

func TestHandleDividendsDelete(t *testing.T) {
	// Create a mock market data manager
	mockManager := mocks.NewMockMarketDataManager()
	mockManager.SetDividendMetadata("AAPL", []types.DividendsMetadata{
		{Ticker: "AAPL", ExDate: "2023-01-01", Amount: 1.0, WithholdingTax: 0.3},
	})

	// Create a test request
	req := httptest.NewRequest("DELETE", "/api/v1/mdata/dividends/AAPL?type=Custom", nil)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create the handler
	handler := HandleDividendsDelete(mockManager)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Dividends metadata deleted successfully")

	// Verify deletion
	_, err := mockManager.GetDividendsMetadata("AAPL")
	assert.Error(t, err)
}

func TestHandleDividendsDelete_MissingType(t *testing.T) {
	mockManager := mocks.NewMockMarketDataManager()
	req := httptest.NewRequest("DELETE", "/api/v1/mdata/dividends/AAPL", nil)
	recorder := httptest.NewRecorder()
	handler := HandleDividendsDelete(mockManager)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Dividend type is required")
}

func TestHandleDividendsDelete_InvalidType(t *testing.T) {
	mockManager := mocks.NewMockMarketDataManager()
	req := httptest.NewRequest("DELETE", "/api/v1/mdata/dividends/AAPL?type=Invalid", nil)
	recorder := httptest.NewRecorder()
	handler := HandleDividendsDelete(mockManager)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid dividend type")
}
