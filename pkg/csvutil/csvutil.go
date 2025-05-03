package csvutil

import (
	"fmt"
	"math"
	"strconv"
)

// TradeHeaders contains the standard CSV headers for trade exports
var TradeHeaders = []string{"TradeDate", "Ticker", "Side", "Quantity", "Price", "Yield", "Trader", "Broker", "Account", "Status", "Fx"}

// The numeric indices of each column in the trade CSV
const (
	TradeDateIdx = iota
	TickerIdx
	SideIdx
	QuantityIdx
	PriceIdx
	YieldIdx
	TraderIdx
	BrokerIdx
	AccountIdx
	StatusIdx
	FxIdx
)

// FormatFloat formats a float to display:
// - 0 decimal places if value is a whole number (e.g., 1.0 -> "1")
// - Up to maxDecimals decimal places if there are significant decimals
func FormatFloat(value float64, maxDecimals int) string {
	// Check if value is effectively a whole number
	if math.Abs(value-math.Round(value)) < 0.0000001 {
		return strconv.FormatInt(int64(value), 10)
	}

	// Format with up to maxDecimals decimal places
	formatted := fmt.Sprintf("%.*f", maxDecimals, value)

	// Trim trailing zeros after decimal point
	for formatted[len(formatted)-1] == '0' {
		formatted = formatted[:len(formatted)-1]
	}

	// Remove decimal point if it's the last character
	if formatted[len(formatted)-1] == '.' {
		formatted = formatted[:len(formatted)-1]
	}

	return formatted
}
