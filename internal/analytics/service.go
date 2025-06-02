package analytics

import (
	"context"
	"fmt"
	"path/filepath"
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
func NewService(sgxClient SGXClient, aiAnalyzer AIAnalyzer, dataDir string) Service {
	return &ServiceImpl{
		sgxClient:  sgxClient,
		aiAnalyzer: aiAnalyzer,
		dataDir:    dataDir,
	}
}

// FetchLatestReport fetches the latest SGX report and analyzes it
func (s *ServiceImpl) FetchLatestReport(ctx context.Context) (*ReportAnalysis, error) {
	// Fetch reports from SGX
	reports, err := s.sgxClient.FetchReports(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reports: %w", err)
	}

	if len(reports.List.Results) == 0 {
		return nil, fmt.Errorf("no reports found")
	}

	// Find the latest report (they should already be sorted by date)
	latestReport := reports.List.Results[0]

	return s.processReport(ctx, latestReport)
}

// FetchLatestReportByType fetches the latest report of a specific type
func (s *ServiceImpl) FetchLatestReportByType(ctx context.Context, reportType string) (*ReportAnalysis, error) {
	// Fetch reports from SGX
	reports, err := s.sgxClient.FetchReports(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reports: %w", err)
	}

	// Filter reports by type and find the latest one
	var filteredReports []SGXReport
	for _, report := range reports.List.Results {
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

	return s.processReport(ctx, filteredReports[0])
}

// processReport downloads and analyzes a single report
func (s *ServiceImpl) processReport(ctx context.Context, report SGXReport) (*ReportAnalysis, error) {
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

	// Download the file
	if err := s.sgxClient.DownloadFile(ctx, fileURL, filePath); err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// Analyze the file
	fileType := strings.TrimPrefix(fileExt, ".")
	analysis, err := s.aiAnalyzer.AnalyzeDocument(ctx, filePath, fileType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze document: %w", err)
	}

	// Update analysis with report metadata
	analysis.ReportDate = report.Data.ReportDate
	analysis.ReportTitle = report.Data.Title
	analysis.DownloadURL = fileURL
	analysis.FilePath = filePath

	// Extract report type from funds flow type
	if len(report.Data.FundsFlowType) > 0 {
		analysis.ReportType = report.Data.FundsFlowType[0].Data.Data.Name
	}

	// Add metadata
	if analysis.Metadata == nil {
		analysis.Metadata = make(map[string]string)
	}
	analysis.Metadata["file_url"] = fileURL
	analysis.Metadata["file_mime"] = fileMime
	analysis.Metadata["report_name"] = report.Data.Report.Data.Name
	analysis.Metadata["media_type"] = report.Data.Report.Data.MediaType
	analysis.Metadata["download_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())

	return analysis, nil
}

// AnalyzeExistingFile analyzes an existing file
func (s *ServiceImpl) AnalyzeExistingFile(ctx context.Context, filePath string) (*ReportAnalysis, error) {
	// Determine file type from extension
	ext := strings.ToLower(filepath.Ext(filePath))
	fileType := strings.TrimPrefix(ext, ".")

	// Analyze the file
	analysis, err := s.aiAnalyzer.AnalyzeDocument(ctx, filePath, fileType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze existing file: %w", err)
	}

	return analysis, nil
}
