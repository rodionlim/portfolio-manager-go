package analytics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Helper function to truncate string for testing
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
