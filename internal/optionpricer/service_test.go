package optionpricer

import (
	"testing"
	"time"

	testifymocks "portfolio-manager/internal/mocks/testify"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func float64Ptr(value float64) *float64 {
	return &value
}

type stubRateProvider struct {
	result resolvedRate
}

func (s stubRateProvider) Resolve(_ float64) resolvedRate {
	return s.result
}

func TestPriceWithProvidedVolatility(t *testing.T) {
	svc := NewService(nil)
	svc.now = func() time.Time {
		return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	volatility := 0.2
	result, err := svc.Price(PriceRequest{
		Ticker:     "AAPL",
		OptionType: "call",
		Spot:       100,
		Strike:     100,
		Expiry:     "2026-12-31",
		Rate:       float64Ptr(0.05),
		Volatility: &volatility,
	})

	require.NoError(t, err)
	assert.Equal(t, "input", result.RateSource)
	assert.Equal(t, "input", result.VolatilitySource)
	assert.InDelta(t, 10.45, result.NPV, 0.2)
	assert.InDelta(t, 0.64, result.Delta, 0.02)
	assert.InDelta(t, 0.019, result.Gamma, 0.002)
	assert.InDelta(t, -0.018, result.Theta, 0.003)
}

func TestPriceEstimatesVolatilityWhenMissing(t *testing.T) {
	mdataSvc := new(testifymocks.MockMarketDataManager)
	svc := NewService(mdataSvc)
	fixedNow := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	lookbackDays := 60
	svc.now = func() time.Time {
		return fixedNow
	}

	from := fixedNow.AddDate(0, 0, -lookbackDays).Unix()
	to := fixedNow.Unix()
	history := make([]*types.AssetData, 0, 80)
	price := 100.0
	for i := 0; i < 80; i++ {
		if i%2 == 0 {
			price *= 1.012
		} else {
			price *= 0.991
		}
		history = append(history, &types.AssetData{
			Ticker:    "AAPL",
			AdjClose:  price,
			Timestamp: fixedNow.AddDate(0, 0, -(80 - i)).Unix(),
		})
	}

	mdataSvc.On("GetHistoricalData", "AAPL", from, to).Return(history, false, nil)

	result, err := svc.Price(PriceRequest{
		Ticker:                 "AAPL",
		OptionType:             "put",
		Spot:                   100,
		Strike:                 95,
		Expiry:                 "2026-09-30",
		Rate:                   float64Ptr(0.03),
		VolatilityLookbackDays: lookbackDays,
	})

	require.NoError(t, err)
	assert.Equal(t, "estimated_historical", result.VolatilitySource)
	assert.Equal(t, lookbackDays, result.VolatilityLookback)
	assert.Greater(t, result.Volatility, 0.01)
	assert.NotZero(t, result.NPV)
	assert.NotZero(t, result.Delta)
	assert.NotZero(t, result.Gamma)
	mdataSvc.AssertExpectations(t)
}

func TestPriceRejectsUnsupportedHistoricalVolLookback(t *testing.T) {
	svc := NewService(nil)
	svc.now = func() time.Time {
		return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	_, err := svc.Price(PriceRequest{
		Ticker:                 "AAPL",
		OptionType:             "call",
		Spot:                   100,
		Strike:                 100,
		Expiry:                 "2026-12-31",
		Rate:                   float64Ptr(0.05),
		VolatilityLookbackDays: 90,
	})

	require.EqualError(t, err, "volatilityLookbackDays must be one of 30, 60, 180, or 360")
}

func TestPriceImpliesVolatilityFromPremium(t *testing.T) {
	svc := NewService(nil)
	svc.now = func() time.Time {
		return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	expectedVolatility := 0.2
	expiry, err := parseExpiry("2026-12-31")
	require.NoError(t, err)
	timeToExpiryYears := expiry.Sub(svc.now().UTC()).Hours() / 24 / calendarDaysPerYear
	premium, _, _, _, err := calculateBlackScholes(
		"call",
		100,
		100,
		timeToExpiryYears,
		0.05,
		0,
		expectedVolatility,
	)
	require.NoError(t, err)

	result, err := svc.Price(PriceRequest{
		Ticker:     "AAPL",
		OptionType: "call",
		Spot:       100,
		Strike:     100,
		Expiry:     "2026-12-31",
		Rate:       float64Ptr(0.05),
		Premium:    &premium,
	})

	require.NoError(t, err)
	assert.Equal(t, "implied_from_premium", result.VolatilitySource)
	if assert.NotNil(t, result.Premium) {
		assert.InDelta(t, premium, *result.Premium, 1e-8)
	}
	assert.InDelta(t, expectedVolatility, result.Volatility, 1e-6)
	assert.InDelta(t, premium, result.NPV, 1e-6)
}

func TestPriceUsesFetchedRateWhenMissing(t *testing.T) {
	svc := NewService(nil)
	svc.now = func() time.Time {
		return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	svc.rateProvider = stubRateProvider{
		result: resolvedRate{
			Rate:      0.0372,
			Source:    "fed_h15_treasury_constant_maturity",
			CurveDate: "2026-04-06",
		},
	}

	volatility := 0.2
	result, err := svc.Price(PriceRequest{
		Ticker:     "AAPL",
		OptionType: "call",
		Spot:       100,
		Strike:     100,
		Expiry:     "2026-12-31",
		Volatility: &volatility,
	})

	require.NoError(t, err)
	assert.Equal(t, "fed_h15_treasury_constant_maturity", result.RateSource)
	assert.Equal(t, "2026-04-06", result.RateCurveDate)
	assert.InDelta(t, 0.0372, result.Rate, 1e-9)
}
