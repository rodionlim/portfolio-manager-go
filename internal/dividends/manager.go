package dividends

import (
	"fmt"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/reference"
	"portfolio-manager/pkg/mdata"
)

type DividendsManager struct {
	db      dal.Database
	rdata   *reference.ReferenceManager
	blotter *blotter.TradeBlotter
	mdata   *mdata.Manager
}

type Dividends struct {
	ExDate         string
	Amount         float64
	AmountPerShare float64
}

func NewDividendsManager(db dal.Database, rdata *reference.ReferenceManager, blotter *blotter.TradeBlotter, mdata *mdata.Manager) *DividendsManager {
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
	for _, trade := range trades {
		for _, dividend := range dividends {
			if trade.TradeDate <= dividend.ExDate {
				allDividends = append(allDividends, Dividends{
					ExDate:         dividend.ExDate,
					Amount:         trade.Quantity * dividend.Amount,
					AmountPerShare: dividend.Amount,
				})
			}
		}
	}

	return allDividends, nil
}
