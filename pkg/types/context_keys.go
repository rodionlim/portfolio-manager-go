package types

// Define unexported types for context keys
type contextKey string

// Define context keys
const (
	LoggerKey contextKey = "logger"
)
