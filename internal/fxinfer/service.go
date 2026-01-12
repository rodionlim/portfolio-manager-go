package fxinfer

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/pkg/csvutil"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"sort"
	"time"
)

// Service is responsible for inferring FX rates for trades and enriching trade data
type Service struct {
	blotterSvc blotter.TradeGetter
	mdataSvc   mdata.MarketDataManager
	rdataSvc   rdata.ReferenceManager
	baseCcy    string
}

// NewFXInferenceService creates a new instance of the FX inference service
func NewFXInferenceService(
	blotterSvc blotter.TradeGetter,
	mdataSvc mdata.MarketDataManager,
	rdataSvc rdata.ReferenceManager,
	baseCcy string,
) *Service {
	return &Service{
		blotterSvc: blotterSvc,
		mdataSvc:   mdataSvc,
		rdataSvc:   rdataSvc,
		baseCcy:    baseCcy,
	}
}

// ExportTradesWithInferredFX exports all trades as CSV with inferred FX rates where missing
// This function also amend the blotter trades in memory when fx rates are 0,
// but users that want to persist the new rates should wipe and reimport the blotter
func (s *Service) ExportTradesWithInferredFX() ([]byte, error) {
	// Get all trades from blotter
	trades := s.blotterSvc.GetTrades()
	if len(trades) == 0 {
		logging.GetLogger().Info("No trades to export")
		return exportTradesToCSV(trades)
	}

	// Sort trades by ticker for better organization
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Ticker < trades[j].Ticker
	})

	// Group trades by currency and collect trade dates
	currencyGroups := make(map[string][]*blotter.Trade)
	currencyDateRanges := make(map[string]struct {
		minDate int64
		maxDate int64
	})

	// Collect all currencies and their date ranges
	for i := range trades {
		// impt!! Skip trades that already have FX rates
		if trades[i].Fx != 0 {
			continue
		}

		// Get reference data for the trade's ticker
		refData, err := s.rdataSvc.GetTicker(trades[i].Ticker)
		if err != nil {
			logging.GetLogger().Errorf("Failed to get reference data for ticker %s: %v", trades[i].Ticker, err)
			continue
		}

		// Skip if currency is base currency
		if refData.Ccy == s.baseCcy {
			trades[i].Fx = 1.0 // Set FX rate to 1.0 for base currency
			continue
		}

		// Group by currency
		if _, exists := currencyGroups[refData.Ccy]; !exists {
			currencyGroups[refData.Ccy] = []*blotter.Trade{}
			currencyDateRanges[refData.Ccy] = struct {
				minDate int64
				maxDate int64
			}{
				minDate: time.Now().Unix(),
				maxDate: 0,
			}
		}

		// Append trade to its currency group
		currencyGroups[refData.Ccy] = append(currencyGroups[refData.Ccy], &trades[i])

		// Parse trade date
		tradeDate, err := time.Parse(time.RFC3339, trades[i].TradeDate)
		if err != nil {
			logging.GetLogger().Errorf("Failed to parse trade date %s: %v", trades[i].TradeDate, err)
			continue
		}

		// Update date range
		tradeEpoch := tradeDate.Unix()
		dateRange := currencyDateRanges[refData.Ccy]
		if tradeEpoch < dateRange.minDate {
			dateRange.minDate = tradeEpoch
		}
		if tradeEpoch > dateRange.maxDate {
			dateRange.maxDate = tradeEpoch
		}
		currencyDateRanges[refData.Ccy] = dateRange
	}

	// Fetch historical FX data for each currency and update trades
	for currency, dateRange := range currencyDateRanges {
		// Construct FX ticker - e.g., USD-SGD
		fxTicker := fmt.Sprintf("%s-%s", currency, s.baseCcy)

		// Get historical FX data
		fxData, _, err := s.mdataSvc.GetHistoricalData(fxTicker, dateRange.minDate, dateRange.maxDate)
		if err != nil {
			logging.GetLogger().Errorf("Failed to get historical data for %s: %v", fxTicker, err)
			continue
		}

		// Create a map of date to FX rate for easy lookup
		fxRatesByDate := make(map[int64]float64)

		for _, data := range fxData {
			// Always map in terms of units of base currency to 1 dollar worth of foreign currency, e.g. 1.3 SGD/USD
			fxRatesByDate[data.Timestamp] = data.Price
		}

		// Update FX rates for trades in this currency group
		for _, trade := range currencyGroups[currency] {
			tradeDate, err := time.Parse(time.RFC3339, trade.TradeDate)
			if err != nil {
				logging.GetLogger().Errorf("Failed to parse trade date %s: %v", trade.TradeDate, err)
				continue
			}

			// Find closest available FX rate to the trade date
			tradeDateEpoch := tradeDate.Unix()
			bestRate := 0.0
			bestDiff := int64(86400 * 30) // 30 days in seconds as max difference

			for dataDate, rate := range fxRatesByDate {
				diff := abs(tradeDateEpoch - dataDate)
				// If we find a rate closer to our date, use it
				if diff < bestDiff {
					bestDiff = diff
					bestRate = rate
				}
			}

			// Update the trade's FX rate
			if bestRate > 0 {
				trade.Fx = bestRate
				logging.GetLogger().Infof("Updated FX rate for %s trade on %s to %.4f", trade.Ticker, trade.TradeDate, bestRate)
			} else {
				logging.GetLogger().Warnf("No FX rate found for %s trade on %s", trade.Ticker, trade.TradeDate)
			}
		}
	}

	// Export updated trades to CSV
	return exportTradesToCSV(trades)
}

// Helper function to export trades to CSV
func exportTradesToCSV(trades []blotter.Trade) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header row using shared headers from csvutil package
	if err := writer.Write(csvutil.TradeHeaders); err != nil {
		return nil, fmt.Errorf("error writing CSV header: %w", err)
	}

	// Write trade rows
	for _, trade := range trades {
		row := []string{
			trade.TradeDate,
			trade.Ticker,
			trade.Side,
			csvutil.FormatFloat(trade.Quantity, 4),
			csvutil.FormatFloat(trade.Price, 4),
			csvutil.FormatFloat(trade.Yield, 4),
			trade.Book,
			trade.Broker,
			trade.Account,
			trade.Status,
			csvutil.FormatFloat(trade.Fx, 4),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("error writing CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper function to calculate absolute difference between two int64 values
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetCurrentFXRates returns the current FX rates for currencies used in the trades
func (s *Service) GetCurrentFXRates() (map[string]float64, error) {
	trades := s.blotterSvc.GetTrades()
	if len(trades) == 0 {
		return map[string]float64{s.baseCcy: 1.0}, nil
	}

	// Store unique currencies we need to fetch
	currencies := make(map[string]bool)
	currencies[s.baseCcy] = true // Always include base currency

	// Process all trades to find unique currencies
	for _, trade := range trades {
		refData, err := s.rdataSvc.GetTicker(trade.Ticker)
		if err != nil {
			logging.GetLogger().Warnf("Failed to get reference data for ticker %s: %v", trade.Ticker, err)
			continue
		}
		currencies[refData.Ccy] = true
	}

	// Create the result map with rates relative to base currency
	result := make(map[string]float64)
	for ccy := range currencies {
		if ccy == s.baseCcy {
			result[ccy] = 1.0 // Base currency always has rate of 1.0
			continue
		}

		// Fetch FX rate from market data service
		fxPair := ccy + "-" + s.baseCcy
		assetData, err := s.mdataSvc.GetAssetPrice(fxPair)
		if err != nil {
			logging.GetLogger().Warnf("Failed to get FX rate for %s: %v", fxPair, err)
			continue
		}

		result[ccy] = assetData.Price
	}

	return result, nil
}
