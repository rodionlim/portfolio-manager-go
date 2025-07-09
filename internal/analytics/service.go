package analytics

import (
	"fmt"
	"os"
	"path/filepath"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
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
		if !strings.HasSuffix(file, ".dat") && !strings.HasSuffix(file, ".DS_Store") && !strings.HasSuffix(file, ".gitignore") { // Exclude .dat and .DS_Store files
			reportFiles = append(reportFiles, file)
		}
	}

	return reportFiles, nil
}

// DownloadLatestNReports downloads the latest N SGX reports of a specific type
func (s *ServiceImpl) DownloadLatestNReports(n int, reportType string) ([]string, error) {
	logger := logging.GetLogger()

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
		logger.Warnf("Requested %d reports, but only %d available. Adjusting to available count.", n, len(filteredReports))
		n = len(filteredReports) // Adjust n if it exceeds available reports
	}

	logger.Infof("Downloading latest %d reports of type '%s'", n, reportType)

	var downloadedFiles []string
	for i := range n {
		filePath, _, _, err := s.downloadReport(filteredReports[i])
		if err != nil {
			return nil, fmt.Errorf("failed to download report %d: %w", i+1, err)
		}
		downloadedFiles = append(downloadedFiles, filePath)
	}

	return downloadedFiles, nil
}

// AnalyzeLatestNReports analyzes the latest N SGX reports and returns their analysis results
func (s *ServiceImpl) AnalyzeLatestNReports(n int, reportType string, forceReanalysis bool) ([]*ReportAnalysis, error) {
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

	var analyses []*ReportAnalysis
	for i := 0; i < n; i++ {
		report := filteredReports[i]

		// First, download the report
		filePath, safeFileName, fileExt, err := s.downloadReport(report)
		if err != nil {
			return nil, fmt.Errorf("failed to download report %d: %w", i+1, err)
		}

		var analysis *ReportAnalysis

		// Check if analysis already exists in database (unless force reanalysis is requested)
		if !forceReanalysis {
			existingAnalysis, err := s.aiAnalyzer.FetchAnalysisByFileName(safeFileName)
			if err == nil && existingAnalysis != nil {
				// Use existing analysis from database
				analysis = existingAnalysis
			}
		}

		// If no existing analysis or force reanalysis is requested, perform new analysis
		if analysis == nil {
			newAnalysis, err := s.analyzeReport(report, filePath, safeFileName, fileExt)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze report %d: %w", i+1, err)
			}
			analysis = newAnalysis
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
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

// ListAndExtractMostTradedStocks filters for SGX Fund Flow reports and extracts the "100 Most Traded Stocks" worksheet
// n - limit results to the latest n reports (0 means no limit)
func (s *ServiceImpl) ListAndExtractMostTradedStocks(n int) ([]*MostTradedStocksReport, error) {
	// Get all files in data directory
	files, err := s.ListReportsInDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	var results []*MostTradedStocksReport

	// Filter for SGX Fund Flow Weekly Tracker files
	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		if strings.Contains(fileName, "SGX_Fund_Flow_Weekly_Tracker") {
			report, err := s.extractMostTradedStocksFromFile(filePath)
			if err != nil {
				// Log error but continue processing other files
				continue
			}
			if report != nil {
				results = append(results, report)
			}
		}
	}

	// Sort reports by report date (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ReportDate > results[j].ReportDate
	})

	// Limit results to latest n reports if n > 0
	if n > 0 && len(results) > n {
		results = results[:n]
	}

	// Calculate changes between consecutive reports
	s.calculateInstitutionNetBuySellChanges(results)

	return results, nil
}

// extractMostTradedStocksFromFile extracts the "100 Most Traded Stocks" data from a single XLSX file
func (s *ServiceImpl) extractMostTradedStocksFromFile(filePath string) (*MostTradedStocksReport, error) {
	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Log the error but don't fail the function
		}
	}()

	// Find the worksheet containing "100 Most Traded Stocks"
	sheetNames := f.GetSheetList()
	var targetSheet string
	for _, sheetName := range sheetNames {
		if strings.Contains(strings.ToLower(sheetName), "100 most traded stocks") {
			targetSheet = sheetName
			break
		}
	}

	if targetSheet == "" {
		return nil, fmt.Errorf("worksheet '100 Most Traded Stocks' not found in file %s", filePath)
	}

	// Get all rows from the target sheet
	rows, err := f.GetRows(targetSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s: %w", targetSheet, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("insufficient data in sheet %s", targetSheet)
	}

	// Extract report date and title from filename
	fileName := filepath.Base(filePath)
	reportDate := extractDateFromSGXFilename(fileName)
	reportTitle := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	report := &MostTradedStocksReport{
		ReportDate:  reportDate,
		ReportTitle: reportTitle,
		FilePath:    filePath,
		Stocks:      []MostTradedStock{},
		ExtractedAt: time.Now().Unix(),
	}

	// Find the header row (contains "Stock Code", "YTD Avg Daily Turnover", etc.)
	headerRowIndex := -1
	for i, row := range rows {
		if len(row) > 1 && strings.Contains(strings.ToLower(strings.Join(row, " ")), "stock code") {
			headerRowIndex = i
			break
		}
	}

	if headerRowIndex == -1 {
		return nil, fmt.Errorf("header row not found in sheet %s", targetSheet)
	}

	// Parse the data rows
	for i := headerRowIndex + 1; i < len(rows); i++ {
		row := rows[i]

		// Skip empty rows or rows with insufficient data
		if len(row) < 6 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		// Skip definition/note rows
		if strings.Contains(strings.ToLower(row[0]), "definition") ||
			strings.Contains(strings.ToLower(row[0]), "note") ||
			strings.Contains(strings.ToLower(row[0]), "all stocks") {
			break
		}

		stock := MostTradedStock{
			StockName: strings.TrimSpace(row[0]),
			Sector:    "",
		}

		// Parse stock code (column 1)
		if len(row) > 1 {
			stock.StockCode = strings.TrimSpace(row[1])
		}

		// Parse YTD Avg Daily Turnover (column 2)
		if len(row) > 2 {
			if val, err := parseFloat(row[2]); err == nil {
				stock.YTDAvgDailyTurnoverSGDM = val
			}
		}

		// Parse YTD Institution Net Buy/Sell (column 3)
		if len(row) > 3 {
			if val, err := parseFloat(row[3]); err == nil {
				stock.YTDInstitutionNetBuySellSGDM = val
			}
		}

		// Parse Past 5 Sessions Institution Net (column 4)
		if len(row) > 4 {
			if val, err := parseFloat(row[4]); err == nil {
				stock.Past5SessionsInstitutionNetSGDM = val
			}
		}

		// Parse Sector (column 5)
		if len(row) > 5 {
			stock.Sector = strings.TrimSpace(row[5])
		}

		// Only add stocks with valid stock codes
		if stock.StockCode != "" {
			report.Stocks = append(report.Stocks, stock)
		}
	}

	return report, nil
}

// extractDateFromSGXFilename extracts date from SGX filename format
// Example: "SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025.xlsx" -> "26 May 2025"
func extractDateFromSGXFilename(filename string) string {
	// Remove extension
	base := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Split by underscore and find the date part
	parts := strings.Split(base, "_")

	// Look for the pattern "Week_of_DD_MMM_YYYY"
	for i := 0; i < len(parts)-3; i++ {
		if parts[i] == "Week" && i+1 < len(parts) && parts[i+1] == "of" {
			if i+4 < len(parts) {
				// Return "DD MMM YYYY" format
				return fmt.Sprintf("%s %s %s", parts[i+2], parts[i+3], parts[i+4])
			}
		}
	}

	return ""
}

// parseFloat parses a string to float64, handling common formatting issues
func parseFloat(s string) (float64, error) {
	// Clean the string
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")  // Remove commas
	s = strings.ReplaceAll(s, "(", "-") // Convert (123) to -123
	s = strings.ReplaceAll(s, ")", "")

	if s == "" || s == "-" {
		return 0, nil
	}

	return strconv.ParseFloat(s, 64)
}

// calculateInstitutionNetBuySellChanges calculates the change in YTDInstitutionNetBuySellSGDM
// between consecutive reports for each stock
func (s *ServiceImpl) calculateInstitutionNetBuySellChanges(reports []*MostTradedStocksReport) {
	if len(reports) < 2 {
		return // Need at least 2 reports to calculate changes
	}

	// Reports are sorted by date descending (newest first)
	// We need to process from oldest to newest to calculate cumulative changes
	for i := len(reports) - 1; i > 0; i-- {
		currentReport := reports[i-1] // newer report
		previousReport := reports[i]  // older report

		// Create a map of stock code to previous values for quick lookup
		previousValues := make(map[string]float64)
		for _, stock := range previousReport.Stocks {
			previousValues[stock.StockCode] = stock.YTDInstitutionNetBuySellSGDM
		}

		// Calculate changes for each stock in the current report
		for j := range currentReport.Stocks {
			stock := &currentReport.Stocks[j]
			if previousValue, exists := previousValues[stock.StockCode]; exists {
				change := stock.YTDInstitutionNetBuySellSGDM - previousValue
				stock.InstitutionNetBuySellChange = &change
			}
			// If stock doesn't exist in previous report, leave change as nil
		}
	}
}

// ListAndExtractSectorFundsFlow filters for SGX Fund Flow reports and extracts the "Institutional" worksheet
// n - limit results to the latest n reports (0 means no limit)
func (s *ServiceImpl) ListAndExtractSectorFundsFlow(n int) ([]*SectorFundsFlowReport, error) {
	// Get all files in data directory
	files, err := s.ListReportsInDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	var results []*SectorFundsFlowReport

	// Filter for SGX Fund Flow Weekly Tracker files
	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		if strings.Contains(fileName, "SGX_Fund_Flow_Weekly_Tracker") {
			report, err := s.extractSectorFundsFlowFromFile(filePath)
			if err != nil {
				// Log error but continue processing other files
				continue
			}
			if report != nil {
				results = append(results, report)
			}
		}
	}

	// Sort reports by report date (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ReportDate > results[j].ReportDate
	})

	// Limit results to latest n reports if n > 0
	if n > 0 && len(results) > n {
		results = results[:n]
	}

	return results, nil
}

// extractSectorFundsFlowFromFile extracts the "Institutional" data from a single XLSX file
func (s *ServiceImpl) extractSectorFundsFlowFromFile(filePath string) (*SectorFundsFlowReport, error) {
	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Log the error but don't fail the function
		}
	}()

	// Find the worksheet containing "Institutional"
	sheetNames := f.GetSheetList()
	var targetSheet string
	for _, sheetName := range sheetNames {
		if strings.Contains(strings.ToLower(sheetName), "institutional") {
			targetSheet = sheetName
			break
		}
	}

	if targetSheet == "" {
		return nil, fmt.Errorf("worksheet 'Institutional' not found in file %s", filePath)
	}

	// Get all rows from the target sheet
	rows, err := f.GetRows(targetSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s: %w", targetSheet, err)
	}

	if len(rows) < 3 {
		return nil, fmt.Errorf("insufficient data in sheet %s", targetSheet)
	}

	// Extract report date and title from filename
	fileName := filepath.Base(filePath)
	reportDate := extractDateFromSGXFilename(fileName)
	reportTitle := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	report := &SectorFundsFlowReport{
		ReportDate:  reportDate,
		ReportTitle: reportTitle,
		FilePath:    filePath,
		SectorFlows: []SectorFlow{},
		ExtractedAt: time.Now().Unix(),
	}

	// Find the header row with sector names
	var sectorHeaders []string
	headerRowIndex := -1

	for i, row := range rows {
		if len(row) > 3 && strings.Contains(strings.ToLower(strings.Join(row, " ")), "consumer cyclicals") {
			headerRowIndex = i
			// Extract sector names from header row (skip first 2 columns: Overall and Date)
			for j := 2; j < len(row); j++ {
				if strings.TrimSpace(row[j]) != "" {
					sectorHeaders = append(sectorHeaders, strings.TrimSpace(row[j]))
				}
			}
			break
		}
	}

	if headerRowIndex == -1 || len(sectorHeaders) == 0 {
		return nil, fmt.Errorf("sector header row not found in sheet %s", targetSheet)
	}

	// Find the last data row before notes/definitions
	var lastDataRow []string
	var weekEndingDate string
	var overallNetBuySell float64

	for i := headerRowIndex + 1; i < len(rows); i++ {
		row := rows[i]

		// Skip empty rows
		if len(row) < 2 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		// Stop if we hit source/definition/note rows
		if strings.Contains(strings.ToLower(row[0]), "source") ||
			strings.Contains(strings.ToLower(row[0]), "definition") ||
			strings.Contains(strings.ToLower(row[0]), "note") ||
			strings.Contains(strings.ToLower(strings.Join(row, " ")), "https://") {
			break
		}

		// This is a data row, keep track of it (we want the last one)
		lastDataRow = row

		// Parse overall net buy/sell (first column)
		if val, err := parseFloat(row[0]); err == nil {
			overallNetBuySell = val
		}

		// Parse date (second column)
		if len(row) > 1 {
			weekEndingDate = strings.TrimSpace(row[1])
		}
	}

	if len(lastDataRow) == 0 {
		return nil, fmt.Errorf("no data rows found in sheet %s", targetSheet)
	}

	// Set the extracted values
	report.WeekEndingDate = weekEndingDate
	report.OverallNetBuySell = overallNetBuySell

	// Parse sector flows from the last data row (starting from column 2)
	for i, sectorName := range sectorHeaders {
		columnIndex := i + 2 // Skip overall and date columns
		if columnIndex < len(lastDataRow) {
			if val, err := parseFloat(lastDataRow[columnIndex]); err == nil {
				sectorFlow := SectorFlow{
					SectorName:     sectorName,
					NetBuySellSGDM: val,
				}
				report.SectorFlows = append(report.SectorFlows, sectorFlow)
			}
		}
	}

	return report, nil
}
