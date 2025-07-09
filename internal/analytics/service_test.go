package analytics

import (
	"fmt"
	"testing"
	"time"

	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSGXClient implements SGXClient interface for testing
type MockSGXClient struct {
	mock.Mock
}

func (m *MockSGXClient) FetchReports() (*SGXReportsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SGXReportsResponse), args.Error(1)
}

func (m *MockSGXClient) DownloadFile(url, filePath string) error {
	args := m.Called(url, filePath)
	return args.Error(0)
}

// MockAIAnalyzer implements AIAnalyzer interface for testing
type MockAIAnalyzer struct {
	mock.Mock
}

func (m *MockAIAnalyzer) AnalyzeDocument(filePath string, fileType string) (*ReportAnalysis, error) {
	args := m.Called(filePath, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func (m *MockAIAnalyzer) FetchAnalysisByFileName(fileName string) (*ReportAnalysis, error) {
	args := m.Called(fileName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReportAnalysis), args.Error(1)
}

func (m *MockAIAnalyzer) SetDatabase(db dal.Database) {
	m.Called(db)
}

func (m *MockAIAnalyzer) GetAllAnalysisKeys() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestServiceImpl_FetchAndAnalyzeLatestReportByType(t *testing.T) {
	// Arrange
	mockSGXClient := new(MockSGXClient)
	mockAIAnalyzer := new(MockAIAnalyzer)
	mockDB := mocks.NewMockDatabase()

	// Set up the SetDatabase expectation
	mockAIAnalyzer.On("SetDatabase", mockDB).Return()

	service := NewService(mockSGXClient, mockAIAnalyzer, "./data", mockDB)

	mockReports := &SGXReportsResponse{
		Data: struct {
			ReportTypes struct {
				Count   int `json:"count"`
				Results []struct {
					Data struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				} `json:"results"`
			} `json:"reportTypes"`
			List struct {
				Count   int         `json:"count"`
				Results []SGXReport `json:"results"`
			} `json:"list"`
		}{
			ReportTypes: struct {
				Count   int `json:"count"`
				Results []struct {
					Data struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				} `json:"results"`
			}{
				Count: 1,
			},
			List: struct {
				Count   int         `json:"count"`
				Results []SGXReport `json:"results"`
			}{
				Count: 1,
				Results: []SGXReport{
					{
						Data: struct {
							Title      string `json:"title"`
							ReportDate int64  `json:"reportDate"`
							Report     struct {
								Data struct {
									MediaType string `json:"mediaType"`
									Name      string `json:"name"`
									Date      int64  `json:"date"`
									File      struct {
										Data struct {
											URL      string `json:"url"`
											FileMime string `json:"filemime"`
										} `json:"data"`
									} `json:"file"`
								} `json:"data"`
							} `json:"report"`
							FundsFlowType []struct {
								Data struct {
									Data struct {
										ID           string  `json:"id"`
										Name         string  `json:"name"`
										Order        string  `json:"order"`
										ParentCode   *string `json:"parentCode"`
										EntityBundle string  `json:"entityBundle"`
									} `json:"data"`
								} `json:"data"`
							} `json:"fundsFlowType"`
						}{
							Title:      "Test Report",
							ReportDate: time.Now().Unix(),
							Report: struct {
								Data struct {
									MediaType string `json:"mediaType"`
									Name      string `json:"name"`
									Date      int64  `json:"date"`
									File      struct {
										Data struct {
											URL      string `json:"url"`
											FileMime string `json:"filemime"`
										} `json:"data"`
									} `json:"file"`
								} `json:"data"`
							}{
								Data: struct {
									MediaType string `json:"mediaType"`
									Name      string `json:"name"`
									Date      int64  `json:"date"`
									File      struct {
										Data struct {
											URL      string `json:"url"`
											FileMime string `json:"filemime"`
										} `json:"data"`
									} `json:"file"`
								}{
									MediaType: "application/xlsx",
									Name:      "Fund Flow Report",
									Date:      time.Now().Unix(),
									File: struct {
										Data struct {
											URL      string `json:"url"`
											FileMime string `json:"filemime"`
										} `json:"data"`
									}{
										Data: struct {
											URL      string `json:"url"`
											FileMime string `json:"filemime"`
										}{
											URL:      "https://example.com/report.xlsx",
											FileMime: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
										},
									},
								},
							},
							FundsFlowType: []struct {
								Data struct {
									Data struct {
										ID           string  `json:"id"`
										Name         string  `json:"name"`
										Order        string  `json:"order"`
										ParentCode   *string `json:"parentCode"`
										EntityBundle string  `json:"entityBundle"`
									} `json:"data"`
								} `json:"data"`
							}{
								{
									Data: struct {
										Data struct {
											ID           string  `json:"id"`
											Name         string  `json:"name"`
											Order        string  `json:"order"`
											ParentCode   *string `json:"parentCode"`
											EntityBundle string  `json:"entityBundle"`
										} `json:"data"`
									}{
										Data: struct {
											ID           string  `json:"id"`
											Name         string  `json:"name"`
											Order        string  `json:"order"`
											ParentCode   *string `json:"parentCode"`
											EntityBundle string  `json:"entityBundle"`
										}{
											ID:   "203",
											Name: "Fund Flow Tracker",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mockAnalysis := &ReportAnalysis{
		Summary:     "Test analysis summary",
		KeyInsights: []string{"Insight 1", "Insight 2"},
	}

	mockSGXClient.On("FetchReports").Return(mockReports, nil)
	mockSGXClient.On("DownloadFile", "https://example.com/report.xlsx", mock.AnythingOfType("string")).Return(nil)
	mockAIAnalyzer.On("AnalyzeDocument", mock.AnythingOfType("string"), "xlsx").Return(mockAnalysis, nil)
	mockAIAnalyzer.On("FetchAnalysisByFileName", mock.AnythingOfType("string")).Return(nil, fmt.Errorf("not found"))

	// Act
	result, err := service.FetchAndAnalyzeLatestReportByType("fund flow")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test analysis summary", result.Summary)
	assert.Equal(t, "Test Report", result.ReportTitle)
	assert.Equal(t, "Fund Flow Tracker", result.ReportType)

	mockSGXClient.AssertExpectations(t)
	mockAIAnalyzer.AssertExpectations(t)
}

func TestServiceImpl_FetchLatestReport_NoReports(t *testing.T) {
	// Arrange
	mockSGXClient := new(MockSGXClient)
	mockAIAnalyzer := new(MockAIAnalyzer)
	mockDB := mocks.NewMockDatabase()

	// Set up the SetDatabase expectation
	mockAIAnalyzer.On("SetDatabase", mockDB).Return()

	service := NewService(mockSGXClient, mockAIAnalyzer, "./data", mockDB)

	mockReports := &SGXReportsResponse{
		Data: struct {
			ReportTypes struct {
				Count   int `json:"count"`
				Results []struct {
					Data struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				} `json:"results"`
			} `json:"reportTypes"`
			List struct {
				Count   int         `json:"count"`
				Results []SGXReport `json:"results"`
			} `json:"list"`
		}{
			List: struct {
				Count   int         `json:"count"`
				Results []SGXReport `json:"results"`
			}{
				Count:   0,
				Results: []SGXReport{},
			},
		},
	}

	mockSGXClient.On("FetchReports").Return(mockReports, nil)

	// Act
	result, err := service.FetchAndAnalyzeLatestReportByType("fund flow")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no reports found")

	mockSGXClient.AssertExpectations(t)
}

func TestServiceImpl_FetchLatestReport_SGXClientError(t *testing.T) {
	// Arrange
	mockSGXClient := new(MockSGXClient)
	mockAIAnalyzer := new(MockAIAnalyzer)
	mockDB := mocks.NewMockDatabase()

	// Set up the SetDatabase expectation
	mockAIAnalyzer.On("SetDatabase", mockDB).Return()

	service := NewService(mockSGXClient, mockAIAnalyzer, "./data", mockDB)

	expectedError := fmt.Errorf("SGX API error")
	mockSGXClient.On("FetchReports").Return((*SGXReportsResponse)(nil), expectedError)

	// Act
	result, err := service.FetchAndAnalyzeLatestReportByType("fund flow")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch reports")

	mockSGXClient.AssertExpectations(t)
}

func TestServiceImpl_AnalyzeExistingFile(t *testing.T) {
	// Arrange
	mockSGXClient := new(MockSGXClient)
	mockAIAnalyzer := new(MockAIAnalyzer)
	mockDB := mocks.NewMockDatabase()

	// Set up the SetDatabase expectation
	mockAIAnalyzer.On("SetDatabase", mockDB).Return()

	service := NewService(mockSGXClient, mockAIAnalyzer, "./data", mockDB)

	mockAnalysis := &ReportAnalysis{
		Summary:     "Analysis of existing file",
		KeyInsights: []string{"Key finding 1", "Key finding 2"},
	}

	filePath := "./data/existing_report.xlsx"
	mockAIAnalyzer.On("AnalyzeDocument", filePath, "xlsx").Return(mockAnalysis, nil)

	// Act
	result, err := service.AnalyzeExistingFile(filePath)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Analysis of existing file", result.Summary)

	mockAIAnalyzer.AssertExpectations(t)
}

func TestServiceImpl_ListAllAnalysis(t *testing.T) {
	// Arrange
	mockSGXClient := new(MockSGXClient)
	mockAIAnalyzer := new(MockAIAnalyzer)
	mockDB := mocks.NewMockDatabase()

	// Set up the SetDatabase expectation
	mockAIAnalyzer.On("SetDatabase", mockDB).Return()

	service := NewService(mockSGXClient, mockAIAnalyzer, "./data", mockDB)

	// Mock analysis keys and data
	mockKeys := []string{
		"ANALYTICS_SUMMARY:SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025.xlsx",
		"ANALYTICS_SUMMARY:SGX_Fund_Flow_Weekly_Tracker_Week_of_19_May_2025.xlsx",
	}

	mockAnalysis1 := &ReportAnalysis{
		Summary:     "Analysis 1",
		KeyInsights: []string{"Insight 1"},
		ReportTitle: "Fund Flow Week of 26 May 2025",
	}

	mockAnalysis2 := &ReportAnalysis{
		Summary:     "Analysis 2",
		KeyInsights: []string{"Insight 2"},
		ReportTitle: "Fund Flow Week of 19 May 2025",
	}

	// Set up expectations
	mockAIAnalyzer.On("GetAllAnalysisKeys").Return(mockKeys, nil)
	mockAIAnalyzer.On("FetchAnalysisByFileName", "SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025.xlsx").Return(mockAnalysis1, nil)
	mockAIAnalyzer.On("FetchAnalysisByFileName", "SGX_Fund_Flow_Weekly_Tracker_Week_of_19_May_2025.xlsx").Return(mockAnalysis2, nil)

	// Act
	result, err := service.ListAllAnalysis()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Analysis 1", result[0].Summary)
	assert.Equal(t, "Analysis 2", result[1].Summary)

	mockAIAnalyzer.AssertExpectations(t)
}

func TestServiceImpl_ExtractSectorFundsFlowFromFile(t *testing.T) {
	// Arrange
	mockSGXClient := new(MockSGXClient)
	mockAIAnalyzer := new(MockAIAnalyzer)
	mockDB := mocks.NewMockDatabase()

	// Set up the SetDatabase expectation
	mockAIAnalyzer.On("SetDatabase", mockDB).Return()

	service := NewService(mockSGXClient, mockAIAnalyzer, "./../../data", mockDB).(*ServiceImpl)

	testFilePath := "./../../data/SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025.xlsx"

	// Act
	report, err := service.extractSectorFundsFlowFromFile(testFilePath)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, report)

	// Verify basic report information
	assert.Equal(t, "26 May 2025", report.ReportDate)
	assert.Equal(t, "SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025", report.ReportTitle)
	assert.Equal(t, testFilePath, report.FilePath)
	assert.NotEmpty(t, report.WeekEndingDate)
	assert.NotZero(t, report.ExtractedAt)

	// Verify sector flows
	assert.NotEmpty(t, report.SectorFlows)
	assert.Equal(t, len(report.SectorFlows), 12, "Should have 12 sectors")

	// Verify each sector flow has required fields
	for _, flow := range report.SectorFlows {
		assert.NotEmpty(t, flow.SectorName, "Sector name should not be empty")
		assert.IsType(t, float64(0), flow.NetBuySellSGDM, "Net buy/sell should be a float64")
	}

	// Test specific extraction - the overall net buy/sell should be a reasonable number
	assert.IsType(t, float64(0), report.OverallNetBuySell)

	assert.Equal(t, "26 May 2025", report.ReportDate, "Report date should match the file name")
	assert.Equal(t, "26-May-25", report.WeekEndingDate, "Week ending date should match the file name")
	assert.Equal(t, -18.8, report.SectorFlows[0].NetBuySellSGDM, "Consumer Cyclicals net buy/sell should match expected value")
	assert.Equal(t, "Consumer Cyclicals", report.SectorFlows[0].SectorName, "First sector should be Consumer Cyclicals")
	assert.Equal(t, 13.7, report.SectorFlows[len(report.SectorFlows)-1].NetBuySellSGDM, "Last sector net buy/sell should match expected value")
	assert.Equal(t, "Utilities", report.SectorFlows[len(report.SectorFlows)-1].SectorName, "Last sector should be Utilities")

	mockAIAnalyzer.AssertExpectations(t)
}
