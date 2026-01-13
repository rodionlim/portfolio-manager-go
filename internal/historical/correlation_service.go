package historical

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"time"
)

// CalculateCorrelation fetches historical data for the given tickers and time range,
// formats it as a wide CSV (date, ticker1, ticker2...), and executes the
// quantlib correlation command with the provided flags.
func (s *Service) CalculateCorrelation(ctx context.Context, tickers []string, fromUnix, toUnix int64, flags []string) ([]byte, error) {
	if len(tickers) < 2 {
		return nil, fmt.Errorf("at least two tickers are required")
	}

	// 1. Fetch Data
	// map[DateString]map[Ticker]PriceString
	priceMap := make(map[string]map[string]string)
	allDates := make(map[string]bool)

	for _, ticker := range tickers {
		data, _, err := s.mdataManager.GetHistoricalData(ticker, fromUnix, toUnix)
		if err != nil {
			s.logger.Errorf("failed to get historical data ticker=%s error=%v", ticker, err)
			return nil, fmt.Errorf("failed to fetch data for %s: %w", ticker, err)
		}

		for _, point := range data {
			// Use YYYY-MM-DD
			d := time.Unix(point.Timestamp, 0).UTC().Format("2006-01-02")
			if _, ok := priceMap[d]; !ok {
				priceMap[d] = make(map[string]string)
			}
			// Use AdjClose
			priceMap[d][ticker] = strconv.FormatFloat(point.AdjClose, 'f', 6, 64)
			allDates[d] = true
		}
	}

	// 2. Format CSV
	var sortedDates []string
	for d := range allDates {
		sortedDates = append(sortedDates, d)
	}
	sort.Strings(sortedDates)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	header := append([]string{"date"}, tickers...)
	if err := w.Write(header); err != nil {
		return nil, err
	}

	// Rows
	for _, d := range sortedDates {
		row := []string{d}
		for _, t := range tickers {
			if val, ok := priceMap[d][t]; ok {
				row = append(row, val)
			} else {
				row = append(row, "") // Missing data
			}
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()

	// 3. Exec quantlib
	// args start with "corr"
	cmdArgs := append([]string{"corr"}, flags...)

	s.logger.Infof("Executing quantlib with args: %v", cmdArgs)

	cmd := exec.CommandContext(ctx, "quantlib", cmdArgs...)
	cmd.Stdin = &buf
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("quantlib failed: %v, stderr: %s", err, errBuf.String())
	}

	return outBuf.Bytes(), nil
}
