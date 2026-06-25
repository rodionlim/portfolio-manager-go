package rdata

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	toolListReferenceData   = "list_reference_data"
	toolGetReferenceData    = "get_reference_data"
	toolAddReferenceData    = "add_reference_data"
	toolUpdateReferenceData = "update_reference_data"
	toolDeleteReferenceData = "delete_reference_data"
)

// RegisterMCPTools registers reference data MCP tools.
func RegisterMCPTools(mcpServer *server.MCPServer, refSvc ReferenceManager) {
	listTool := mcp.NewTool(toolListReferenceData,
		mcp.WithDescription("List ticker reference data. Use filters or ids to keep the response compact when possible."),
		mcp.WithString("asset_class", mcp.Description("Optional asset class filter, for example eq, bond, cash, cmdty, crypto, or fx.")),
		mcp.WithString("asset_sub_class", mcp.Description("Optional asset sub-class filter, for example stock, etf, reit, option, future, govies, cash, or spot.")),
		mcp.WithString("category", mcp.Description("Optional category filter, for example technology, finance, reits, healthcare, or energy.")),
		mcp.WithArray("ids",
			mcp.Description("Optional ticker reference ids to return. Prefer this over listing all data when the user names tickers."),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithInteger("limit", mcp.Description("Optional maximum number of records to return after filtering.")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
	mcpServer.AddTool(listTool, createHandleListReferenceData(refSvc))

	getTool := mcp.NewTool(toolGetReferenceData,
		mcp.WithDescription("Get one ticker reference data record by id."),
		mcp.WithString("id", mcp.Description("Ticker reference id, for example AAPL or C31.SI."), mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
	mcpServer.AddTool(getTool, createHandleGetReferenceData(refSvc))

	addTool := mcp.NewTool(toolAddReferenceData,
		mcp.WithDescription("Add ticker reference data. This writes to the database. ALWAYS ask the user to explicitly confirm before calling this tool with confirm='yes'."),
		mcp.WithObject("ticker", tickerReferenceSchemaOptions()...),
		mcp.WithString("confirm", mcp.Description("Must be set to 'yes' to perform the insertion. If omitted, the tool returns a confirmation prompt.")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	)
	mcpServer.AddTool(addTool, createHandleAddReferenceData(refSvc))

	updateTool := mcp.NewTool(toolUpdateReferenceData,
		mcp.WithDescription("Update ticker reference data by id. This writes to the database. ALWAYS ask the user to explicitly confirm before calling this tool with confirm='yes'."),
		mcp.WithObject("ticker", tickerReferenceSchemaOptions()...),
		mcp.WithString("confirm", mcp.Description("Must be set to 'yes' to perform the update. If omitted, the tool returns a confirmation prompt.")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	)
	mcpServer.AddTool(updateTool, createHandleUpdateReferenceData(refSvc))

	deleteTool := mcp.NewTool(toolDeleteReferenceData,
		mcp.WithDescription("Delete ticker reference data by ids. This is destructive. ALWAYS ask the user to explicitly confirm before calling this tool with confirm='yes'."),
		mcp.WithArray("ids",
			mcp.Required(),
			mcp.Description("Ticker reference ids to delete."),
			mcp.Items(map[string]any{"type": "string"}),
			mcp.MinItems(1),
			mcp.UniqueItems(true),
		),
		mcp.WithString("confirm", mcp.Description("Must be set to 'yes' to perform the deletion. If omitted, the tool returns a confirmation prompt.")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	)
	mcpServer.AddTool(deleteTool, createHandleDeleteReferenceData(refSvc))
}

func createHandleListReferenceData(refSvc ReferenceManager) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		allTickers, err := refSvc.GetAllTickers()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list reference data: %v", err)), nil
		}

		ids := normalizeIDs(request.GetStringSlice("ids", nil))
		idFilter := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			idFilter[id] = struct{}{}
		}

		assetClass := request.GetString("asset_class", "")
		assetSubClass := request.GetString("asset_sub_class", "")
		category := request.GetString("category", "")
		limit := request.GetInt("limit", 0)

		keys := make([]string, 0, len(allTickers))
		for id := range allTickers {
			keys = append(keys, id)
		}
		sort.Strings(keys)

		tickers := make(map[string]TickerReferenceWithSGXMapped)
		for _, id := range keys {
			ticker := allTickers[id]
			if len(idFilter) > 0 {
				if _, ok := idFilter[strings.ToUpper(id)]; !ok {
					continue
				}
			}
			if assetClass != "" && ticker.AssetClass != assetClass {
				continue
			}
			if assetSubClass != "" && ticker.AssetSubClass != assetSubClass {
				continue
			}
			if category != "" && ticker.Category != category {
				continue
			}
			tickers[id] = ticker
			if limit > 0 && len(tickers) >= limit {
				break
			}
		}

		return referenceDataToolResult(map[string]any{
			"total_reference_data": len(tickers),
			"reference_data":       tickers,
			"filter": map[string]any{
				"ids":             ids,
				"asset_class":     assetClass,
				"asset_sub_class": assetSubClass,
				"category":        category,
				"limit":           limit,
			},
		})
	}
}

func createHandleGetReferenceData(refSvc ReferenceManager) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil || strings.TrimSpace(id) == "" {
			return mcp.NewToolResultError("missing required parameter: id"), nil
		}
		ticker, err := refSvc.GetTicker(strings.ToUpper(strings.TrimSpace(id)))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get reference data for id %s: %v", id, err)), nil
		}
		return referenceDataToolResult(ticker)
	}
}

func createHandleAddReferenceData(refSvc ReferenceManager) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, err := bindReferenceDataWriteArgs(request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if args.Confirm != "yes" {
			return mcp.NewToolResultText(fmt.Sprintf("You are requesting to add reference data (id=%s). This writes to the database. If you wish to proceed, call the tool again with confirm='yes'.", args.Ticker.ID)), nil
		}
		id, err := refSvc.AddTicker(args.Ticker)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to add reference data: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully added reference data (id=%s)", id)), nil
	}
}

func createHandleUpdateReferenceData(refSvc ReferenceManager) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, err := bindReferenceDataWriteArgs(request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if args.Confirm != "yes" {
			return mcp.NewToolResultText(fmt.Sprintf("You are requesting to update reference data (id=%s). This writes to the database. If you wish to proceed, call the tool again with confirm='yes'.", args.Ticker.ID)), nil
		}
		if err := refSvc.UpdateTicker(&args.Ticker); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update reference data: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully updated reference data (id=%s)", args.Ticker.ID)), nil
	}
}

func createHandleDeleteReferenceData(refSvc ReferenceManager) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ids := normalizeIDs(request.GetStringSlice("ids", nil))
		if len(ids) == 0 {
			return mcp.NewToolResultError("missing required parameter: ids"), nil
		}
		confirm := request.GetString("confirm", "")
		if confirm != "yes" {
			return mcp.NewToolResultText(fmt.Sprintf("You are requesting to delete reference data ids %v. This is DESTRUCTIVE and cannot be undone. If you wish to proceed, call the tool again with confirm='yes'.", ids)), nil
		}
		for _, id := range ids {
			if err := refSvc.DeleteTicker(id); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to delete reference data for id %s: %v", id, err)), nil
			}
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted reference data ids %v", ids)), nil
	}
}

type referenceDataWriteArgs struct {
	Ticker  TickerReference `json:"ticker"`
	Confirm string          `json:"confirm"`
}

func bindReferenceDataWriteArgs(request mcp.CallToolRequest) (referenceDataWriteArgs, error) {
	var args referenceDataWriteArgs
	if err := request.BindArguments(&args); err != nil {
		return args, fmt.Errorf("invalid arguments: %w", err)
	}
	args.Ticker.ID = strings.ToUpper(strings.TrimSpace(args.Ticker.ID))
	if args.Ticker.ID == "" {
		return args, fmt.Errorf("missing required ticker.id")
	}
	return args, nil
}

func referenceDataToolResult(data any) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal reference data response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

func normalizeIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		normalized := strings.ToUpper(strings.TrimSpace(id))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func tickerReferenceSchemaOptions() []mcp.PropertyOption {
	return []mcp.PropertyOption{
		mcp.Required(),
		mcp.Description("Ticker reference data payload."),
		mcp.Properties(map[string]any{
			"id":                  map[string]any{"type": "string", "description": "Ticker reference id, for example AAPL or C31.SI."},
			"name":                map[string]any{"type": "string"},
			"underlying_ticker":   map[string]any{"type": "string"},
			"yahoo_ticker":        map[string]any{"type": "string"},
			"google_ticker":       map[string]any{"type": "string"},
			"dividends_sg_ticker": map[string]any{"type": "string"},
			"nasdaq_ticker":       map[string]any{"type": "string"},
			"barchart_ticker":     map[string]any{"type": "string"},
			"asset_class":         map[string]any{"type": "string", "enum": []string{AssetClassBonds, AssetClassCash, AssetClassCommodities, AssetClassCrypto, AssetClassEquities, AssetClassFX}},
			"asset_sub_class":     map[string]any{"type": "string", "enum": []string{AssetSubClassBond, AssetSubClassCash, AssetSubClassETF, AssetSubClassFuture, AssetSubClassGovies, AssetSubClassOption, AssetSubClassReit, AssetSubClassSpot, AssetSubClassStock}},
			"category":            map[string]any{"type": "string", "enum": []string{CategoryAgriculture, CategoryConsumerCyclicals, CategoryConsumerNonCyclicals, CategoryCrypto, CategoryEnergy, CategoryFinance, CategoryFuneral, CategoryHealthcare, CategoryIndustrials, CategoryMaterials, CategoryRealEstate, CategoryREITs, CategoryTelecommunications, CategoryTechnology, CategoryUtilities}},
			"sub_category":        map[string]any{"type": "string"},
			"ccy":                 map[string]any{"type": "string"},
			"domicile":            map[string]any{"type": "string"},
			"coupon_rate":         map[string]any{"type": "number"},
			"maturity_date":       map[string]any{"type": "string", "description": "YYYY-MM-DD when applicable."},
			"strike_price":        map[string]any{"type": "number"},
			"call_put":            map[string]any{"type": "string", "enum": []string{"call", "put", ""}},
		}),
		mcp.AdditionalProperties(false),
	}
}
