package metrics

import (
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"sort"
	"strings"
	"time"

	"github.com/maksim77/goxirr"
)

// MetricsService provides portfolio metrics calculations such as IRR
type MetricsService struct {
	blotterSvc       blotter.TradeGetter
	portfolioSvc     portfolio.PortfolioReader
	dividendsManager dividends.Manager
	mdataSvc         mdata.MarketDataManager
	rdataSvc         rdata.ReferenceManager
}

// NewMetricsService creates a new MetricsService
func NewMetricsService(
	blotterSvc blotter.TradeGetter,
	portfolioSvc portfolio.PortfolioReader,
	dividendsManager dividends.Manager,
	mdataSvc mdata.MarketDataManager,
	rdataSvc rdata.ReferenceManager,
) *MetricsService {
	return &MetricsService{
		blotterSvc:       blotterSvc,
		portfolioSvc:     portfolioSvc,
		dividendsManager: dividendsManager,
		mdataSvc:         mdataSvc,
		rdataSvc:         rdataSvc,
	}
}

// CalculatePortfolioMetrics computes the XIRR for the portfolio using all trades, dividends, and current market value as final cash flow
// It also stores other metrics such as price paid, market value of portfolio and total dividends
// If book_filter is specified, it filters trades by the given book
func (m *MetricsService) CalculatePortfolioMetrics(book_filter string) (MetricResultsWithCashFlows, error) {
	var cashflows goxirr.Transactions
	var result MetricResultsWithCashFlows
	result.Metrics = MetricsResult{}
	result.CashFlows = []CashFlow{}

	var pricePaid float64
	var totalDividends float64

	book_filter = strings.ToLower(book_filter)

	// Set metrics label if book_filter is provided
	if book_filter != "" {
		result.Label = book_filter
	}

	// 1. Add all filtered trades as cash flows (buys are negative, sells are positive)
	trades := m.blotterSvc.GetTrades()
	if book_filter != "" {
		// Filter trades by book if specified
		var filteredTrades []blotter.Trade
		for _, trade := range trades {
			if strings.ToLower(trade.Book) == book_filter {
				filteredTrades = append(filteredTrades, trade)
			}
		}
		trades = filteredTrades
	}

	for _, trade := range trades {
		tradeDate, err := time.Parse(time.RFC3339, trade.TradeDate)
		if err != nil {
			continue // skip invalid dates
		}
		amount := trade.Quantity * trade.Price * trade.Fx
		flowType := CashFlowTypeSell
		if trade.Side == blotter.TradeSideBuy {
			amount = -amount
			flowType = CashFlowTypeBuy
		}
		pricePaid += amount

		// Add to goxirr transactions for calculation
		cashflows = append(cashflows, goxirr.Transaction{
			Date: tradeDate,
			Cash: amount,
		})

		// Add to result cash flows for returning
		result.CashFlows = append(result.CashFlows, CashFlow{
			Date:        tradeDate.Format(time.RFC3339),
			Cash:        amount,
			Ticker:      trade.Ticker,
			Description: flowType,
		})
	}
	result.Metrics.PricePaid = -pricePaid

	// 2. Add all dividends as positive cash flows
	var divs map[string][]dividends.Dividends
	var err error
	if book_filter != "" {
		divs, err = m.dividendsManager.CalculateDividendsForSingleBook(book_filter)
	} else {
		divs, err = m.dividendsManager.CalculateDividendsForAllTickers()
	}
	if err != nil {
		return MetricResultsWithCashFlows{}, err
	}

	for ticker, dividendsList := range divs {
		// Get ticker reference to obtain currency
		tickerRef, err := m.rdataSvc.GetTicker(ticker)
		if err != nil {
			logging.GetLogger().Errorf("Failed to get ticker reference for %s: %v", ticker, err)
			continue
		}

		// Cache for currency rates to avoid repeated lookups
		fxRates := make(map[string]float64)

		for _, div := range dividendsList {
			divDate, err := time.Parse("2006-01-02", div.ExDate)
			if err != nil {
				logging.GetLogger().Errorf("Failed to parse %s for dividend date %s: %v", ticker, div.ExDate, err)
				continue
			}

			// Get FX rate if it's not SGD and not already in our cache
			fxRate := 1.0 // Default for SGD
			if tickerRef.Ccy != "SGD" {
				if rate, exists := fxRates[tickerRef.Ccy]; exists {
					fxRate = rate
				} else {
					// Try to get FX rate from market data service
					fxTicker := tickerRef.Ccy + "-SGD" // Format for currency pair (e.g., USD-SGD)
					fxData, err := m.mdataSvc.GetAssetPrice(fxTicker)
					if err != nil {
						logging.GetLogger().Warnf("Could not get FX rate for %s to SGD, using 1.0: %v", tickerRef.Ccy, err)
					} else if fxData != nil {
						fxRate = fxData.Price
						// Cache the rate for future use
						fxRates[tickerRef.Ccy] = fxRate
					}
				}
			}

			// Apply FX rate to dividend amount
			sgdAmount := div.Amount * fxRate
			totalDividends += sgdAmount

			// Add to goxirr transactions for calculation
			cashflows = append(cashflows, goxirr.Transaction{
				Date: divDate,
				Cash: sgdAmount,
			})

			// Add to result cash flows for returning
			result.CashFlows = append(result.CashFlows, CashFlow{
				Date:        divDate.Format(time.RFC3339),
				Cash:        sgdAmount,
				Ticker:      ticker,
				Description: CashFlowTypeDividend,
			})
		}
	}
	result.Metrics.TotalDividends = totalDividends

	// 3. Add final cash flow as current market value (positive, at current date)
	positions, err := m.portfolioSvc.GetAllPositions()
	if err != nil {
		return MetricResultsWithCashFlows{}, err
	}
	if book_filter != "" {
		var filteredPositions []*portfolio.Position
		for _, position := range positions {
			if strings.ToLower(position.Book) == book_filter {
				filteredPositions = append(filteredPositions, position)
			}
		}
		positions = filteredPositions
	}

	totalMarketValue := 0.0
	for _, position := range positions {
		totalMarketValue += position.Mv * position.FxRate
	}
	result.Metrics.MV = totalMarketValue

	now := time.Now()
	// Add to goxirr transactions for calculation
	cashflows = append(cashflows, goxirr.Transaction{
		Date: now,
		Cash: totalMarketValue,
	})

	// Add to result cash flows for returning
	result.CashFlows = append(result.CashFlows, CashFlow{
		Date:        now.Format(time.RFC3339),
		Cash:        totalMarketValue,
		Ticker:      "Portfolio",
		Description: CashFlowTypePortfolioValue,
	})

	// 4. Sort cashflows by date ascending
	sort.Slice(cashflows, func(i, j int) bool {
		return cashflows[i].Date.Before(cashflows[j].Date)
	})

	// DEBUG: Log cashflows for IRR calculation
	for i, cf := range cashflows {
		logging.GetLogger().Infof("IRR cashflow[%d]: date=%s, cash=%.2f", i, cf.Date.Format("2006-01-02"), cf.Cash)
	}

	// 5. Calculate XIRR
	r := goxirr.Xirr(cashflows)
	result.Metrics.IRR = r / 100 // goxirr.Xirr returns percent, so divide by 100 for decimal
	return result, nil
}
