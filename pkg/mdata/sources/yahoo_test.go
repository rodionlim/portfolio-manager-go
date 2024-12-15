//go:build integration

package sources_test

import (
	"testing"
	"time"

	"portfolio-manager/pkg/mdata/sources"

	"github.com/stretchr/testify/assert"
)

func TestYahooFinance_GetQuoteNasdaq_Integration(t *testing.T) {
	yf := sources.NewYahooFinance()

	quote, err := yf.GetAssetPrice("AAPL")
	assert.NoError(t, err)
	assert.NotZero(t, quote.Price)
	assert.NotEmpty(t, quote.Ticker)
}

func TestYahooFinance_GetQuoteSGX_Integration(t *testing.T) {
	yf := sources.NewYahooFinance()

	quote, err := yf.GetAssetPrice("ES3.SI")
	assert.NoError(t, err)
	assert.NotZero(t, quote.Price)
	assert.NotEmpty(t, quote.Ticker)
}

func TestYahooFinance_GetHistoricalData_Integration(t *testing.T) {
	yf := sources.NewYahooFinance()

	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0) // Get 1 month of data

	end := endTime.Unix()
	start := startTime.Unix()

	data, err := yf.GetHistoricalData("AAPL", start, end)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Basic sanity checks
	for _, d := range data {
		assert.NotZero(t, d.Price)
		assert.True(t, d.Timestamp >= start, "Date should be after start time")
		assert.True(t, d.Timestamp <= end, "Date should be before end time")
	}
}
