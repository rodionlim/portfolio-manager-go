package metrics_test

import (
	"errors"
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/dividends"
	"portfolio-manager/internal/metrics"
	"portfolio-manager/internal/mocks/testify"
	"portfolio-manager/internal/portfolio"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateIRR(t *testing.T) {
	// Setup mocks
	mockBlotter := new(testify.MockBlotterTradeGetter)
	mockPortfolio := new(testify.MockPortfolioGetter)
	mockDividends := new(testify.MockDividendsManager)
	mockMdataSvc := new(testify.MockMarketDataManager)
	mockRdataSvc := new(testify.MockReferenceManager)

	// Create sample data
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	sixMonthsAgo := now.AddDate(0, -6, 0)

	trades := []blotter.Trade{
		{
			TradeID:   "trade1",
			Ticker:    "AAPL",
			Side:      blotter.TradeSideBuy,
			Quantity:  10,
			Price:     150.0,
			Fx:        1.0,
			TradeDate: oneYearAgo.Format(time.RFC3339),
		},
		{
			TradeID:   "trade2",
			Ticker:    "GOOGL",
			Side:      blotter.TradeSideBuy,
			Quantity:  5,
			Price:     1000.0,
			Fx:        1.0,
			TradeDate: sixMonthsAgo.Format(time.RFC3339),
		},
	}

	dividendsMap := map[string][]dividends.Dividends{
		"AAPL": {
			{
				ExDate:         sixMonthsAgo.Format("2006-01-02"),
				Amount:         50.0,
				AmountPerShare: 5.0,
				Qty:            10,
			},
		},
	}

	positions := []*portfolio.Position{
		{
			Ticker: "AAPL",
			Qty:    10,
			Mv:     2000.0,
			FxRate: 1.0,
		},
		{
			Ticker: "GOOGL",
			Qty:    5,
			Mv:     6000.0,
			FxRate: 1.0,
		},
	}

	// Setup ticker reference and FX rate data
	aaplRef := rdata.TickerReference{
		ID:  "AAPL",
		Ccy: "USD", // Non-SGD currency to test FX conversion
	}

	usdSgdRate := &types.AssetData{
		Ticker: "USD-SGD",
		Price:  1.35, // Sample SGD to USD rate
	}

	// Configure mock expectations
	mockBlotter.On("GetTrades").Return(trades)
	mockDividends.On("CalculateDividendsForAllTickers").Return(dividendsMap, nil)
	mockPortfolio.On("GetAllPositions").Return(positions, nil)
	mockRdataSvc.On("GetTicker", "AAPL").Return(aaplRef, nil)
	mockMdataSvc.On("GetAssetPrice", "USD-SGD").Return(usdSgdRate, nil)

	// Create the service
	service := metrics.NewMetricsService(mockBlotter, mockPortfolio, mockDividends, mockMdataSvc, mockRdataSvc)

	// Test the CalculatePortfolioMetrics method
	res, err := service.CalculatePortfolioMetrics("")

	// Verify expectations
	mockBlotter.AssertExpectations(t)
	mockDividends.AssertExpectations(t)
	mockPortfolio.AssertExpectations(t)
	mockRdataSvc.AssertExpectations(t)
	mockMdataSvc.AssertExpectations(t)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, res.Metrics.IRR)
	assert.True(t, res.Metrics.IRR > 0, "IRR should be positive for a profitable portfolio")

	// Verify that the dividend amount has been multiplied by the FX rate
	var dividendCashFlow *metrics.CashFlow
	for _, cf := range res.CashFlows {
		if cf.Description == metrics.CashFlowTypeDividend && cf.Ticker == "AAPL" {
			dividendCashFlow = &cf
			break
		}
	}

	assert.NotNil(t, dividendCashFlow, "Should have a dividend cash flow for AAPL")
	if dividendCashFlow != nil {
		// The original amount was 50.0, and with FX rate of 1.35, it should be 67.5
		assert.InDelta(t, 67.5, dividendCashFlow.Cash, 0.01, "Dividend should be converted to SGD")
	}
}

func TestCalculateIRR_Error(t *testing.T) {
	// Setup mocks
	mockBlotter := new(testify.MockBlotterTradeGetter)
	mockPortfolio := new(testify.MockPortfolioGetter)
	mockDividends := new(testify.MockDividendsManager)
	mockMdataSvc := new(testify.MockMarketDataManager)
	mockRdataSvc := new(testify.MockReferenceManager)

	// Create sample data
	trades := []blotter.Trade{}

	// Configure mock expectations
	mockBlotter.On("GetTrades").Return(trades)
	mockDividends.On("CalculateDividendsForAllTickers").Return(nil, errors.New("dividends error"))

	// Create the service
	service := metrics.NewMetricsService(mockBlotter, mockPortfolio, mockDividends, mockMdataSvc, mockRdataSvc)

	// Test the CalculateIRR method
	irr, err := service.CalculatePortfolioMetrics("")

	// Verify expectations
	mockBlotter.AssertExpectations(t)
	mockDividends.AssertExpectations(t)
	mockPortfolio.AssertNotCalled(t, "GetAllPositions")

	// Assert results
	assert.Error(t, err)
	assert.Zero(t, irr)
}

func TestCalculateIRR_SimpleProfitWithDividend(t *testing.T) {
	mockBlotter := new(testify.MockBlotterTradeGetter)
	mockPortfolio := new(testify.MockPortfolioGetter)
	mockDividends := new(testify.MockDividendsManager)
	mockMdataSvc := new(testify.MockMarketDataManager)
	mockRdataSvc := new(testify.MockReferenceManager)

	buyDate := time.Now().AddDate(-1, 0, 0)
	trades := []blotter.Trade{
		{
			TradeID:   "trade1",
			Ticker:    "AAPL",
			Side:      blotter.TradeSideBuy,
			Quantity:  1,
			Price:     100.0,
			Fx:        1.0,
			TradeDate: buyDate.Format(time.RFC3339),
		},
	}
	positions := []*portfolio.Position{
		{
			Ticker: "AAPL",
			Qty:    1,
			Mv:     110.0,
			FxRate: 1.0,
		},
	}
	dividendsMap := map[string][]dividends.Dividends{
		"AAPL": {
			{
				ExDate:         buyDate.AddDate(0, 6, 0).Format("2006-01-02"),
				Amount:         10.0,
				AmountPerShare: 10.0,
				Qty:            1,
			},
		},
	}
	aaplRef := rdata.TickerReference{ID: "AAPL", Ccy: "USD"}
	usdSgdRate := &types.AssetData{Ticker: "USD-SGD", Price: 1.0}

	mockBlotter.On("GetTrades").Return(trades)
	mockDividends.On("CalculateDividendsForAllTickers").Return(dividendsMap, nil)
	mockPortfolio.On("GetAllPositions").Return(positions, nil)
	mockRdataSvc.On("GetTicker", "AAPL").Return(aaplRef, nil)
	mockMdataSvc.On("GetAssetPrice", "USD-SGD").Return(usdSgdRate, nil)

	service := metrics.NewMetricsService(mockBlotter, mockPortfolio, mockDividends, mockMdataSvc, mockRdataSvc)
	res, err := service.CalculatePortfolioMetrics("")

	mockBlotter.AssertExpectations(t)
	mockDividends.AssertExpectations(t)
	mockPortfolio.AssertExpectations(t)
	mockRdataSvc.AssertExpectations(t)
	mockMdataSvc.AssertExpectations(t)

	assert.NoError(t, err)
	assert.InDelta(t, 0.20, res.Metrics.IRR, 0.01, "IRR should be approximately 20% for 10% capital gain + 10% dividend")
}
