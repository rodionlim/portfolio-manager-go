package blotter

import (
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"time"

	"github.com/google/uuid"
)

// Supported asset classes
const (
	AssetClassFX          = "fx"
	AssetClassEquities    = "eq"
	AssetClassCrypto      = "crypto"
	AssetClassCommodities = "cmdty"
	AssetClassCash        = "cash"
	AssetClassBonds       = "bond"
)

// TradeSide represents the side of a trade (buy or sell).
const (
	TradeSideBuy  = "buy"
	TradeSideSell = "sell"
)

// TradeBlotter represents a service for managing trades.
type TradeBlotter struct {
	trades             []Trade
	tradesByID         map[string]*Trade
	tradesByTicker     map[string][]Trade
	tradesByAssetClass map[string][]Trade
	currentSeqNum      int // used as a pointer to the head of the blotter
	db                 dal.Database
}

// NewBlotter creates a new TradeBlotter instance.
func NewBlotter(db dal.Database) *TradeBlotter {
	var currentSeqNum int
	err := db.Get(string(types.HeadSequenceKey), currentSeqNum)
	if err != nil {
		currentSeqNum = -1
	}

	return &TradeBlotter{
		trades:             []Trade{},
		tradesByID:         make(map[string]*Trade),
		tradesByTicker:     make(map[string][]Trade),
		tradesByAssetClass: make(map[string][]Trade),
		currentSeqNum:      currentSeqNum,
		db:                 db,
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
		err = b.AddTrade(trade)
		if err != nil {
			return err
		}
	}

	logging.GetLogger().Info("Loaded trades from database")

	return nil
}

// AddTrade adds a new trade to the blotter and writes it to the database.
func (b *TradeBlotter) AddTrade(trade Trade) error {
	trade.SeqNum = b.getNextSeqNum()

	// Write trade to the database
	tradeKey := generateTradeKey(trade)
	err := b.db.Put(tradeKey, trade)
	if err != nil {
		return err
	}

	// Check if the trade already exists
	if existingTrade, exists := b.tradesByID[trade.TradeID]; exists {
		// Remove the existing trade from the trades slice
		for i, t := range b.trades {
			if t.TradeID == existingTrade.TradeID {
				b.trades = append(b.trades[:i], b.trades[i+1:]...)
				break
			}
		}
	}

	// Add trade to the trades slice and indexes
	b.trades = append(b.trades, trade)
	b.tradesByID[trade.TradeID] = &trade
	b.tradesByTicker[trade.Ticker] = append(b.tradesByTicker[trade.Ticker], trade)
	b.tradesByAssetClass[trade.AssetClass] = append(b.tradesByAssetClass[trade.AssetClass], trade)

	return nil
}

// RemoveTrade removes a trade from the blotter and deletes it from the database.
func (b *TradeBlotter) RemoveTrade(tradeID string) error {
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
	b.tradesByAssetClass[trade.AssetClass] = removeTradeFromSlice(b.tradesByAssetClass[trade.AssetClass], tradeID)

	// Remove trade from the database
	tradeKey := generateTradeKey(*trade)
	err := b.db.Delete(tradeKey)
	if err != nil {
		logging.GetLogger().Error("Failed to delete trade from database", err)
		return err
	}

	return nil
}

// GetTrades returns all trades in the blotter.
func (b *TradeBlotter) GetTrades() []Trade {
	return b.trades
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

// generateTradeKey generates a unique key for the trade.
func generateTradeKey(trade Trade) string {
	return fmt.Sprintf("%s%s:%s:%s", types.TradeKeyPrefix, trade.AssetClass, trade.Ticker, trade.TradeID)
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
	b.db.Put(string(types.HeadSequenceKey), seqNum)
}

// Trade represents a trade in the blotter.
type Trade struct {
	TradeID       string  `json:"TradeID"`       // Unique identifier for the trade
	TradeDate     string  `json:"TradeDate"`     // Date and time of the trade
	Ticker        string  `json:"Ticker"`        // Ticker symbol of the asset
	Side          string  `json:"Side"`          // Buy or Sell
	Quantity      float64 `json:"Quantity"`      // Quantity of the asset
	AssetClass    string  `json:"AssetClass"`    // e.g., Equity, Fixed Income, Commodity
	AssetSubClass string  `json:"AssetSubclass"` // e.g., Stock, Bond, Gold
	Price         float64 `json:"Price"`         // Price per unit of the asset
	Yield         float64 `json:"Yield"`         // Yield of the asset
	Trader        string  `json:"Trader"`        // Trader who executed the trade
	Broker        string  `json:"Broker"`        // Broker who executed the trade
	SeqNum        int     `json:"SeqNum"`        // Sequence number
}

// NewTrade creates a new Trade instance.
func NewTrade(side string, quantity float64, assetClass, assetSubClass, ticker, trader, broker string, price float64, yield float64, tradeDate time.Time) (*Trade, error) {
	if !isValidAssetClass(assetClass) {
		return nil, errors.New("unsupported asset class")
	}

	if !isValidSide(side) {
		return nil, errors.New("side must be either 'buy' or 'sell'")
	}

	return &Trade{
		TradeID:       uuid.New().String(),
		TradeDate:     tradeDate.Format(time.RFC3339),
		Ticker:        ticker,
		Side:          side,
		Quantity:      quantity,
		AssetClass:    assetClass,
		AssetSubClass: assetSubClass,
		Price:         price,
		Yield:         yield,
		Trader:        trader,
		Broker:        broker,
	}, nil
}

// isValidAssetClass checks if the provided asset class is supported.
func isValidAssetClass(assetClass string) bool {
	switch assetClass {
	case AssetClassFX, AssetClassEquities, AssetClassCrypto, AssetClassCommodities, AssetClassCash, AssetClassBonds:
		return true
	default:
		return false
	}
}

// isValidSide checks if the provided side is valid.
func isValidSide(side string) bool {
	return side == TradeSideBuy || side == TradeSideSell
}
