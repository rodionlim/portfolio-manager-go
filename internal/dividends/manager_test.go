package dividends

import (
	"portfolio-manager/internal/blotter"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReferenceManager struct{}

func (m *mockReferenceManager) GetTicker(ticker string) (*rdata.TickerReference, error) {
	return &reference.TickerReference{
		DividendsSgTicker: "SGX:DIVIDEND",
	}, nil
}

type mockMdataManager struct{}

func (m *mockMdataManager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	return []types.DividendsMetadata{
		{ExDate: "2023-01-01", Amount: 1.0, WithholdingTax: 0.1},
		{ExDate: "2023-02-01", Amount: 2.0, WithholdingTax: 0.1},
	}, nil
}

type mockTradeBlotter struct{}

func (m *mockTradeBlotter) GetTradesByTicker(ticker string) ([]blotter.Trade, error) {
	return []blotter.Trade{
		{TradeDate: "2022-12-31", Quantity: 100},
		{TradeDate: "2023-01-15", Quantity: 200},
	}, nil
}

func TestCalculateDividendsForSingleTicker(t *testing.T) {
	db := &mocks.MockDatabase{}
	rdata := &mocks.MockReferenceManager{}
	blotter := &mockTradeBlotter{}
	mdata := &mockMdataManager{}

	dm := NewDividendsManager(db, rdata, blotter, mdata)

	dividends, err := dm.CalculateDividendsForSingleTicker("AAPL")
	assert.NoError(t, err)
	assert.Len(t, dividends, 2)

	expectedDividends := []Dividends{
		{ExDate: "2023-01-01", Amount: 90.0, AmountPerShare: 1.0},
		{ExDate: "2023-02-01", Amount: 360.0, AmountPerShare: 2.0},
	}

	assert.Equal(t, expectedDividends, dividends)
}
