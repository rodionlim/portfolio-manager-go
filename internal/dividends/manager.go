package dividends

import (
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

func NewDividendsManager(db dal.Database, rdata *reference.ReferenceManager, blotter *blotter.TradeBlotter, mdata *mdata.Manager) *DividendsManager {
	return &DividendsManager{
		db:      db,
		rdata:   rdata,
		blotter: blotter,
		mdata:   mdata,
	}
}

func (dm *DividendsManager) CalculateDividendsForSingleTicker(ticker string) (float64, error) {
	// Fetch dividends data from mdata service
	dividends, err := dm.mdata.GetDividendsMetadata(ticker)
	if err != nil {
		return 0, err
	}

	// Fetch trades for the ticker from the blotter
	trades, err := dm.blotter.GetTradesByTicker(ticker)
	if err != nil {
		return 0, err
	}

	// Calculate total dividends based on trades and dividends data
	var totalDividends float64
	for _, trade := range trades {
		for _, dividend := range dividends {
			if trade.TradeDate <= dividend.ExDate {
				totalDividends += trade.Quantity * dividend.Amount
			}
		}
	}

	return totalDividends, nil
}
