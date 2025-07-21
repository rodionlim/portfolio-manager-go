package sources

import (
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"sort"

	"github.com/patrickmn/go-cache"
)

// BaseDividendSource provides common functionality for mdata sources that support dividends
type BaseDividendSource struct {
	db    dal.Database
	cache *cache.Cache
}

// GetDividendsMetadata retrieves either custom or official dividends metadata for a given ticker
func (base *BaseDividendSource) getSingleDividendsMetadata(ticker string, isCustom bool) ([]types.DividendsMetadata, error) {
	var err error
	if base.db != nil {
		var dividends []types.DividendsMetadata
		if !isCustom {
			err = base.db.Get(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), &dividends)
		} else {
			err = base.db.Get(fmt.Sprintf("%s:%s", types.DividendsCustomKeyPrefix, ticker), &dividends)
		}
		if err != nil {
			dividends = []types.DividendsMetadata{}
		}
		return dividends, err
	}

	return []types.DividendsMetadata{}, err
}

// StoreDividends stores either custom or official dividends metadata for a given ticker
func (base *BaseDividendSource) StoreDividendsMetadata(ticker string, dividends []types.DividendsMetadata, isCustom bool) ([]types.DividendsMetadata, error) {
	logger := logging.GetLogger()
	var err error

	if base.db == nil {
		logger.Warn("database is not initialized, skipping storing dividends")
		return nil, fmt.Errorf("database is not initialized")
	}

	// Custom dividends are dividends metadata manually added by user, either ways, storage means a cache invalidation
	// Upstream should not call this method if data has not been changed
	if !isCustom {
		err = base.db.Put(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), dividends)
		customDividendsMetadata, _ := base.getSingleDividendsMetadata(ticker, true)
		dividends = base.mergeAndSortDividendsMetadata(dividends, customDividendsMetadata)
	} else {
		err = base.db.Put(fmt.Sprintf("%s:%s", types.DividendsCustomKeyPrefix, ticker), dividends)
		officialdDividendsMetadata, _ := base.getSingleDividendsMetadata(ticker, false)
		dividends = base.mergeAndSortDividendsMetadata(officialdDividendsMetadata, dividends)
	}

	base.cache.Set(fmt.Sprintf("%s:%s", types.DividendsKeyPrefix, ticker), dividends, cache.DefaultExpiration)

	return dividends, err
}

// mergeAndSortDividends concatenates two slices of DividendsMetadata and sorts them by ExDate.
func (base *BaseDividendSource) mergeAndSortDividendsMetadata(dividendsLeft []types.DividendsMetadata, dividendsRight []types.DividendsMetadata) []types.DividendsMetadata {
	merged := append(dividendsLeft, dividendsRight...)
	sort.Slice(merged, func(i, j int) bool {
		// ExDate is in "yyyy-mm-dd" format, so lexicographical comparison works correctly.
		return merged[i].ExDate < merged[j].ExDate
	})
	return merged
}

// FetchBenchmarkInterestRates provides a default implementation for data sources that don't support interest rates
func (base *BaseDividendSource) FetchBenchmarkInterestRates(country string, points int) ([]types.InterestRates, error) {
	return nil, fmt.Errorf("benchmark interest rates not supported for this data source")
}
