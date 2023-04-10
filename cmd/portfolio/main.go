package main

import (
	"log"
	"os"

	"portfolio-manager/internal/config"
	"portfolio-manager/internal/logging"
)

func main() {
	// Load configuration

	config, err := config.NewConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	// Setup logger
	logger, err := logging.InitializeLogger(config.VerboseLogging, config.LogFilePath)
	if err != nil {
		log.Fatalf("Failed to setup logger: %s", err)
	}

	// Log some sample messages
	logger.Info("Starting application with configuration: %v", config)
	logger.Debug("Debugging information")
	logger.Warn("Warning message")

	// Exit
	os.Exit(0)
}
