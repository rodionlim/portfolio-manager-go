package dividends

import (
	"fmt"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
)

type DividendsManager struct {
	db      dal.Database
	mdata   mdata.MarketDataManager
	rdata   rdata.ReferenceManager
	blotter blotter.TradeGetter
}

type Dividends struct {
	ExDate         string
	Amount         float64
	AmountPerShare float64
}

func NewDividendsManager(db dal.Database, mdata mdata.MarketDataManager, rdata rdata.ReferenceManager, blotter blotter.TradeGetter) *DividendsManager {
	return &DividendsManager{
		db:      db,
		mdata:   mdata,
		rdata:   rdata,
		blotter: blotter,
	}
}

func (dm *DividendsManager) CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error) {
	// Get dividends.sg ticker from ticker reference
	tickerRef, err := dm.rdata.GetTicker(ticker)
	if err != nil {
		return nil, err
	}

	if tickerRef.DividendsSgTicker == "" && !common.IsSSB(ticker) {
		return nil, fmt.Errorf("no dividends.sg ticker found for the given ticker %s", ticker)
	}

	// Fetch dividends data from mdata service
	dividends, err := dm.mdata.GetDividendsMetadataFromTickerRef(tickerRef)
	if err != nil {
		return nil, err
	}

	// Fetch trades for the ticker from the blotter
	trades, err := dm.blotter.GetTradesByTicker(ticker)
	if err != nil {
		return nil, err
	}

	// Calculate total dividends based on trades and dividends data
	var allDividends []Dividends
	for _, dividend := range dividends {
		// If dividend.ExDate is in the future (after today), skip
		if common.IsFutureDate(dividend.ExDate) {
			continue
		}

		// Use binary search to find the first trade with TradeDate >= ExDate
		idx := SearchEarliestTradeIndexAfterExDate(trades, dividend.ExDate)

		// Calculate total dividend amount for trades with TradeDate < ExDate
		totalQty := 0.0
		for i := 0; i < idx; i++ {
			if trades[i].Side == blotter.TradeSideBuy {
				totalQty += trades[i].Quantity
			} else {
				totalQty -= trades[i].Quantity
			}
		}

		totalAmount := totalQty * dividend.Amount * (1 - dividend.WithholdingTax)
		if totalAmount > 0 {
			allDividends = append(allDividends, Dividends{
				ExDate:         dividend.ExDate,
				Amount:         totalAmount,
				AmountPerShare: dividend.Amount,
			})
		}
	}

	return allDividends, nil
}
