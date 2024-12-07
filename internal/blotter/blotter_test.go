package blotter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"portfolio-manager/internal/dal"

	"github.com/stretchr/testify/assert"
)

func setupTempDB(t *testing.T) (dal.Database, string) {
	dbPath := filepath.Join(os.TempDir(), "testdb_"+t.Name())
	db, err := dal.NewLevelDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	return db, dbPath
}

func cleanupTempDB(t *testing.T, db dal.Database, dbPath string) {
	err := db.Close()
	if err != nil {
		t.Fatalf("Failed to close temp database: %v", err)
	}
	err = os.RemoveAll(dbPath)
	if err != nil {
		t.Fatalf("Failed to remove temp database: %v", err)
	}
}

func createTestTrade() (*Trade, error) {
	return NewTrade("buy", 100, AssetClassEquities, "stock", "AAPL", "traderA", 150.0, 0.0, time.Now())
}

func TestAddTrade(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := NewBlotter(db)

	trade, err := createTestTrade()
	assert.NoError(t, err)

	err = blotter.AddTrade(*trade)
	assert.NoError(t, err)

	trades := blotter.GetTrades()
	assert.Equal(t, 1, len(trades))
}

func TestGetTradeByID(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := NewBlotter(db)

	trade, err := createTestTrade()
	assert.NoError(t, err)

	err = blotter.AddTrade(*trade)
	assert.NoError(t, err)

	retrievedTrade, err := blotter.GetTradeByID(trade.TradeID)
	assert.NoError(t, err)
	assert.Equal(t, trade.TradeID, retrievedTrade.TradeID)
}

func TestRemoveTrade(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := NewBlotter(db)

	trade, err := createTestTrade()
	assert.NoError(t, err)

	err = blotter.AddTrade(*trade)
	assert.NoError(t, err)

	err = blotter.RemoveTrade(trade.TradeID)
	assert.NoError(t, err)

	trades := blotter.GetTrades()
	assert.Equal(t, 0, len(trades))
}

func TestCreateTradeWithInvalidSide(t *testing.T) {
	trade, err := NewTrade("buysell", 100, AssetClassEquities, "stock", "AAPL", "traderA", 150.0, 0.0, time.Now())
	assert.Error(t, err)
	assert.Nil(t, trade)
}
