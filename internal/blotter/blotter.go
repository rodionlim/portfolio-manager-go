package blotter

import (
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
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
	AssetClassCash        = "cash"
	AssetClassBonds       = "bond"
)

// TradeSide represents the side of a trade (buy or sell).
const (
	TradeSideBuy  = "buy"
	TradeSideSell = "sell"
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
			series.New([]string{}, series.String, "TradeID"),
			series.New([]string{}, series.String, "TradeDate"),
			series.New([]string{}, series.String, "Ticker"),
			series.New([]string{}, series.String, "Side"),
			series.New([]float64{}, series.Float, "Quantity"),
			series.New([]string{}, series.String, "AssetClass"),
			series.New([]string{}, series.String, "AssetSubClass"),
			series.New([]float64{}, series.Float, "Price"),
			series.New([]float64{}, series.Float, "Yield"),
		),
		db: db,
	}
}

func (b *Blotter) LoadFromDB() error {
	tradeKeys, err := b.db.GetAllKeysWithPrefix("trade:")
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
func (b *Blotter) AddTrade(trade Trade) error {
	// Write trade to the database
	tradeKey := generateTradeKey(trade)
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
			if b.trades.Elem(i, common.IndexOf(b.trades.Names(), "TradeID")).String() == trade.TradeID {
				b.trades = b.trades.Drop(i)
				break
			}
		}
	}

	// Add trade to the DataFrame
	newRow := dataframe.LoadStructs([]Trade{trade})
	if newRow.Error() != nil {
		return newRow.Error()
	}

	b.trades = b.trades.RBind(newRow)

	return nil
}

// RemoveTrade removes a trade from the blotter and deletes it from the database.
func (b *Blotter) RemoveTrade(tradeID string) error {
	// Check if the trade exists in the DataFrame
	row := b.trades.Filter(dataframe.F{
		Colname:    "TradeID",
		Comparator: "==",
		Comparando: tradeID,
	})
	if row.Nrow() == 0 {
		return errors.New("trade not found")
	}

	// Remove trade from the DataFrame
	b.trades = b.trades.Filter(dataframe.F{
		Colname:    "TradeID",
		Comparator: "!=",
		Comparando: tradeID,
	})

	// Remove trade from the database
	trade, err := b.createTradeFromRow(row)
	if err != nil {
		return err
	}

	tradeKey := generateTradeKey(*trade)
	err = b.db.Delete(tradeKey)
	if err != nil {
		logging.GetLogger().Error("Failed to delete trade from database", err)

		// Add the trade back to the DataFrame (ROLLBACK)
		newRow := dataframe.LoadStructs([]Trade{*trade})
		b.trades = b.trades.RBind(newRow)
	}

	return nil
}

// GetTrades returns all trades in the blotter.
func (b *Blotter) GetTradesDf() dataframe.DataFrame {
	return b.trades
}

// GetTradeByID returns a trade with the given ID.
func (b *Blotter) GetTradeByID(tradeID string) (*Trade, error) {
	row := b.trades.Filter(dataframe.F{
		Colname:    "TradeID",
		Comparator: "==",
		Comparando: tradeID,
	})
	if row.Nrow() == 0 {
		return nil, errors.New("trade not found")
	}

	return b.createTradeFromRow(row)
}

// GetTradesByTicker returns all trades for the given ticker.
func (b *Blotter) GetTradesByTicker(ticker string) ([]Trade, error) {
	rows := b.trades.Filter(dataframe.F{
		Colname:    "Ticker",
		Comparator: "==",
		Comparando: ticker,
	})
	if rows.Nrow() == 0 {
		return nil, errors.New("no trades found for the given ticker")
	}

	var trades []Trade
	for i := 0; i < rows.Nrow(); i++ {
		trade, _ := b.createTradeFromRow(rows.Subset(i))
		trades = append(trades, *trade)
	}

	return trades, nil
}

// generateTradeKey generates a unique key for the trade.
func generateTradeKey(trade Trade) string {
	return fmt.Sprintf("trade:%s:%s:%s", trade.AssetClass, trade.Ticker, trade.TradeID)
}

// createTradeFromRow creates a Trade instance from a dataframe row.
func (b *Blotter) createTradeFromRow(row dataframe.DataFrame) (*Trade, error) {
	if row.Nrow() > 1 {
		return nil, errors.New("more than one row found")
	}

	dfCols := b.trades.Names()

	return &Trade{
		TradeID:       row.Elem(0, common.IndexOf(dfCols, "TradeID")).String(),
		TradeDate:     row.Elem(0, common.IndexOf(dfCols, "TradeDate")).String(),
		Ticker:        row.Elem(0, common.IndexOf(dfCols, "Ticker")).String(),
		Side:          row.Elem(0, common.IndexOf(dfCols, "Side")).String(),
		Quantity:      row.Elem(0, common.IndexOf(dfCols, "Quantity")).Float(),
		AssetClass:    row.Elem(0, common.IndexOf(dfCols, "AssetClass")).String(),
		AssetSubClass: row.Elem(0, common.IndexOf(dfCols, "AssetSubClass")).String(),
		Price:         row.Elem(0, common.IndexOf(dfCols, "Price")).Float(),
		Yield:         row.Elem(0, common.IndexOf(dfCols, "Yield")).Float(),
	}, nil
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
}

// NewTrade creates a new Trade instance.
func NewTrade(side string, quantity float64, assetClass, assetSubClass, ticker string, price float64, yield float64, tradeDate time.Time) (*Trade, error) {
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
