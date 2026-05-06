package sources

import (
	"fmt"
	"testing"
	"time"

	testifydb "portfolio-manager/internal/mocks/testify/database"
	"portfolio-manager/pkg/types"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreDividendsMetadata_Custom_Overlap(t *testing.T) {
	// Setup
	mockDB := new(testifydb.MockDatabase)
	c := cache.New(5*time.Minute, 10*time.Minute)
	base := &BaseDividendSource{
		db:    mockDB,
		cache: c,
	}

	ticker := "AAPL"
	isCustom := true

	// Existing official dividends
	officialDividends := []types.DividendsMetadata{
		{ExDate: "2023-01-01", Amount: 0.5, Source: types.DividendSourceOfficial},
	}

	// New custom dividends (overlapping date)
	customDividends := []types.DividendsMetadata{
		{ExDate: "2023-01-01", Amount: 1.0, Source: types.DividendSourceCustom},
	}

	// Expectations
	// 1. Put custom dividends
	mockDB.On("Put", fmt.Sprintf("%s:%s", types.DividendsCustomKeyPrefix, ticker), customDividends).Return(nil)

	// 2. Get official dividends
	// Note: The code calls getSingleDividendsMetadata(ticker, false) which calls db.Get(DividendsKeyPrefix...)
	mockDB.On("Get", fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]types.DividendsMetadata)
		*arg = officialDividends
	})

	// Execute
	result, err := base.StoreDividendsMetadata(ticker, customDividends, isCustom)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Verify only custom exists (custom takes precedence)
	assert.Equal(t, customDividends[0], result[0])

	// Verify cache was set
	cachedVal, found := c.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker))
	assert.True(t, found)
	assert.Equal(t, result, cachedVal)

	mockDB.AssertExpectations(t)
}

func TestUpsertOfficialDividendsMetadata_PreservesHistoryAndAddsNewerRows(t *testing.T) {
	mockDB := new(testifydb.MockDatabase)
	c := cache.New(5*time.Minute, 10*time.Minute)
	base := &BaseDividendSource{
		db:    mockDB,
		cache: c,
	}

	ticker := "BS6"
	existingOfficial := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2008-05-06", Amount: 0.0157, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.065, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.12, Source: types.DividendSourceOfficial},
	}
	fetchedOfficial := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.065, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.12, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2026-05-06", Amount: 0.20, Source: types.DividendSourceOfficial},
	}
	customDividends := []types.DividendsMetadata{}
	expected := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2008-05-06", Amount: 0.0157, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.065, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.12, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2026-05-06", Amount: 0.20, Source: types.DividendSourceOfficial},
	}

	mockDB.On("Get", fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), mock.Anything).
		Return(nil).
		Once().
		Run(func(args mock.Arguments) {
			arg := args.Get(1).(*[]types.DividendsMetadata)
			*arg = existingOfficial
		})
	mockDB.On("Get", fmt.Sprintf("%s:%s", types.DividendsCustomKeyPrefix, ticker), mock.Anything).
		Return(nil).
		Once().
		Run(func(args mock.Arguments) {
			arg := args.Get(1).(*[]types.DividendsMetadata)
			*arg = customDividends
		})
	mockDB.On("Put", fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), expected).Return(nil).Once()

	result, err := base.upsertOfficialDividendsMetadata(ticker, fetchedOfficial)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	cachedVal, found := c.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker))
	assert.True(t, found)
	assert.Equal(t, result, cachedVal)

	mockDB.AssertExpectations(t)
}

func TestUpsertOfficialDividendsMetadata_UpdatesOverlapAndHonorsCustomZero(t *testing.T) {
	mockDB := new(testifydb.MockDatabase)
	c := cache.New(5*time.Minute, 10*time.Minute)
	base := &BaseDividendSource{
		db:    mockDB,
		cache: c,
	}

	ticker := "BS6"
	existingOfficial := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2008-05-06", Amount: 0.0157, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.065, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.12, Source: types.DividendSourceOfficial},
	}
	fetchedOfficial := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.07, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.12, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2026-05-06", Amount: 0.20, Source: types.DividendSourceOfficial},
	}
	customDividends := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.0, Source: types.DividendSourceCustom},
	}
	expectedStoredOfficial := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2008-05-06", Amount: 0.0157, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.07, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.12, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2026-05-06", Amount: 0.20, Source: types.DividendSourceOfficial},
	}
	expectedReturned := []types.DividendsMetadata{
		{Ticker: ticker, ExDate: "2008-05-06", Amount: 0.0157, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2024-04-29", Amount: 0.07, Source: types.DividendSourceOfficial},
		{Ticker: ticker, ExDate: "2025-05-05", Amount: 0.0, Source: types.DividendSourceCustom},
		{Ticker: ticker, ExDate: "2026-05-06", Amount: 0.20, Source: types.DividendSourceOfficial},
	}

	mockDB.On("Get", fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), mock.Anything).
		Return(nil).
		Once().
		Run(func(args mock.Arguments) {
			arg := args.Get(1).(*[]types.DividendsMetadata)
			*arg = existingOfficial
		})
	mockDB.On("Get", fmt.Sprintf("%s:%s", types.DividendsCustomKeyPrefix, ticker), mock.Anything).
		Return(nil).
		Once().
		Run(func(args mock.Arguments) {
			arg := args.Get(1).(*[]types.DividendsMetadata)
			*arg = customDividends
		})
	mockDB.On("Put", fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), expectedStoredOfficial).Return(nil).Once()

	result, err := base.upsertOfficialDividendsMetadata(ticker, fetchedOfficial)

	assert.NoError(t, err)
	assert.Equal(t, expectedReturned, result)

	cachedVal, found := c.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker))
	assert.True(t, found)
	assert.Equal(t, result, cachedVal)

	mockDB.AssertExpectations(t)
}
