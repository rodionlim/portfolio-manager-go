package analytics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestConvertXLSXToText(t *testing.T) {
	// Test with the sample XLSX file
	filePath := "../../data/SGX_Fund_Flow_Weekly_Tracker_Week_of_26_May_2025.xlsx"

	result, err := convertXLSXToText(filePath)

	require.NoError(t, err, "XLSX conversion should not fail")
	require.NotEmpty(t, result, "Converted text should not be empty")

	// Check that the result contains sheet separators (indicating multi-sheet support)
	assert.Contains(t, result, "SHEET:", "Should contain sheet headers")
	assert.Contains(t, result, "====", "Should contain sheet separators")

	// Count the number of sheets
	sheetCount := strings.Count(result, "SHEET:")
	t.Logf("Number of sheets found: %d", sheetCount)

	// Print the first 2500 characters to see more structure
	t.Logf("XLSX Conversion Result (first 2500 chars):\n%s", truncateString(result, 2500))

	// Basic checks for CSV-like structure
	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 5, "Should have multiple lines of data")
}

// TestSGXFundFlow6October2025 tests the specific file with double underscore in name
// Note: October 2025+ files have a different format with only 3 sheets (Weekly Top 10, Institutional, Retail)
// compared to earlier 2025 files that had 5 sheets (including 100 Most Traded Stocks, Sector Funds Flow)
func TestSGXFundFlow6October2025(t *testing.T) {
	filePath := "../../data/SGX_Fund_Flow_Weekly_Tracker_Week_of__6_October_2025.xlsx"

	t.Run("OpenFile", func(t *testing.T) {
		f, err := excelize.OpenFile(filePath)
		require.NoError(t, err, "Should be able to open the Excel file with double underscore in filename")
		defer f.Close()

		sheets := f.GetSheetList()
		t.Logf("Sheets found: %v", sheets)
		assert.Len(t, sheets, 3, "October 2025 files have 3 sheets (new format)")
	})

	t.Run("VerifyNewFormatSheets", func(t *testing.T) {
		f, err := excelize.OpenFile(filePath)
		require.NoError(t, err)
		defer f.Close()

		sheets := f.GetSheetList()
		expectedSheets := []string{"Weekly Top 10", "Institutional", "Retail"}
		for _, expected := range expectedSheets {
			found := false
			for _, sheet := range sheets {
				if strings.EqualFold(sheet, expected) {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find '%s' sheet in new format", expected)
		}
	})

	t.Run("FindWeeklyTop10Sheet", func(t *testing.T) {
		f, err := excelize.OpenFile(filePath)
		require.NoError(t, err)
		defer f.Close()

		sheets := f.GetSheetList()
		var targetSheet string
		for _, sheetName := range sheets {
			if strings.Contains(strings.ToLower(sheetName), "weekly top 10") {
				targetSheet = sheetName
				break
			}
		}
		assert.NotEmpty(t, targetSheet, "Should find 'Weekly Top 10' worksheet")
		t.Logf("Found target sheet: %s", targetSheet)
	})

	t.Run("MostTradedStocksNotInNewFormat", func(t *testing.T) {
		f, err := excelize.OpenFile(filePath)
		require.NoError(t, err)
		defer f.Close()

		sheets := f.GetSheetList()
		var targetSheet string
		for _, sheetName := range sheets {
			if strings.Contains(strings.ToLower(sheetName), "100 most traded stocks") {
				targetSheet = sheetName
				break
			}
		}
		// This is expected NOT to be found in October 2025+ files
		assert.Empty(t, targetSheet, "'100 Most Traded Stocks' worksheet is not present in October 2025+ format")
		t.Log("As expected: '100 Most Traded Stocks' sheet not found in new format")
	})

	t.Run("SectorFundsFlowNotInNewFormat", func(t *testing.T) {
		f, err := excelize.OpenFile(filePath)
		require.NoError(t, err)
		defer f.Close()

		sheets := f.GetSheetList()
		var targetSheet string
		for _, sheetName := range sheets {
			if strings.Contains(strings.ToLower(sheetName), "sector funds flow") {
				targetSheet = sheetName
				break
			}
		}
		// This is expected NOT to be found in October 2025+ files
		assert.Empty(t, targetSheet, "'Sector Funds Flow' worksheet is not present in October 2025+ format")
		t.Log("As expected: 'Sector Funds Flow' sheet not found in new format")
	})

	t.Run("ReadWeeklyTop10Data", func(t *testing.T) {
		f, err := excelize.OpenFile(filePath)
		require.NoError(t, err)
		defer f.Close()

		targetSheet := "Weekly Top 10"
		rows, err := f.GetRows(targetSheet)
		require.NoError(t, err, "Should be able to read rows from Weekly Top 10")
		assert.Greater(t, len(rows), 5, "Should have data rows")

		t.Logf("Total rows in Weekly Top 10: %d", len(rows))
		for i := 0; i < min(5, len(rows)); i++ {
			t.Logf("Row %d: %v", i+1, rows[i])
		}
	})

	t.Run("ConvertXLSXToText", func(t *testing.T) {
		result, err := convertXLSXToText(filePath)
		require.NoError(t, err, "XLSX conversion should not fail")
		require.NotEmpty(t, result, "Converted text should not be empty")

		// Check for expected sheets
		assert.Contains(t, result, "SHEET:", "Should contain sheet headers")

		sheetCount := strings.Count(result, "SHEET:")
		assert.Equal(t, 3, sheetCount, "Should have 3 sheets in new format")
		t.Logf("Number of sheets found: %d", sheetCount)
	})

	t.Run("ExtractDateFromFilename", func(t *testing.T) {
		fileName := "SGX_Fund_Flow_Weekly_Tracker_Week_of__6_October_2025.xlsx"
		date := extractDateFromSGXFilename(fileName)
		t.Logf("Extracted date: '%s'", date)

		// Should correctly handle double underscore and normalize full month name to abbreviated
		assert.Equal(t, "6 Oct 2025", date, "Should extract date with abbreviated month name")
		assert.NotEmpty(t, date, "Should extract date from filename")
		assert.Contains(t, date, "Oct", "Should contain abbreviated month name")
		assert.Contains(t, date, "6", "Should contain day number")
	})
}

// Helper function to truncate string for testing
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
