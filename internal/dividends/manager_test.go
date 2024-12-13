package dividends

import (
	"portfolio-manager/internal/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDividendsForSingleTicker(t *testing.T) {
	db := mocks.NewMockDatabase()
	rdata := mocks.NewMockReferenceManager()
	blotterTradeGetter := mocks.NewMockTradeGetterBlotter(db)
	mdata := mocks.NewMockMarketDataManager()

	dm := NewDividendsManager(db, mdata, rdata, blotterTradeGetter)

	dividends, err := dm.CalculateDividendsForSingleTicker("AAPL")
	assert.NoError(t, err)
	assert.Len(t, dividends, 2)

	expectedDividends := []Dividends{
		{ExDate: "2023-01-01", Amount: 90.0, AmountPerShare: 1.0},
		{ExDate: "2023-02-01", Amount: 360.0, AmountPerShare: 2.0},
	}

	assert.Equal(t, expectedDividends, dividends)
}
