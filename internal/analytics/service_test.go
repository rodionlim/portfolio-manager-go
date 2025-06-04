package analytics

import (
	"context"
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

func (m *MockSGXClient) FetchReports(ctx context.Context) (*SGXReportsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SGXReportsResponse), args.Error(1)
}

func (m *MockSGXClient) DownloadFile(ctx context.Context, url, filePath string) error {
	args := m.Called(ctx, url, filePath)
	return args.Error(0)
}

// MockAIAnalyzer implements AIAnalyzer interface for testing
type MockAIAnalyzer struct {
	mock.Mock
}

func (m *MockAIAnalyzer) AnalyzeDocument(ctx context.Context, filePath string, fileType string) (*ReportAnalysis, error) {
	args := m.Called(ctx, filePath, fileType)
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

func TestServiceImpl_FetchLatestReportByType(t *testing.T) {
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

	ctx := context.Background()
	mockSGXClient.On("FetchReports", ctx).Return(mockReports, nil)
	mockSGXClient.On("DownloadFile", ctx, "https://example.com/report.xlsx", mock.AnythingOfType("string")).Return(nil)
	mockAIAnalyzer.On("AnalyzeDocument", ctx, mock.AnythingOfType("string"), "xlsx").Return(mockAnalysis, nil)
	mockAIAnalyzer.On("FetchAnalysisByFileName", mock.AnythingOfType("string")).Return(nil, fmt.Errorf("not found"))

	// Act
	result, err := service.FetchLatestReportByType(ctx, "fund flow")

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

	ctx := context.Background()
	mockSGXClient.On("FetchReports", ctx).Return(mockReports, nil)

	// Act
	result, err := service.FetchLatestReportByType(ctx, "fund flow")

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

	ctx := context.Background()
	expectedError := fmt.Errorf("SGX API error")
	mockSGXClient.On("FetchReports", ctx).Return((*SGXReportsResponse)(nil), expectedError)

	// Act
	result, err := service.FetchLatestReportByType(ctx, "fund flow")

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

	ctx := context.Background()
	filePath := "./data/existing_report.xlsx"
	mockAIAnalyzer.On("AnalyzeDocument", ctx, filePath, "xlsx").Return(mockAnalysis, nil)

	// Act
	result, err := service.AnalyzeExistingFile(ctx, filePath)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Analysis of existing file", result.Summary)

	mockAIAnalyzer.AssertExpectations(t)
}
