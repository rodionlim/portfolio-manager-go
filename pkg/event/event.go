package event

import (
	"sync"

	"github.com/google/uuid"
)

// Event represents an event with a name and data.
type Event struct {
	Name string
	Data interface{}
}

// EventHandler is a struct that consists of a function that handles an event.
type EventHandler struct {
	id       uuid.UUID
	callback func(Event)
}

// NewEventHandler creates a new EventHandler instance.
func NewEventHandler(callback func(Event)) EventHandler {
	return EventHandler{
		callback: callback,
	}
}

// EventBus manages event subscriptions and publishing.
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe adds a new event handler for a specific event name.
func (bus *EventBus) Subscribe(eventName string, handler EventHandler) uuid.UUID {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	handler.id = uuid.New()
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)
	return handler.id
}

// Unsubscribe removes an event handler for a specific event name.
func (bus *EventBus) Unsubscribe(eventName string, handlerID uuid.UUID) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	handlers := bus.handlers[eventName]
	for i, h := range handlers {
		if h.id == handlerID {
			bus.handlers[eventName] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish sends an event to all subscribed handlers.
func (bus *EventBus) Publish(event Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	if handlers, found := bus.handlers[event.Name]; found {
		for _, handler := range handlers {
			go handler.callback(event)
		}
	}
}
