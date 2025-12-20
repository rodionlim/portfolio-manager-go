package testify

import (
	"portfolio-manager/internal/analytics"
	"portfolio-manager/internal/dal"

	"github.com/stretchr/testify/mock"
)

// MockSGXClient is a testify mock for analytics.SGXClient.
type MockSGXClient struct {
	mock.Mock
}

func (m *MockSGXClient) FetchReports() (*analytics.SGXReportsResponse, error) {
	args := m.Called()
	resp, _ := args.Get(0).(*analytics.SGXReportsResponse)
	return resp, args.Error(1)
}

func (m *MockSGXClient) DownloadFile(url, filePath string) error {
	args := m.Called(url, filePath)
	return args.Error(0)
}

// MockAIAnalyzer is a testify mock for analytics.AIAnalyzer.
type MockAIAnalyzer struct {
	mock.Mock
}

func (m *MockAIAnalyzer) AnalyzeDocument(filePath string, fileType string) (*analytics.ReportAnalysis, error) {
	args := m.Called(filePath, fileType)
	analysis, _ := args.Get(0).(*analytics.ReportAnalysis)
	return analysis, args.Error(1)
}

func (m *MockAIAnalyzer) SetDatabase(db dal.Database) {
	m.Called(db)
}

func (m *MockAIAnalyzer) FetchAnalysisByFileName(fileName string) (*analytics.ReportAnalysis, error) {
	args := m.Called(fileName)
	analysis, _ := args.Get(0).(*analytics.ReportAnalysis)
	return analysis, args.Error(1)
}

func (m *MockAIAnalyzer) GetAllAnalysisKeys() ([]string, error) {
	args := m.Called()
	keys, _ := args.Get(0).([]string)
	return keys, args.Error(1)
}
