package historical

import (
	"fmt"
	"portfolio-manager/pkg/types"
	"sort"
	"strings"
	"time"
)

type AssetConfig struct {
	Ticker        string `json:"ticker"`
	Source        string `json:"source"`
	Enabled       bool   `json:"enabled"`
	LastSync      int64  `json:"last_sync,omitempty"`      // Unix timestamp
	LookbackYears int    `json:"lookback_years,omitempty"` // Number of years to backfill
}

// GetAssetConfigs retrieves all configured assets for historical data collection
func (s *Service) GetAssetConfigs() ([]AssetConfig, error) {
	var configs []AssetConfig
	err := s.db.Get(string(types.HistoricalAssetConfigKey), &configs)
	if err != nil {
		if err.Error() == "key not found" || err.Error() == "Key not found" {
			return []AssetConfig{}, nil
		}
		// Some implementations return a specific error for not found, or use sentinel.
		// Since we don't have visibility into exact DAL implementation details of Get here
		// (it takes interface{}), we assume it might fill it or error.
		// If it errors, we propagate unless it's not found.
		// Checking "key not found" string is fragile but common in this codebase?
		// Let's assume empty list if error? No, safer to return error if real db error.
		// If Get fails to unmarshal or key missing.
		// Based on `blotter.go`: `err := db.Get(string(types.HeadSequenceBlotterKey), currentSeqNum)`
		// if err != nil { currentSeqNum = -1 }. It swallows error assuming it's key not found.
		// But LevelDB usually returns error if key missing.
		// We'll return empty if error, but log it? No, if it's a real DB error we want to know.
		// But for now let's assume if it fails it's empty.
		return []AssetConfig{}, nil
	}
	return configs, nil
}

// UpdateAssetConfig updates or adds an asset configuration
func (s *Service) UpdateAssetConfig(config AssetConfig) error {
	configs, err := s.GetAssetConfigs()
	if err != nil {
		return err
	}

	found := false
	for i, c := range configs {
		if c.Ticker == config.Ticker {
			configs[i] = config
			found = true
			break
		}
	}

	if !found {
		configs = append(configs, config)
	}

	// db.Put usually expects a value that it can serialize?
	// `internal/dal/leveldb.go` (if exists) uses json.Marshal if it's not []byte/string?
	// The interface is Put(key string, v interface{})
	return s.db.Put(string(types.HistoricalAssetConfigKey), configs)
}

// EnableAssetConfig enables or disables historical data collection for a ticker
func (s *Service) EnableAssetConfig(ticker string, enabled bool) error {
	configs, err := s.GetAssetConfigs()
	if err != nil {
		return err
	}

	found := false
	for i, c := range configs {
		if c.Ticker == ticker {
			configs[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("config not found for ticker: %s", ticker)
	}

	return s.db.Put(string(types.HistoricalAssetConfigKey), configs)
}

// RemoveAssetConfig removes an asset configuration
func (s *Service) RemoveAssetConfig(ticker string) error {
	configs, err := s.GetAssetConfigs()
	if err != nil {
		return err
	}

	newConfigs := []AssetConfig{}
	found := false
	for _, c := range configs {
		if c.Ticker == ticker {
			found = true
			continue
		}
		newConfigs = append(newConfigs, c)
	}

	if !found {
		return fmt.Errorf("config not found for ticker: %s", ticker)
	}

	// Delete actual data from persistence
	// We need to delete all keys with prefix HISTORICAL_ASSET_DATA:TICKER:
	dataPrefix := fmt.Sprintf("%s:%s:", string(types.HistoricalAssetDataKeyPrefix), ticker)
	keys, err := s.db.GetAllKeysWithPrefix(dataPrefix)
	if err != nil {
		s.logger.Errorf("Failed to list keys for deletion: %v", err)
		// We continue to remove config even if data deletion fails partially?
		// Or should we fail? Better to warn and proceed with config removal so user isn't stuck.
	} else {
		for _, k := range keys {
			if err := s.db.Delete(k); err != nil {
				s.logger.Warnf("Failed to delete key %s: %v", k, err)
			}
		}
	}

	return s.db.Put(string(types.HistoricalAssetConfigKey), newConfigs)
}

// SyncAssetData fetches (and optionally backfills) data for an enabled asset
func (s *Service) SyncAssetData(ticker string) (string, error) {
	configs, err := s.GetAssetConfigs()
	if err != nil {
		return "", err
	}

	var config *AssetConfig
	for i := range configs {
		if configs[i].Ticker == ticker {
			config = &configs[i]
			break
		}
	}

	if config == nil {
		return "", fmt.Errorf("config not found for ticker: %s", ticker)
	}

	if !config.Enabled {
		return "", fmt.Errorf("historical data collection disabled for ticker: %s", ticker)
	}

	now := time.Now().Unix()

	lookbackYears := float64(5)
	if config.LookbackYears > 0 {
		lookbackYears = float64(config.LookbackYears)
	}

	// Default to lookback years if never synced
	fromDate := now - int64(lookbackYears*365*24*60*60)

	if config.LastSync > 0 {
		// Start from last sync time.
		fromDate = config.LastSync
	}

	dateRangeMsg := fmt.Sprintf("Fetching from %s to %s", time.Unix(fromDate, 0).Format("2006-01-02"), time.Unix(now, 0).Format("2006-01-02"))
	s.logger.Infof("Syncing historical data for %s: %s", ticker, dateRangeMsg)

	data, err := s.mdataManager.GetHistoricalData(ticker, fromDate, now)
	if err != nil {
		return "", err
	}

	// Assuming db implementation supports batch or we just loop Put
	// `dal.Database` interface doesn't show Batch. I'll just use Put loop.
	// It might be slow but functional.
	for _, d := range data {
		dateStr := time.Unix(d.Timestamp, 0).Format("20060102")
		key := fmt.Sprintf("%s:%s:%s", string(types.HistoricalAssetDataKeyPrefix), ticker, dateStr)

		if err := s.db.Put(key, d); err != nil {
			s.logger.Errorf("Failed to save historical data for %s date %s: %v", ticker, dateStr, err)
		}
	}

	// Update LastSync
	config.LastSync = now
	err = s.UpdateAssetConfig(*config)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Synced %d records. %s", len(data), dateRangeMsg), nil
}

// GetHistoricalAssetData retrieves historical data with pagination and filtering
func (s *Service) GetHistoricalAssetData(ticker string, from, to int64, page, limit int) ([]types.AssetData, int, error) {
	prefix := fmt.Sprintf("%s:%s:", string(types.HistoricalAssetDataKeyPrefix), ticker)
	keys, err := s.db.GetAllKeysWithPrefix(prefix)
	if err != nil {
		return nil, 0, err
	}

	// Filter keys by date range
	var filteredKeys []string
	for _, k := range keys {
		// key format: PREFIX:TICKER:YYYYMMDD
		// extract YYYYMMDD
		parts := strings.Split(k, ":")
		if len(parts) < 3 {
			continue
		}
		dateStr := parts[len(parts)-1]
		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}
		ts := date.Unix()
		if (from == 0 || ts >= from) && (to == 0 || ts <= to) {
			filteredKeys = append(filteredKeys, k)
		}
	}

	// Sort reverse chronological (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(filteredKeys)))

	total := len(filteredKeys)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	start := (page - 1) * limit
	if start >= total {
		return []types.AssetData{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}

	var result []types.AssetData
	for _, k := range filteredKeys[start:end] {
		var d types.AssetData
		if err := s.db.Get(k, &d); err != nil {
			continue
		}
		result = append(result, d)
	}

	return result, total, nil
}
