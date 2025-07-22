package historical

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHistoricalService struct {
	mock.Mock
}

func (m *mockHistoricalService) GetMetrics(book_filter string) ([]TimestampedMetrics, error) {
	args := m.Called(book_filter)
	return args.Get(0).([]TimestampedMetrics), args.Error(1)
}

func (m *mockHistoricalService) GetMetricsByDateRange(book_filter string, start, end time.Time) ([]TimestampedMetrics, error) {
	args := m.Called(book_filter, start, end)
	return args.Get(0).([]TimestampedMetrics), args.Error(1)
}

func TestHandleGetMetrics_Success(t *testing.T) {
	mockSvc := new(mockHistoricalService)
	metrics := []TimestampedMetrics{{}}
	mockSvc.On("GetMetrics", "").Return(metrics, nil)

	req := httptest.NewRequest("GET", "/api/v1/historical/metrics", nil)
	rr := httptest.NewRecorder()
	handler := HandleGetMetrics(mockSvc) // change to accept interface
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandleGetMetrics_Error(t *testing.T) {
	mockSvc := new(mockHistoricalService)
	// Return an empty slice for the first return value, not nil
	mockSvc.On("GetMetrics", "").Return([]TimestampedMetrics{}, assert.AnError)

	req := httptest.NewRequest("GET", "/api/v1/historical/metrics", nil)
	rr := httptest.NewRecorder()
	handler := HandleGetMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandleGetMetrics_EmptyResult(t *testing.T) {
	mockSvc := new(mockHistoricalService)
	// Return an empty slice, not nil
	mockSvc.On("GetMetrics", "").Return([]TimestampedMetrics{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/historical/metrics", nil)
	rr := httptest.NewRecorder()
	handler := HandleGetMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Check that the response is an empty array `[]`, not `null`
	assert.Equal(t, "[]\n", rr.Body.String())
	mockSvc.AssertExpectations(t)
}

// Test for the new trigger metrics endpoint
func TestHandleTriggerMetricsCollection_Success(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)
	mockSvc.On("StoreCurrentMetrics", "tactical").Return(nil)

	reqBody := `{"bookFilter": "tactical"}`
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/trigger", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleTriggerMetricsCollection(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Metrics collection triggered successfully")
	mockSvc.AssertExpectations(t)
}

func TestHandleTriggerMetricsCollection_EntirePortfolio(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)
	mockSvc.On("StoreCurrentMetrics", "").Return(nil)

	reqBody := `{"bookFilter": ""}`
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/trigger", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleTriggerMetricsCollection(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Metrics collection triggered successfully")
	mockSvc.AssertExpectations(t)
}
