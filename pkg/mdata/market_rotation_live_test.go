//go:build integration

package mdata

import (
	"encoding/json"
	"testing"
	"time"

	"portfolio-manager/pkg/mdata/sources"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/require"
)

func TestLiveMarketRotationDryRun(t *testing.T) {
	manager := &Manager{screenerSources: map[string]types.ScreenerSource{sources.TradingView: sources.NewTradingView()}}
	brief, err := manager.ScreenDailyMarketRotation(MarketRotationOptions{PersistHistory: false, MaxStockCandidates: 5})
	require.NoError(t, err)
	require.NotEmpty(t, brief.SectorFundFlows)
	require.LessOrEqual(t, len(brief.StockCandidates), 5)
	require.False(t, brief.DataQuality.Persisted)

	payload, err := json.Marshal(brief)
	require.NoError(t, err)
	require.Less(t, len(payload), 25_000, "MCP payload should remain compact")
	t.Logf("live dry run at %s (%d bytes):\n%s", time.Now().Format(time.RFC3339), len(payload), payload)
}
