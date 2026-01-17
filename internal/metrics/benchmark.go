package metrics

import (
	"fmt"
	"math"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"
	"strings"
	"time"

	"github.com/maksim77/goxirr"
)

type benchmarkPosition struct {
	Ticker   string
	Quantity float64
}

type benchmarkPositionChange struct {
	Date     time.Time
	QtyDelta float64
}

func (m *MetricsService) BenchmarkPortfolioPerformance(req BenchmarkRequest) (BenchmarkComparisonResult, error) {
	portfolioMetrics, err := m.CalculatePortfolioMetrics(req.BookFilter)
	if err != nil {
		return BenchmarkComparisonResult{}, err
	}

	if len(req.BenchmarkTickers) == 0 {
		return BenchmarkComparisonResult{}, fmt.Errorf("benchmark_tickers is required")
	}

	mode := req.Mode
	if mode != BenchmarkModeBuyAtStart && mode != BenchmarkModeMatchTrades {
		return BenchmarkComparisonResult{}, fmt.Errorf("unsupported benchmark mode: %s", mode)
	}

	weights, err := normalizeBenchmarkWeights(req.BenchmarkTickers)
	if err != nil {
		return BenchmarkComparisonResult{}, err
	}

	trades := m.blotterSvc.GetTrades()
	if req.BookFilter != "" {
		bookFilter := strings.ToLower(req.BookFilter)
		var filtered []blotter.Trade
		for _, trade := range trades {
			if strings.ToLower(trade.Book) == bookFilter {
				filtered = append(filtered, trade)
			}
		}
		trades = filtered
	}
	if len(trades) == 0 {
		return BenchmarkComparisonResult{}, fmt.Errorf("no trades available for benchmarking")
	}

	if mode == BenchmarkModeBuyAtStart && req.Notional <= 0 {
		return BenchmarkComparisonResult{}, fmt.Errorf("notional must be provided for buy_at_start mode")
	}

	benchmarkCashflows := []CashFlow{}
	benchmarkTransactions := goxirr.Transactions{}
	positions := map[string]*benchmarkPosition{}
	positionChanges := map[string][]benchmarkPositionChange{}
	feeTotal := 0.0
	pricePaid := 0.0

	startDate, endDate, err := tradeDateRange(trades)
	if err != nil {
		return BenchmarkComparisonResult{}, err
	}
	endDate = time.Now()

	priceCache, err := buildBenchmarkPriceCache(m, weights, startDate, endDate)
	if err != nil {
		return BenchmarkComparisonResult{}, err
	}

	appendCashflow := func(date time.Time, cash float64, ticker string, desc CashFlowType) {
		benchmarkTransactions = append(benchmarkTransactions, goxirr.Transaction{Date: date, Cash: cash})
		benchmarkCashflows = append(benchmarkCashflows, CashFlow{
			Date:        date.Format(time.RFC3339),
			Cash:        cash,
			Ticker:      ticker,
			Description: desc,
		})
	}

	if mode == BenchmarkModeBuyAtStart {
		fee := computeBenchmarkFee(req.BenchmarkCost, req.Notional)
		feeTotal += fee
		pricePaid += req.Notional + fee

		for _, w := range weights {
			alloc := req.Notional * w.Weight
			allocFee := fee * w.Weight
			price, err := priceCache.priceInSGDAt(w.Ticker, startDate)
			if err != nil {
				return BenchmarkComparisonResult{}, err
			}
			qty := 0.0
			if price > 0 {
				qty = alloc / price
			}
			positions[w.Ticker] = &benchmarkPosition{Ticker: w.Ticker, Quantity: qty}
			positionChanges[w.Ticker] = append(positionChanges[w.Ticker], benchmarkPositionChange{
				Date:     startDate,
				QtyDelta: qty,
			})

			appendCashflow(startDate, -(alloc + allocFee), w.Ticker, CashFlowTypeBuy)
		}
	} else {
		for _, trade := range trades {
			tradeDate, err := time.Parse(time.RFC3339, trade.TradeDate)
			if err != nil {
				continue
			}

			notional := math.Abs(trade.Quantity * trade.Price * trade.Fx)
			if notional == 0 {
				continue
			}

			fee := computeBenchmarkFee(req.BenchmarkCost, notional)
			feeTotal += fee

			isBuy := trade.Side == blotter.TradeSideBuy
			if isBuy {
				pricePaid += notional + fee
			} else {
				pricePaid -= notional - fee
			}

			for _, w := range weights {
				alloc := notional * w.Weight
				allocFee := fee * w.Weight
				price, err := priceCache.priceInSGDAt(w.Ticker, tradeDate)
				if err != nil {
					return BenchmarkComparisonResult{}, err
				}
				qty := 0.0
				if price > 0 {
					qty = alloc / price
				}
				pos := positions[w.Ticker]
				if pos == nil {
					pos = &benchmarkPosition{Ticker: w.Ticker}
					positions[w.Ticker] = pos
				}

				if isBuy {
					pos.Quantity += qty
					positionChanges[w.Ticker] = append(positionChanges[w.Ticker], benchmarkPositionChange{
						Date:     tradeDate,
						QtyDelta: qty,
					})
					appendCashflow(tradeDate, -(alloc + allocFee), w.Ticker, CashFlowTypeBuy)
				} else {
					pos.Quantity -= qty
					positionChanges[w.Ticker] = append(positionChanges[w.Ticker], benchmarkPositionChange{
						Date:     tradeDate,
						QtyDelta: -qty,
					})
					appendCashflow(tradeDate, alloc-allocFee, w.Ticker, CashFlowTypeSell)
				}
			}
		}
	}

	// Add benchmark dividends (cash, no reinvest)
	if err := m.appendBenchmarkDividends(weights, positionChanges, appendCashflow, priceCache); err != nil {
		return BenchmarkComparisonResult{}, err
	}

	// Final benchmark market value
	finalMV := 0.0
	for _, w := range weights {
		pos := positions[w.Ticker]
		if pos == nil {
			continue
		}
		price, err := priceCache.priceInSGDAt(w.Ticker, endDate)
		if err != nil {
			return BenchmarkComparisonResult{}, err
		}
		finalMV += pos.Quantity * price
	}

	now := time.Now()
	appendCashflow(now, finalMV, "Benchmark", CashFlowTypePortfolioValue)

	sort.Slice(benchmarkTransactions, func(i, j int) bool {
		return benchmarkTransactions[i].Date.Before(benchmarkTransactions[j].Date)
	})

	benchmarkIRR := goxirr.Xirr(benchmarkTransactions) / 100

	result := BenchmarkComparisonResult{
		PortfolioMetrics: portfolioMetrics.Metrics,
		BenchmarkMetrics: BenchmarkMetrics{
			IRR:       benchmarkIRR,
			PricePaid: pricePaid,
			MV:        finalMV,
			Fees:      feeTotal,
		},
		PortfolioIRR:       portfolioMetrics.Metrics.IRR,
		BenchmarkIRR:       benchmarkIRR,
		IRRDifference:      portfolioMetrics.Metrics.IRR - benchmarkIRR,
		BenchmarkCashFlows: benchmarkCashflows,
	}

	if result.IRRDifference > 0 {
		result.Winner = "portfolio"
	} else if result.IRRDifference < 0 {
		result.Winner = "benchmark"
	} else {
		result.Winner = "tie"
	}

	return result, nil
}

func normalizeBenchmarkWeights(weights []BenchmarkTickerWeight) ([]BenchmarkTickerWeight, error) {
	sum := 0.0
	for _, w := range weights {
		if w.Ticker == "" {
			return nil, fmt.Errorf("benchmark ticker cannot be empty")
		}
		if w.Weight <= 0 {
			return nil, fmt.Errorf("benchmark weight must be > 0 for %s", w.Ticker)
		}
		sum += w.Weight
	}
	if sum == 0 {
		return nil, fmt.Errorf("benchmark weights must sum to > 0")
	}

	out := make([]BenchmarkTickerWeight, 0, len(weights))
	for _, w := range weights {
		out = append(out, BenchmarkTickerWeight{Ticker: w.Ticker, Weight: w.Weight / sum})
	}
	return out, nil
}

func computeBenchmarkFee(cost BenchmarkCost, notional float64) float64 {
	feePct := cost.Pct * notional
	if feePct > cost.Absolute {
		return feePct
	}
	return cost.Absolute
}

func tradeDateRange(trades []blotter.Trade) (time.Time, time.Time, error) {
	var earliest time.Time
	var latest time.Time
	for _, trade := range trades {
		tradeDate, err := time.Parse(time.RFC3339, trade.TradeDate)
		if err != nil {
			continue
		}
		if earliest.IsZero() || tradeDate.Before(earliest) {
			earliest = tradeDate
		}
		if latest.IsZero() || tradeDate.After(latest) {
			latest = tradeDate
		}
	}
	if earliest.IsZero() || latest.IsZero() {
		return time.Time{}, time.Time{}, fmt.Errorf("no trades available to determine date range")
	}
	return earliest, latest, nil
}

type benchmarkPriceCache struct {
	series    map[string][]*types.AssetData
	fxSeries  map[string][]*types.AssetData
	tickerCcy map[string]string
}

func buildBenchmarkPriceCache(m *MetricsService, weights []BenchmarkTickerWeight, startDate, endDate time.Time) (*benchmarkPriceCache, error) {
	series := map[string][]*types.AssetData{}
	fxSeries := map[string][]*types.AssetData{}
	tickerCcy := map[string]string{}

	from := startDate.AddDate(0, 0, -5).Unix()
	to := endDate.AddDate(0, 0, 5).Unix()

	for _, w := range weights {
		if _, ok := series[w.Ticker]; ok {
			continue
		}
		data, _, err := m.mdataSvc.GetHistoricalData(w.Ticker, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get historical data for %s: %w", w.Ticker, err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("no historical data for %s", w.Ticker)
		}
		sort.Slice(data, func(i, j int) bool { return data[i].Timestamp < data[j].Timestamp })
		series[w.Ticker] = data

		ref, err := m.rdataSvc.GetTicker(w.Ticker)
		if err == nil {
			tickerCcy[w.Ticker] = ref.Ccy
		} else {
			tickerCcy[w.Ticker] = data[0].Currency
		}
	}

	for ticker, ccy := range tickerCcy {
		if strings.EqualFold(ccy, "SGD") || ccy == "" {
			continue
		}
		fxTicker := strings.ToUpper(ccy) + "-SGD"
		if _, ok := fxSeries[fxTicker]; ok {
			continue
		}
		data, _, err := m.mdataSvc.GetHistoricalData(fxTicker, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get FX data for %s: %w", fxTicker, err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("no FX data for %s", fxTicker)
		}
		sort.Slice(data, func(i, j int) bool { return data[i].Timestamp < data[j].Timestamp })
		fxSeries[fxTicker] = data
		_ = ticker
	}

	return &benchmarkPriceCache{series: series, fxSeries: fxSeries, tickerCcy: tickerCcy}, nil
}

func (c *benchmarkPriceCache) priceInSGDAt(ticker string, date time.Time) (float64, error) {
	data, ok := c.series[ticker]
	if !ok || len(data) == 0 {
		return 0, fmt.Errorf("no cached data for %s", ticker)
	}
	closest, err := closestByTimestamp(data, date)
	if err != nil {
		return 0, err
	}
	ccy := c.tickerCcy[ticker]
	return c.convertToSGD(closest.Price, ccy, date)
}

func (c *benchmarkPriceCache) convertToSGD(amount float64, ccy string, date time.Time) (float64, error) {
	if strings.EqualFold(ccy, "SGD") || ccy == "" {
		return amount, nil
	}
	fxTicker := strings.ToUpper(ccy) + "-SGD"
	series, ok := c.fxSeries[fxTicker]
	if !ok || len(series) == 0 {
		return 0, fmt.Errorf("no cached FX data for %s", fxTicker)
	}
	closest, err := closestByTimestamp(series, date)
	if err != nil {
		return 0, err
	}
	return amount * closest.Price, nil
}

func closestByTimestamp(series []*types.AssetData, date time.Time) (*types.AssetData, error) {
	if len(series) == 0 {
		return nil, fmt.Errorf("empty series")
	}

	closest := series[0]
	minDiff := math.Abs(float64(closest.Timestamp - date.Unix()))
	for _, item := range series {
		diff := math.Abs(float64(item.Timestamp - date.Unix()))
		if diff < minDiff {
			closest = item
			minDiff = diff
		}
	}
	return closest, nil
}

func (m *MetricsService) appendBenchmarkDividends(
	weights []BenchmarkTickerWeight,
	positionChanges map[string][]benchmarkPositionChange,
	appendCashflow func(date time.Time, cash float64, ticker string, desc CashFlowType),
	priceCache *benchmarkPriceCache,
) error {
	weightSet := map[string]struct{}{}
	for _, w := range weights {
		weightSet[w.Ticker] = struct{}{}
		if len(positionChanges[w.Ticker]) > 0 {
			sort.Slice(positionChanges[w.Ticker], func(i, j int) bool {
				return positionChanges[w.Ticker][i].Date.Before(positionChanges[w.Ticker][j].Date)
			})
		}
	}

	for ticker := range weightSet {
		ref, err := m.rdataSvc.GetTicker(ticker)
		if err != nil {
			logging.GetLogger().Errorf("Failed to get ticker reference for %s: %v", ticker, err)
			continue
		}

		dividendsList, err := m.mdataSvc.GetDividendsMetadataFromTickerRef(ref.TickerReference)
		if err != nil {
			logging.GetLogger().Warnf("Failed to get dividends metadata for %s: %v", ticker, err)
			continue
		}
		if len(dividendsList) == 0 {
			continue
		}

		changes := positionChanges[ticker]
		for _, div := range dividendsList {
			if common.IsFutureDate(div.ExDate) {
				continue
			}
			divDate, err := time.Parse("2006-01-02", div.ExDate)
			if err != nil {
				logging.GetLogger().Errorf("Failed to parse %s for dividend date %s: %v", ticker, div.ExDate, err)
				continue
			}

			qty := benchmarkQuantityAtDate(changes, divDate)
			if qty <= 0 {
				continue
			}

			amount := qty * div.Amount * (1 - div.WithholdingTax)
			sgdAmount, err := priceCache.convertToSGD(amount, ref.Ccy, divDate)
			if err != nil {
				logging.GetLogger().Warnf("Failed to convert dividend to SGD for %s: %v", ticker, err)
				continue
			}

			appendCashflow(divDate, sgdAmount, ticker, CashFlowTypeDividend)
		}
	}

	return nil
}

func benchmarkQuantityAtDate(changes []benchmarkPositionChange, date time.Time) float64 {
	if len(changes) == 0 {
		return 0
	}

	qty := 0.0
	for _, change := range changes {
		if change.Date.After(date) {
			break
		}
		qty += change.QtyDelta
	}
	return qty
}
