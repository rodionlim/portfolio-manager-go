//go:build integration

package sources_test

import (
	"testing"

	"portfolio-manager/pkg/mdata/sources"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNasdaq_FetchDividends(t *testing.T) {
	nasdaq := sources.NewNasdaq(nil)
	dividends, err := nasdaq.GetDividendsMetadata("MSFT", 0.30)
	require.NoError(t, err)

	// Verify we got some dividend data
	assert.Greater(t, len(dividends), 0, "should have received some dividends")

	// Verify dividend values are positive
	for _, metadata := range dividends {
		assert.Greater(t, metadata.Amount, 0.0, "dividend amount should be positive for date %v", metadata.ExDate)
		assert.Equal(t, "MSFT", metadata.Ticker)
		assert.Equal(t, 0.30, metadata.WithholdingTax)
	}

	// Verify dates are in YYYY-MM-DD format
	for _, metadata := range dividends {
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, metadata.ExDate, "date should be in YYYY-MM-DD format")
	}
}
