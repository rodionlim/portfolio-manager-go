package common

import "github.com/google/uuid"

// Generate a unique ID string
func GenerateTradeID() string {
	return uuid.New().String()
}
