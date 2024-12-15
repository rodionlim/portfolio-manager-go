package common

import (
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
