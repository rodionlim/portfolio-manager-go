package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"portfolio-manager/internal/config"
	"portfolio-manager/internal/server"

	"portfolio-manager/pkg/logging"
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

	// Create context with logger
	ctx := context.WithValue(context.Background(), "logger", logger)

	// Log out configurations
	logger.Info("Starting application with configuration:", config)

	// Start the http server to serve requests
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	srv := server.NewServer(addr)
	if err := srv.Start(ctx); err != nil {
		logger.Error("Failed to start server:", err)
	}

	// Exit
	os.Exit(0)
}
