package blotter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	toolQueryBlotterTrades = "query_blotter_trades"
	toolInsertBlotterTrade = "insert_blotter_trade"
)

// RegisterMCPTools registers all blotter related MCP tools
func RegisterMCPTools(mcpServer *server.MCPServer, blotter *TradeBlotter) {
	// Query blotter trades tool
	queryBlotterTool := mcp.NewTool(toolQueryBlotterTrades,
		mcp.WithDescription("Query blotter trades based on criteria like ticker, date range, or trade type"),
		mcp.WithString("ticker",
			mcp.Description("Filter by ticker symbol (optional)"),
		),
		mcp.WithString("start_date",
			mcp.Description("Start date for filtering trades in YYYY-MM-DD format (optional)"),
		),
		mcp.WithString("end_date",
			mcp.Description("End date for filtering trades in YYYY-MM-DD format (optional)"),
		),
		mcp.WithString("trade_type",
			mcp.Description("Filter by trade type: buy or sell (optional)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Limit the number of results returned (default: 100)"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	mcpServer.AddTool(queryBlotterTool, createHandleQueryBlotterTrades(blotter))

	insertBlotterTool := mcp.NewTool(toolInsertBlotterTrade,
		mcp.WithDescription("Insert a blotter trade from natural-language details. This writes to the database. ALWAYS ask the user to explicitly confirm before calling this tool with confirm='yes'. If confirm is omitted, the tool returns a preview populated with sensible defaults from the latest same-ticker trade, or UI defaults when no similar trade exists. For non-SGD assets, missing fx is inferred through market data using the trade date, then current FX as fallback."),
		mcp.WithString("confirm", mcp.Description("Must be set to 'yes' to perform the insertion. If omitted, the tool returns a confirmation preview.")),
		mcp.WithString("ticker", mcp.Description("Ticker symbol or option underlying ticker. For option trades, ticker can be omitted if underlying_ticker is supplied.")),
		mcp.WithString("trade_date", mcp.Description("Trade date in YYYY-MM-DD, YYYYMMDD, or RFC3339 format. Defaults to today.")),
		mcp.WithString("side", mcp.Required(), mcp.Description("Trade side: buy or sell")),
		mcp.WithNumber("quantity", mcp.Required(), mcp.Description("Trade quantity")),
		mcp.WithNumber("price", mcp.Description("Trade price. Required unless value is supplied.")),
		mcp.WithNumber("value", mcp.Description("Gross trade value. Used to derive price when price is omitted.")),
		mcp.WithNumber("fx", mcp.Description("FX rate to SGD. If omitted, the tool infers it from reference data and market data.")),
		mcp.WithNumber("yield", mcp.Description("Trade yield, if relevant")),
		mcp.WithString("book", mcp.Description("Book. Defaults from the latest same-ticker trade, otherwise Main.")),
		mcp.WithString("broker", mcp.Description("Broker. Defaults from the latest same-ticker trade, otherwise DBS.")),
		mcp.WithString("account", mcp.Description("Account. Defaults from the latest same-ticker trade, otherwise CDP.")),
		mcp.WithString("status", mcp.Description("Status: open, autoclosed, or closed. Defaults from the latest same-ticker trade, otherwise open.")),
		mcp.WithString("orig_trade_id", mcp.Description("Original trade ID for closure or linked trades.")),
		mcp.WithString("instrument_type", mcp.Description("Instrument type: outright, option, or future. Defaults from supplied option fields or latest same-ticker trade.")),
		mcp.WithString("underlying_ticker", mcp.Description("Underlying ticker for options. Also used for FX lookup for option trades.")),
		mcp.WithNumber("underlying_spot_ref", mcp.Description("Underlying spot reference for options. If omitted for options, market data is used to infer it.")),
		mcp.WithString("expiry_date", mcp.Description("Option expiry date in YYYY-MM-DD, YYYYMMDD, or RFC3339 format.")),
		mcp.WithNumber("strike_price", mcp.Description("Option strike price")),
		mcp.WithString("call_put", mcp.Description("Option type: call or put")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	)

	mcpServer.AddTool(insertBlotterTool, createHandleInsertBlotterTrade(blotter))
}

// createHandleQueryBlotterTrades creates a handler for querying blotter trades
func createHandleQueryBlotterTrades(blotter *TradeBlotter) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ticker, _ := request.RequireString("ticker")
		ticker = strings.ToUpper(strings.TrimSpace(ticker))
		startDateStr, _ := request.RequireString("start_date")
		endDateStr, _ := request.RequireString("end_date")
		tradeTypeStr, _ := request.RequireString("trade_type")
		tradeTypeStr = strings.ToLower(strings.TrimSpace(tradeTypeStr))
		limit := request.GetInt("limit", 100)
		if limit <= 0 {
			limit = 100
		}

		// Get all trades from blotter
		trades := blotter.GetTrades()

		// Apply filters
		var filteredTrades []Trade
		for _, trade := range trades {
			// Filter by ticker
			if ticker != "" && trade.Ticker != ticker {
				continue
			}

			// Filter by trade type (Side field in blotter.Trade)
			if tradeTypeStr != "" && trade.Side != tradeTypeStr {
				continue
			}

			// Filter by date range
			if startDateStr != "" {
				startDate, err := time.Parse("2006-01-02", startDateStr)
				if err == nil {
					tradeDate, parseErr := parseMCPTradeDate(trade.TradeDate)
					if parseErr == nil && tradeDate.Before(startDate) {
						continue
					}
				}
			}

			if endDateStr != "" {
				endDate, err := time.Parse("2006-01-02", endDateStr)
				if err == nil {
					tradeDate, parseErr := parseMCPTradeDate(trade.TradeDate)
					if parseErr == nil && tradeDate.After(endDate) {
						continue
					}
				}
			}

			filteredTrades = append(filteredTrades, trade)

			// Apply limit
			if len(filteredTrades) >= limit {
				break
			}
		}

		// Prepare response
		response := map[string]interface{}{
			"total_trades": len(filteredTrades),
			"trades":       filteredTrades,
			"filters": map[string]interface{}{
				"ticker":     ticker,
				"start_date": startDateStr,
				"end_date":   endDateStr,
				"trade_type": tradeTypeStr,
				"limit":      limit,
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// createHandleInsertBlotterTrade creates a handler for inserting blotter trades with a confirmation gate.
func createHandleInsertBlotterTrade(blotter *TradeBlotter) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		draft, err := buildMCPTradeDraft(blotter, request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		trade, err := blotter.BuildTrade(draft.input)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to build trade: %v", err)), nil
		}

		response := map[string]any{
			"trade":              trade,
			"source_defaults":    draft.sourceDefaults,
			"similar_trade":      draft.similarTrade,
			"fx_source":          draft.fxSource,
			"requires_confirm":   request.GetString("confirm", "") != "yes",
			"confirmation_hint":  "This writes a new blotter trade. Ask the user to confirm the preview, then call again with confirm='yes' to insert it.",
			"confirmation_value": "yes",
		}

		if request.GetString("confirm", "") != "yes" {
			return jsonToolResult(response)
		}

		if err := blotter.AddTrade(*trade); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to insert trade: %v", err)), nil
		}

		response["inserted"] = true
		response["requires_confirm"] = false
		response["confirmation_hint"] = "Trade inserted."
		return jsonToolResult(response)
	}
}

type mcpTradeDraft struct {
	input          TradeInput
	sourceDefaults string
	similarTrade   *Trade
	fxSource       string
}

func buildMCPTradeDraft(blotter *TradeBlotter, request mcp.CallToolRequest) (*mcpTradeDraft, error) {
	ticker := strings.ToUpper(strings.TrimSpace(request.GetString("ticker", "")))
	underlyingTicker := strings.ToUpper(strings.TrimSpace(request.GetString("underlying_ticker", "")))
	if ticker == "" {
		ticker = underlyingTicker
	}
	if ticker == "" {
		return nil, fmt.Errorf("ticker is required")
	}

	side := strings.ToLower(strings.TrimSpace(request.GetString("side", "")))
	if side == "" {
		return nil, fmt.Errorf("side is required")
	}

	quantity := request.GetFloat("quantity", 0)
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}

	price := request.GetFloat("price", 0)
	value := request.GetFloat("value", 0)
	if price <= 0 && value <= 0 {
		return nil, fmt.Errorf("either price or value must be greater than 0")
	}
	if price > 0 && value > 0 {
		return nil, fmt.Errorf("specify either price or value, not both")
	}
	if price <= 0 {
		price = value / quantity
	}

	tradeDate, err := parseMCPTradeDate(request.GetString("trade_date", ""))
	if err != nil {
		return nil, err
	}

	defaultTrade := findLatestSimilarTrade(blotter.GetTrades(), ticker, underlyingTicker)
	input := TradeInput{
		TradeDate: tradeDate,
		Ticker:    ticker,
		Side:      side,
		Quantity:  quantity,
		Price:     price,
		Fx:        request.GetFloat("fx", 0),
		Yield:     request.GetFloat("yield", 0),
		Book:      "Main",
		Broker:    "DBS",
		Account:   "CDP",
		Status:    StatusOpen,
		Attributes: TradeAttributes{
			InstrumentType:    request.GetString("instrument_type", ""),
			UnderlyingTicker:  underlyingTicker,
			UnderlyingSpotRef: request.GetFloat("underlying_spot_ref", 0),
			ExpiryDate:        request.GetString("expiry_date", ""),
			StrikePrice:       request.GetFloat("strike_price", 0),
			CallPut:           request.GetString("call_put", ""),
		},
	}

	sourceDefaults := "ui_defaults"
	if defaultTrade != nil {
		sourceDefaults = "latest_same_asset_trade"
		input.Book = defaultTrade.Book
		input.Broker = defaultTrade.Broker
		input.Account = defaultTrade.Account
		input.Status = defaultTrade.Status
		input.Fx = defaultTrade.Fx
		input.Attributes.InstrumentType = defaultTrade.InstrumentType
		input.Attributes.UnderlyingTicker = defaultTrade.UnderlyingTicker
		input.Attributes.UnderlyingSpotRef = defaultTrade.UnderlyingSpotRef
		input.Attributes.ExpiryDate = defaultTrade.ExpiryDate
		input.Attributes.StrikePrice = defaultTrade.StrikePrice
		input.Attributes.CallPut = defaultTrade.CallPut
	}

	if value := strings.TrimSpace(request.GetString("book", "")); value != "" {
		input.Book = value
	}
	if value := strings.TrimSpace(request.GetString("broker", "")); value != "" {
		input.Broker = value
	}
	if value := strings.TrimSpace(request.GetString("account", "")); value != "" {
		input.Account = value
	}
	if value := strings.ToLower(strings.TrimSpace(request.GetString("status", ""))); value != "" {
		input.Status = value
	}
	if value := strings.TrimSpace(request.GetString("orig_trade_id", "")); value != "" {
		input.OrigTradeID = value
	}
	if value := strings.TrimSpace(request.GetString("instrument_type", "")); value != "" {
		input.Attributes.InstrumentType = value
	}
	if underlyingTicker != "" {
		input.Attributes.UnderlyingTicker = underlyingTicker
	}
	if value := request.GetFloat("underlying_spot_ref", 0); value > 0 {
		input.Attributes.UnderlyingSpotRef = value
	}
	if value := strings.TrimSpace(request.GetString("expiry_date", "")); value != "" {
		input.Attributes.ExpiryDate = value
	}
	if value := request.GetFloat("strike_price", 0); value > 0 {
		input.Attributes.StrikePrice = value
	}
	if value := strings.TrimSpace(request.GetString("call_put", "")); value != "" {
		input.Attributes.CallPut = value
	}

	fxSource := "provided"
	if request.GetFloat("fx", 0) <= 0 {
		fx, source, err := inferMCPTradeFX(blotter, input)
		if err != nil {
			return nil, err
		}
		input.Fx = fx
		fxSource = source
	}

	var similarTrade *Trade
	if defaultTrade != nil {
		copy := defaultTrade.Clone()
		similarTrade = &copy
	}

	return &mcpTradeDraft{
		input:          input,
		sourceDefaults: sourceDefaults,
		similarTrade:   similarTrade,
		fxSource:       fxSource,
	}, nil
}

func parseMCPTradeDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	if parsed, err := time.Parse("20060102", trimmed); err == nil {
		return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, fmt.Errorf("invalid trade_date %q: use YYYY-MM-DD, YYYYMMDD, or RFC3339", value)
}

func findLatestSimilarTrade(trades []Trade, ticker, underlyingTicker string) *Trade {
	var latest *Trade
	for i := range trades {
		trade := trades[i]
		if !sameMCPTradeAsset(trade, ticker, underlyingTicker) {
			continue
		}
		if latest == nil || trade.TradeDate > latest.TradeDate || (trade.TradeDate == latest.TradeDate && trade.SeqNum > latest.SeqNum) {
			copy := trade.Clone()
			latest = &copy
		}
	}
	return latest
}

func sameMCPTradeAsset(trade Trade, ticker, underlyingTicker string) bool {
	tradeTicker := strings.ToUpper(strings.TrimSpace(trade.Ticker))
	tradeUnderlying := strings.ToUpper(strings.TrimSpace(trade.UnderlyingTicker))
	return tradeTicker == ticker || (underlyingTicker != "" && (tradeTicker == underlyingTicker || tradeUnderlying == underlyingTicker))
}

func inferMCPTradeFX(blotter *TradeBlotter, input TradeInput) (float64, string, error) {
	if blotter.rdataSvc == nil {
		if input.Fx > 0 {
			return input.Fx, "latest_same_asset_trade", nil
		}
		return 0, "", fmt.Errorf("fx was not specified and reference data service is not configured")
	}

	fxLookupTicker := input.Ticker
	if input.Attributes.UnderlyingTicker != "" && InferInstrumentType(input.Ticker, input.Attributes) == InstrumentTypeOption {
		fxLookupTicker = input.Attributes.UnderlyingTicker
	}

	ref, err := blotter.rdataSvc.GetTicker(fxLookupTicker)
	if err != nil {
		if input.Fx > 0 {
			return input.Fx, "latest_same_asset_trade", nil
		}
		return 0, "", fmt.Errorf("fx was not specified and failed to resolve reference currency for %s: %w", fxLookupTicker, err)
	}

	quoteCcy := strings.ToUpper(strings.TrimSpace(ref.Ccy))
	if quoteCcy == "" || quoteCcy == "SGD" {
		return 1, "reference_data_sgd", nil
	}
	if blotter.mdataSvc == nil {
		if input.Fx > 0 {
			return input.Fx, "latest_same_asset_trade", nil
		}
		return 0, "", fmt.Errorf("fx was not specified and market data service is not configured for %s-SGD", quoteCcy)
	}

	fxTicker := quoteCcy + "-SGD"
	startOfDay := time.Date(input.TradeDate.Year(), input.TradeDate.Month(), input.TradeDate.Day(), 0, 0, 0, 0, time.UTC).Unix()
	endOfDay := time.Date(input.TradeDate.Year(), input.TradeDate.Month(), input.TradeDate.Day(), 23, 59, 59, 0, time.UTC).Unix()
	historical, _, historicalErr := blotter.mdataSvc.GetHistoricalData(fxTicker, startOfDay, endOfDay)
	if historicalErr == nil && len(historical) > 0 {
		for i := len(historical) - 1; i >= 0; i-- {
			if historical[i] != nil && historical[i].Price > 0 {
				return historical[i].Price, "historical_mdata_" + fxTicker, nil
			}
		}
	}

	current, currentErr := blotter.mdataSvc.GetAssetPrice(fxTicker)
	if currentErr == nil && current != nil && current.Price > 0 {
		return current.Price, "current_mdata_" + fxTicker, nil
	}

	if historicalErr != nil {
		return 0, "", fmt.Errorf("failed to infer FX for %s: historical lookup failed: %v, current lookup failed: %w", fxTicker, historicalErr, currentErr)
	}
	return 0, "", fmt.Errorf("failed to infer FX for %s: no valid historical price, current lookup failed: %w", fxTicker, currentErr)
}

func jsonToolResult(response any) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}
