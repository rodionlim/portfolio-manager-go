package dividends

import (
	"portfolio-manager/internal/blotter"
	"sort"
)

func SearchEarliestTradeIndexAfterExDate(trades []blotter.Trade, exDate string) int {
	// Search for a trade with TradeDate >= ExDate and return the corresponding index
	return sort.Search(len(trades), func(i int) bool {
		return trades[i].TradeDate >= exDate
	})
}
