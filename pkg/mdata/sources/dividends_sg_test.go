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
	assert.Greater(t, assetData.Price, 0, "should have received some price")
}
