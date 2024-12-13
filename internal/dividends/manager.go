package dividends

import (
	"fmt"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"sort"
	"time"
)

type DividendsManager struct {
	db      dal.Database
	rdata   rdata.ReferenceManager
	blotter *blotter.TradeBlotter
	mdata   *mdata.Manager
}

type Dividends struct {
	ExDate         string
	Amount         float64
	AmountPerShare float64
}

func NewDividendsManager(db dal.Database, rdata rdata.ReferenceManager, blotter *blotter.TradeBlotter, mdata *mdata.Manager) *DividendsManager {
	return &DividendsManager{
		db:      db,
		rdata:   rdata,
		blotter: blotter,
		mdata:   mdata,
	}
}

func (dm *DividendsManager) CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error) {
	// Get dividends.sg ticker from ticker reference
	tickerRef, err := dm.rdata.GetTicker(ticker)
	if err != nil {
		return nil, err
	}

	if tickerRef.DividendsSgTicker == "" {
		return nil, fmt.Errorf("no dividends.sg ticker found for the given ticker %s", ticker)
	}

	// Fetch dividends data from mdata service
	dividends, err := dm.mdata.GetDividendsMetadata(tickerRef.DividendsSgTicker)
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
		exDate, err := time.Parse("2006-01-02", dividend.ExDate)
		if err != nil {
			return nil, err
		}

		// Use binary search to find the first trade with TradeDate >= ExDate
		idx := sort.Search(len(trades), func(i int) bool {
			tradeDate, _ := time.Parse("2006-01-02", trades[i].TradeDate)
			return tradeDate.After(exDate) || tradeDate.Equal(exDate)
		})

		// Calculate total amount for trades with TradeDate < ExDate
		totalAmount := 0.0
		for i := 0; i < idx; i++ {
			totalAmount += trades[i].Quantity * dividend.Amount * (1 - dividend.WithholdingTax)
		}

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
