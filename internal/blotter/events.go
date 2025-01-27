package blotter

import "portfolio-manager/pkg/event"

// Define event names
const (
	NewTradeEvent    = "NewTrade"
	RemoveTradeEvent = "RemoveTrade"
	UpdateTradeEvent = "UpdateTrade"
)

// NewTradeEventPayload represents the payload for a new trade event.
type TradeEventPayload struct {
	Trade         Trade
	OriginalTrade Trade // only used for trade updates
}

// PublishNewTradeEvent publishes a new trade event.
func (b *TradeBlotter) PublishNewTradeEvent(trade Trade) {
	event := event.Event{
		Name: NewTradeEvent,
		Data: TradeEventPayload{Trade: trade},
	}
	b.eventBus.Publish(event)
}

// PublishRemoveTradeEvent publishes a remove trade event.
func (b *TradeBlotter) PublishRemoveTradeEvent(trade Trade) {
	event := event.Event{
		Name: RemoveTradeEvent,
		Data: TradeEventPayload{Trade: trade},
	}
	b.eventBus.Publish(event)
}

// PublishUpdateTradeEvent publishes a update trade event.
func (b *TradeBlotter) PublishUpdateTradeEvent(trade Trade, originalTrade Trade) {
	event := event.Event{
		Name: UpdateTradeEvent,
		Data: TradeEventPayload{Trade: trade, OriginalTrade: originalTrade},
	}
	b.eventBus.Publish(event)
}
