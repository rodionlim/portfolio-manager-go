package sources

import (
	"fmt"
	"testing"
	"time"

	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/types"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreDividendsMetadata_Custom_Overlap(t *testing.T) {
	// Setup
	mockDB := new(mocks.MockDatabase)
	c := cache.New(5*time.Minute, 10*time.Minute)
	base := &BaseDividendSource{
		db:    mockDB,
		cache: c,
	}

	ticker := "AAPL"
	isCustom := true

	// Existing official dividends
	officialDividends := []types.DividendsMetadata{
		{ExDate: "2023-01-01", Amount: 0.5},
	}

	// New custom dividends (overlapping date)
	customDividends := []types.DividendsMetadata{
		{ExDate: "2023-01-01", Amount: 1.0},
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
