package portfolio

import (
	"fmt"
	"testing"
	"time"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestPortfolio() (*Portfolio, *mocks.MockDatabase) {
	mockDB := new(mocks.MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("Get", string(types.HeadSequenceBlotterKey), mock.Anything).Return(nil)
	mockDB.On("Get", fmt.Sprintf("%s:%s", types.ReferenceDataKeyPrefix, "C31.SI"), mock.AnythingOfType("*rdata.TickerReference")).Return(nil).Run(func(args mock.Arguments) {
		ref := args.Get(1).(*rdata.TickerReference)
		*ref = rdata.TickerReference{
			ID:         "C31.SI",
			Name:       "CapitaLand",
			Ccy:        "SGD",
			AssetClass: rdata.AssetClassEquities,
		}
	})
	mockDB.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("*rdata.TickerReference")).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Delete", mock.Anything, mock.Anything).Return(nil)

	rdataMgr, _ := rdata.NewManager(mockDB, "")
	mdataMgr, _ := mdata.NewManager(mockDB, rdataMgr)
	dividendsMgr := dividends.NewDividendsManager(mockDB, mdataMgr, rdataMgr, nil)

	return NewPortfolio(mockDB, mdataMgr, rdataMgr, dividendsMgr), mockDB
}

func createTestPortfolioWithDb(mockDB *mocks.MockDatabase) *Portfolio {
	rdataMgr, _ := rdata.NewManager(mockDB, "")
	mdataMgr, _ := mdata.NewManager(mockDB, rdataMgr)
	dividendsMgr := dividends.NewDividendsManager(mockDB, mdataMgr, rdataMgr, nil)

	return NewPortfolio(mockDB, mdataMgr, rdataMgr, dividendsMgr)
}

func TestNewPortfolio(t *testing.T) {
	p, _ := createTestPortfolio()

	assert.NotNil(t, p)
	assert.Equal(t, 0, p.currentSeqNum)
	assert.Empty(t, p.positions)
}

func TestUpdatePosition(t *testing.T) {
	p, _ := createTestPortfolio()

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		time.Now(),
	)

	err := p.updatePosition(trade)
	assert.NoError(t, err)

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(100), position.Qty)
}

func TestAvgPriceOnUpdatePosition(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 150.0, 1, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 50, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 200.0, 1, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.InDelta(t, 166.67, position.AvgPx, 0.01) // Allowing a small delta of 0.01
}

func TestClosingPosition(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 150.0, 1, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideSell, 100, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 200.0, 1, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(0), position.Qty)
	assert.Equal(t, float64(0), position.PnL)
	assert.Equal(t, float64(0), position.Mv)
}

func TestUpdateBuyAndSellPosition(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 150.0, 1, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideSell, 50, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 200.0, 1, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(50), position.Qty)
	assert.InDelta(t, 100, position.AvgPx, 0.01)
}

func TestGetPositions(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 150.0, 1, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 50, "GOOGL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 2500.0, 1, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 75, "MSFT", "book2", "broker1", "cdp", blotter.StatusOpen, "", 300.0, 1, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	// Test GetPositions for book1
	book1Positions, err := p.GetPositions("book1")
	assert.NoError(t, err)
	assert.Len(t, book1Positions, 2)

	// Test GetAllPositions
	allPositions, err := p.GetAllPositions()
	assert.NoError(t, err)
	assert.Len(t, allPositions, 3)

	// Test GetPositions for "" book
	emptyBookPositions, err := p.GetPositions("")
	assert.NoError(t, err)
	assert.Len(t, emptyBookPositions, 3)
}

func TestLoadPositions(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("*rdata.TickerReference")).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.PositionKeyPrefix)).Return([]string{
		string(types.PositionKeyPrefix) + ":book1:AAPL",
	}, nil)

	position := &Position{
		Ticker: "AAPL",
		Book:   "book1",
		Qty:    100,
		Mv:     15000,
		PnL:    1000,
		AvgPx:  150.0,
	}
	mockDB.On("Get", string(types.PositionKeyPrefix)+":book1:AAPL", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		pos := args.Get(1).(*Position)
		*pos = *position
	})

	p := createTestPortfolioWithDb(mockDB)
	err := p.LoadPositions()
	assert.NoError(t, err)

	loadedPosition, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, loadedPosition)
	assert.Equal(t, position.Qty, loadedPosition.Qty)
	assert.Equal(t, position.Mv, loadedPosition.Mv)
	assert.Equal(t, position.PnL, loadedPosition.PnL)
	assert.Equal(t, position.AvgPx, loadedPosition.AvgPx)
}

func TestSubscribeToBlotter(t *testing.T) {
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		time.Now(),
	)

	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, 100.0, position.Qty)
}

func TestSubscribeToBlotterWithBuyAndClosePosition(t *testing.T) {
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		700,
		"C31.SI",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		3.5068,
		1,
		0.0,
		time.Now(),
	)
	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	trade, _ = blotter.NewTrade(
		blotter.TradeSideSell,
		700,
		"C31.SI",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		3.8231,
		1,
		0.0,
		time.Now(),
	)
	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("book1", "C31.SI")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(0), position.Qty)
	assert.Equal(t, 221, int(position.PnL))
}

func TestSubscribeToBlotterWithBuyAndClosePositionTwice(t *testing.T) {
	// This was initially a bug with the p&l calculation
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		700,
		"C31.SI",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		3.5068,
		1,
		0.0,
		time.Now(),
	)
	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	trade, _ = blotter.NewTrade(
		blotter.TradeSideSell,
		700,
		"C31.SI",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		3.8231,
		1,
		0.0,
		time.Now(),
	)
	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	trade, _ = blotter.NewTrade(
		blotter.TradeSideBuy,
		600,
		"C31.SI",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		3.6194,
		1,
		0.0,
		time.Now(),
	)
	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	trade, _ = blotter.NewTrade(
		blotter.TradeSideSell,
		600,
		"C31.SI",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		3.5507,
		1,
		0.0,
		time.Now(),
	)
	err = blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("book1", "C31.SI")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(0), position.Qty)
	assert.Equal(t, 180, int(position.PnL))
}

func TestSubscribeToBlotterWithTradeDeletion(t *testing.T) {
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		time.Now(),
	)

	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	err = blotterSvc.RemoveTrades([]string{trade.TradeID})
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), position.Qty)
	assert.Equal(t, float64(0), position.AvgPx)
}

func TestSubscribeToBlotterWithTradeUpdate(t *testing.T) {
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		time.Now(),
	)

	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	trade, _ = blotter.NewTradeWithID(
		trade.TradeID,
		blotter.TradeSideBuy,
		200,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		1,
		time.Now(),
	)

	err = blotterSvc.UpdateTrade(*trade)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.Equal(t, 200.0, position.Qty)
	assert.Equal(t, 150.0, position.AvgPx)
}

func TestSubscribeToBlotterWithBookUpdate(t *testing.T) {
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		time.Now(),
	)

	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	trade2, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		300,
		"AAPL",
		"book1",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		time.Now(),
	)

	err = blotterSvc.AddTrade(*trade2)
	assert.NoError(t, err)

	trade, _ = blotter.NewTradeWithID(
		trade.TradeID,
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"book2",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		1,
		time.Now(),
	)

	err = blotterSvc.UpdateTrade(*trade)
	assert.NoError(t, err)

	trade2, _ = blotter.NewTradeWithID(
		trade2.TradeID,
		blotter.TradeSideBuy,
		300,
		"AAPL",
		"book2",
		"broker1",
		"cdp",
		blotter.StatusOpen,
		"",
		150.0,
		1,
		0.0,
		1,
		time.Now(),
	)

	err = blotterSvc.UpdateTrade(*trade2)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("book2", "AAPL")
	assert.NoError(t, err)
	assert.Equal(t, 400.0, position.Qty)

	position, err = p.GetPosition("book1", "AAPL")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, position.Qty)
}

// Helper function to handle error in test data setup
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
