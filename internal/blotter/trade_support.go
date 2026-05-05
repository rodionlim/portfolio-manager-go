package blotter

import (
	"fmt"
	"strings"
	"time"

	"portfolio-manager/pkg/csvutil"
	"portfolio-manager/pkg/rdata"
)

func (b *TradeBlotter) BuildTrade(input TradeInput) (*Trade, error) {
	input.Ticker = strings.ToUpper(strings.TrimSpace(input.Ticker))

	normalizedAttributes, err := NormalizeTradeAttributes(input.Ticker, input.Attributes)
	if err != nil {
		return nil, err
	}
	normalizedAttributes.InstrumentType = InferInstrumentType(input.Ticker, normalizedAttributes)
	input.Attributes = normalizedAttributes

	if input.Attributes.InstrumentType == InstrumentTypeOption {
		if b.rdataSvc == nil {
			return nil, fmt.Errorf("reference data service is not configured for option trades")
		}
		if input.Attributes.UnderlyingTicker == "" {
			return nil, fmt.Errorf("underlying ticker is required for option trades")
		}

		if input.Attributes.UnderlyingSpotRef <= 0 {
			spotRef, err := b.lookupUnderlyingSpotReference(input.Attributes.UnderlyingTicker, input.TradeDate)
			if err != nil {
				return nil, err
			}
			input.Attributes.UnderlyingSpotRef = spotRef
		}

		optionTicker, err := BuildOptionTicker(
			input.Attributes.UnderlyingTicker,
			input.Attributes.ExpiryDate,
			input.Attributes.StrikePrice,
			input.Attributes.CallPut,
		)
		if err != nil {
			return nil, err
		}
		input.Ticker = optionTicker

		underlyingRef, err := b.rdataSvc.GetTicker(input.Attributes.UnderlyingTicker)
		if err != nil {
			return nil, fmt.Errorf("failed to get underlying reference data for %s: %w", input.Attributes.UnderlyingTicker, err)
		}

		if err := b.ensureOptionReference(optionTicker, underlyingRef, input.Attributes); err != nil {
			return nil, err
		}
	}

	if input.ID == "" {
		return NewTrade(
			input.Side,
			input.Quantity,
			input.Ticker,
			input.Book,
			input.Broker,
			input.Account,
			input.Status,
			input.OrigTradeID,
			input.Price,
			input.Fx,
			input.Yield,
			input.TradeDate,
			input.Attributes,
		)
	}

	return NewTradeWithID(
		input.ID,
		input.Side,
		input.Quantity,
		input.Ticker,
		input.Book,
		input.Broker,
		input.Account,
		input.Status,
		input.OrigTradeID,
		input.Price,
		input.Fx,
		input.Yield,
		input.SeqNum,
		input.TradeDate,
		input.Attributes,
	)
}

func (b *TradeBlotter) lookupUnderlyingSpotReference(underlyingTicker string, tradeDate time.Time) (float64, error) {
	if b.mdataSvc == nil {
		return 0, fmt.Errorf("market data service is not configured to infer underlying spot reference")
	}

	startOfDay := time.Date(tradeDate.Year(), tradeDate.Month(), tradeDate.Day(), 0, 0, 0, 0, time.UTC).Unix()
	endOfDay := time.Date(tradeDate.Year(), tradeDate.Month(), tradeDate.Day(), 23, 59, 59, 0, time.UTC).Unix()

	historical, _, historicalErr := b.mdataSvc.GetHistoricalData(underlyingTicker, startOfDay, endOfDay)
	if historicalErr == nil && len(historical) > 0 && historical[len(historical)-1] != nil {
		return historical[len(historical)-1].Price, nil
	}

	assetPrice, currentErr := b.mdataSvc.GetAssetPrice(underlyingTicker)
	if currentErr != nil {
		if historicalErr != nil {
			return 0, fmt.Errorf("failed to infer underlying spot reference for %s: historical lookup failed: %v, current lookup failed: %w", underlyingTicker, historicalErr, currentErr)
		}
		return 0, fmt.Errorf("failed to infer underlying spot reference for %s: %w", underlyingTicker, currentErr)
	}

	return assetPrice.Price, nil
}

func (b *TradeBlotter) ensureOptionReference(ticker string, underlyingRef rdata.TickerReferenceWithSGXMapped, attrs TradeAttributes) error {
	ref := rdata.TickerReference{
		ID:               ticker,
		Name:             buildOptionReferenceName(underlyingRef.Name, attrs),
		UnderlyingTicker: attrs.UnderlyingTicker,
		AssetClass:       underlyingRef.AssetClass,
		AssetSubClass:    rdata.AssetSubClassOption,
		Category:         underlyingRef.Category,
		SubCategory:      underlyingRef.SubCategory,
		Ccy:              underlyingRef.Ccy,
		Domicile:         underlyingRef.Domicile,
		MaturityDate:     attrs.ExpiryDate,
		StrikePrice:      attrs.StrikePrice,
		CallPut:          attrs.CallPut,
	}

	if _, err := b.rdataSvc.GetTicker(ticker); err == nil {
		return b.rdataSvc.UpdateTicker(&ref)
	}

	_, err := b.rdataSvc.AddTicker(ref)
	return err
}

func buildOptionReferenceName(underlyingName string, attrs TradeAttributes) string {
	if strings.TrimSpace(underlyingName) == "" {
		underlyingName = attrs.UnderlyingTicker
	}

	callPut := strings.ToUpper(attrs.CallPut)
	if callPut == strings.ToUpper(CallPutCall) {
		callPut = "CALL"
	} else if callPut == strings.ToUpper(CallPutPut) {
		callPut = "PUT"
	}

	return fmt.Sprintf("%s %s %s %s", underlyingName, attrs.ExpiryDate, callPut, csvutil.FormatFloat(attrs.StrikePrice, 4))
}
