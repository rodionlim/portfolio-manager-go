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

func (m *MockService) ListAndExtractMostTradedStocks(n int) ([]*MostTradedStocksReport, error) {
	args := m.Called(n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*MostTradedStocksReport), args.Error(1)
}

func (m *MockService) ListAndExtractTop10Stocks(n int) ([]*Top10WeeklyReport, error) {
	args := m.Called(n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Top10WeeklyReport), args.Error(1)
}

func (m *MockService) ListAndExtractSectorFundsFlow(n int) ([]*SectorFundsFlowReport, error) {
	args := m.Called(n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*SectorFundsFlowReport), args.Error(1)
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

func TestHandleListAndExtractSectorFundsFlow(t *testing.T) {
	// Arrange
	mockService := new(MockService)

	// Mock data
	mockReport := &SectorFundsFlowReport{
		ReportDate:        "26 May 2025",
		ReportTitle:       "SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025",
		FilePath:          "./data/test.xlsx",
		WeekEndingDate:    "26-May-25",
		OverallNetBuySell: 40.6,
		SectorFlows: []SectorFlow{
			{SectorName: "Financial Services", NetBuySellSGDM: 103.1},
			{SectorName: "Industrials", NetBuySellSGDM: 42.9},
		},
		ExtractedAt: 1625097600,
	}

	// Set up expectations
	mockService.On("ListAndExtractSectorFundsFlow", 1).Return([]*SectorFundsFlowReport{mockReport}, nil)

	// Create handler
	handler := HandleListAndExtractSectorFundsFlow(mockService)

	// Create test request
	req, err := http.NewRequest("GET", "/api/v1/analytics/sector_funds_flow?n=1", nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Verify response contains valid JSON
	var response []SectorFundsFlowReport
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	// Verify structure of first report
	report := response[0]
	assert.Equal(t, "26 May 2025", report.ReportDate)
	assert.Equal(t, "SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025", report.ReportTitle)
	assert.NotEmpty(t, report.SectorFlows)
	assert.Equal(t, 2, len(report.SectorFlows))

	mockService.AssertExpectations(t)
}

func TestHandleListAndExtractSectorFundsFlow_InvalidParam(t *testing.T) {
	// Arrange
	mockService := new(MockService)

	// Create handler
	handler := HandleListAndExtractSectorFundsFlow(mockService)

	// Create test request with invalid n parameter
	req, err := http.NewRequest("GET", "/api/v1/analytics/sector_funds_flow?n=invalid", nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	mockService.AssertExpectations(t)
}
