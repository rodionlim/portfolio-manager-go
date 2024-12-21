package common

import (
	"fmt"
	"log"
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
