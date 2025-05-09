package common

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

func IsFutureDate(date string) bool {
	// date is in the format yyyy-mm-dd
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		log.Println("Error parsing date:", err)
		return false
	}
	return parsedDate.After(time.Now())
}

// ConvertDateFormat converts a date string from one format to another.
func ConvertDateFormat(dateStr, fromFormat, toFormat string) (string, error) {
	date, err := time.Parse(fromFormat, dateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse date: %w", err)
	}
	return date.Format(toFormat), nil
}

// ParseDateToEpoch converts a date string in YYYYMMDD format to an epoch timestamp.
// The timestamp represents the start of the day (00:00:00) in UTC.
func ParseDateToEpoch(dateStr string) (int64, error) {
	if len(dateStr) != 8 {
		return 0, fmt.Errorf("date must be in YYYYMMDD format")
	}

	year, err := strconv.Atoi(dateStr[0:4])
	if err != nil {
		return 0, err
	}

	month, err := strconv.Atoi(dateStr[4:6])
	if err != nil || month < 1 || month > 12 {
		return 0, fmt.Errorf("invalid month")
	}

	day, err := strconv.Atoi(dateStr[6:8])
	if err != nil || day < 1 || day > 31 {
		return 0, fmt.Errorf("invalid day")
	}

	// Create date with time at start of day (00:00:00)
	return GetEpochFromDate(year, month, day), nil
}

// GetEpochFromDate creates an epoch timestamp for the given year, month, and day.
// The time component is set to 00:00:00 UTC.
func GetEpochFromDate(year, month, day int) int64 {
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return date.Unix()
}

// GetCurrentEpochTime returns the current time as an epoch timestamp
func GetCurrentEpochTime() int64 {
	return time.Now().Unix()
}

// Helper function to check if a date string is in RFC3339 format
func IsRFC3339Format(dateStr string) bool {
	_, err := time.Parse(time.RFC3339, dateStr)
	return err == nil
}

// Helper function to parse dates in various formats
func ParseFlexibleDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",          // YYYY-MM-DD
		"2006/01/02",          // YYYY/MM/DD
		"02-01-2006",          // DD-MM-YYYY
		"02/01/2006",          // DD/MM/YYYY
		"2006-01-02 15:04:05", // YYYY-MM-DD HH:MM:SS
		"2006/01/02 15:04:05", // YYYY/MM/DD HH:MM:SS
	}

	for _, format := range formats {
		if parsedDate, err := time.Parse(format, dateStr); err == nil {
			return parsedDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date string: %s", dateStr)
}
