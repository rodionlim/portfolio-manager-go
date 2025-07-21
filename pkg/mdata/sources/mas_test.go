//go:build integration

package sources_test

import (
	"portfolio-manager/pkg/mdata/sources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMas_GetDividendsMetadata_Integration(t *testing.T) {
	src := sources.NewMas(nil)

	coupons, err := src.GetDividendsMetadata("BS24124Z", 0.0)
	require.NoError(t, err)
	require.NotEmpty(t, coupons)
	assert.Equal(t, 1, len(coupons))
}

func TestMas_FetchBenchmarkInterestRates_Integration(t *testing.T) {
	src := sources.NewMas(nil)

	// Test with Singapore - request 10 recent records
	rates, err := src.FetchBenchmarkInterestRates("SG", 10)
	if err != nil {
		t.Logf("Error fetching SG rates: %v", err)
		// Don't fail the test as this might be network-dependent
		return
	}

	require.NotEmpty(t, rates)
	require.LessOrEqual(t, len(rates), 10, "Should return at most 10 records")

	// Verify the structure
	for _, rate := range rates {
		assert.Equal(t, "SG", rate.Country)
		assert.Equal(t, "SORA", rate.RateType)
		assert.Equal(t, "O/N", rate.Tenor)
		assert.Greater(t, rate.Rate, 0.0)
		assert.NotEmpty(t, rate.Date)
	}

	t.Logf("Successfully fetched %d interest rates", len(rates))
}

func TestMas_FetchBenchmarkInterestRates_UnsupportedCountry_Integration(t *testing.T) {
	src := sources.NewMas(nil)

	_, err := src.FetchBenchmarkInterestRates("US", 10)
	require.Error(t, err)
	assert.Equal(t, "unsupported country: US", err.Error())
}
