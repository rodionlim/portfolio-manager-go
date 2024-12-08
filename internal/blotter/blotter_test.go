package blotter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/event"

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
	return NewTrade("buy", 100, AssetClassEquities, "stock", "AAPL", "traderA", "dbs", 150.0, 0.0, time.Now())
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
	trade, err := NewTrade("buysell", 100, AssetClassEquities, "stock", "AAPL", "traderA", "dbs", 150.0, 0.0, time.Now())
	assert.Error(t, err)
	assert.Nil(t, trade)
}

func TestTradeSequenceNumber(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := NewBlotter(db)

	trade1, err := createTestTrade()
	assert.NoError(t, err)

	err = blotter.AddTrade(*trade1)
	assert.NoError(t, err)

	trade2, err := createTestTrade()
	assert.NoError(t, err)

	err = blotter.AddTrade(*trade2)
	assert.NoError(t, err)

	trades := blotter.GetTrades()
	assert.Equal(t, 2, len(trades))
	assert.Equal(t, 0, trades[0].SeqNum)
	assert.Equal(t, 1, trades[1].SeqNum)
}

func TestEventPublishingOnAddTrade(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotterSvc := NewBlotter(db)
	trade, err := createTestTrade()
	assert.NoError(t, err)

	eventPublished := false
	blotterSvc.Subscribe(NewTradeEvent, event.NewEventHandler(func(handler event.Event) {
		if handler.Data.(NewTradeEventPayload).Trade.TradeID == trade.TradeID {
			eventPublished = true
		}
	}))

	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	// Wait for the event to be published
	time.Sleep(50 * time.Millisecond)

	assert.True(t, eventPublished, "Expected event to be published when trade is added")
}

func TestEventPublishingOnRemoveTrade(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotterSvc := NewBlotter(db)
	trade, err := createTestTrade()
	assert.NoError(t, err)

	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	eventPublished := false
	blotterSvc.Subscribe(RemoveTradeEvent, event.NewEventHandler(func(handler event.Event) {
		if handler.Data.(NewTradeEventPayload).Trade.TradeID == trade.TradeID {
			eventPublished = true
		}
	}))

	err = blotterSvc.RemoveTrade(trade.TradeID)
	assert.NoError(t, err)

	// Wait for the event to be published
	time.Sleep(50 * time.Millisecond)

	assert.True(t, eventPublished, "Expected event to be published when trade is removed")
}

func TestGetTradesBySeqNumRange(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := NewBlotter(db)

	// Create and add multiple trades
	for i := 0; i < 5; i++ {
		trade, err := createTestTrade()
		assert.NoError(t, err)
		err = blotter.AddTrade(*trade)
		assert.NoError(t, err)
	}

	// Test valid range
	trades := blotter.GetTradesBySeqNumRange(1, 3)
	assert.Equal(t, 3, len(trades))
	for _, trade := range trades {
		assert.True(t, trade.SeqNum >= 1 && trade.SeqNum <= 3)
	}

	// Test empty range
	trades = blotter.GetTradesBySeqNumRange(10, 15)
	assert.Empty(t, trades)

	// Test invalid range (start > end)
	trades = blotter.GetTradesBySeqNumRange(3, 1)
	assert.Empty(t, trades)

	// Test single sequence number
	trades = blotter.GetTradesBySeqNumRange(2, 2)
	assert.Equal(t, 1, len(trades))
	assert.Equal(t, 2, trades[0].SeqNum)
}
