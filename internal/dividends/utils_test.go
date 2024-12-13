package dividends_test

import (
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dividends"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchEarliestTradeIndexAfterExDate(t *testing.T) {
	trades := []blotter.Trade{
		{TradeDate: "2015-08-10T09:00:00Z", Quantity: 100},
		{TradeDate: "2015-09-10T09:00:00Z", Quantity: 200},
		{TradeDate: "2023-02-01T09:00:00Z", Quantity: 150},
		{TradeDate: "2023-02-15T09:00:00Z", Quantity: 300},
	}

	idx := dividends.SearchEarliestTradeIndexAfterExDate(trades, "2023-02-16")

	assert.Equal(t, 4, idx)
}

func TestSearchEarliestTradeIndexAfterExDateForNoMatch(t *testing.T) {
	trades := []blotter.Trade{
		{TradeDate: "2015-08-10T09:00:00Z", Quantity: 100},
		{TradeDate: "2015-09-10T09:00:00Z", Quantity: 200},
		{TradeDate: "2023-02-01T09:00:00Z", Quantity: 150},
		{TradeDate: "2023-02-15T09:00:00Z", Quantity: 300},
	}

	idx := dividends.SearchEarliestTradeIndexAfterExDate(trades, "2002-02-01")

	assert.Equal(t, 0, idx)
}
