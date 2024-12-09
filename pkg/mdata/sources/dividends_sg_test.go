//go:build integration

package sources_test

import (
	"testing"

	"portfolio-manager/pkg/mdata/sources"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDividendsSg_FetchDividends(t *testing.T) {
	ds := sources.NewDividendsSg()
	dividends, err := ds.FetchDividends("ES3")
	require.NoError(t, err)

	// Verify we got some dividend data
	assert.Greater(t, len(dividends), 0, "should have received some dividends")

	// Verify dividend values are positive
	for date, amount := range dividends {
		assert.Greater(t, amount, 0.0, "dividend amount should be positive for date %v", date)
	}
}
