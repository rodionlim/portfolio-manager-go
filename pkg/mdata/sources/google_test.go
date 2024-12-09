//go:build integration

package sources_test

import (
	"testing"

	"portfolio-manager/pkg/mdata/sources"

	"github.com/stretchr/testify/assert"
)

func TestGoogleFinance_GetQuoteNasdaq_Integration(t *testing.T) {
	gf := sources.NewGoogleFinance()

	quote, err := gf.GetStockPrice("AAPL:NASDAQ")
	assert.NoError(t, err)
	assert.NotZero(t, quote.Price)
	assert.NotEmpty(t, quote.Ticker)
}

func TestGoogleFinance_GetQuoteSGX_Integration(t *testing.T) {
	gf := sources.NewGoogleFinance()

	quote, err := gf.GetStockPrice("D05:SGX")
	assert.NoError(t, err)
	assert.NotZero(t, quote.Price)
	assert.NotEmpty(t, quote.Ticker)
}
