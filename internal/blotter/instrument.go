package blotter

import (
	"fmt"
	"math"
	"strings"
	"time"

	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/csvutil"
)

const (
	InstrumentTypeOutright = "outright"
	InstrumentTypeOption   = "option"
	InstrumentTypeFuture   = "future"

	CallPutCall = "call"
	CallPutPut  = "put"
)

type TradeAttributes struct {
	InstrumentType    string  `json:"InstrumentType"`
	UnderlyingTicker  string  `json:"UnderlyingTicker"`
	UnderlyingSpotRef float64 `json:"UnderlyingSpotRef"`
	ExpiryDate        string  `json:"ExpiryDate"`
	StrikePrice       float64 `json:"StrikePrice"`
	CallPut           string  `json:"CallPut"`
}

type TradeInput struct {
	ID          string
	TradeDate   time.Time
	Ticker      string
	Side        string
	Quantity    float64
	Price       float64
	Fx          float64
	Yield       float64
	Book        string
	Broker      string
	Account     string
	Status      string
	OrigTradeID string
	SeqNum      int
	Attributes  TradeAttributes
}

func NormalizeInstrumentType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "out", "outright", "spot", "stock", "cash":
		return InstrumentTypeOutright
	case "opt", "option", "options":
		return InstrumentTypeOption
	case "fut", "future", "futures":
		return InstrumentTypeFuture
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func NormalizeCallPut(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "c", CallPutCall:
		return CallPutCall
	case "p", CallPutPut:
		return CallPutPut
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func CanonicalizeExpiryDate(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}

	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return parsed.Format("2006-01-02"), nil
	}

	parsed, err := common.ParseFlexibleDate(trimmed)
	if err != nil {
		return "", err
	}

	return parsed.Format("2006-01-02"), nil
}

func NormalizeTradeAttributes(ticker string, attrs TradeAttributes) (TradeAttributes, error) {
	attrs.InstrumentType = NormalizeInstrumentType(attrs.InstrumentType)
	attrs.UnderlyingTicker = strings.ToUpper(strings.TrimSpace(attrs.UnderlyingTicker))
	attrs.CallPut = NormalizeCallPut(attrs.CallPut)

	if attrs.InstrumentType == "" {
		attrs.InstrumentType = InstrumentTypeOutright
	}

	if attrs.InstrumentType == InstrumentTypeOutright && attrs.UnderlyingTicker == "" {
		attrs.UnderlyingTicker = strings.ToUpper(strings.TrimSpace(ticker))
	}

	if attrs.ExpiryDate != "" {
		canonicalExpiry, err := CanonicalizeExpiryDate(attrs.ExpiryDate)
		if err != nil {
			return TradeAttributes{}, fmt.Errorf("invalid expiry date: %w", err)
		}
		attrs.ExpiryDate = canonicalExpiry
	}

	return attrs, nil
}

func InferInstrumentType(ticker string, attrs TradeAttributes) string {
	if normalized := NormalizeInstrumentType(attrs.InstrumentType); normalized != InstrumentTypeOutright {
		return normalized
	}

	if attrs.ExpiryDate != "" || attrs.StrikePrice > 0 || NormalizeCallPut(attrs.CallPut) != "" {
		return InstrumentTypeOption
	}

	if common.IsFutures(ticker) {
		return InstrumentTypeFuture
	}

	return InstrumentTypeOutright
}

func BuildOptionTicker(underlyingTicker, expiryDate string, strikePrice float64, callPut string) (string, error) {
	underlyingTicker = strings.ToUpper(strings.TrimSpace(underlyingTicker))
	if underlyingTicker == "" {
		return "", fmt.Errorf("underlying ticker is required")
	}

	canonicalExpiry, err := CanonicalizeExpiryDate(expiryDate)
	if err != nil {
		return "", fmt.Errorf("invalid expiry date: %w", err)
	}
	if canonicalExpiry == "" {
		return "", fmt.Errorf("expiry date is required")
	}
	if strikePrice <= 0 {
		return "", fmt.Errorf("strike price must be greater than 0")
	}
	if math.Abs(strikePrice-math.Round(strikePrice)) > 0.0000001 {
		return "", fmt.Errorf("strike price must be a whole number")
	}

	callPut = NormalizeCallPut(callPut)
	if callPut != CallPutCall && callPut != CallPutPut {
		return "", fmt.Errorf("call put must be either '%s' or '%s'", CallPutCall, CallPutPut)
	}

	cpCode := "c"
	if callPut == CallPutPut {
		cpCode = "p"
	}

	strikeComponent := strings.ReplaceAll(csvutil.FormatFloat(strikePrice, 4), ".", "p")
	expiryComponent := strings.ReplaceAll(canonicalExpiry, "-", "")

	return strings.ToUpper(fmt.Sprintf("%s_%s_%s_%s", underlyingTicker, expiryComponent, strikeComponent, cpCode)), nil
}
