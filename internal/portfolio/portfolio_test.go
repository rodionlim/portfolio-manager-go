package portfolio

import (
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
		"trader1",
		"broker1",
		"cdp",
		150.0,
		0.0,
		time.Now(),
	)

	err := p.updatePosition(trade)
	assert.NoError(t, err)

	position, err := p.GetPosition("trader1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(100), position.Qty)
}

func TestAvgPriceOnUpdatePosition(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "trader1", "broker1", "cdp", 150.0, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 50, "AAPL", "trader1", "broker1", "cdp", 200.0, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position, err := p.GetPosition("trader1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.InDelta(t, 166.67, position.AvgPx, 0.01) // Allowing a small delta of 0.01
}

func TestUpdateBuyAndSellPosition(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "trader1", "broker1", "cdp", 150.0, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideSell, 50, "AAPL", "trader1", "broker1", "cdp", 200.0, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position, err := p.GetPosition("trader1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, float64(50), position.Qty)
	assert.InDelta(t, 100, position.AvgPx, 0.01)
}

func TestGetPositions(t *testing.T) {
	p, _ := createTestPortfolio()

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "trader1", "broker1", "cdp", 150.0, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 50, "GOOGL", "trader1", "broker1", "cdp", 2500.0, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 75, "MSFT", "trader2", "broker1", "cdp", 300.0, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	// Test GetPositions for trader1
	trader1Positions, err := p.GetPositions("trader1")
	assert.NoError(t, err)
	assert.Len(t, trader1Positions, 2)

	// Test GetAllPositions
	allPositions, err := p.GetAllPositions()
	assert.NoError(t, err)
	assert.Len(t, allPositions, 3)
}

func TestLoadPositions(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("*rdata.TickerReference")).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.PositionKeyPrefix)).Return([]string{
		string(types.PositionKeyPrefix) + ":trader1:AAPL",
	}, nil)

	position := &Position{
		Ticker: "AAPL",
		Trader: "trader1",
		Qty:    100,
		Mv:     15000,
		PnL:    1000,
		AvgPx:  150.0,
	}
	mockDB.On("Get", string(types.PositionKeyPrefix)+":trader1:AAPL", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		pos := args.Get(1).(*Position)
		*pos = *position
	})

	p := createTestPortfolioWithDb(mockDB)
	err := p.LoadPositions()
	assert.NoError(t, err)

	loadedPosition, err := p.GetPosition("trader1", "AAPL")
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
		"trader1",
		"broker1",
		"cdp",
		150.0,
		0.0,
		time.Now(),
	)

	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("trader1", "AAPL")
	assert.NoError(t, err)
	assert.NotNil(t, position)
	assert.Equal(t, 100.0, position.Qty)
}

func TestSubscribeToBlotterWithTradeDeletion(t *testing.T) {
	p, mockDB := createTestPortfolio()
	blotterSvc := blotter.NewBlotter(mockDB)

	p.SubscribeToBlotter(blotterSvc)

	trade, _ := blotter.NewTrade(
		blotter.TradeSideBuy,
		100,
		"AAPL",
		"trader1",
		"broker1",
		"cdp",
		150.0,
		0.0,
		time.Now(),
	)

	err := blotterSvc.AddTrade(*trade)
	assert.NoError(t, err)

	err = blotterSvc.RemoveTrades([]string{trade.TradeID})
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("trader1", "AAPL")
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
		"trader1",
		"broker1",
		"cdp",
		150.0,
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
		"trader1",
		"broker1",
		"cdp",
		150.0,
		0.0,
		time.Now(),
	)

	err = blotterSvc.UpdateTrade(*trade)
	assert.NoError(t, err)

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	position, err := p.GetPosition("trader1", "AAPL")
	assert.NoError(t, err)
	assert.Equal(t, 200.0, position.Qty)
	assert.Equal(t, 150.0, position.AvgPx)
}

// Helper function to handle error in test data setup
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
