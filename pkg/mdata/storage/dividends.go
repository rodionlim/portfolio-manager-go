package storage

import (
	"fmt"
	"sort"

	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/types"
)

// LoadDividends reads persisted history only, with custom dates overriding official
// dates. Database errors and missing history must not become zero dividend income.
func LoadDividends(db dal.Database, ticker string) ([]types.DividendsMetadata, error) {
	if db == nil {
		return nil, fmt.Errorf("dividend database is unavailable")
	}
	byDate := make(map[string][]types.DividendsMetadata)
	for _, source := range []struct{ prefix, name string }{
		{string(types.DividendsKeyPrefix), types.DividendSourceOfficial},
		{string(types.DividendsCustomKeyPrefix), types.DividendSourceCustom},
	} {
		key := source.prefix + ":" + ticker
		keys, err := db.GetAllKeysWithPrefix(key)
		if err != nil {
			return nil, err
		}
		for _, found := range keys {
			if found != key {
				continue
			}
			var rows []types.DividendsMetadata
			if err := db.Get(key, &rows); err != nil {
				return nil, err
			}
			grouped := make(map[string][]types.DividendsMetadata)
			for _, row := range rows {
				row.Source = source.name
				grouped[row.ExDate] = append(grouped[row.ExDate], row)
			}
			for date, rows := range grouped {
				byDate[date] = rows
			}
		}
	}
	var result []types.DividendsMetadata
	for _, rows := range byDate {
		result = append(result, rows...)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no stored dividend history for %s; load or import its history before marking it complete", ticker)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ExDate < result[j].ExDate })
	return result, nil
}
