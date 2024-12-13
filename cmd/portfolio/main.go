package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/config"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/internal/server"

	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
)

// @title Portfolio Manager API
// @version 1.0
// @description This is a server for a portfolio manager.

// @host localhost:8080
// @BasePath /

func main() {
	// Define a command-line flag for the configuration file path
	configFilePath := flag.String("config", "./config.yaml", "Path to the configuration file")
	flag.Parse()

	// Load configuration
	config, err := config.GetOrCreateConfig(*configFilePath)
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
	logger.Info("Starting application with configuration:", *configFilePath, config)

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

	// Create a new reference data manager
	rdata, err := rdata.NewManager(db, config.RefDataSeedPath)
	if err != nil {
		logging.GetLogger().Fatalf("Failed to create reference data manager")
	}

	// Create a new market data manager
	mdata, err := mdata.NewManager(db, rdata)
	if err != nil {
		logging.GetLogger().Fatalf("Failed to create market data manager")
	}

	// Create a new dividends manager
	dividendsSvc := dividends.NewDividendsManager(db, mdata, rdata, blotterSvc)

	// Create a new portfolio service
	portfolioSvc := portfolio.NewPortfolio(db, mdata, rdata, dividendsSvc)
	err = portfolioSvc.LoadPositions()
	if err != nil {
		logger.Fatalf("Failed to create portfolio service: %s", err)
	}
	portfolioSvc.SubscribeToBlotter(blotterSvc)

	// Start the http server to serve requests
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	srv := server.NewServer(addr, blotterSvc, portfolioSvc)

	if err := srv.Start(ctx); err != nil {
		logger.Error("Failed to start server:", err)
	}

	// Exit
	os.Exit(0)
}
