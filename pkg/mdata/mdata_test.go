package mdata

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"portfolio-manager/internal/config"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/mdata/sources"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
)

type stubDividendDataSource struct {
	stored  map[string][]types.DividendsMetadata
	deleted map[string]bool
}

func newStubDividendDataSource() *stubDividendDataSource {
	return &stubDividendDataSource{
		stored:  make(map[string][]types.DividendsMetadata),
		deleted: make(map[string]bool),
	}
}

func (s *stubDividendDataSource) GetAssetPrice(string) (*types.AssetData, error) {
	return nil, nil
}

func (s *stubDividendDataSource) GetHistoricalData(string, int64, int64) ([]*types.AssetData, bool, error) {
	return nil, false, nil
}

func (s *stubDividendDataSource) GetDividendsMetadata(ticker string, _ float64) ([]types.DividendsMetadata, error) {
	return s.stored[ticker], nil
}

func (s *stubDividendDataSource) StoreDividendsMetadata(ticker string, dividends []types.DividendsMetadata, _ bool) ([]types.DividendsMetadata, error) {
	s.stored[ticker] = dividends
	return dividends, nil
}

func (s *stubDividendDataSource) DeleteDividendsMetadata(ticker string, isCustom bool) error {
	s.deleted[ticker] = isCustom
	delete(s.stored, ticker)
	return nil
}

func (s *stubDividendDataSource) FetchBenchmarkInterestRates(string, int) ([]types.InterestRates, error) {
	return nil, nil
}

func TestMapDomicileToWitholdingTax(t *testing.T) {
	// Mock configuration
	cfg := &config.Config{
		Dividends: config.DividendsConfig{
			WithholdingTaxSG: 0.1,
			WithholdingTaxUS: 0.2,
			WithholdingTaxHK: 0.15,
			WithholdingTaxIE: 0.05,
		},
	}
	config.SetConfig(cfg)

	manager := &Manager{}

	tests := []struct {
		domicile string
		expected float64
	}{
		{"SG", 0.1},
		{"US", 0.2},
		{"HK", 0.15},
		{"IE", 0.05},
		{"UNKNOWN", 0.0},
	}

	for _, test := range tests {
		t.Run(test.domicile, func(t *testing.T) {
			result := manager.MapDomicileToWitholdingTax(test.domicile)
			if result != test.expected {
				t.Errorf("expected %f, got %f", test.expected, result)
			}
		})
	}
}

func TestStoreAndDeleteCustomDividendsMetadata_UpdatesNasdaqAndYahooSources(t *testing.T) {
	mockRdata := mocks.NewMockReferenceManager()
	_ = mockRdata.UpdateTicker(&rdata.TickerReference{
		ID:           "AAPL",
		NasdaqTicker: "AAPL",
		YahooTicker:  "AAPL",
	})

	nasdaqSrc := newStubDividendDataSource()
	yahooSrc := newStubDividendDataSource()

	manager := &Manager{
		rdata: mockRdata,
		sources: map[string]types.DataSource{
			sources.Nasdaq:       nasdaqSrc,
			sources.YahooFinance: yahooSrc,
		},
	}

	customDividends := []types.DividendsMetadata{
		{Ticker: "AAPL", ExDate: "2025-02-10", Amount: 0.25, Source: types.DividendSourceCustom},
	}

	err := manager.StoreCustomDividendsMetadata("AAPL", customDividends)
	assert.NoError(t, err)
	assert.Equal(t, customDividends, nasdaqSrc.stored["AAPL"])
	assert.Equal(t, customDividends, yahooSrc.stored["AAPL"])

	err = manager.DeleteDividendsMetadata("AAPL", true)
	assert.NoError(t, err)
	assert.True(t, nasdaqSrc.deleted["AAPL"])
	assert.True(t, yahooSrc.deleted["AAPL"])
}
