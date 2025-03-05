package dividends

import (
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() (*DividendsManager, *mocks.MockMarketDataManager, *mocks.MockTradeGetterBlotter, error) {
	db := mocks.NewMockDatabase()
	mdataMgr := mocks.NewMockMarketDataManager()
	rdataMgr := mocks.NewMockReferenceManager()
	blotterMgr := mocks.NewMockTradeGetterBlotter()

	mdataMgr.SetDividendMetadata("AAPL", []types.DividendsMetadata{
		{Ticker: "AAPL", ExDate: "2023-01-01", Amount: 1.0, WithholdingTax: 0.3},
		{Ticker: "AAPL", ExDate: "2023-02-01", Amount: 2.0, WithholdingTax: 0.3},
		{Ticker: "AAPL", ExDate: "2099-02-01", Amount: 5.0, WithholdingTax: 0.3},
	})

	rdataMgr.AddTicker(rdata.TickerReference{
		ID:                "AAPL",
		DividendsSgTicker: "AAPL",
	})

	rdataMgr.AddTicker(rdata.TickerReference{
		ID: "SBFEB50",
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
		{ExDate: "2023-01-01", Amount: 70.0, AmountPerShare: 1.0, Qty: 100},
		{ExDate: "2023-02-01", Amount: 420.0, AmountPerShare: 2.0, Qty: 300},
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
		{ExDate: "2023-01-01", Amount: 70.0, AmountPerShare: 1.0, Qty: 100},
	}

	assert.Equal(t, expectedDividends, dividends)
}

func TestCalculateDividendsForSSB(t *testing.T) {
	dm, mdataMgr, blotterMgr, err := setup()
	assert.NoError(t, err)

	blotterMgr.SetTrades("SBFEB50", []blotter.Trade{
		{Ticker: "SBFEB50", TradeDate: "2050-02-01", Quantity: 100, TradeID: "1", Side: blotter.TradeSideBuy},
	})

	mdataMgr.SetDividendMetadata("SBFEB50",
		[]types.DividendsMetadata{
			{Ticker: "SBFEB50", ExDate: "2050-08-01", Amount: 1.98, Interest: 1.98, AvgInterest: 1.98, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2051-02-01", Amount: 1.98, Interest: 1.98, AvgInterest: 1.98, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2051-08-01", Amount: 1.98, Interest: 1.98, AvgInterest: 1.98, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2052-02-01", Amount: 1.98, Interest: 1.98, AvgInterest: 1.98, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2052-08-01", Amount: 1.98, Interest: 1.98, AvgInterest: 1.98, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2053-02-01", Amount: 1.98, Interest: 1.98, AvgInterest: 1.98, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2053-08-01", Amount: 2.09, Interest: 2.09, AvgInterest: 2.01, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2054-02-01", Amount: 2.09, Interest: 2.09, AvgInterest: 2.01, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2054-08-01", Amount: 2.16, Interest: 2.16, AvgInterest: 2.04, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2055-02-01", Amount: 2.16, Interest: 2.16, AvgInterest: 2.04, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2055-08-01", Amount: 2.21, Interest: 2.21, AvgInterest: 2.06, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2056-02-01", Amount: 2.21, Interest: 2.21, AvgInterest: 2.06, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2056-08-01", Amount: 2.3, Interest: 2.3, AvgInterest: 2.1, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2057-02-01", Amount: 2.3, Interest: 2.3, AvgInterest: 2.1, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2057-08-01", Amount: 2.38, Interest: 2.38, AvgInterest: 2.13, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2058-02-01", Amount: 2.38, Interest: 2.38, AvgInterest: 2.13, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2058-08-01", Amount: 2.46, Interest: 2.46, AvgInterest: 2.16, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2059-02-01", Amount: 2.46, Interest: 2.46, AvgInterest: 2.16, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2059-08-01", Amount: 2.53, Interest: 2.53, AvgInterest: 2.2, WithholdingTax: 0},
			{Ticker: "SBFEB50", ExDate: "2060-02-01", Amount: 2.53, Interest: 2.53, AvgInterest: 2.2, WithholdingTax: 0},
		})

	dividends, err := dm.CalculateDividendsForSingleTicker("SBFEB50")
	assert.NoError(t, err)
	assert.Len(t, dividends, 0)
}
