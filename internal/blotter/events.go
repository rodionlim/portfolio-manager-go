package blotter

import "portfolio-manager/pkg/event"

// Define event names
const (
	NewTradeEvent    = "NewTrade"
	RemoveTradeEvent = "RemoveTrade"
)

// NewTradeEventPayload represents the payload for a new trade event.
type NewTradeEventPayload struct {
	Trade Trade
}

// PublishNewTradeEvent publishes a new trade event.
func (b *TradeBlotter) PublishNewTradeEvent(trade Trade) {
	event := event.Event{
		Name: NewTradeEvent,
		Data: NewTradeEventPayload{Trade: trade},
	}
	b.eventBus.Publish(event)
}

// PublishNewTradeEvent publishes a new trade event.
func (b *TradeBlotter) PublishRemoveTradeEvent(trade Trade) {
	event := event.Event{
		Name: RemoveTradeEvent,
		Data: NewTradeEventPayload{Trade: trade},
	}
	b.eventBus.Publish(event)
}
