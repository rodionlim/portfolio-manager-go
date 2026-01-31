package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"portfolio-manager/internal/analytics"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/cli"
	"portfolio-manager/internal/config"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/fxinfer"
	"portfolio-manager/internal/historical"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/internal/migrations"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/internal/server"
	"portfolio-manager/internal/user"

	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/scheduler"
	"portfolio-manager/pkg/types"
)

// @title Portfolio Manager API
// @version 1.0
// @description This is a server for a portfolio manager.

// @host localhost:8080
// @BasePath /

func main() {
	// Check if this is a CLI command
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "backup", "restore-from-backup", "-v", "--version":
			cliHandler := cli.NewCLI()
			if err := cliHandler.ParseAndExecute(os.Args); err != nil {
				log.Fatalf("CLI command failed: %s", err)
			}
			return
		case "-h", "--help", "help":
			cliHandler := cli.NewCLI()
			cliHandler.ShowHelp()
			return
		}
	}

	// Continue with normal server startup
	// Define a command-line flag for the configuration file path
	configFilePath := flag.String("config", "./config.yaml", "Path to the configuration file")
	urlFlag := flag.String("url", "http://localhost:8080", "Base URL for CLI commands")
	flag.Parse()

	// Check if this is a CLI command
	args := flag.Args()
	if len(args) > 0 {
		// This is a CLI command, handle it
		if err := cli.RunCLI(args, *urlFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			cli.PrintUsage()
			os.Exit(1)
		}
		return
	}

	// Continue with normal server startup if no CLI command provided
	startServer(*configFilePath)
}

func startServer(configFilePath string) {

	// Load configuration
	config, err := config.GetOrCreateConfig(configFilePath)
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
	logger.Info("Starting application with configuration:", configFilePath, config)

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

	// Start version check to ensure schema is updated to the latest supported by the application
	migrator := migrations.NewMigrator(db)
	if migrator.Migrate() != nil {
		logger.Fatalf("Failed to migrate schema to the latest version: %s", err)
	}

	// Create a new scheduler
	sched := scheduler.NewScheduler()
	sched.Start(ctx)
	defer sched.Stop()

	// Create a new blotter service
	blotterSvc := blotter.NewBlotter(db)
	err = blotterSvc.LoadFromDB()
	if err != nil {
		logger.Fatalf("Failed to create blotter service: %s", err)
	}

	// Create a new confirmation service
	confirmationSvc := blotter.NewConfirmationService(db)

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

	// Configure market data rate limiting
	if config.MarketData.RateLimitMs > 0 {
		common.SetRateLimitInterval(config.MarketData.RateLimitMs)
		logger.Infof("Set market data rate limit (yahoo mdata source) to %dms between requests", config.MarketData.RateLimitMs)
	}

	// Create a new dividends manager
	dividendsSvc := dividends.NewDividendsManager(db, mdata, rdata, blotterSvc)

	// Create a new FX inference service
	fxInferSvc := fxinfer.NewFXInferenceService(blotterSvc, mdata, rdata, config.BaseCcy)

	// Create a new portfolio service
	portfolioSvc := portfolio.NewPortfolio(db, mdata, rdata, dividendsSvc)
	err = portfolioSvc.LoadPositions()
	if err != nil {
		logger.Fatalf("Failed to create portfolio service: %s", err)
	}
	portfolioSvc.SubscribeToBlotter(blotterSvc)

	// Create a new metrics service
	metricsSvc := metrics.NewMetricsService(blotterSvc, portfolioSvc, dividendsSvc, mdata, rdata)

	// Create analytics service if API key is configured
	var analyticsSvc analytics.Service
	geminiAPIKey := config.Analytics.GeminiAPIKey

	if geminiAPIKey != "" {
		sgxClient := analytics.NewSGXClient()
		aiAnalyzer, err := analytics.NewGeminiAnalyzer(ctx, geminiAPIKey, config.Analytics.GeminiModel)
		if err != nil {
			logger.Error("Failed to create Gemini analyzer, analytics will be disabled:", err)
		} else {
			analyticsSvc = analytics.NewService(sgxClient, aiAnalyzer, config.Analytics.DataDir, db)
			logger.Info("Analytics service initialized with Gemini AI")
		}
	} else {
		logger.Info("Gemini API key not configured, analytics service disabled")
	}

	// Create a new user service
	userSvc := user.NewService(db)

	// Create a new historical metrics service
	historicalSvc := historical.NewService(metricsSvc, analyticsSvc, db, sched, mdata)

	// Start metrics collection schedule if configured
	if config.Metrics.Schedule != "" {
		// Start metrics collection without book filters to collect metrics for aggregated portfolio
		stopFn := historicalSvc.StartMetricsCollection(config.Metrics.Schedule, "")
		defer stopFn()
	}

	customMetricsJob, err := historicalSvc.ListMetricsJobs()
	if err != nil {
		logger.Error("Failed to list custom metrics jobs:", err)
	} else {
		logger.Infof("Custom metrics jobs: %v", customMetricsJob)
		// Start each custom metrics job
		for _, job := range customMetricsJob {
			stopFn := historicalSvc.StartMetricsCollection(job.CronExpr, job.BookFilter)
			defer stopFn()
		}
	}

	// Start analytics schedule if configured and Gemini API key is set
	if config.Analytics.Schedule != "" && config.Analytics.GeminiAPIKey != "" {
		stopFn := historicalSvc.StartSGXReportCollection(config.Analytics.Schedule)
		defer stopFn()
	} else {
		logger.Info("Analytics collection schedule not configured or Gemini API key not set, skipping")
	}

	// Create MCP server if enabled
	var mcpServer *server.MCPServer
	mcpAddr := fmt.Sprintf("%s:%s", config.MCP.Host, config.MCP.Port)
	if config.MCP.Enabled {
		mcpServer = server.NewMCPServer(mcpAddr, blotterSvc, portfolioSvc, mdata)
		logger.Infof("MCP server enabled and will start on %s", mcpAddr)
	} else {
		logger.Info("MCP server disabled")
	}

	// Start the http server to serve requests
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	srv := server.NewServer(addr, blotterSvc, confirmationSvc, portfolioSvc, fxInferSvc, metricsSvc, historicalSvc, analyticsSvc, userSvc, mcpServer)

	if err := srv.Start(ctx); err != nil {
		logger.Error("Failed to start server:", err)
	}

	// Exit
	os.Exit(0)
}
