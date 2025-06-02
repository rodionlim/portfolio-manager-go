package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService implements Service interface for testing
type MockService struct {
	mock.Mock
}

func (m *MockService) FetchLatestReport(ctx context.Context) (*ReportAnalysis, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func (m *MockService) FetchLatestReportByType(ctx context.Context, reportType string) (*ReportAnalysis, error) {
	args := m.Called(ctx, reportType)
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func (m *MockService) AnalyzeExistingFile(ctx context.Context, filePath string) (*ReportAnalysis, error) {
	args := m.Called(ctx, filePath)
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func TestHandleGetLatestReport(t *testing.T) {
	// Arrange
	mockService := new(MockService)
	mockAnalysis := &ReportAnalysis{
		Summary:     "Test summary",
		KeyInsights: []string{"Insight 1", "Insight 2"},
		ReportTitle: "Test Report",
	}

	mockService.On("FetchLatestReport", mock.Anything).Return(mockAnalysis, nil)

	// Create handler
	handler := HandleGetLatestReport(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/latest", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ReportAnalysis
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Test summary", response.Summary)
	assert.Equal(t, "Test Report", response.ReportTitle)

	mockService.AssertExpectations(t)
}

func TestHandleAnalyzeExistingFile(t *testing.T) {
	// Arrange
	mockService := new(MockService)
	mockAnalysis := &ReportAnalysis{
		Summary:     "File analysis summary",
		KeyInsights: []string{"Key finding"},
		FilePath:    "./data/test.xlsx",
	}

	mockService.On("AnalyzeExistingFile", mock.Anything, "./data/test.xlsx").Return(mockAnalysis, nil)

	// Create handler
	handler := HandleAnalyzeExistingFile(mockService)

	// Create request body
	requestBody := AnalyzeFileRequest{
		FilePath: "./data/test.xlsx",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/analyze", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ReportAnalysis
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "File analysis summary", response.Summary)
	assert.Equal(t, "./data/test.xlsx", response.FilePath)

	mockService.AssertExpectations(t)
}

func TestHandleAnalyzeExistingFile_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockService)

	// Create handler
	handler := HandleAnalyzeExistingFile(mockService)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/analyze", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetLatestReport_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockService)
	expectedError := fmt.Errorf("service error")

	mockService.On("FetchLatestReport", mock.Anything).Return((*ReportAnalysis)(nil), expectedError)

	// Create handler
	handler := HandleGetLatestReport(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/latest", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}
