//go:build integration

package sources_test

import (
	"testing"

	"portfolio-manager/pkg/mdata/sources"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDividendsSg_FetchDividends(t *testing.T) {
	ds := sources.NewDividendsSg(nil)
	dividends, err := ds.GetDividendsMetadata("ES3", 0.0)
	require.NoError(t, err)

	// Verify we got some dividend data
	assert.Greater(t, len(dividends), 0, "should have received some dividends")

	// Verify dividend values are positive
	for date, metadata := range dividends {
		assert.Greater(t, metadata.Amount, 0.0, "dividend amount should be positive for date %v", date)
	}
}

func TestDividendsSg_GetAssetPrice(t *testing.T) {
	ds := sources.NewDividendsSg(nil)
	assetData, err := ds.GetAssetPrice("TEMB")
	require.NoError(t, err)

	// Verify we got some dividend data
	assert.Greater(t, assetData.Price, float64(0), "should have received some price")
}

func TestDividendsSg_CurrentEquityAndBondPages(t *testing.T) {
	for _, ticker := range []string{"CLR", "TEMB", "6AZB"} {
		t.Run(ticker, func(t *testing.T) {
			ds := sources.NewDividendsSg(nil)
			quote, err := ds.GetAssetPrice(ticker)
			require.NoError(t, err)
			require.Positive(t, quote.Price)
			dividends, err := ds.GetDividendsMetadata(ticker, 0)
			require.NoError(t, err)
			require.NotEmpty(t, dividends)
			for _, dividend := range dividends {
				require.Positive(t, dividend.Amount)
			}
		})
	}
}
