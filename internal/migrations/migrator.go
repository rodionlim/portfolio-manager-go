package migrations

import (
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"strconv"
	"strings"
)

// Migrator handles database schema migrations
type Migrator struct {
	db                   dal.Database
	lastMigrationApplied string
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db dal.Database) *Migrator {
	migrator := &Migrator{
		db:                   db,
		lastMigrationApplied: "",
	}

	// Load the last migration applied from database
	migrator.loadLastMigrationFromDB()

	return migrator
}

// Migrate runs all necessary migrations for the current version
func (m *Migrator) Migrate() error {
	logging.GetLogger().Info("Starting database migrations")

	currentVersion := m.lastMigrationApplied
	logging.GetLogger().Infof("Current migration version: %s", currentVersion)

	// Check if v1.7.0 migration should be run
	if m.shouldRunMigration("1.7.0", currentVersion) {
		if err := m.migrateV170(); err != nil {
			return fmt.Errorf("failed to run v1.7.0 migrations: %w", err)
		}
	} else {
		logging.GetLogger().Info("Skipping v1.7.0 migration (already applied or version is higher)")
	}

	// Subsequent migrations should be added here as needed

	logging.GetLogger().Infof("Database migrations completed successfully. Last migration applied: %s", m.lastMigrationApplied)
	return nil
}

// GetLastMigrationApplied returns the version of the last migration that was applied
func (m *Migrator) GetLastMigrationApplied() string {
	return m.lastMigrationApplied
}

// loadLastMigrationFromDB loads the last migration applied from database
func (m *Migrator) loadLastMigrationFromDB() {
	migrationKey := fmt.Sprintf("%s:last_applied", types.MigrationKeyPrefix)
	var lastMigration string
	err := m.db.Get(migrationKey, &lastMigration)
	if err == nil {
		m.lastMigrationApplied = lastMigration
		logging.GetLogger().Infof("Loaded last migration from database: %s", lastMigration)
	} else {
		m.lastMigrationApplied = ""
		logging.GetLogger().Info("No previous migration found in database")
	}
}

// saveLastMigrationToDB saves the last migration applied to database
func (m *Migrator) saveLastMigrationToDB() error {
	migrationKey := fmt.Sprintf("%s:last_applied", types.MigrationKeyPrefix)
	return m.db.Put(migrationKey, m.lastMigrationApplied)
}

// shouldRunMigration determines if a migration should be run based on version comparison
func (m *Migrator) shouldRunMigration(targetVersion, currentVersion string) bool {
	// If no previous migration, run the migration
	if currentVersion == "" {
		return true
	}

	// Compare versions
	return compareVersions(targetVersion, currentVersion) > 0
}

// compareVersions compares two semantic version strings (major.minor.patch)
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if v1 == v2
func compareVersions(v1, v2 string) int {
	v1Parts := parseVersion(v1)
	v2Parts := parseVersion(v2)

	for i := 0; i < 3; i++ {
		if v1Parts[i] > v2Parts[i] {
			return 1
		}
		if v1Parts[i] < v2Parts[i] {
			return -1
		}
	}
	return 0
}

// parseVersion parses a version string into [major, minor, patch] integers
func parseVersion(version string) [3]int {
	parts := strings.Split(version, ".")
	result := [3]int{0, 0, 0}

	for i := 0; i < len(parts) && i < 3; i++ {
		if val, err := strconv.Atoi(parts[i]); err == nil {
			result[i] = val
		}
	}

	return result
}

// migrateV170 handles all migrations for version 1.7.0
func (m *Migrator) migrateV170() error {
	logging.GetLogger().Info("Running v1.7.0 migrations")

	// Migrate Trader field to Book field in blotter trades
	if err := m.migrateTraderToBookInBlotter(); err != nil {
		return fmt.Errorf("failed to migrate Trader to Book in blotter: %w", err)
	}

	// Migrate Trader field to Book field in portfolio positions
	if err := m.migrateTraderToBookInPosition(); err != nil {
		return fmt.Errorf("failed to migrate Trader to Book in positions: %w", err)
	}

	m.lastMigrationApplied = "1.7.0"

	// Save to database
	if err := m.saveLastMigrationToDB(); err != nil {
		return fmt.Errorf("failed to save migration version to database: %w", err)
	}

	return nil
}

// migrateTraderToBook migrates old "Trader" field to new "Book" field in trade records
func (m *Migrator) migrateTraderToBookInBlotter() error {
	tradeKeys, err := m.db.GetAllKeysWithPrefix(string(types.TradeKeyPrefix))
	if err != nil {
		return err
	}

	if len(tradeKeys) == 0 {
		logging.GetLogger().Info("No trade records found, skipping Trader->Book migration")
		return nil
	}

	migratedCount := 0
	for _, key := range tradeKeys {
		// Define old trade structure with Trader field
		var oldTrade struct {
			TradeID     string  `json:"TradeID"`
			TradeDate   string  `json:"TradeDate"`
			Ticker      string  `json:"Ticker"`
			Side        string  `json:"Side"`
			Quantity    float64 `json:"Quantity"`
			Price       float64 `json:"Price"`
			Fx          float64 `json:"Fx"`
			Yield       float64 `json:"Yield"`
			Trader      string  `json:"Trader"` // Old field name
			Book        string  `json:"Book"`   // New field name (may be empty in old records)
			Broker      string  `json:"Broker"`
			Account     string  `json:"Account"`
			Status      string  `json:"Status"`
			OrigTradeID string  `json:"OrigTradeID"`
			SeqNum      int     `json:"SeqNum"`
		}

		// Try to load the trade data
		err := m.db.Get(key, &oldTrade)
		if err != nil {
			return fmt.Errorf("failed to load trade from key %s: %w", key, err)
		}

		// Check if migration is needed (Book is empty but Trader has value)
		if oldTrade.Book == "" && oldTrade.Trader != "" {
			// Create the migrated trade structure
			migratedTrade := struct {
				TradeID     string  `json:"TradeID"`
				TradeDate   string  `json:"TradeDate"`
				Ticker      string  `json:"Ticker"`
				Side        string  `json:"Side"`
				Quantity    float64 `json:"Quantity"`
				Price       float64 `json:"Price"`
				Fx          float64 `json:"Fx"`
				Yield       float64 `json:"Yield"`
				Book        string  `json:"Book"`
				Broker      string  `json:"Broker"`
				Account     string  `json:"Account"`
				Status      string  `json:"Status"`
				OrigTradeID string  `json:"OrigTradeID"`
				SeqNum      int     `json:"SeqNum"`
			}{
				TradeID:     oldTrade.TradeID,
				TradeDate:   oldTrade.TradeDate,
				Ticker:      oldTrade.Ticker,
				Side:        oldTrade.Side,
				Quantity:    oldTrade.Quantity,
				Price:       oldTrade.Price,
				Fx:          oldTrade.Fx,
				Yield:       oldTrade.Yield,
				Book:        oldTrade.Trader, // Migrate Trader -> Book
				Broker:      oldTrade.Broker,
				Account:     oldTrade.Account,
				Status:      oldTrade.Status,
				OrigTradeID: oldTrade.OrigTradeID,
				SeqNum:      oldTrade.SeqNum,
			}

			// Save the migrated trade back to database
			err = m.db.Put(key, migratedTrade)
			if err != nil {
				return fmt.Errorf("failed to save migrated trade %s: %w", oldTrade.TradeID, err)
			}

			migratedCount++
			logging.GetLogger().Infof("Migrated trade %s: Trader '%s' -> Book '%s'", oldTrade.TradeID, oldTrade.Trader, migratedTrade.Book)
		}
	}

	if migratedCount > 0 {
		logging.GetLogger().Infof("Migration completed: %d trades migrated from Trader to Book field", migratedCount)
	} else {
		logging.GetLogger().Info("No trades required migration from Trader to Book field")
	}

	return nil
}

// migrateTraderToBookInPosition migrates old "Trader" field to new "Book" field in position records
func (m *Migrator) migrateTraderToBookInPosition() error {
	// REF: fmt.Sprintf("%s:%s:%s", types.PositionKeyPrefix, trade.Trader/Book, trade.Ticker)
	positionKeys, err := m.db.GetAllKeysWithPrefix(string(types.PositionKeyPrefix))
	if err != nil {
		return err
	}

	if len(positionKeys) == 0 {
		logging.GetLogger().Info("No position records found, skipping Trader->Book migration")
		return nil
	}

	migratedCount := 0
	for _, key := range positionKeys {
		// Define old position structure with Trader field
		var oldPosition struct {
			Ticker        string  `json:"Ticker"`
			Trader        string  `json:"Trader"` // Old field name
			Book          string  `json:"Book"`   // New field name (may be empty in old records)
			Ccy           string  `json:"Ccy"`
			AssetClass    string  `json:"AssetClass"`
			AssetSubClass string  `json:"AssetSubClass"`
			Qty           float64 `json:"Qty"`
			Mv            float64 `json:"Mv"`
			PnL           float64 `json:"PnL"`
			Dividends     float64 `json:"Dividends"`
			AvgPx         float64 `json:"AvgPx"`
			Px            float64 `json:"Px"`
			TotalPaid     float64 `json:"TotalPaid"`
			FxRate        float64 `json:"FxRate"`
		}

		// Try to load the position data
		err := m.db.Get(key, &oldPosition)
		if err != nil {
			return fmt.Errorf("failed to load position from key %s: %w", key, err)
		}

		// Check if migration is needed (Book is empty but Trader has value)
		if oldPosition.Book == "" && oldPosition.Trader != "" {
			// Create the migrated position structure
			migratedPosition := struct {
				Ticker        string  `json:"Ticker"`
				Book          string  `json:"Book"`
				Ccy           string  `json:"Ccy"`
				AssetClass    string  `json:"AssetClass"`
				AssetSubClass string  `json:"AssetSubClass"`
				Qty           float64 `json:"Qty"`
				Mv            float64 `json:"Mv"`
				PnL           float64 `json:"PnL"`
				Dividends     float64 `json:"Dividends"`
				AvgPx         float64 `json:"AvgPx"`
				Px            float64 `json:"Px"`
				TotalPaid     float64 `json:"TotalPaid"`
				FxRate        float64 `json:"FxRate"`
			}{
				Ticker:        oldPosition.Ticker,
				Book:          oldPosition.Trader, // Migrate Trader -> Book
				Ccy:           oldPosition.Ccy,
				AssetClass:    oldPosition.AssetClass,
				AssetSubClass: oldPosition.AssetSubClass,
				Qty:           oldPosition.Qty,
				Mv:            oldPosition.Mv,
				PnL:           oldPosition.PnL,
				Dividends:     oldPosition.Dividends,
				AvgPx:         oldPosition.AvgPx,
				Px:            oldPosition.Px,
				TotalPaid:     oldPosition.TotalPaid,
				FxRate:        oldPosition.FxRate,
			}

			// Save the migrated position back to database
			err = m.db.Put(key, migratedPosition)
			if err != nil {
				return fmt.Errorf("failed to save migrated position %s: %w", oldPosition.Ticker, err)
			}

			migratedCount++
			logging.GetLogger().Infof("Migrated position %s: Trader '%s' -> Book '%s'", oldPosition.Ticker, oldPosition.Trader, migratedPosition.Book)
		}
	}

	if migratedCount > 0 {
		logging.GetLogger().Infof("Migration completed: %d positions migrated from Trader to Book field", migratedCount)
	} else {
		logging.GetLogger().Info("No positions required migration from Trader to Book field")
	}

	return nil
}
