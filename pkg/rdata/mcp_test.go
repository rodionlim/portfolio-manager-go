package rdata

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

func TestReferenceDataMCPToolRegistration(t *testing.T) {
	server := mcpserver.NewMCPServer("test", "1.0.0")
	RegisterMCPTools(server, &mcpTestReferenceManager{})

	tools := server.ListTools()
	for _, name := range []string{
		toolListReferenceData,
		toolGetReferenceData,
		toolAddReferenceData,
		toolUpdateReferenceData,
		toolDeleteReferenceData,
	} {
		require.Contains(t, tools, name)
	}
	require.Contains(t, tools[toolAddReferenceData].Tool.Description, "confirm='yes'")
	require.Contains(t, tools[toolDeleteReferenceData].Tool.Description, "ALWAYS ask")
}

func TestListReferenceDataMCPHandlerFiltersAndLimits(t *testing.T) {
	manager := &mcpTestReferenceManager{tickers: map[string]TickerReference{
		"AAPL":   {ID: "AAPL", Name: "Apple", AssetClass: AssetClassEquities, AssetSubClass: AssetSubClassStock, Category: CategoryTechnology},
		"MSFT":   {ID: "MSFT", Name: "Microsoft", AssetClass: AssetClassEquities, AssetSubClass: AssetSubClassStock, Category: CategoryTechnology},
		"C31.SI": {ID: "C31.SI", Name: "CapitaLand", AssetClass: AssetClassEquities, AssetSubClass: AssetSubClassReit, Category: CategoryREITs},
	}}
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"category": CategoryTechnology,
		"limit":    float64(1),
	}}}

	result, err := createHandleListReferenceData(manager)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)
	content := result.Content[0].(mcp.TextContent)

	var response struct {
		TotalReferenceData int                                     `json:"total_reference_data"`
		ReferenceData      map[string]TickerReferenceWithSGXMapped `json:"reference_data"`
	}
	require.NoError(t, json.Unmarshal([]byte(content.Text), &response))
	require.Equal(t, 1, response.TotalReferenceData)
	require.Len(t, response.ReferenceData, 1)
	require.Contains(t, response.ReferenceData, "AAPL")
}

func TestReferenceDataWriteMCPHandlersRequireConfirmation(t *testing.T) {
	manager := &mcpTestReferenceManager{tickers: map[string]TickerReference{}}
	request := referenceDataWriteRequest(TickerReference{ID: "aapl", Name: "Apple"})

	addResult, err := createHandleAddReferenceData(manager)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, addResult.IsError)
	require.Empty(t, manager.tickers)

	updateResult, err := createHandleUpdateReferenceData(manager)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, updateResult.IsError)
	require.Empty(t, manager.tickers)

	deleteRequest := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"ids": []string{"AAPL"},
	}}}
	deleteResult, err := createHandleDeleteReferenceData(manager)(context.Background(), deleteRequest)
	require.NoError(t, err)
	require.False(t, deleteResult.IsError)
	require.Empty(t, manager.deletedIDs)
}

func TestReferenceDataWriteMCPHandlersApplyWithConfirmation(t *testing.T) {
	manager := &mcpTestReferenceManager{tickers: map[string]TickerReference{}}
	addRequest := referenceDataWriteRequest(TickerReference{ID: "aapl", Name: "Apple"})
	addRequest.Params.Arguments.(map[string]any)["confirm"] = "yes"

	result, err := createHandleAddReferenceData(manager)(context.Background(), addRequest)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Contains(t, manager.tickers, "AAPL")

	updateRequest := referenceDataWriteRequest(TickerReference{ID: "AAPL", Name: "Apple Inc."})
	updateRequest.Params.Arguments.(map[string]any)["confirm"] = "yes"
	result, err = createHandleUpdateReferenceData(manager)(context.Background(), updateRequest)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Equal(t, "Apple Inc.", manager.tickers["AAPL"].Name)

	deleteRequest := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"ids":     []string{"aapl"},
		"confirm": "yes",
	}}}
	result, err = createHandleDeleteReferenceData(manager)(context.Background(), deleteRequest)
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.NotContains(t, manager.tickers, "AAPL")
	require.Equal(t, []string{"AAPL"}, manager.deletedIDs)
}

func TestReferenceDataWriteMCPHandlerNormalizesCanonicalFields(t *testing.T) {
	manager := &mcpTestReferenceManager{tickers: map[string]TickerReference{}}
	request := referenceDataWriteRequest(TickerReference{
		ID:           "dram",
		Name:         "Global X Data Center REITs & Digital Infrastructure ETF",
		YahooTicker:  "dram",
		GoogleTicker: "NASDAQ:DRAM",
	})
	request.Params.Arguments.(map[string]any)["confirm"] = "yes"

	result, err := createHandleAddReferenceData(manager)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)

	ticker := manager.tickers["DRAM"]
	require.Equal(t, "DRAM", ticker.UnderlyingTicker)
	require.Equal(t, "DRAM", ticker.YahooTicker)
	require.Equal(t, "DRAM:BATS", ticker.GoogleTicker)
	require.Equal(t, "USD", ticker.Ccy)
	require.Equal(t, "US", ticker.Domicile)
}

func TestReferenceDataWriteMCPHandlerInfersCurrencyAndDomicileFromTickerSuffix(t *testing.T) {
	manager := &mcpTestReferenceManager{tickers: map[string]TickerReference{}}
	request := referenceDataWriteRequest(TickerReference{
		ID:          "c31.si",
		Name:        "CapitaLand Integrated Commercial Trust",
		YahooTicker: "c31.si",
	})
	request.Params.Arguments.(map[string]any)["confirm"] = "yes"

	result, err := createHandleAddReferenceData(manager)(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError)

	ticker := manager.tickers["C31.SI"]
	require.Equal(t, "C31.SI", ticker.UnderlyingTicker)
	require.Equal(t, "SGD", ticker.Ccy)
	require.Equal(t, "SG", ticker.Domicile)
}

func TestReferenceDataWriteMCPHandlerRejectsInvalidCurrencyAndDomicile(t *testing.T) {
	manager := &mcpTestReferenceManager{tickers: map[string]TickerReference{}}

	invalidCurrencyRequest := referenceDataWriteRequest(TickerReference{
		ID:       "AAPL",
		Name:     "Apple",
		Ccy:      "US",
		Domicile: "US",
	})
	invalidCurrencyRequest.Params.Arguments.(map[string]any)["confirm"] = "yes"
	result, err := createHandleAddReferenceData(manager)(context.Background(), invalidCurrencyRequest)
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Contains(t, result.Content[0].(mcp.TextContent).Text, "ticker.ccy")

	invalidDomicileRequest := referenceDataWriteRequest(TickerReference{
		ID:       "AAPL",
		Name:     "Apple",
		Ccy:      "USD",
		Domicile: "USA",
	})
	invalidDomicileRequest.Params.Arguments.(map[string]any)["confirm"] = "yes"
	result, err = createHandleAddReferenceData(manager)(context.Background(), invalidDomicileRequest)
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Contains(t, result.Content[0].(mcp.TextContent).Text, "ticker.domicile")
}

func TestGetReferenceDataMCPHandlerReturnsError(t *testing.T) {
	manager := &mcpTestReferenceManager{getErr: errors.New("not found")}
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"id": "UNKNOWN",
	}}}

	result, err := createHandleGetReferenceData(manager)(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError)
	content := result.Content[0].(mcp.TextContent)
	require.Contains(t, content.Text, "not found")
}

func referenceDataWriteRequest(ticker TickerReference) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"ticker": map[string]any{
			"id":                  ticker.ID,
			"name":                ticker.Name,
			"underlying_ticker":   ticker.UnderlyingTicker,
			"yahoo_ticker":        ticker.YahooTicker,
			"google_ticker":       ticker.GoogleTicker,
			"dividends_sg_ticker": ticker.DividendsSgTicker,
			"nasdaq_ticker":       ticker.NasdaqTicker,
			"barchart_ticker":     ticker.BarchartTicker,
			"asset_class":         ticker.AssetClass,
			"asset_sub_class":     ticker.AssetSubClass,
			"category":            ticker.Category,
			"sub_category":        ticker.SubCategory,
			"ccy":                 ticker.Ccy,
			"domicile":            ticker.Domicile,
			"coupon_rate":         ticker.CouponRate,
			"maturity_date":       ticker.MaturityDate,
			"strike_price":        ticker.StrikePrice,
			"call_put":            ticker.CallPut,
		},
	}}}
}

type mcpTestReferenceManager struct {
	tickers    map[string]TickerReference
	deletedIDs []string
	getErr     error
}

func (m *mcpTestReferenceManager) AddTicker(ticker TickerReference) (string, error) {
	m.ensureTickers()
	m.tickers[ticker.ID] = ticker
	return ticker.ID, nil
}

func (m *mcpTestReferenceManager) UpdateTicker(ticker *TickerReference) error {
	m.ensureTickers()
	m.tickers[ticker.ID] = *ticker
	return nil
}

func (m *mcpTestReferenceManager) DeleteTicker(id string) error {
	m.ensureTickers()
	delete(m.tickers, id)
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func (m *mcpTestReferenceManager) GetTicker(id string) (TickerReferenceWithSGXMapped, error) {
	if m.getErr != nil {
		return TickerReferenceWithSGXMapped{}, m.getErr
	}
	m.ensureTickers()
	ticker, ok := m.tickers[id]
	if !ok {
		return TickerReferenceWithSGXMapped{}, errors.New("not found")
	}
	return ticker.ToSGXMapped(), nil
}

func (m *mcpTestReferenceManager) GetAllTickers() (map[string]TickerReferenceWithSGXMapped, error) {
	m.ensureTickers()
	result := make(map[string]TickerReferenceWithSGXMapped, len(m.tickers))
	for id, ticker := range m.tickers {
		result[id] = ticker.ToSGXMapped()
	}
	return result, nil
}

func (m *mcpTestReferenceManager) ExportToYamlBytes() ([]byte, error) {
	return nil, nil
}

func (m *mcpTestReferenceManager) ensureTickers() {
	if m.tickers == nil {
		m.tickers = make(map[string]TickerReference)
	}
}
