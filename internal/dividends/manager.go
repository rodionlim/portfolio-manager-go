package dividends

import (
	"fmt"
	"maps"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"strings"
	"sync"
	"time"
)

type DividendsManager struct {
	db                dal.Database
	mdata             mdata.MarketDataManager
	rdata             rdata.ReferenceManager
	blotter           blotter.TradeGetter
	cachedDividends   map[string][]Dividends
	cachedDividendsAt time.Time
	cacheMutex        sync.RWMutex
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

// CalculateDividendsForAllTickers calculates dividends for all tickers across the portfolio
// Results are cached for 24 hours to improve performance
func (dm *DividendsManager) CalculateDividendsForAllTickers() (map[string][]Dividends, error) {
	const cacheDuration = 24 * time.Hour

	// Check cache first
	dm.cacheMutex.RLock()
	if dm.cachedDividends != nil && time.Since(dm.cachedDividendsAt) < cacheDuration {
		// Create a copy of the cached data to avoid concurrent map access issues
		result := make(map[string][]Dividends, len(dm.cachedDividends))
		maps.Copy(result, dm.cachedDividends)
		dm.cacheMutex.RUnlock()
		return result, nil
	}
	dm.cacheMutex.RUnlock()

	// Cache miss or expired - calculate fresh results
	tickers, err := dm.blotter.GetAllTickers()
	if err != nil {
		return nil, err
	}

	dividendsMap := make(map[string][]Dividends)
	for _, ticker := range tickers {
		logging.GetLogger().Infof("Fetching dividends for ticker: %s", ticker)
		dividends, err := dm.CalculateDividendsForSingleTicker(ticker)
		if err != nil {
			return nil, err
		}
		dividendsMap[ticker] = dividends
	}

	// Update cache
	dm.cacheMutex.Lock()
	dm.cachedDividends = dividendsMap
	dm.cachedDividendsAt = time.Now()
	dm.cacheMutex.Unlock()

	return dividendsMap, nil
}

func (dm *DividendsManager) CalculateDividendsForSingleTicker(ticker string) ([]Dividends, error) {
	ticker = strings.ToUpper(ticker)

	// Get ticker reference
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
	dividends, err := dm.mdata.GetDividendsMetadataFromTickerRef(tickerRef.TickerReference)
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

// CalculateDividendsForSingleBook calculates dividends for all tickers within a specific book
func (dm *DividendsManager) CalculateDividendsForSingleBook(book string) (map[string][]Dividends, error) {
	// Get all trades from the blotter
	allTrades := dm.blotter.GetTrades()

	// Filter trades by book and collect unique valid tickers
	bookTrades := make(map[string][]blotter.Trade)
	validTickers := make(map[string]bool)

	for _, trade := range allTrades {
		if strings.EqualFold(trade.Book, book) {
			ticker := strings.ToUpper(trade.Ticker)
			bookTrades[ticker] = append(bookTrades[ticker], trade)

			// Check if ticker is valid (has reference data for dividends)
			tickerRef, err := dm.rdata.GetTicker(ticker)
			if err == nil && (tickerRef.DividendsSgTicker != "" || tickerRef.YahooTicker != "" || common.IsSSB(ticker) || common.IsSgTBill(ticker)) {
				validTickers[ticker] = true
			}
		}
	}

	// Calculate dividends for each valid ticker in the book
	dividendsMap := make(map[string][]Dividends)
	for ticker := range validTickers {
		trades := bookTrades[ticker]

		// Get ticker reference
		tickerRef, err := dm.rdata.GetTicker(ticker)
		if err != nil {
			continue // Skip if no reference data
		}

		// Fetch dividends data from mdata service
		dividends, err := dm.mdata.GetDividendsMetadataFromTickerRef(tickerRef.TickerReference)
		if err != nil {
			continue // Skip if can't get dividends data
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

		if len(allDividends) > 0 {
			dividendsMap[ticker] = allDividends
		}
	}

	return dividendsMap, nil
}
