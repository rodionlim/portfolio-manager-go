package blotter

import (
	"bytes"
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/event"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"
	"strings"
	"sync"
	"time"

	"encoding/csv"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TradeSide represents the side of a trade (buy or sell).
const (
	TradeSideBuy  = "buy"
	TradeSideSell = "sell"
)

const (
	StatusOpen      = "open"
	StatusAutoClose = "autoclosed"
	StatusClosed    = "closed"
)

// TradeBlotter represents a service for managing trades.
type TradeBlotter struct {
	trades         []Trade
	tradesByID     map[string]*Trade
	tradesByTicker map[string][]Trade
	currentSeqNum  int // used as a pointer to the head of the blotter
	db             dal.Database
	eventBus       *event.EventBus
	mu             sync.Mutex
}

// NewBlotter creates a new TradeBlotter instance.
func NewBlotter(db dal.Database) *TradeBlotter {
	var currentSeqNum int
	err := db.Get(string(types.HeadSequenceBlotterKey), currentSeqNum)
	if err != nil {
		currentSeqNum = -1
	}

	return &TradeBlotter{
		trades:         []Trade{},
		tradesByID:     make(map[string]*Trade),
		tradesByTicker: make(map[string][]Trade),
		currentSeqNum:  currentSeqNum,
		db:             db,
		eventBus:       event.NewEventBus(),
	}
}

func (b *TradeBlotter) LoadFromDB() error {
	tradeKeys, err := b.db.GetAllKeysWithPrefix(string(types.TradeKeyPrefix))
	if err != nil {
		return err
	}

	for _, key := range tradeKeys {
		var trade Trade
		err := b.db.Get(key, &trade)
		if err != nil {
			return err
		}
		err = b.AddTradePreloaded(trade)
		if err != nil {
			return err
		}
	}

	b.sortTrades()

	logging.GetLogger().Infof("Loaded %d trades from database", len(tradeKeys))

	return nil
}

// SortTrades sorts the trades and tradesByTicker by TradeDate.
func (b *TradeBlotter) sortTrades() {
	logging.GetLogger().Info("Sorting trades (ascending) within the blotter")
	// Sort the trades slice
	sort.Slice(b.trades, func(i, j int) bool {
		return b.trades[i].TradeDate < b.trades[j].TradeDate
	})

	// Sort the tradesByTicker map
	for ticker, trades := range b.tradesByTicker {
		sort.Slice(trades, func(i, j int) bool {
			return trades[i].TradeDate < trades[j].TradeDate
		})
		b.tradesByTicker[ticker] = trades
	}
}

// AddTrade adds a new trade to the blotter and writes it to the database.
func (b *TradeBlotter) AddTrade(trade Trade) error {
	return b.addTrade(trade, false)
}

// AddTrade adds trade from database to the blotter
func (b *TradeBlotter) AddTradePreloaded(trade Trade) error {
	return b.addTrade(trade, true)
}

func (b *TradeBlotter) addTrade(trade Trade, isPreLoadFromDB bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !isPreLoadFromDB {
		trade.SeqNum = b.getNextSeqNum()

		// Write trade to the database
		tradeKey := generateTradeKey(trade)
		err := b.db.Put(tradeKey, trade)
		if err != nil {
			return err
		}

		// Check if the trade already exists
		if _, exists := b.tradesByID[trade.TradeID]; exists {
			// Remove the existing trade from the trades slice
			return errors.New("trade already exists. call RemoveTrade instead")
		}
	}

	// Add trade to the trades slice and indexes
	b.trades = append(b.trades, trade)
	b.tradesByID[trade.TradeID] = &trade
	b.tradesByTicker[trade.Ticker] = append(b.tradesByTicker[trade.Ticker], trade)

	// Publish a new trade event
	if !isPreLoadFromDB {
		b.PublishNewTradeEvent(trade)
	}

	return nil
}

func (b *TradeBlotter) UpdateTrade(trade Trade) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if the trade exists
	t, exists := b.tradesByID[trade.TradeID]
	originalTrade := *t // will be deleted, copy it out first

	if !exists {
		return errors.New("trade not found")
	}

	// Remove trade from the trades slice
	for i, t := range b.trades {
		if t.TradeID == trade.TradeID {
			b.trades = append(b.trades[:i], b.trades[i+1:]...)
			break
		}
	}

	// Remove trade from the indexes
	delete(b.tradesByID, trade.TradeID)
	b.tradesByTicker[trade.Ticker] = removeTradeFromSlice(b.tradesByTicker[trade.Ticker], trade.TradeID)

	// Add trade to the trades slice and indexes
	b.trades = append(b.trades, trade)
	b.tradesByID[trade.TradeID] = &trade
	b.tradesByTicker[trade.Ticker] = append(b.tradesByTicker[trade.Ticker], trade)

	// Publish an update trade event
	b.PublishUpdateTradeEvent(trade, originalTrade)

	return nil
}

func (b *TradeBlotter) RemoveTrades(tradeIDs []string) error {
	var err error
	errorCount := 0
	var failedTradeIDs []string

	for _, tradeID := range tradeIDs {
		if e := b.RemoveTrade(tradeID); e != nil {
			logging.GetLogger().Errorf("Failed to remove trade with ID %s, err %v", tradeID, e)
			errorCount++
			failedTradeIDs = append(failedTradeIDs, tradeID)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("failed to remove %d trades: %s", errorCount, strings.Join(failedTradeIDs, ", "))
	}

	return err
}

// RemoveTrade removes a trade from the blotter and deletes it from the database.
func (b *TradeBlotter) RemoveTrade(tradeID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if the trade exists
	trade, exists := b.tradesByID[tradeID]
	if !exists {
		return errors.New("trade not found")
	}

	// Remove trade from the trades slice
	for i, t := range b.trades {
		if t.TradeID == tradeID {
			b.trades = append(b.trades[:i], b.trades[i+1:]...)
			break
		}
	}

	// Remove trade from the indexes
	delete(b.tradesByID, tradeID)
	b.tradesByTicker[trade.Ticker] = removeTradeFromSlice(b.tradesByTicker[trade.Ticker], tradeID)

	// Remove trade from the database
	tradeKey := generateTradeKey(*trade)
	err := b.db.Delete(tradeKey)
	if err != nil {
		logging.GetLogger().Error("Failed to delete trade from database", err)
		return err
	}

	b.PublishRemoveTradeEvent(*trade)

	return nil
}

// GetTrades returns all trades in the blotter.
func (b *TradeBlotter) GetTrades() []Trade {
	return b.trades
}

// GetTradesBySeqNumRange returns all trades within the range provided
func (b *TradeBlotter) GetTradesBySeqNumRange(startSeqNum, endSeqNum int) []Trade {
	var trades []Trade
	for _, trade := range b.trades {
		if trade.SeqNum >= startSeqNum && trade.SeqNum <= endSeqNum {
			trades = append(trades, trade)
		}
	}
	return trades
}

// GetTradesBySeqNumRangeWithCallback allow to get trades within the range provided and call a callback function, locking the blotter to prevent races
func (b *TradeBlotter) GetTradesBySeqNumRangeWithCallback(startSeqNum, endSeqNum int, callback func(trade Trade)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, trade := range b.trades {
		if trade.SeqNum >= startSeqNum && trade.SeqNum <= endSeqNum {
			callback(trade)
		}
	}
}

// GetTradeByID returns a trade with the given ID.
func (b *TradeBlotter) GetTradeByID(tradeID string) (*Trade, error) {
	trade, exists := b.tradesByID[tradeID]
	if !exists {
		return nil, errors.New("trade not found")
	}
	return trade, nil
}

// GetTradesByTicker returns all trades for the given ticker.
func (b *TradeBlotter) GetTradesByTicker(ticker string) ([]Trade, error) {
	trades, exists := b.tradesByTicker[ticker]
	if !exists {
		return nil, errors.New("no trades found for the given ticker")
	}
	return trades, nil
}

// GetCurrentSeqNum returns the current sequence number.
func (b *TradeBlotter) GetCurrentSeqNum() int {
	return b.currentSeqNum
}

// Subscribe allows other packages to subscribe to blotter events.
// It returns the current sequence number so that the subscriber can request for older trades if necessary in order to catch up
func (tb *TradeBlotter) Subscribe(eventName string, handler event.EventHandler) {
	tb.eventBus.Subscribe(eventName, handler)
}

// Unsubscribe allows other packages to unsubscribe from blotter events.
func (tb *TradeBlotter) Unsubscribe(eventName string, corrId uuid.UUID) {
	tb.eventBus.Unsubscribe(eventName, corrId)
}

// generateTradeKey generates a unique key for the trade.
func generateTradeKey(trade Trade) string {
	return fmt.Sprintf("%s:%s:%d:%s", types.TradeKeyPrefix, trade.Ticker, trade.SeqNum, trade.TradeID)
}

// removeTradeFromSlice removes a trade from a slice of trades by trade ID.
func removeTradeFromSlice(trades []Trade, tradeID string) []Trade {
	for i, t := range trades {
		if t.TradeID == tradeID {
			return append(trades[:i], trades[i+1:]...)
		}
	}
	return trades
}

// getNextSeqNum returns the next sequence number.
func (b *TradeBlotter) getNextSeqNum() int {
	b.currentSeqNum++
	b.saveSeqNumToDAL(b.currentSeqNum)
	return b.currentSeqNum
}

// saveSeqNumToDAL saves the current sequence number to the DAL database.
func (b *TradeBlotter) saveSeqNumToDAL(seqNum int) {
	// Implement the logic to save seqNum to the DAL database
	b.db.Put(string(types.HeadSequenceBlotterKey), seqNum)
}

// Trade represents a trade in the blotter.
type Trade struct {
	TradeID     string  `json:"TradeID"`                       // Unique identifier for the trade
	TradeDate   string  `json:"TradeDate" validate:"required"` // Date and time of the trade
	Ticker      string  `json:"Ticker" validate:"required"`    // Ticker symbol of the asset
	Side        string  `json:"Side" validate:"required"`      // Buy or Sell
	Quantity    float64 `json:"Quantity" validate:"required"`  // Quantity of the asset
	Price       float64 `json:"Price" validate:"required"`     // Price per unit of the asset
	Yield       float64 `json:"Yield"`                         // Yield of the asset
	Trader      string  `json:"Trader" validate:"required"`    // Trader who executed the trade
	Broker      string  `json:"Broker" validate:"required"`    // Broker who executed the trade
	Account     string  `json:"Account" validate:"required"`   // Account associated with the trade (CDP, MIP, Custodian)
	Status      string  `json:"Status"`                        // Status of the trade (e.g. Open, AutoClosed, Closed), autoclosed if the trade is closed by the system automatically upon expiry (e.g. MAS Bills), closed if the trade is closed manually
	OrigTradeID string  `json:"OrigTradeID"`                   // Original trade ID to link auto closed trades to the original trade
	SeqNum      int     `json:"SeqNum"`                        // Sequence number
}

// NewTrade creates a new Trade instance.
func NewTrade(side string, quantity float64, ticker, trader, broker, account, status, origTradeId string, price float64, yield float64, tradeDate time.Time) (*Trade, error) {

	if !isValidSide(side) {
		return nil, errors.New("side must be either 'buy' or 'sell'")
	}

	if !isValidStatus(status) {
		return nil, fmt.Errorf("status must be either '%s', '%s' or '%s'", StatusOpen, StatusAutoClose, StatusClosed)
	}

	trade := Trade{
		TradeID:     common.GenerateTradeID(),
		TradeDate:   tradeDate.Format(time.RFC3339),
		Ticker:      ticker,
		Side:        side,
		Quantity:    quantity,
		Price:       price,
		Yield:       yield,
		Trader:      trader,
		Broker:      broker,
		Account:     account,
		Status:      status,
		OrigTradeID: origTradeId,
	}

	err := validateTrade(trade)
	return &trade, err
}

// NewTradeWithID creates a new Trade instance with a given trade ID, mainly for updating purposes
func NewTradeWithID(tradeID string, side string, quantity float64, ticker, trader, broker, account, status, origTradeId string, price float64, yield float64, seqNum int, tradeDate time.Time) (*Trade, error) {

	if !isValidSide(side) {
		return nil, errors.New("side must be either 'buy' or 'sell'")
	}

	if !isValidStatus(status) {
		return nil, fmt.Errorf("status must be either '%s', '%s' or '%s'", StatusOpen, StatusAutoClose, StatusClosed)
	}

	trade := Trade{
		TradeID:     tradeID,
		TradeDate:   tradeDate.Format(time.RFC3339),
		Ticker:      ticker,
		Side:        side,
		Quantity:    quantity,
		Price:       price,
		Yield:       yield,
		Trader:      trader,
		Broker:      broker,
		Account:     account,
		Status:      status,
		OrigTradeID: origTradeId,
		SeqNum:      seqNum,
	}

	err := validateTrade(trade)
	return &trade, err
}

// isValidSide checks if the provided side is valid.
func isValidSide(side string) bool {
	return side == TradeSideBuy || side == TradeSideSell
}

// isValidStatus checks if the provided status is valid.
func isValidStatus(status string) bool {
	return status == StatusOpen || status == StatusAutoClose || status == StatusClosed
}

// validateTrade validates the trade struct according to predefined rules.
func validateTrade(trade Trade) error {
	validate := validator.New()
	return validate.Struct(trade)
}

// ** Import / Export CSV Section **
// All import and export functionalities for CSV has no concept of autoclosed nor an original trade id, since assumption is for migrating to a different system
// For backup purposes without intention of migrating to a different system, the author suggests to backup the leveldb instead of migrating via csv files

// ImportFromCSV imports trades from a CSV file and adds them to the blotter.
// Expected CSV format: TradeDate,Ticker,Side,Quantity,Price,Yield,Trader,Broker
func (b *TradeBlotter) ImportFromCSVFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return b.ImportFromCSVReader(reader)
}

// ImportFromCSVReader imports trades from a CSV reader and adds them to the blotter.
func (b *TradeBlotter) ImportFromCSVReader(reader *csv.Reader) error {
	logging.GetLogger().Info("Importing trades from CSV")

	// Read and validate header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	expectedHeaders := []string{"TradeDate", "Ticker", "Side", "Quantity", "Price", "Yield", "Trader", "Broker", "Account"}
	if len(header) != len(expectedHeaders) {
		return fmt.Errorf("invalid CSV format: expected %d columns, got %d", len(expectedHeaders), len(header))
	}

	for i, h := range expectedHeaders {
		if header[i] != h {
			return fmt.Errorf("invalid CSV header: expected %s at position %d, got %s", h, i, header[i])
		}
	}

	// Read all rows and create trades
	var trades []*Trade
	lineNum := 1
	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("error reading CSV line %d: %w", lineNum, err)
		}

		quantity, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return fmt.Errorf("invalid quantity at line %d: %w", lineNum, err)
		}

		price, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return fmt.Errorf("invalid price at line %d: %w", lineNum, err)
		}

		var yield float64
		if row[5] != "" {
			yield, err = strconv.ParseFloat(row[5], 64)
			if err != nil {
				return fmt.Errorf("invalid yield at line %d: %w", lineNum, err)
			}
		}

		tradeDate, err := time.Parse(time.RFC3339, row[0])
		if err != nil {
			return fmt.Errorf("invalid trade date at line %d: %w", lineNum, err)
		}

		trade, err := NewTrade(
			row[2], // Side
			quantity,
			row[1], // Ticker
			row[6], // Trader
			row[7], // Broker
			row[8], // Account
			StatusOpen,
			"", // OrigTradeID
			price,
			yield,
			tradeDate,
		)
		if err != nil {
			return fmt.Errorf("error creating trade at line %d: %w", lineNum, err)
		}

		trades = append(trades, trade)
		lineNum++
	}

	// Add all trades after validation
	for _, trade := range trades {
		if err := b.AddTrade(*trade); err != nil {
			return fmt.Errorf("error adding trades: %w", err)
		}
	}

	b.sortTrades()

	return nil
}

// ExportToCSVBytes exports all trades to a CSV file in memory and returns it as a byte slice.
func (b *TradeBlotter) ExportToCSVBytes() ([]byte, error) {
	logging.GetLogger().Info("Exporting trades to CSV in memory")

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	err := writer.Write([]string{"TradeDate", "Ticker", "Side", "Quantity", "Price", "Yield", "Trader", "Broker", "Account"})
	if err != nil {
		return nil, fmt.Errorf("error writing CSV header: %w", err)
	}

	// Write trades
	for _, trade := range b.trades {
		err = writer.Write([]string{
			trade.TradeDate,
			trade.Ticker,
			trade.Side,
			strconv.FormatFloat(trade.Quantity, 'f', -1, 64),
			strconv.FormatFloat(trade.Price, 'f', -1, 64),
			strconv.FormatFloat(trade.Yield, 'f', -1, 64),
			trade.Trader,
			trade.Broker,
			trade.Account,
		})
		if err != nil {
			return nil, fmt.Errorf("error writing trade to CSV: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}
