package blotter

import (
	"bytes"
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/csvutil"
	"portfolio-manager/pkg/event"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"sort"
	"strings"
	"sync"
	"time"

	"encoding/csv"
	"os"
	"strconv"

	"slices"

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

const csvImportDefaultBroker = "csv-import"

var tradeCSVHeaderAliases = map[string]string{
	"TradeDate":         "TradeDate",
	"Date":              "TradeDate",
	"Ticker":            "Ticker",
	"Side":              "Side",
	"Quantity":          "Quantity",
	"Price":             "Price",
	"Yield":             "Yield",
	"Book":              "Book",
	"Broker":            "Broker",
	"Account":           "Account",
	"Status":            "Status",
	"Fx":                "Fx",
	"InstrumentType":    "InstrumentType",
	"Type":              "InstrumentType",
	"UnderlyingTicker":  "UnderlyingTicker",
	"Underlying":        "UnderlyingTicker",
	"ExpiryDate":        "ExpiryDate",
	"Expiry":            "ExpiryDate",
	"StrikePrice":       "StrikePrice",
	"Strike":            "StrikePrice",
	"CallPut":           "CallPut",
	"CP":                "CallPut",
	"UnderlyingSpotRef": "UnderlyingSpotRef",
	"TradeID":           "",
	"Trade ID":          "",
	"Name":              "",
	"Value":             "",
	"Asset Class":       "",
	"Asset Sub Class":   "",
	"CCY":               "",
	"Category":          "",
	"Sub Category":      "",
	"Domicile":          "",
	"Confirmation":      "",
}

type tradeCSVLayout struct {
	indexes map[string]int
}

func newTradeCSVLayout(header []string) (*tradeCSVLayout, error) {
	indexes := make(map[string]int, len(header))
	for i, rawHeader := range header {
		headerName := strings.TrimSpace(rawHeader)
		canonical, ok := tradeCSVHeaderAliases[headerName]
		if !ok {
			return nil, fmt.Errorf("invalid CSV header: unsupported column %s at position %d", rawHeader, i)
		}
		if canonical == "" {
			continue
		}
		if _, exists := indexes[canonical]; exists {
			return nil, fmt.Errorf("invalid CSV header: duplicate column %s", rawHeader)
		}
		indexes[canonical] = i
	}

	for _, required := range []string{"TradeDate", "Ticker", "Side", "Quantity", "Price", "Book", "Account"} {
		if _, ok := indexes[required]; !ok {
			return nil, fmt.Errorf("invalid CSV header: missing required column %s", required)
		}
	}

	return &tradeCSVLayout{indexes: indexes}, nil
}

func (l *tradeCSVLayout) get(row []string, field string) string {
	idx, ok := l.indexes[field]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func (l *tradeCSVLayout) has(field string) bool {
	_, ok := l.indexes[field]
	return ok
}

// TradeBlotter represents a service for managing trades.
type TradeBlotter struct {
	trades         []Trade
	tradesByID     map[string]*Trade
	tradesByTicker map[string][]Trade
	currentSeqNum  int // used as a pointer to the head of the blotter
	db             dal.Database
	rdataSvc       rdata.ReferenceManager
	mdataSvc       mdata.MarketDataManager
	eventBus       *event.EventBus
	mu             sync.Mutex
}

// NewBlotter creates a new TradeBlotter instance.
func NewBlotter(db dal.Database) *TradeBlotter {
	var currentSeqNum int
	err := db.Get(string(types.HeadSequenceBlotterKey), &currentSeqNum)
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

func (b *TradeBlotter) SetTradeSupportServices(rdataSvc rdata.ReferenceManager, mdataSvc mdata.MarketDataManager) {
	b.rdataSvc = rdataSvc
	b.mdataSvc = mdataSvc
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
	if isPreLoadFromDB && trade.SeqNum > b.currentSeqNum {
		b.currentSeqNum = trade.SeqNum
	}

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
			b.trades = slices.Delete(b.trades, i, i+1)
			break
		}
	}

	// Remove trade from the indexes
	delete(b.tradesByID, trade.TradeID)
	b.tradesByTicker[originalTrade.Ticker] = removeTradeFromSlice(b.tradesByTicker[originalTrade.Ticker], trade.TradeID)

	// Add trade to the trades slice and indexes
	b.trades = append(b.trades, trade)
	b.tradesByID[trade.TradeID] = &trade
	b.tradesByTicker[trade.Ticker] = append(b.tradesByTicker[trade.Ticker], trade)

	// Write updated trade to the database
	tradeKey := generateTradeKey(trade)
	err := b.db.Put(tradeKey, trade)
	if err != nil {
		return err
	}

	// Publish an update trade event
	b.PublishUpdateTradeEvent(trade, originalTrade)

	return nil
}

func (b *TradeBlotter) RemoveAllTrades() error {
	tradeIDs := make([]string, len(b.trades))
	for i, trade := range b.trades {
		tradeIDs[i] = trade.TradeID
	}

	// Then remove them
	err := b.RemoveTrades(tradeIDs)
	if err != nil {
		return err
	}

	b.currentSeqNum = -1
	b.saveSeqNumToDAL(-1)
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

	logging.GetLogger().Infof("Removing trade with ID %s", tradeID)

	// Remove trade from the trades slice
	for i, t := range b.trades {
		if t.TradeID == tradeID {
			b.trades = slices.Delete(b.trades, i, i+1)
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

// GetAllTickers returns all unique tickers that has ever traded in the blotter.
func (b *TradeBlotter) GetAllTickers() ([]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	tickers := make([]string, 0, len(b.tradesByTicker))
	for ticker := range b.tradesByTicker {
		tickers = append(tickers, ticker)
	}

	return tickers, nil
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
			return slices.Delete(trades, i, i+1)
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
	TradeID           string  `json:"TradeID" validate:"required"`   // Unique identifier for the trade
	TradeDate         string  `json:"TradeDate" validate:"required"` // Date and time of the trade
	Ticker            string  `json:"Ticker" validate:"required"`    // Ticker symbol of the asset
	Side              string  `json:"Side" validate:"required"`      // Buy or Sell
	Quantity          float64 `json:"Quantity" validate:"required"`  // Quantity of the asset
	Price             float64 `json:"Price" validate:"gte=0"`        // Price per unit of the asset
	Fx                float64 `json:"Fx"`                            // FX rate for the trade
	Yield             float64 `json:"Yield"`                         // Yield of the asset
	Book              string  `json:"Book" validate:"required"`      // Book associated with the trade
	Broker            string  `json:"Broker" validate:"required"`    // Broker who executed the trade
	Account           string  `json:"Account" validate:"required"`   // Account associated with the trade (CDP, MIP, Custodian)
	Status            string  `json:"Status"`                        // Status of the trade (e.g. Open, AutoClosed, Closed), autoclosed if the trade is closed by the system automatically upon expiry (e.g. MAS Bills), closed if the trade is closed manually
	OrigTradeID       string  `json:"OrigTradeID"`                   // Original trade ID to link auto closed trades to the original trade
	SeqNum            int     `json:"SeqNum"`                        // Sequence number
	InstrumentType    string  `json:"InstrumentType"`
	UnderlyingTicker  string  `json:"UnderlyingTicker"`
	UnderlyingSpotRef float64 `json:"UnderlyingSpotRef"`
	ExpiryDate        string  `json:"ExpiryDate"`
	StrikePrice       float64 `json:"StrikePrice"`
	CallPut           string  `json:"CallPut"`
}

// Clone returns a deep copy of the Trade.
func (t Trade) Clone() Trade {
	return Trade{
		TradeID:           t.TradeID,
		TradeDate:         t.TradeDate,
		Ticker:            t.Ticker,
		Side:              t.Side,
		Quantity:          t.Quantity,
		Price:             t.Price,
		Fx:                t.Fx,
		Yield:             t.Yield,
		Book:              t.Book,
		Broker:            t.Broker,
		Account:           t.Account,
		Status:            t.Status,
		OrigTradeID:       t.OrigTradeID,
		SeqNum:            t.SeqNum,
		InstrumentType:    t.InstrumentType,
		UnderlyingTicker:  t.UnderlyingTicker,
		UnderlyingSpotRef: t.UnderlyingSpotRef,
		ExpiryDate:        t.ExpiryDate,
		StrikePrice:       t.StrikePrice,
		CallPut:           t.CallPut,
	}
}

// NewTrade creates a new Trade instance.
func NewTrade(side string, quantity float64, ticker, book, broker, account, status, origTradeId string, price, fx float64, yield float64, tradeDate time.Time, attributes ...TradeAttributes) (*Trade, error) {
	if !isValidSide(side) {
		return nil, errors.New("side must be either 'buy' or 'sell'")
	}

	if !isValidStatus(status) {
		return nil, fmt.Errorf("status must be either '%s', '%s' or '%s'", StatusOpen, StatusAutoClose, StatusClosed)
	}

	tradeAttributes := TradeAttributes{}
	if len(attributes) > 0 {
		tradeAttributes = attributes[0]
	}
	tradeAttributes, err := NormalizeTradeAttributes(ticker, tradeAttributes)
	if err != nil {
		return nil, err
	}

	trade := Trade{
		TradeID:           common.GenerateTradeID(),
		TradeDate:         tradeDate.Format(time.RFC3339),
		Ticker:            ticker,
		Side:              side,
		Quantity:          quantity,
		Price:             price,
		Fx:                fx,
		Yield:             yield,
		Book:              book,
		Broker:            broker,
		Account:           account,
		Status:            status,
		OrigTradeID:       origTradeId,
		InstrumentType:    tradeAttributes.InstrumentType,
		UnderlyingTicker:  tradeAttributes.UnderlyingTicker,
		UnderlyingSpotRef: tradeAttributes.UnderlyingSpotRef,
		ExpiryDate:        tradeAttributes.ExpiryDate,
		StrikePrice:       tradeAttributes.StrikePrice,
		CallPut:           tradeAttributes.CallPut,
	}

	err = validateTrade(trade)
	return &trade, err
}

// NewTradeWithID creates a new Trade instance with a given trade ID, mainly for updating purposes
func NewTradeWithID(tradeID string, side string, quantity float64, ticker, book, broker, account, status, origTradeId string, price, fx float64, yield float64, seqNum int, tradeDate time.Time, attributes ...TradeAttributes) (*Trade, error) {

	if !isValidSide(side) {
		return nil, errors.New("side must be either 'buy' or 'sell'")
	}

	if !isValidStatus(status) {
		return nil, fmt.Errorf("status must be either '%s', '%s' or '%s'", StatusOpen, StatusAutoClose, StatusClosed)
	}

	tradeAttributes := TradeAttributes{}
	if len(attributes) > 0 {
		tradeAttributes = attributes[0]
	}
	tradeAttributes, err := NormalizeTradeAttributes(ticker, tradeAttributes)
	if err != nil {
		return nil, err
	}

	trade := Trade{
		TradeID:           tradeID,
		TradeDate:         tradeDate.Format(time.RFC3339),
		Ticker:            ticker,
		Side:              side,
		Quantity:          quantity,
		Price:             price,
		Fx:                fx,
		Yield:             yield,
		Book:              book,
		Broker:            broker,
		Account:           account,
		Status:            status,
		OrigTradeID:       origTradeId,
		SeqNum:            seqNum,
		InstrumentType:    tradeAttributes.InstrumentType,
		UnderlyingTicker:  tradeAttributes.UnderlyingTicker,
		UnderlyingSpotRef: tradeAttributes.UnderlyingSpotRef,
		ExpiryDate:        tradeAttributes.ExpiryDate,
		StrikePrice:       tradeAttributes.StrikePrice,
		CallPut:           tradeAttributes.CallPut,
	}

	err = validateTrade(trade)
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
	if err := validate.Struct(trade); err != nil {
		return err
	}

	trade.InstrumentType = NormalizeInstrumentType(trade.InstrumentType)
	trade.CallPut = NormalizeCallPut(trade.CallPut)
	if trade.InstrumentType == "" {
		trade.InstrumentType = InstrumentTypeOutright
	}

	if trade.InstrumentType == InstrumentTypeOption {
		if strings.TrimSpace(trade.UnderlyingTicker) == "" {
			return fmt.Errorf("underlying ticker is required for option trades")
		}
		if trade.ExpiryDate == "" {
			return fmt.Errorf("expiry date is required for option trades")
		}
		if trade.StrikePrice <= 0 {
			return fmt.Errorf("strike price must be greater than 0 for option trades")
		}
		if trade.CallPut != CallPutCall && trade.CallPut != CallPutPut {
			return fmt.Errorf("call put must be either '%s' or '%s' for option trades", CallPutCall, CallPutPut)
		}
	}

	return nil
}

// ** Import / Export CSV Section **
// All import and export functionalities for CSV has no concept of autoclosed nor an original trade id, since assumption is for migrating to a different system
// For backup purposes without intention of migrating to a different system, the author suggests to backup the leveldb instead of migrating via csv files

// ImportFromCSV imports trades from a CSV file and adds them to the blotter.
// Expected CSV format: TradeDate,Ticker,Side,Quantity,Price,Yield,Book,Broker,Account,Status,Fx
func (b *TradeBlotter) ImportFromCSVFile(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("error opening CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return b.ImportFromCSVReader(reader)
}

// ImportFromCSVReader imports trades from a CSV reader and adds them to the blotter.
func (b *TradeBlotter) ImportFromCSVReader(reader *csv.Reader) (int, error) {
	logging.GetLogger().Info("Importing trades from CSV")

	// Read and validate header
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("error reading CSV header: %w", err)
	}

	layout, err := newTradeCSVLayout(header)
	if err != nil {
		return 0, err
	}

	// Read all rows and create trades
	var trades []*Trade
	lineNum := 1
	for {
		cnt := lineNum - 1
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return cnt, fmt.Errorf("error reading CSV line %d: %w", lineNum, err)
		}

		quantity, err := strconv.ParseFloat(layout.get(row, "Quantity"), 64)
		if err != nil {
			return cnt, fmt.Errorf("invalid quantity at line %d: %w", lineNum, err)
		}

		price, err := strconv.ParseFloat(layout.get(row, "Price"), 64)
		if err != nil {
			return cnt, fmt.Errorf("invalid price at line %d: %w", lineNum, err)
		}

		fx := 0.0
		if layout.has("Fx") && layout.get(row, "Fx") != "" {
			fx, err = strconv.ParseFloat(layout.get(row, "Fx"), 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid fx rate at line %d: %w", lineNum, err)
			}
		}

		var yield float64
		if layout.has("Yield") && layout.get(row, "Yield") != "" {
			yield, err = strconv.ParseFloat(layout.get(row, "Yield"), 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid yield at line %d: %w", lineNum, err)
			}
		}

		tradeDate, err := time.Parse(time.RFC3339, layout.get(row, "TradeDate"))
		if err != nil {
			return cnt, fmt.Errorf("invalid trade date at line %d: %w", lineNum, err)
		}

		status := StatusOpen
		if layout.has("Status") && layout.get(row, "Status") != "" {
			status = layout.get(row, "Status")
		}

		attributes := TradeAttributes{}
		attributes.InstrumentType = layout.get(row, "InstrumentType")
		attributes.UnderlyingTicker = layout.get(row, "UnderlyingTicker")
		attributes.ExpiryDate = layout.get(row, "ExpiryDate")
		attributes.CallPut = layout.get(row, "CallPut")
		if layout.has("StrikePrice") && layout.get(row, "StrikePrice") != "" {
			attributes.StrikePrice, err = strconv.ParseFloat(layout.get(row, "StrikePrice"), 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid strike price at line %d: %w", lineNum, err)
			}
		}
		if layout.has("UnderlyingSpotRef") && layout.get(row, "UnderlyingSpotRef") != "" {
			attributes.UnderlyingSpotRef, err = strconv.ParseFloat(layout.get(row, "UnderlyingSpotRef"), 64)
			if err != nil {
				return cnt, fmt.Errorf("invalid underlying spot reference at line %d: %w", lineNum, err)
			}
		}

		broker := layout.get(row, "Broker")
		if broker == "" {
			broker = csvImportDefaultBroker
		}

		trade, err := b.BuildTrade(TradeInput{
			TradeDate:   tradeDate,
			Ticker:      layout.get(row, "Ticker"),
			Side:        layout.get(row, "Side"),
			Quantity:    quantity,
			Price:       price,
			Fx:          fx,
			Yield:       yield,
			Book:        layout.get(row, "Book"),
			Broker:      broker,
			Account:     layout.get(row, "Account"),
			Status:      status,
			OrigTradeID: "",
			Attributes:  attributes,
		})
		if err != nil {
			return cnt, fmt.Errorf("error creating trade at line %d: %w", lineNum, err)
		}

		trades = append(trades, trade)
		lineNum++
	}

	// Add all trades after validation
	for i, trade := range trades {
		if err := b.AddTrade(*trade); err != nil {
			return i, fmt.Errorf("error adding trades: %w", err)
		}
	}

	b.sortTrades()

	return len(trades), nil
}

// ExportToCSVBytes exports all trades to a CSV file in memory and returns it as a byte slice.
func (b *TradeBlotter) ExportToCSVBytes() ([]byte, error) {
	logging.GetLogger().Info("Exporting trades to CSV in memory")

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	err := writer.Write(csvutil.TradeHeaders)
	if err != nil {
		return nil, fmt.Errorf("error writing CSV header: %w", err)
	}

	// Write trades
	for _, trade := range b.trades {
		err = writer.Write([]string{
			trade.TradeDate,
			trade.Ticker,
			trade.Side,
			csvutil.FormatFloat(trade.Quantity, 4),
			csvutil.FormatFloat(trade.Price, 4),
			csvutil.FormatFloat(trade.Yield, 4),
			trade.Book,
			trade.Broker,
			trade.Account,
			trade.Status,
			csvutil.FormatFloat(trade.Fx, 4),
			trade.InstrumentType,
			trade.UnderlyingTicker,
			trade.ExpiryDate,
			csvutil.FormatFloat(trade.StrikePrice, 4),
			trade.CallPut,
			csvutil.FormatFloat(trade.UnderlyingSpotRef, 4),
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
