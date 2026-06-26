package blotter

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

func TestBlotterMCPToolRegistration(t *testing.T) {
	server := mcpserver.NewMCPServer("test", "1.0.0")
	RegisterMCPTools(server, NewBlotter(newMCPTestDB(t)))

	tools := server.ListTools()
	require.Contains(t, tools, toolQueryBlotterTrades)
	require.Contains(t, tools, toolInsertBlotterTrade)
	require.Contains(t, tools[toolInsertBlotterTrade].Tool.Description, "confirm='yes'")
	require.Contains(t, tools[toolInsertBlotterTrade].Tool.Description, "ALWAYS ask")
}

func TestQueryBlotterTradesMCPFiltersRFC3339DatesAndNumericLimit(t *testing.T) {
	blotter := NewBlotter(newMCPTestDB(t))
	first, err := NewTrade("buy", 10, "AAPL", "Growth", "IBKR", "Margin", StatusOpen, "", 100, 1, 0, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	second, err := NewTrade("sell", 3, "AAPL", "Growth", "IBKR", "Margin", StatusOpen, "", 150, 1, 0, time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NoError(t, blotter.AddTrade(*first))
	require.NoError(t, blotter.AddTrade(*second))

	result, err := createHandleQueryBlotterTrades(blotter)(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"ticker":     "aapl",
		"start_date": "2026-06-01",
		"limit":      1.0,
	}}})
	require.NoError(t, err)
	require.False(t, result.IsError)

	content := result.Content[0].(mcp.TextContent)
	var response struct {
		TotalTrades int     `json:"total_trades"`
		Trades      []Trade `json:"trades"`
	}
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, 1, response.TotalTrades)
	require.Equal(t, "sell", response.Trades[0].Side)
}

func TestInsertBlotterTradeMCPRequiresConfirmationAndUsesDefaults(t *testing.T) {
	blotter := NewBlotter(newMCPTestDB(t))
	previous, err := NewTrade("buy", 10, "AAPL", "Growth", "IBKR", "Margin", StatusOpen, "", 100, 1.31, 0, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NoError(t, blotter.AddTrade(*previous))
	blotter.SetTradeSupportServices(
		&mcpReferenceManager{refs: map[string]rdata.TickerReference{
			"AAPL": {ID: "AAPL", Name: "Apple", Ccy: "USD", Domicile: "US", AssetClass: rdata.AssetClassEquities, AssetSubClass: rdata.AssetSubClassStock},
		}},
		&mcpMarketDataManager{historical: map[string][]*types.AssetData{
			"USD-SGD": {{Ticker: "USD-SGD", Price: 1.36}},
		}},
	)

	request := mcpInsertTradeRequest(map[string]any{
		"ticker":     "aapl",
		"trade_date": "2026-06-25",
		"side":       "BUY",
		"quantity":   5.0,
		"price":      200.0,
	})

	result, err := createHandleInsertBlotterTrade(blotter)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, blotter.GetTrades(), 1)

	response := decodeMCPTextResult(t, result)
	require.True(t, response.RequiresConfirm)
	require.Equal(t, "latest_same_asset_trade", response.SourceDefaults)
	require.Equal(t, "historical_mdata_USD-SGD", response.FXSource)
	require.Equal(t, "Growth", response.Trade.Book)
	require.Equal(t, "IBKR", response.Trade.Broker)
	require.Equal(t, "Margin", response.Trade.Account)
	require.Equal(t, 1.36, response.Trade.Fx)

	request.Params.Arguments.(map[string]any)["confirm"] = "yes"
	result, err = createHandleInsertBlotterTrade(blotter)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, blotter.GetTrades(), 2)
	response = decodeMCPTextResult(t, result)
	require.False(t, response.RequiresConfirm)
	require.True(t, response.Inserted)
}

func TestInsertBlotterTradeMCPUsesCurrentFXFallback(t *testing.T) {
	blotter := NewBlotter(newMCPTestDB(t))
	blotter.SetTradeSupportServices(
		&mcpReferenceManager{refs: map[string]rdata.TickerReference{
			"MSFT": {ID: "MSFT", Name: "Microsoft", Ccy: "USD", Domicile: "US", AssetClass: rdata.AssetClassEquities, AssetSubClass: rdata.AssetSubClassStock},
		}},
		&mcpMarketDataManager{
			historicalErr: map[string]error{"USD-SGD": assertErr("missing historical")},
			current:       map[string]*types.AssetData{"USD-SGD": {Ticker: "USD-SGD", Price: 1.37}},
		},
	)

	result, err := createHandleInsertBlotterTrade(blotter)(context.Background(), mcpInsertTradeRequest(map[string]any{
		"ticker":     "MSFT",
		"trade_date": "2026-06-25",
		"side":       "sell",
		"quantity":   2.0,
		"value":      600.0,
		"confirm":    "yes",
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)

	response := decodeMCPTextResult(t, result)
	require.Equal(t, "current_mdata_USD-SGD", response.FXSource)
	require.Equal(t, 300.0, response.Trade.Price)
	require.Equal(t, 1.37, response.Trade.Fx)
	require.Len(t, blotter.GetTrades(), 1)
}

type mcpInsertTradeResponse struct {
	Trade           Trade  `json:"trade"`
	SourceDefaults  string `json:"source_defaults"`
	FXSource        string `json:"fx_source"`
	RequiresConfirm bool   `json:"requires_confirm"`
	Inserted        bool   `json:"inserted"`
	SimilarTrade    *Trade `json:"similar_trade"`
}

func mcpInsertTradeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func decodeMCPTextResult(t *testing.T, result *mcp.CallToolResult) mcpInsertTradeResponse {
	t.Helper()
	content := result.Content[0].(mcp.TextContent)
	var response mcpInsertTradeResponse
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	return response
}

func newMCPTestDB(t *testing.T) dal.Database {
	t.Helper()
	db, err := dal.NewLevelDB(filepath.Join(t.TempDir(), "leveldb"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	return db
}

type mcpReferenceManager struct {
	refs map[string]rdata.TickerReference
}

func (m *mcpReferenceManager) AddTicker(ticker rdata.TickerReference) (string, error) {
	if m.refs == nil {
		m.refs = map[string]rdata.TickerReference{}
	}
	m.refs[ticker.ID] = ticker
	return ticker.ID, nil
}

func (m *mcpReferenceManager) UpdateTicker(ticker *rdata.TickerReference) error {
	if m.refs == nil {
		m.refs = map[string]rdata.TickerReference{}
	}
	m.refs[ticker.ID] = *ticker
	return nil
}

func (m *mcpReferenceManager) DeleteTicker(id string) error {
	delete(m.refs, id)
	return nil
}

func (m *mcpReferenceManager) GetTicker(id string) (rdata.TickerReferenceWithSGXMapped, error) {
	ref, ok := m.refs[id]
	if !ok {
		return rdata.TickerReferenceWithSGXMapped{}, assertErr("reference not found")
	}
	return ref.ToSGXMapped(), nil
}

func (m *mcpReferenceManager) GetAllTickers() (map[string]rdata.TickerReferenceWithSGXMapped, error) {
	result := make(map[string]rdata.TickerReferenceWithSGXMapped, len(m.refs))
	for id, ref := range m.refs {
		result[id] = ref.ToSGXMapped()
	}
	return result, nil
}

func (m *mcpReferenceManager) ExportToYamlBytes() ([]byte, error) {
	return nil, nil
}

type mcpMarketDataManager struct {
	historical    map[string][]*types.AssetData
	historicalErr map[string]error
	current       map[string]*types.AssetData
	currentErr    map[string]error
}

func (m *mcpMarketDataManager) GetAssetPrice(ticker string) (*types.AssetData, error) {
	if err := m.currentErr[ticker]; err != nil {
		return nil, err
	}
	if data := m.current[ticker]; data != nil {
		return data, nil
	}
	return nil, assertErr("current price not found")
}

func (m *mcpMarketDataManager) GetHistoricalData(ticker string, fromDate, toDate int64) ([]*types.AssetData, bool, error) {
	if err := m.historicalErr[ticker]; err != nil {
		return nil, false, err
	}
	return m.historical[ticker], false, nil
}

func (m *mcpMarketDataManager) GetDividendsMetadata(ticker string) ([]types.DividendsMetadata, error) {
	return nil, nil
}

func (m *mcpMarketDataManager) GetDividendsMetadataFromTickerRef(tickerRef rdata.TickerReference) ([]types.DividendsMetadata, error) {
	return nil, nil
}

func (m *mcpMarketDataManager) ImportCustomDividendsFromCSVReader(reader *csv.Reader) (int, error) {
	return 0, nil
}

func (m *mcpMarketDataManager) StoreCustomDividendsMetadata(ticker string, dividends []types.DividendsMetadata) error {
	return nil
}

func (m *mcpMarketDataManager) DeleteDividendsMetadata(ticker string, isCustom bool) error {
	return nil
}

func (m *mcpMarketDataManager) FetchBenchmarkInterestRates(country string, points int) ([]types.InterestRates, error) {
	return nil, nil
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
