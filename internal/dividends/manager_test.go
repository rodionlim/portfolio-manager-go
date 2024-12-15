package dividends

import (
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() (*DividendsManager, mdata.MarketDataManager, *mocks.MockTradeGetterBlotter, error) {
	db := mocks.NewMockDatabase()
	mdataMgr := mocks.NewMockMarketDataManager()
	rdataMgr := mocks.NewMockReferenceManager()
	blotterMgr := mocks.NewMockTradeGetterBlotter()

	mdataMgr.SetDividendMetadata("AAPL", []types.DividendsMetadata{
		{Ticker: "AAPL", ExDate: "2023-01-01", Amount: 1.0, WithholdingTax: 0.3},
		{Ticker: "AAPL", ExDate: "2023-02-01", Amount: 2.0, WithholdingTax: 0.3},
	})

	rdataMgr.AddTicker(rdata.TickerReference{
		ID:                "AAPL",
		DividendsSgTicker: "AAPL",
	})

	blotterMgr.SetTrades("AAPL", []blotter.Trade{
		{Ticker: "AAPL", TradeDate: "2022-12-31", Quantity: 100, TradeID: "1", Side: blotter.TradeSideBuy},
		{Ticker: "AAPL", TradeDate: "2023-01-15", Quantity: 200, TradeID: "2", Side: blotter.TradeSideBuy},
	})

	dm := NewDividendsManager(db, mdataMgr, rdataMgr, blotterMgr)
	return dm, mdataMgr, blotterMgr, nil
}

func TestCalculateDividendsForSingleTickerOnlyBuys(t *testing.T) {
	dm, _, _, err := setup()
	assert.NoError(t, err)

	dividends, err := dm.CalculateDividendsForSingleTicker("AAPL")
	assert.NoError(t, err)
	assert.Len(t, dividends, 2)

	expectedDividends := []Dividends{
		{ExDate: "2023-01-01", Amount: 70.0, AmountPerShare: 1.0},
		{ExDate: "2023-02-01", Amount: 420.0, AmountPerShare: 2.0},
	}

	assert.Equal(t, expectedDividends, dividends)
}

func TestCalculateDividendsForSingleTickerBuysAndSells(t *testing.T) {
	dm, _, blotterMgr, err := setup()

	blotterMgr.SetTrades("AAPL", []blotter.Trade{
		{Ticker: "AAPL", TradeDate: "2022-12-31", Quantity: 100, TradeID: "1", Side: blotter.TradeSideBuy},
		{Ticker: "AAPL", TradeDate: "2023-01-15", Quantity: 200, TradeID: "2", Side: blotter.TradeSideBuy},
		{Ticker: "AAPL", TradeDate: "2023-01-16", Quantity: 300, TradeID: "3", Side: blotter.TradeSideSell},
	})

	assert.NoError(t, err)

	dividends, err := dm.CalculateDividendsForSingleTicker("AAPL")
	assert.NoError(t, err)
	assert.Len(t, dividends, 1)

	expectedDividends := []Dividends{
		{ExDate: "2023-01-01", Amount: 70.0, AmountPerShare: 1.0},
	}

	assert.Equal(t, expectedDividends, dividends)
}
