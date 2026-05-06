package portfolio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dal"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/pkg/mdata"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStartupRebuildPortfolio(t *testing.T) (dal.Database, *Portfolio, *blotter.TradeBlotter, string) {
	t.Helper()

	dbPath := filepath.Join(os.TempDir(), "portfolio_startup_rebuild_"+t.Name())
	db, err := dal.NewLevelDB(dbPath)
	require.NoError(t, err)

	rdataMgr, err := rdata.NewManager(db, "")
	require.NoError(t, err)

	mdataMgr, err := mdata.NewManager(db, rdataMgr)
	require.NoError(t, err)

	dividendsMgr := dividends.NewDividendsManager(db, mdataMgr, rdataMgr, nil)
	portfolioSvc := NewPortfolio(db, mdataMgr, rdataMgr, dividendsMgr)
	blotterSvc := blotter.NewBlotter(db)

	return db, portfolioSvc, blotterSvc, dbPath
}

func teardownStartupRebuildPortfolio(t *testing.T, db dal.Database, dbPath string) {
	t.Helper()
	assert.NoError(t, db.Close())
	assert.NoError(t, os.RemoveAll(dbPath))
}

func TestSubscribeToBlotterRebuildsWhenSequenceHeadMissing(t *testing.T) {
	db, portfolioSvc, blotterSvc, dbPath := setupStartupRebuildPortfolio(t)
	defer teardownStartupRebuildPortfolio(t, db, dbPath)

	portfolioSvc.SubscribeToBlotter(blotterSvc)

	seedTrades := []*blotter.Trade{
		must(blotter.NewTrade(blotter.TradeSideBuy, 100, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 150.0, 1, 0.0, time.Now())),
		must(blotter.NewTrade(blotter.TradeSideSell, 40, "AAPL", "book1", "broker1", "cdp", blotter.StatusOpen, "", 160.0, 1, 0.0, time.Now())),
	}

	for _, trade := range seedTrades {
		require.NoError(t, blotterSvc.AddTrade(*trade))
	}

	positions := portfolioSvc.GetAllPositionsWithoutEnrichment()
	require.Len(t, positions, 1)
	assert.Equal(t, 60.0, positions[0].Qty)

	require.NoError(t, db.Delete(string(types.HeadSequencePortfolioKey)))

	restartedBlotter := blotter.NewBlotter(db)
	require.NoError(t, restartedBlotter.LoadFromDB())

	rdataMgr, err := rdata.NewManager(db, "")
	require.NoError(t, err)
	mdataMgr, err := mdata.NewManager(db, rdataMgr)
	require.NoError(t, err)
	dividendsMgr := dividends.NewDividendsManager(db, mdataMgr, rdataMgr, nil)
	restartedPortfolio := NewPortfolio(db, mdataMgr, rdataMgr, dividendsMgr)
	require.NoError(t, restartedPortfolio.LoadPositions())

	loadedPositions := restartedPortfolio.GetAllPositionsWithoutEnrichment()
	require.Len(t, loadedPositions, 1)
	assert.Equal(t, 60.0, loadedPositions[0].Qty)

	restartedPortfolio.SubscribeToBlotter(restartedBlotter)

	rebuiltPositions := restartedPortfolio.GetAllPositionsWithoutEnrichment()
	require.Len(t, rebuiltPositions, 1)
	assert.Equal(t, 60.0, rebuiltPositions[0].Qty)

	var savedSeqNum int
	require.NoError(t, db.Get(string(types.HeadSequencePortfolioKey), &savedSeqNum))
	assert.Equal(t, restartedBlotter.GetCurrentSeqNum(), savedSeqNum)
}
