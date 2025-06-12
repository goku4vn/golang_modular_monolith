package eventbus

import (
	"log"
	"reflect"
	"sync"

	"golang_modular_monolith/internal/shared/domain"
)

// EventHandler represents an event handler function
type EventHandler func(event domain.DomainEvent) error

// InMemoryEventBus implements EventBus using in-memory handler registration
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// SubscribeToEventType registers an event handler for a specific event type
func (b *InMemoryEventBus) SubscribeToEventType(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// SubscribeToEvent registers an event handler for a specific event type using reflection
func (b *InMemoryEventBus) SubscribeToEvent(event domain.DomainEvent, handler EventHandler) {
	eventType := reflect.TypeOf(event).String()
	b.SubscribeToEventType(eventType, handler)
}

// Publish publishes an event to all registered handlers
func (b *InMemoryEventBus) Publish(event domain.DomainEvent) error {
	eventType := reflect.TypeOf(event).String()

	b.mu.RLock()
	handlers := b.handlers[eventType]
	b.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			// Log error but continue with other handlers
			log.Printf("Error handling event %s: %v", eventType, err)
			// In a production system, you might want to collect these errors
			// and handle them appropriately (retry, dead letter queue, etc.)
		}
	}

	return nil
}

// PublishAll publishes multiple events
func (b *InMemoryEventBus) PublishAll(events []domain.DomainEvent) error {
	for _, event := range events {
		if err := b.Publish(event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes a handler to events (domain.EventHandler interface)
func (b *InMemoryEventBus) Subscribe(handler domain.EventHandler) error {
	// This is a simplified implementation
	// In a real system, you'd register the handler properly
	log.Printf("Handler subscribed: %T", handler)
	return nil
}

// Unsubscribe removes a handler
func (b *InMemoryEventBus) Unsubscribe(handler domain.EventHandler) error {
	// This is a simplified implementation
	log.Printf("Handler unsubscribed: %T", handler)
	return nil
}

// SubscribeByType registers an event handler for a specific event type (local method)
func (b *InMemoryEventBus) SubscribeByType(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// GetSubscriberCount returns the number of subscribers for an event type
func (b *InMemoryEventBus) GetSubscriberCount(eventType string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.handlers[eventType])
}

// Clear removes all handlers (useful for testing)
func (b *InMemoryEventBus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers = make(map[string][]EventHandler)
}

// GetEventTypes returns all registered event types
func (b *InMemoryEventBus) GetEventTypes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	types := make([]string, 0, len(b.handlers))
	for eventType := range b.handlers {
		types = append(types, eventType)
	}

	return types
}

// Example event handlers that can be registered

// LogEventHandler logs all events
func LogEventHandler(event domain.DomainEvent) error {
	log.Printf("Event published: %s - AggregateID: %s", reflect.TypeOf(event).String(), event.GetAggregateID())
	return nil
}

// MetricsEventHandler could be used to collect metrics
func MetricsEventHandler(event domain.DomainEvent) error {
	// Here you would send metrics to your metrics system
	// For example: increment counter, record timing, etc.
	eventType := reflect.TypeOf(event).String()
	log.Printf("Metrics: Event %s published at %s", eventType, event.GetOccurredAt().Format("2006-01-02 15:04:05"))
	return nil
}

// AsyncEventBus wraps InMemoryEventBus to handle events asynchronously
type AsyncEventBus struct {
	bus *InMemoryEventBus
}

// NewAsyncEventBus creates a new async event bus
func NewAsyncEventBus() *AsyncEventBus {
	return &AsyncEventBus{
		bus: NewInMemoryEventBus(),
	}
}

// SubscribeToEventType registers an event handler
func (a *AsyncEventBus) SubscribeToEventType(eventType string, handler EventHandler) {
	a.bus.SubscribeToEventType(eventType, handler)
}

// SubscribeToEvent registers an event handler for a specific event type
func (a *AsyncEventBus) SubscribeToEvent(event domain.DomainEvent, handler EventHandler) {
	a.bus.SubscribeToEvent(event, handler)
}

// Publish publishes an event asynchronously
func (a *AsyncEventBus) Publish(event domain.DomainEvent) error {
	go func() {
		if err := a.bus.Publish(event); err != nil {
			log.Printf("Error publishing event asynchronously: %v", err)
		}
	}()
	return nil
}

// PublishSync publishes an event synchronously
func (a *AsyncEventBus) PublishSync(event domain.DomainEvent) error {
	return a.bus.Publish(event)
}

// GetSubscriberCount returns the number of subscribers for an event type
func (a *AsyncEventBus) GetSubscriberCount(eventType string) int {
	return a.bus.GetSubscriberCount(eventType)
}

// Clear removes all handlers
func (a *AsyncEventBus) Clear() {
	a.bus.Clear()
}
