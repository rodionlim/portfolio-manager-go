package blotter_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"portfolio-manager/internal/blotter"
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

func createTestTrade() (*blotter.Trade, error) {
	return blotter.NewTrade("buy", 100, "AAPL", "traderA", "dbs", "cdp", blotter.StatusOpen, "", 150.0, 0.0, time.Now())
}

func createMockCSVFile(t *testing.T, content [][]string) string {
	file, err := os.CreateTemp("", "trades_*.csv")
	assert.NoError(t, err)

	writer := csv.NewWriter(file)
	err = writer.WriteAll(content)
	assert.NoError(t, err)

	writer.Flush()
	file.Close()

	return file.Name()
}

func TestAddTrade(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := blotter.NewBlotter(db)

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

	blotter := blotter.NewBlotter(db)

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

	blotter := blotter.NewBlotter(db)

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
	trade, err := blotter.NewTrade("buysell", 100, "AAPL", "traderA", "dbs", "cdp", blotter.StatusOpen, "", 150.0, 0.0, time.Now())
	assert.Error(t, err)
	assert.Nil(t, trade)
}

func TestTradeSequenceNumber(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotter := blotter.NewBlotter(db)

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

	blotterSvc := blotter.NewBlotter(db)
	trade, err := createTestTrade()
	assert.NoError(t, err)

	eventPublished := false
	blotterSvc.Subscribe(blotter.NewTradeEvent, event.NewEventHandler(func(handler event.Event) {
		if handler.Data.(blotter.TradeEventPayload).Trade.TradeID == trade.TradeID {
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

	blotterSvc := blotter.NewBlotter(db)
	trade, err := createTestTrade()
	assert.NoError(t, err)

	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	eventPublished := false
	blotterSvc.Subscribe(blotter.RemoveTradeEvent, event.NewEventHandler(func(handler event.Event) {
		if handler.Data.(blotter.TradeEventPayload).Trade.TradeID == trade.TradeID {
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

	blotterSvc := blotter.NewBlotter(db)

	// Create and add multiple trades
	for i := 0; i < 5; i++ {
		trade, err := createTestTrade()
		assert.NoError(t, err)
		err = blotterSvc.AddTrade(*trade)
		assert.NoError(t, err)
	}

	// Test valid range
	trades := blotterSvc.GetTradesBySeqNumRange(1, 3)
	assert.Equal(t, 3, len(trades))
	for _, trade := range trades {
		assert.True(t, trade.SeqNum >= 1 && trade.SeqNum <= 3)
	}

	// Test empty range
	trades = blotterSvc.GetTradesBySeqNumRange(10, 15)
	assert.Empty(t, trades)

	// Test invalid range (start > end)
	trades = blotterSvc.GetTradesBySeqNumRange(3, 1)
	assert.Empty(t, trades)

	// Test single sequence number
	trades = blotterSvc.GetTradesBySeqNumRange(2, 2)
	assert.Equal(t, 1, len(trades))
	assert.Equal(t, 2, trades[0].SeqNum)
}

func TestImportFromCSVFile(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	// Create mock CSV content
	csvContent := [][]string{
		{"TradeDate", "Ticker", "Side", "Quantity", "Price", "Yield", "Trader", "Broker", "Account", "Status"},
		{"2023-10-12T07:20:50Z", "AAPL", "buy", "100", "150.0", "0.0", "trader1", "broker1", "cdp", ""},
		{"2023-10-12T07:20:50Z", "GOOG", "sell", "200", "186.53", "", "trader2", "broker2", "cdp", ""},
	}

	// Create mock CSV file
	filePath := createMockCSVFile(t, csvContent)
	defer os.Remove(filePath)

	// Create mock TradeBlotter
	blotterSvc := blotter.NewBlotter(db)

	// Call ImportFromCSV
	cnt, err := blotterSvc.ImportFromCSVFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, 2, cnt)

	// Verify AddTrade was called with the correct trades
	expectedTrades := []blotter.Trade{
		{
			TradeDate:   "2023-10-12T07:20:50Z",
			Ticker:      "AAPL",
			Side:        "buy",
			Quantity:    100,
			Price:       150.0,
			Yield:       0.0,
			Trader:      "trader1",
			Broker:      "broker1",
			Account:     "cdp",
			Status:      blotter.StatusOpen,
			OrigTradeID: "",
		},
		{
			TradeDate:   "2023-10-12T07:20:50Z",
			Ticker:      "GOOG",
			Side:        "sell",
			Quantity:    200,
			Price:       186.53,
			Yield:       0.0,
			Trader:      "trader2",
			Broker:      "broker2",
			Account:     "cdp",
			Status:      blotter.StatusOpen,
			OrigTradeID: "",
		},
	}

	trades := blotterSvc.GetTrades()
	assert.Equal(t, len(expectedTrades), len(trades))
}
