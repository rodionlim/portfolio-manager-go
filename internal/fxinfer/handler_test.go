package fxinfer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/mocks/testify"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandleGetCurrentFXRates(t *testing.T) {
	// Setup mock services
	mockBlotterSvc := new(testify.MockTradeBlotter)
	mockMdataSvc := new(testify.MockMarketDataManager)
	mockRdataSvc := new(testify.MockReferenceManager)

	// Create test trades with different tickers
	trades := []blotter.Trade{
		{
			TradeID:   "trade1",
			Ticker:    "AAPL",
			TradeDate: time.Now().Format(time.RFC3339),
			Side:      "buy",
			Quantity:  100,
			Price:     150.0,
			Book:      "book1",
			Broker:    "broker1",
			Account:   "account1",
		},
		{
			TradeID:   "trade2",
			Ticker:    "MSFT",
			TradeDate: time.Now().Format(time.RFC3339),
			Side:      "buy",
			Quantity:  50,
			Price:     250.0,
			Book:      "book1",
			Broker:    "broker1",
			Account:   "account1",
		},
		{
			TradeID:   "trade3",
			Ticker:    "AAPL", // Duplicate ticker to test caching
			TradeDate: time.Now().Format(time.RFC3339),
			Side:      "sell",
			Quantity:  25,
			Price:     160.0,
			Book:      "book2",
			Broker:    "broker2",
			Account:   "account2",
		},
	}

	// Setup mock blotter to return our test trades
	mockBlotterSvc.On("GetTrades").Return(trades)

	// Setup reference data for tickers
	tickerReferenceAAPL := rdata.TickerReferenceWithSGXMapped{
		TickerReference: rdata.TickerReference{
			ID:  "AAPL",
			Ccy: "USD",
		},
	}
	mockRdataSvc.On("GetTicker", "AAPL").Return(tickerReferenceAAPL, nil)

	tickerReferenceMSFT := rdata.TickerReferenceWithSGXMapped{
		TickerReference: rdata.TickerReference{
			ID:  "MSFT",
			Ccy: "USD",
		},
	}
	mockRdataSvc.On("GetTicker", "MSFT").Return(tickerReferenceMSFT, nil)

	// Setup market data responses for FX rates
	usdSgdAssetData := &types.AssetData{
		Ticker:    "USD-SGD",
		Price:     1.33,
		Currency:  "SGD",
		Timestamp: time.Now().Unix(),
	}
	mockMdataSvc.On("GetAssetPrice", "USD-SGD").Return(usdSgdAssetData, nil)

	// Create service with the mocks
	service := NewFXInferenceService(
		mockBlotterSvc,
		mockMdataSvc,
		mockRdataSvc,
		"SGD",
	)

	// Create HTTP request and response recorder
	req := httptest.NewRequest("GET", "/api/v1/blotter/fx", nil)
	res := httptest.NewRecorder()

	// Call the handler
	handler := HandleGetCurrentFXRates(service)
	handler.ServeHTTP(res, req)

	// Assert response status code
	assert.Equal(t, http.StatusOK, res.Code)

	// Parse the response body
	var fxRates map[string]float64
	err := json.Unmarshal(res.Body.Bytes(), &fxRates)
	assert.NoError(t, err)

	// Assert the expected rates
	assert.Equal(t, float64(1.0), fxRates["SGD"], "Base currency SGD should have rate 1.0")
	assert.InDelta(t, 1.33, fxRates["USD"], 0.0001, "USD rate should be approximately 1.33")

	// Verify that all expected methods were called
	mockBlotterSvc.AssertExpectations(t)
	mockMdataSvc.AssertExpectations(t)
	mockRdataSvc.AssertExpectations(t)
}
