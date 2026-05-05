package blotter_test

import (
	"errors"
	"testing"
	"time"

	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/mocks/testify"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildTradeCreatesCanonicalOptionTickerAndReference(t *testing.T) {
	db, dbPath := setupTempDB(t)
	defer cleanupTempDB(t, db, dbPath)

	blotterSvc := blotter.NewBlotter(db)
	mockRdataSvc := new(testify.MockReferenceManager)
	mockMdataSvc := new(testify.MockMarketDataManager)
	blotterSvc.SetTradeSupportServices(mockRdataSvc, mockMdataSvc)

	tradeDate := time.Date(2026, time.April, 9, 9, 0, 0, 0, time.UTC)
	underlyingRef := rdata.TickerReferenceWithSGXMapped{
		TickerReference: rdata.TickerReference{
			ID:            "AAPL",
			Name:          "Apple Inc",
			AssetClass:    rdata.AssetClassEquities,
			AssetSubClass: rdata.AssetSubClassStock,
			Category:      rdata.CategoryTechnology,
			Ccy:           "USD",
			Domicile:      "US",
		},
	}
	optionTicker, err := blotter.BuildOptionTicker("AAPL", "2026-06-19", 200, blotter.CallPutCall)
	assert.NoError(t, err)

	mockMdataSvc.
		On("GetHistoricalData", "AAPL", mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]*types.AssetData{{Ticker: "AAPL", Price: 189.5}}, false, nil)
	mockRdataSvc.On("GetTicker", "AAPL").Return(underlyingRef, nil)
	mockRdataSvc.On("GetTicker", optionTicker).Return(rdata.TickerReferenceWithSGXMapped{}, errors.New("not found"))
	mockRdataSvc.On("AddTicker", mock.MatchedBy(func(ref rdata.TickerReference) bool {
		return ref.ID == optionTicker &&
			ref.Name == "Apple Inc 2026-06-19 CALL 200" &&
			ref.UnderlyingTicker == "AAPL" &&
			ref.AssetSubClass == rdata.AssetSubClassOption &&
			ref.MaturityDate == "2026-06-19" &&
			ref.StrikePrice == 200 &&
			ref.CallPut == blotter.CallPutCall
	})).Return(optionTicker, nil)

	trade, err := blotterSvc.BuildTrade(blotter.TradeInput{
		TradeDate: tradeDate,
		Side:      blotter.TradeSideBuy,
		Quantity:  2,
		Price:     12.5,
		Fx:        1,
		Book:      "Growth",
		Broker:    "DBS",
		Account:   "CDP",
		Status:    blotter.StatusOpen,
		Attributes: blotter.TradeAttributes{
			InstrumentType:   blotter.InstrumentTypeOption,
			UnderlyingTicker: "AAPL",
			ExpiryDate:       "2026-06-19",
			StrikePrice:      200,
			CallPut:          blotter.CallPutCall,
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, optionTicker, trade.Ticker)
	assert.Equal(t, blotter.InstrumentTypeOption, trade.InstrumentType)
	assert.Equal(t, "AAPL", trade.UnderlyingTicker)
	assert.Equal(t, 189.5, trade.UnderlyingSpotRef)
	assert.Equal(t, "2026-06-19", trade.ExpiryDate)

	mockRdataSvc.AssertExpectations(t)
	mockMdataSvc.AssertExpectations(t)
}

func TestBuildOptionTickerRejectsDecimalStrike(t *testing.T) {
	optionTicker, err := blotter.BuildOptionTicker("AAPL", "2026-06-19", 200.5, blotter.CallPutCall)
	assert.ErrorContains(t, err, "strike price must be a whole number")
	assert.Empty(t, optionTicker)
}
