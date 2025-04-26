package dividends

import (
	"fmt"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"strings"
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
	Qty            float64
}

func NewDividendsManager(db dal.Database, mdata mdata.MarketDataManager, rdata rdata.ReferenceManager, blotter blotter.TradeGetter) *DividendsManager {
	return &DividendsManager{
		db:      db,
		mdata:   mdata,
		rdata:   rdata,
		blotter: blotter,
	}
}

func (dm *DividendsManager) CalculateDividendsForAllTickers() (map[string][]Dividends, error) {
	// Get all tickers from the database
	tickers, err := dm.blotter.GetAllTickers()
	if err != nil {
		return nil, err
	}

	dividendsMap := make(map[string][]Dividends)
	for _, ticker := range tickers {
		dividends, err := dm.CalculateDividendsForSingleTicker(ticker)
		if err != nil {
			return nil, err
		}
		dividendsMap[ticker] = dividends
	}

	return dividendsMap, nil
}

func (dm *DividendsManager) CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error) {
	ticker = strings.ToUpper(ticker)

	// Get dividends.sg ticker from ticker reference
	tickerRef, err := dm.rdata.GetTicker(ticker)
	if err != nil {
		return nil, err
	}

	// If no dividends.sg ticker found, check if the ticker is a special case
	if tickerRef.DividendsSgTicker == "" && (!common.IsSSB(ticker) && !common.IsSgTBill(ticker)) {
		// yahoo finance also has historical dividends
		if tickerRef.YahooTicker == "" {
			return nil, fmt.Errorf("no dividends.sg / yahoo finance ticker found for the given ticker %s", ticker)
		}
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
		for i := range idx {
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
				Qty:            totalQty,
			})
		}
	}

	return allDividends, nil
}
