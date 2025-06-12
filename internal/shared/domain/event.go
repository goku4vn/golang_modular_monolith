package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event that occurred
type DomainEvent interface {
	// GetEventID returns unique event identifier
	GetEventID() string

	// GetAggregateID returns the ID of the aggregate that produced this event
	GetAggregateID() string

	// GetAggregateType returns the type of aggregate
	GetAggregateType() string

	// GetEventType returns the type of event
	GetEventType() string

	// GetEventVersion returns the version of the event schema
	GetEventVersion() int

	// GetOccurredAt returns when the event occurred
	GetOccurredAt() time.Time

	// GetEventData returns the event payload
	GetEventData() interface{}
}

// BaseDomainEvent provides common implementation for domain events
type BaseDomainEvent struct {
	EventID       string      `json:"event_id"`
	AggregateID   string      `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	EventType     string      `json:"event_type"`
	EventVersion  int         `json:"event_version"`
	OccurredAt    time.Time   `json:"occurred_at"`
	EventData     interface{} `json:"event_data"`
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(aggregateID, aggregateType, eventType string, eventData interface{}) BaseDomainEvent {
	return BaseDomainEvent{
		EventID:       uuid.New().String(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		EventVersion:  1,
		OccurredAt:    time.Now(),
		EventData:     eventData,
	}
}

// GetEventID returns unique event identifier
func (e BaseDomainEvent) GetEventID() string {
	return e.EventID
}

// GetAggregateID returns the ID of the aggregate that produced this event
func (e BaseDomainEvent) GetAggregateID() string {
	return e.AggregateID
}

// GetAggregateType returns the type of aggregate
func (e BaseDomainEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetEventType returns the type of event
func (e BaseDomainEvent) GetEventType() string {
	return e.EventType
}

// GetEventVersion returns the version of the event schema
func (e BaseDomainEvent) GetEventVersion() int {
	return e.EventVersion
}

// GetOccurredAt returns when the event occurred
func (e BaseDomainEvent) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// GetEventData returns the event payload
func (e BaseDomainEvent) GetEventData() interface{} {
	return e.EventData
}

// EventHandler defines how to handle domain events
type EventHandler interface {
	Handle(event DomainEvent) error
	CanHandle(eventType string) bool
}

// EventBus defines the interface for publishing and subscribing to domain events
type EventBus interface {
	// Publish publishes a single event
	Publish(event DomainEvent) error

	// PublishAll publishes multiple events
	PublishAll(events []DomainEvent) error

	// Subscribe subscribes a handler to events
	Subscribe(handler EventHandler) error

	// Unsubscribe removes a handler
	Unsubscribe(handler EventHandler) error
}
