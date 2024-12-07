package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/config"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/server"

	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
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
	defer logger.CloseLogger()

	// Create context with logger
	ctx := context.WithValue(context.Background(), types.LoggerKey, logger)

	// Log out configurations
	logger.Info("Starting application with configuration:", config)

	// Initialize the database
	var db dal.Database
	switch config.Db {
	case dal.LDB:
		db, err = dal.NewLevelDB("./portfolio-manager.db")
		if err != nil {
			logger.Fatalf("Failed to initialize %s: %s", dal.LDB, err)
		}
	case dal.RDB:
		// Add RocksDB initialization here when implemented
		logger.Fatalf("%s is not yet implemented", dal.RDB)
	default:
		logger.Fatalf("Unsupported database type: %s", config.Db)
	}
	defer db.Close()

	// Create a new blotter service
	blotterSvc := blotter.NewBlotter(db)
	err = blotterSvc.LoadFromDB()
	if err != nil {
		logger.Fatalf("Failed to create blotter service: %s", err)
	}

	// Start the http server to serve requests
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	srv := server.NewServer(addr, blotterSvc)
	if err := srv.Start(ctx); err != nil {
		logger.Error("Failed to start server:", err)
	}

	// Exit
	os.Exit(0)
}
