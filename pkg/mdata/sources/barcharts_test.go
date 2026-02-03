//go:build integration

package sources_test

import (
	"testing"
	"time"

	"portfolio-manager/pkg/mdata/sources"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBarcharts_GetHistoricalData(t *testing.T) {
	barcharts := sources.NewBarcharts()
	end := time.Now().Unix()
	start := end - int64(30*24*60*60)

	data, err := barcharts.GetHistoricalData("HEJ26", start, end)
	require.NoError(t, err)

	assert.Greater(t, len(data), 0, "should have received some historical data")
	for _, entry := range data {
		assert.Equal(t, "HEJ26", entry.Ticker)
		assert.Greater(t, entry.LastPrice, 0.0, "last price should be positive")
		assert.Greater(t, entry.Timestamp, int64(0))
	}
}
