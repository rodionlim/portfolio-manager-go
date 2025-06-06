package analytics

import (
	"bytes"
	"encoding/json"
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

func (m *MockService) DownloadLatestNReports(n int, reportType string) ([]string, error) {
	args := m.Called(n, reportType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockService) AnalyzeLatestNReports(n int, reportType string, force bool) ([]*ReportAnalysis, error) {
	args := m.Called(n, reportType, force)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReportAnalysis), args.Error(1)
}

func (m *MockService) FetchAndAnalyzeLatestReportByType(reportType string) (*ReportAnalysis, error) {
	args := m.Called(reportType)
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func (m *MockService) ListReportsInDataDir() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockService) AnalyzeExistingFile(filePath string) (*ReportAnalysis, error) {
	args := m.Called(filePath)
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func (m *MockService) ListAllAnalysis() ([]*ReportAnalysis, error) {
	args := m.Called()
	return args.Get(0).([]*ReportAnalysis), args.Error(1)
}

func TestHandleAnalyzeExistingFile(t *testing.T) {
	// Arrange
	mockService := new(MockService)
	mockAnalysis := &ReportAnalysis{
		Summary:     "File analysis summary",
		KeyInsights: []string{"Key finding"},
		FilePath:    "./data/test.xlsx",
	}

	mockService.On("AnalyzeExistingFile", "./data/test.xlsx").Return(mockAnalysis, nil)

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
