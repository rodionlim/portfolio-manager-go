package dividends

import (
	"portfolio-manager/internal/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

// type mockTradeBlotter struct{}

// func (m *mockTradeBlotter) GetTradesByTicker(ticker string) ([]blotter.Trade, error) {
// 	return []blotter.Trade{
// 		{TradeDate: "2022-12-31", Quantity: 100},
// 		{TradeDate: "2023-01-15", Quantity: 200},
// 	}, nil
// }

func TestCalculateDividendsForSingleTicker(t *testing.T) {
	db := &mocks.MockDatabase{}
	rdata := &mocks.MockReferenceManager{}
	// blotter := &mockTradeBlotter{}
	mdata := &mocks.MockMarketDataManager{}

	dm := NewDividendsManager(db, mdata, rdata, nil)

	dividends, err := dm.CalculateDividendsForSingleTicker("AAPL")
	assert.NoError(t, err)
	assert.Len(t, dividends, 2)

	expectedDividends := []Dividends{
		{ExDate: "2023-01-01", Amount: 90.0, AmountPerShare: 1.0},
		{ExDate: "2023-02-01", Amount: 360.0, AmountPerShare: 2.0},
	}

	assert.Equal(t, expectedDividends, dividends)
}
