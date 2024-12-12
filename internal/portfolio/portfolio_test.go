package portfolio

import (
	"testing"
	"time"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase implements dal.Database for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Put(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockDatabase) Get(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockDatabase) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockDatabase) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	args := m.Called(prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDatabase) Close() error { return nil }

func TestNewPortfolio(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)

	p := NewPortfolio(mockDB, "")
	assert.NotNil(t, p)
	assert.Equal(t, 0, p.currentSeqNum)
	assert.Empty(t, p.positions)
}

func TestUpdatePosition(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	p := NewPortfolio(mockDB, "")

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

	position := p.GetPosition("trader1", "AAPL")
	assert.NotNil(t, position)
	assert.Equal(t, float64(100), position.Qty)
}

func TestAvgPriceOnUpdatePosition(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	p := NewPortfolio(mockDB, "")

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "trader1", "broker1", "cdp", 150.0, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideBuy, 50, "AAPL", "trader1", "broker1", "cdp", 200.0, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position := p.GetPosition("trader1", "AAPL")
	assert.NotNil(t, position)
	assert.InDelta(t, 166.67, position.AvgPx, 0.01) // Allowing a small delta of 0.01
}

func TestUpdateBuyAndSellPosition(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	p := NewPortfolio(mockDB, "")

	// Add multiple trades
	trades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "trader1", "broker1", "cdp", 150.0, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideSell, 50, "AAPL", "trader1", "broker1", "cdp", 200.0, 0.0, time.Now())),
	}

	for _, trade := range trades {
		err := p.updatePosition(trade)
		assert.NoError(t, err)
	}

	position := p.GetPosition("trader1", "AAPL")
	assert.NotNil(t, position)
	assert.Equal(t, float64(50), position.Qty)
	assert.InDelta(t, 100, position.AvgPx, 0.01)
}

func TestGetPositions(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	p := NewPortfolio(mockDB, "")

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
	trader1Positions := p.GetPositions("trader1")
	assert.Len(t, trader1Positions, 2)

	// Test GetAllPositions
	allPositions := p.GetAllPositions()
	assert.Len(t, allPositions, 3)
}

func TestLoadPositions(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
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

	p := NewPortfolio(mockDB, "")
	err := p.LoadPositions()
	assert.NoError(t, err)

	loadedPosition := p.GetPosition("trader1", "AAPL")
	assert.NotNil(t, loadedPosition)
	assert.Equal(t, position.Qty, loadedPosition.Qty)
	assert.Equal(t, position.Mv, loadedPosition.Mv)
	assert.Equal(t, position.PnL, loadedPosition.PnL)
	assert.Equal(t, position.AvgPx, loadedPosition.AvgPx)
}

func TestSubscribeToBlotter(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", string(types.HeadSequencePortfolioKey), mock.Anything).Return(nil)
	mockDB.On("Get", string(types.HeadSequenceBlotterKey), mock.Anything).Return(nil)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix), mock.Anything).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	p := NewPortfolio(mockDB, "")
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

	position := p.GetPosition("trader1", "AAPL")
	assert.NotNil(t, position)
	assert.Equal(t, float64(100), position.Qty)
}

// Helper function to handle error in test data setup
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
