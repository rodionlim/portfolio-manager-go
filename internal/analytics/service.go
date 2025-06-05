package analytics

import (
	"fmt"
	"os"
	"path/filepath"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/types"
	"sort"
	"strings"
	"time"
)

// ServiceImpl implements the analytics Service interface
type ServiceImpl struct {
	sgxClient  SGXClient
	aiAnalyzer AIAnalyzer
	dataDir    string
}

// NewService creates a new analytics service
func NewService(sgxClient SGXClient, aiAnalyzer AIAnalyzer, dataDir string, db dal.Database) Service {
	// Set the database on the AI analyzer
	aiAnalyzer.SetDatabase(db)

	return &ServiceImpl{
		sgxClient:  sgxClient,
		aiAnalyzer: aiAnalyzer,
		dataDir:    dataDir,
	}
}

// FetchReports fetches all available SGX reports and lists them
func (s *ServiceImpl) ListReportsInDataDir() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(s.dataDir, "*"))
	if err != nil {
		return nil, err
	}

	var reportFiles []string
	for _, file := range files {
		if !strings.HasSuffix(file, ".dat") && !strings.HasSuffix(file, ".DS_Store") { // Exclude .dat and .DS_Store files
			reportFiles = append(reportFiles, file)
		}
	}

	return reportFiles, nil
}

// DownloadLatestNReports downloads the latest N SGX reports of a specific type
func (s *ServiceImpl) DownloadLatestNReports(n int, reportType string) ([]string, error) {
	// Fetch reports from SGX
	reports, err := s.sgxClient.FetchReports()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reports: %w", err)
	}

	var filteredReports []SGXReport

	// If reportType is provided, filter by type; otherwise include all reports
	if reportType != "" {
		// Filter reports by type
		for _, report := range reports.Data.List.Results {
			for _, flowType := range report.Data.FundsFlowType {
				if strings.Contains(strings.ToLower(flowType.Data.Data.Name), strings.ToLower(reportType)) {
					filteredReports = append(filteredReports, report)
					break
				}
			}
		}
	} else {
		// Include all reports if no type filter is specified
		filteredReports = reports.Data.List.Results
	}

	// Sort reports by report date (descending)
	sort.Slice(filteredReports, func(i, j int) bool {
		return filteredReports[i].Data.ReportDate > filteredReports[j].Data.ReportDate
	})

	if n > len(filteredReports) {
		n = len(filteredReports) // Adjust n if it exceeds available reports
	}

	var downloadedFiles []string
	for i := 0; i < n; i++ {
		filePath, _, _, err := s.downloadReport(filteredReports[i])
		if err != nil {
			return nil, fmt.Errorf("failed to download report %d: %w", i+1, err)
		}
		downloadedFiles = append(downloadedFiles, filePath)
	}

	return downloadedFiles, nil
}

// FetchAndAnalyzeLatestReportByType fetches the latest report of a specific type and analyzes it
func (s *ServiceImpl) FetchAndAnalyzeLatestReportByType(reportType string) (*ReportAnalysis, error) {
	// Fetch reports from SGX
	reports, err := s.sgxClient.FetchReports()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reports: %w", err)
	}

	// Filter reports by type and find the latest one
	var filteredReports []SGXReport
	for _, report := range reports.Data.List.Results {
		for _, flowType := range report.Data.FundsFlowType {
			if strings.Contains(strings.ToLower(flowType.Data.Data.Name), strings.ToLower(reportType)) {
				filteredReports = append(filteredReports, report)
				break
			}
		}
	}

	if len(filteredReports) == 0 {
		return nil, fmt.Errorf("no reports found for type: %s", reportType)
	}

	// Sort by report date (descending)
	sort.Slice(filteredReports, func(i, j int) bool {
		return filteredReports[i].Data.ReportDate > filteredReports[j].Data.ReportDate
	})

	filePath, safeFileName, fileExt, err := s.downloadReport(filteredReports[0])
	if err != nil {
		return nil, fmt.Errorf("failed to download report: %w", err)
	}

	return s.analyzeReport(filteredReports[0], filePath, safeFileName, fileExt)
}

// downloadReport downloads a report and returns its file path, name and extension
func (s *ServiceImpl) downloadReport(report SGXReport) (string, string, string, error) {
	// Extract file information
	fileURL := report.Data.Report.Data.File.Data.URL
	fileMime := report.Data.Report.Data.File.Data.FileMime

	// Determine file extension
	var fileExt string
	switch fileMime {
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		fileExt = ".xlsx"
	case "application/pdf":
		fileExt = ".pdf"
	case "text/csv":
		fileExt = ".csv"
	default:
		fileExt = ".dat"
	}

	// Generate safe filename
	safeFileName := generateSafeFileName(report.Data.Title, fileExt)
	filePath := filepath.Join(s.dataDir, safeFileName)

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return filePath, safeFileName, fileExt, nil // Return existing file path if it exists
	}

	// Download the file
	if err := s.sgxClient.DownloadFile(fileURL, filePath); err != nil {
		return "", "", "", fmt.Errorf("failed to download file: %w", err)
	}

	return filePath, safeFileName, fileExt, nil
}

// analyzeReport analyzes a single on disk report
func (s *ServiceImpl) analyzeReport(report SGXReport, filePath, safeFileName, fileExt string) (*ReportAnalysis, error) {

	// Check if analysis has already been done
	analysis, err := s.aiAnalyzer.FetchAnalysisByFileName(safeFileName)
	if err == nil && analysis != nil {
		return analysis, nil // Return existing analysis if available
	}

	// Analyze the file
	fileType := strings.TrimPrefix(fileExt, ".")
	analysis, err = s.aiAnalyzer.AnalyzeDocument(filePath, fileType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze document: %w", err)
	}

	// Update analysis with report metadata
	analysis.ReportDate = report.Data.ReportDate
	analysis.ReportTitle = report.Data.Title
	analysis.FilePath = filePath

	// Extract report type from funds flow type
	if len(report.Data.FundsFlowType) > 0 {
		analysis.ReportType = report.Data.FundsFlowType[0].Data.Data.Name
	}

	// Add metadata
	if analysis.Metadata == nil {
		analysis.Metadata = make(map[string]string)
	}
	analysis.Metadata["report_name"] = report.Data.Report.Data.Name
	analysis.Metadata["media_type"] = report.Data.Report.Data.MediaType
	analysis.Metadata["download_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())

	return analysis, nil
}

// AnalyzeExistingFile analyzes an existing file
func (s *ServiceImpl) AnalyzeExistingFile(filePath string) (*ReportAnalysis, error) {
	// Determine file type from extension
	ext := strings.ToLower(filepath.Ext(filePath))
	fileType := strings.TrimPrefix(ext, ".")

	// Analyze the file
	analysis, err := s.aiAnalyzer.AnalyzeDocument(filePath, fileType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze existing file: %w", err)
	}

	return analysis, nil
}

// ListAllAnalysis lists all available analysis reports that were previously stored in database
func (s *ServiceImpl) ListAllAnalysis() ([]*ReportAnalysis, error) {
	// Get all keys with the analytics summary prefix
	keys, err := s.aiAnalyzer.GetAllAnalysisKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis keys: %w", err)
	}

	var analyses []*ReportAnalysis
	for _, key := range keys {
		// Extract filename from key (remove prefix)
		prefix := fmt.Sprintf("%s:", types.AnalyticsSummaryKeyPrefix)
		fileName := strings.TrimPrefix(key, prefix)

		analysis, err := s.aiAnalyzer.FetchAnalysisByFileName(fileName)
		if err != nil {
			// Log error but continue with other analyses
			continue
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}
