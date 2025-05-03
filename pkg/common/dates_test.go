package common

import (
	"testing"
	"time"
)

func TestParseFlexibleDate(t *testing.T) {
	// Test with date string "2025-02-18" (YYYY-MM-DD format)
	dateStr := "2025-02-18"

	// Expected result: February 18, 2025 at 00:00:00
	expectedYear := 2025
	expectedMonth := time.February
	expectedDay := 18

	// Call the function
	parsedDate, err := ParseFlexibleDate(dateStr)

	// Check for errors
	if err != nil {
		t.Errorf("ParseFlexibleDate failed with error: %v", err)
	}

	// Check if the parsed date matches our expectations
	if parsedDate.Year() != expectedYear {
		t.Errorf("Expected year: %d, got: %d", expectedYear, parsedDate.Year())
	}

	if parsedDate.Month() != expectedMonth {
		t.Errorf("Expected month: %s, got: %s", expectedMonth, parsedDate.Month())
	}

	if parsedDate.Day() != expectedDay {
		t.Errorf("Expected day: %d, got: %d", expectedDay, parsedDate.Day())
	}

	// Also verify the time is zeroed out
	if parsedDate.Hour() != 0 || parsedDate.Minute() != 0 || parsedDate.Second() != 0 {
		t.Errorf("Expected time to be 00:00:00, got: %02d:%02d:%02d",
			parsedDate.Hour(), parsedDate.Minute(), parsedDate.Second())
	}

	// Test if the function returns the result in the correct format
	formatted := parsedDate.Format(time.RFC3339)
	expectedFormatted := "2025-02-18T00:00:00Z" // RFC3339 format with time zeroed in UTC

	if formatted != expectedFormatted {
		t.Errorf("Expected formatted date: %s, got: %s", expectedFormatted, formatted)
	}
}

func TestParseFlexibleDateWithTime(t *testing.T) {
	// Test with date string that includes time (YYYY-MM-DD HH:MM:SS format)
	dateStr := "2025-02-18 09:00:00"

	// Expected result: February 18, 2025 at 09:00:00
	expectedYear := 2025
	expectedMonth := time.February
	expectedDay := 18
	expectedHour := 9
	expectedMinute := 0
	expectedSecond := 0

	// Call the function
	parsedDate, err := ParseFlexibleDate(dateStr)

	// Check for errors
	if err != nil {
		t.Errorf("ParseFlexibleDate failed with error: %v", err)
	}

	// Check if the parsed date matches our expectations
	if parsedDate.Year() != expectedYear {
		t.Errorf("Expected year: %d, got: %d", expectedYear, parsedDate.Year())
	}

	if parsedDate.Month() != expectedMonth {
		t.Errorf("Expected month: %s, got: %s", expectedMonth, parsedDate.Month())
	}

	if parsedDate.Day() != expectedDay {
		t.Errorf("Expected day: %d, got: %d", expectedDay, parsedDate.Day())
	}

	// Verify the time component is correctly parsed
	if parsedDate.Hour() != expectedHour {
		t.Errorf("Expected hour: %d, got: %d", expectedHour, parsedDate.Hour())
	}

	if parsedDate.Minute() != expectedMinute {
		t.Errorf("Expected minute: %d, got: %d", expectedMinute, parsedDate.Minute())
	}

	if parsedDate.Second() != expectedSecond {
		t.Errorf("Expected second: %d, got: %d", expectedSecond, parsedDate.Second())
	}

	// Test if the function returns the result in the correct format
	formatted := parsedDate.Format(time.RFC3339)
	expectedFormatted := "2025-02-18T09:00:00Z" // RFC3339 format with time in UTC

	if formatted != expectedFormatted {
		t.Errorf("Expected formatted date: %s, got: %s", expectedFormatted, formatted)
	}
}
