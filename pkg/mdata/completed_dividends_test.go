package mdata

import (
	"errors"
	"path/filepath"
	"testing"

	"portfolio-manager/internal/config"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/mdata/sources"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/require"
)

type completionTestSource struct {
	types.DataSource
	calls int
}

func (s *completionTestSource) GetDividendsMetadata(string, float64) ([]types.DividendsMetadata, error) {
	s.calls++
	return nil, errors.New("upstream unavailable")
}

func TestCompletedBondUsesStoredHistoryAndCanResume(t *testing.T) {
	config.SetConfig(&config.Config{})
	path := filepath.Join(t.TempDir(), "history")
	db, err := dal.NewLevelDB(path)
	require.NoError(t, err)
	ref := rdata.TickerReference{ID: "BOND", AssetClass: rdata.AssetClassBonds, DividendsSgTicker: "6AZB", DividendHistoryComplete: true}
	refs, err := rdata.NewManager(db, "")
	require.NoError(t, err)
	_, err = refs.AddTicker(ref)
	require.ErrorIs(t, err, rdata.ErrDividendCompletion, "missing history must not silently become zero")
	require.NoError(t, db.Put(string(types.DividendsKeyPrefix)+":6AZB", []types.DividendsMetadata{
		{Ticker: "6AZB", ExDate: "2021-09-10", Amount: 0.015},
		{Ticker: "6AZB", ExDate: "2026-03-10", Amount: 0.015},
	}))
	require.NoError(t, db.Put(string(types.DividendsCustomKeyPrefix)+":6AZB", []types.DividendsMetadata{
		{Ticker: "6AZB", ExDate: "2026-03-10", Amount: 0},
	}))
	_, err = refs.AddTicker(ref)
	require.NoError(t, err)
	// Reopen the actual database: both the decision and historical rows must survive restart.
	require.NoError(t, db.Close())
	db, err = dal.NewLevelDB(path)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	refs, err = rdata.NewManager(db, "")
	require.NoError(t, err)
	source := &completionTestSource{}
	manager := &Manager{db: db, rdata: refs, sources: map[string]types.DataSource{sources.DividendsSingapore: source}}
	rows, err := manager.GetDividendsMetadata("BOND")
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "2021-09-10", rows[0].ExDate)
	require.Equal(t, types.DividendSourceOfficial, rows[0].Source)
	require.Zero(t, rows[1].Amount)
	require.Equal(t, types.DividendSourceCustom, rows[1].Source)
	require.Zero(t, source.calls, "completed bonds must not contact upstream")

	ref.DividendHistoryComplete = false
	require.NoError(t, refs.UpdateTicker(&ref))
	_, err = manager.GetDividendsMetadata("BOND")
	require.Error(t, err)
	require.Equal(t, 1, source.calls, "resuming must restore normal source fetching")

	ref.AssetClass = rdata.AssetClassEquities
	ref.DividendHistoryComplete = true
	require.ErrorIs(t, refs.UpdateTicker(&ref), rdata.ErrDividendCompletion)
	// Defend against invalid flags already in data or passed by a caller.
	_, err = manager.GetDividendsMetadataFromTickerRef(ref)
	require.Error(t, err)
	require.Equal(t, 2, source.calls, "equities must not bypass upstream")

	ref.AssetClass = rdata.AssetClassBonds
	require.NoError(t, db.Put(string(types.DividendsKeyPrefix)+":6AZB", "corrupt history"))
	_, err = manager.GetDividendsMetadataFromTickerRef(ref)
	require.ErrorContains(t, err, "unmarshal")
	require.Equal(t, 2, source.calls, "a completed bond must report unusable stored history without fetching")
}
