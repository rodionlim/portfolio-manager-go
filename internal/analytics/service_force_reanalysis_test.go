package analytics

import (
	"os"
	"path/filepath"
	"testing"

	"portfolio-manager/internal/mocks/testify"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnalyzeLatestNReports_ForceReanalysisBypassesCachedAnalysis(t *testing.T) {
	dataDir := t.TempDir()

	mockSGX := new(testify.MockSGXClient)
	mockAI := new(testify.MockAIAnalyzer)

	// NewService always sets db on analyzer; we don't need a real db in this test.
	mockAI.On("SetDatabase", mock.Anything).Return()

	report := SGXReport{}
	report.Data.Title = "SGX_Fund_Flow_Weekly_Tracker_Week_of_6_October_2025"
	report.Data.ReportDate = 1
	report.Data.Report.Data.File.Data.URL = "https://example.com/report.xlsx"
	report.Data.Report.Data.File.Data.FileMime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	report.Data.FundsFlowType = []struct {
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
		{},
	}

	resp := &SGXReportsResponse{}
	resp.Data.List.Results = []SGXReport{report}

	mockSGX.On("FetchReports").Return(resp, nil)

	// When downloading, create a dummy file at the provided path.
	mockSGX.On("DownloadFile", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		filePath := args.String(1)
		_ = os.MkdirAll(filepath.Dir(filePath), 0o755)
		_ = os.WriteFile(filePath, []byte("dummy"), 0o644)
	})

	cached := &ReportAnalysis{Summary: "cached"}

	// Without forcing, the service should return cached analysis and NOT call AnalyzeDocument.
	mockAI.On("FetchAnalysisByFileName", mock.Anything).Return(cached, nil).Once()

	svc := NewService(mockSGX, mockAI, dataDir, nil)
	analyses, err := svc.AnalyzeLatestNReports(1, "", false)
	assert.NoError(t, err)
	if assert.Len(t, analyses, 1) {
		assert.Equal(t, "cached", analyses[0].Summary)
	}

	mockAI.AssertNotCalled(t, "AnalyzeDocument", mock.Anything, mock.Anything)

	// With forcing, the service should bypass cache and call AnalyzeDocument.
	fresh := &ReportAnalysis{Summary: "fresh"}
	mockAI.On("AnalyzeDocument", mock.Anything, mock.Anything).Return(fresh, nil).Once()

	analyses, err = svc.AnalyzeLatestNReports(1, "", true)
	assert.NoError(t, err)
	if assert.Len(t, analyses, 1) {
		assert.Equal(t, "fresh", analyses[0].Summary)
	}

	mockAI.AssertExpectations(t)
	mockSGX.AssertExpectations(t)
}
