package blotter

import (
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/google/uuid"
)

// Supported asset classes
const (
	AssetClassFX          = "fx"
	AssetClassEquities    = "eq"
	AssetClassCrypto      = "crypto"
	AssetClassCommodities = "cmdty"
)

// Blotter represents a service for managing trades.
type Blotter struct {
	trades dataframe.DataFrame
	db     dal.Database
}

// NewBlotter creates a new Blotter instance.
func NewBlotter(db dal.Database) *Blotter {
	return &Blotter{
		trades: dataframe.New(
			series.New([]string{}, series.String, "side"),
			series.New([]float64{}, series.Float, "quantity"),
			series.New([]string{}, series.String, "asset_class"),
			series.New([]string{}, series.String, "asset_subclass"),
			series.New([]float64{}, series.Float, "price"),
			series.New([]string{}, series.String, "ticker"),
			series.New([]time.Time{}, series.String, "trade_date"),
			series.New([]string{}, series.String, "trade_id"),
		),
		db: db,
	}
}

// NewBlotterFromDB creates a new Blotter instance and loads trades from the database.
func NewBlotterFromDB(db dal.Database) (*Blotter, error) {
	blotter := NewBlotter(db)

	// Load trades from the database
	tradeKeys, err := db.GetAllKeysWithPrefix("trade:")
	if err != nil {
		return nil, err
	}

	for _, key := range tradeKeys {
		var trade Trade
		err := db.Get(key, &trade)
		if err != nil {
			return nil, err
		}
		blotter.AddTrade(trade)
	}

	return blotter, nil
}

// AddTrade adds a new trade to the blotter and writes it to the database.
func (b *Blotter) AddTrade(trade Trade) error {
	// Write trade to the database
	tradeKey := fmt.Sprintf("trade:%s:%s:%s", trade.AssetClass, trade.Ticker, trade.TradeID)
	err := b.db.Put(tradeKey, trade)
	if err != nil {
		return err
	}

	// Check if the trade already exists in the DataFrame
	existingRow := b.trades.Filter(dataframe.F{
		Colname:    "TradeID",
		Comparator: "==",
		Comparando: trade.TradeID,
	})

	if existingRow.Nrow() > 0 {
		// Update the existing row
		for i := 0; i < b.trades.Nrow(); i++ {
			if b.trades.Elem(i, common.IndexOf(b.trades.Names(), "trade_id")).String() == trade.TradeID {
				b.trades = b.trades.Drop(i)
				break
			}
		}
	}

	// Add trade to the DataFrame
	newRow := dataframe.LoadStructs([]Trade{trade})
	b.trades = b.trades.RBind(newRow)

	return nil
}

// GetTrades returns all trades in the blotter.
func (b *Blotter) GetTrades() dataframe.DataFrame {
	return b.trades
}

// Trade represents a trade in the blotter.
type Trade struct {
	Side          string  `json:"side"`           // Buy or Sell
	Quantity      float64 `json:"quantity"`       // Quantity of the asset
	AssetClass    string  `json:"asset_class"`    // e.g., Equity, Fixed Income, Commodity
	AssetSubClass string  `json:"asset_subclass"` // e.g., Stock, Bond, Gold
	Price         float64 `json:"price"`          // Price per unit of the asset
	Ticker        string  `json:"ticker"`         // Ticker symbol of the asset
	TradeDate     string  `json:"trade_date"`     // Date and time of the trade
	TradeID       string  `json:"trade_id"`       // Unique identifier for the trade
}

// NewTrade creates a new Trade instance.
func NewTrade(side string, quantity float64, assetClass, assetSubClass, ticker string, price float64, tradeDate time.Time) (*Trade, error) {
	if !isValidAssetClass(assetClass) {
		return nil, errors.New("unsupported asset class")
	}

	return &Trade{
		Side:          side,
		Quantity:      quantity,
		AssetClass:    assetClass,
		AssetSubClass: assetSubClass,
		Price:         price,
		Ticker:        ticker,
		TradeDate:     tradeDate.Format(time.RFC3339),
		TradeID:       uuid.New().String(),
	}, nil
}

// isValidAssetClass checks if the provided asset class is supported.
func isValidAssetClass(assetClass string) bool {
	switch assetClass {
	case AssetClassFX, AssetClassEquities, AssetClassCrypto, AssetClassCommodities:
		return true
	default:
		return false
	}
}
