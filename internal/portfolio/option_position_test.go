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

	"github.com/stretchr/testify/assert"
)

func setupOptionPortfolio(t *testing.T) (*Portfolio, dal.Database, string) {
	dbPath := filepath.Join(os.TempDir(), "portfolio_option_"+t.Name())
	db, err := dal.NewLevelDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create temp database: %v", err)
	}

	rdataMgr, err := rdata.NewManager(db, "")
	if err != nil {
		t.Fatalf("failed to create reference manager: %v", err)
	}

	_, err = rdataMgr.AddTicker(rdata.TickerReference{
		ID:            "AAPL",
		Name:          "Apple Inc",
		AssetClass:    rdata.AssetClassEquities,
		AssetSubClass: rdata.AssetSubClassStock,
		Category:      rdata.CategoryTechnology,
		Ccy:           "USD",
		Domicile:      "US",
	})
	if err != nil {
		t.Fatalf("failed to add underlying ticker: %v", err)
	}

	optionTicker, err := blotter.BuildOptionTicker("AAPL", "2026-06-19", 200, blotter.CallPutCall)
	if err != nil {
		t.Fatalf("failed to build option ticker: %v", err)
	}

	_, err = rdataMgr.AddTicker(rdata.TickerReference{
		ID:               optionTicker,
		Name:             "Apple Inc 2026-06-19 CALL 200.0000",
		UnderlyingTicker: "AAPL",
		AssetClass:       rdata.AssetClassEquities,
		AssetSubClass:    rdata.AssetSubClassOption,
		Category:         rdata.CategoryTechnology,
		Ccy:              "USD",
		Domicile:         "US",
		MaturityDate:     "2026-06-19",
		StrikePrice:      200,
		CallPut:          blotter.CallPutCall,
	})
	if err != nil {
		t.Fatalf("failed to add option ticker: %v", err)
	}

	mdataMgr, err := mdata.NewManager(db, rdataMgr)
	if err != nil {
		t.Fatalf("failed to create market data manager: %v", err)
	}

	dividendsMgr := dividends.NewDividendsManager(db, mdataMgr, rdataMgr, nil)
	return NewPortfolio(db, mdataMgr, rdataMgr, dividendsMgr), db, dbPath
}

func teardownOptionPortfolio(t *testing.T, db dal.Database, dbPath string) {
	assert.NoError(t, db.Close())
	assert.NoError(t, os.RemoveAll(dbPath))
}

func TestOpenOptionPositionHasZeroMarketValueAndGrouping(t *testing.T) {
	portfolioSvc, db, dbPath := setupOptionPortfolio(t)
	defer teardownOptionPortfolio(t, db, dbPath)

	optionTicker, _ := blotter.BuildOptionTicker("AAPL", "2026-06-19", 200, blotter.CallPutCall)
	trade, err := blotter.NewTrade(
		blotter.TradeSideBuy,
		2,
		optionTicker,
		"Growth",
		"DBS",
		"CDP",
		blotter.StatusOpen,
		"",
		12.5,
		1,
		0,
		time.Now(),
		blotter.TradeAttributes{
			InstrumentType:    blotter.InstrumentTypeOption,
			UnderlyingTicker:  "AAPL",
			UnderlyingSpotRef: 189.5,
			ExpiryDate:        "2026-06-19",
			StrikePrice:       200,
			CallPut:           blotter.CallPutCall,
		},
	)
	assert.NoError(t, err)

	assert.NoError(t, portfolioSvc.updatePosition(trade))

	position, err := portfolioSvc.GetPosition("Growth", optionTicker)
	assert.NoError(t, err)
	assert.Equal(t, blotter.InstrumentTypeOption, position.InstrumentType)
	assert.Equal(t, "AAPL", position.UnderlyingTicker)
	assert.Equal(t, "AAPL", position.UnderlyingGroup)
	assert.Equal(t, 0.0, position.Mv)
	assert.Equal(t, 0.0, position.PnL)
}

func TestClosedOptionPositionRealizesPnLFromCashFlows(t *testing.T) {
	portfolioSvc, db, dbPath := setupOptionPortfolio(t)
	defer teardownOptionPortfolio(t, db, dbPath)

	optionTicker, _ := blotter.BuildOptionTicker("AAPL", "2026-06-19", 200, blotter.CallPutCall)
	openTrade, err := blotter.NewTrade(
		blotter.TradeSideBuy,
		1,
		optionTicker,
		"Growth",
		"DBS",
		"CDP",
		blotter.StatusOpen,
		"",
		10,
		1,
		0,
		time.Now(),
		blotter.TradeAttributes{InstrumentType: blotter.InstrumentTypeOption, UnderlyingTicker: "AAPL", ExpiryDate: "2026-06-19", StrikePrice: 200, CallPut: blotter.CallPutCall},
	)
	assert.NoError(t, err)
	closeTrade, err := blotter.NewTrade(
		blotter.TradeSideSell,
		1,
		optionTicker,
		"Growth",
		"DBS",
		"CDP",
		blotter.StatusClosed,
		"",
		15,
		1,
		0,
		time.Now(),
		blotter.TradeAttributes{InstrumentType: blotter.InstrumentTypeOption, UnderlyingTicker: "AAPL", ExpiryDate: "2026-06-19", StrikePrice: 200, CallPut: blotter.CallPutCall},
	)
	assert.NoError(t, err)

	assert.NoError(t, portfolioSvc.updatePosition(openTrade))
	assert.NoError(t, portfolioSvc.updatePosition(closeTrade))

	position, err := portfolioSvc.GetPosition("Growth", optionTicker)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, position.Qty)
	assert.Equal(t, 5.0, position.PnL)
	assert.Equal(t, 0.0, position.Mv)
}
