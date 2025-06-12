package domain

import (
	"time"

	"github.com/google/uuid"
)

// AggregateRoot represents the base aggregate root
type AggregateRoot interface {
	// GetID returns the aggregate ID
	GetID() string

	// GetVersion returns the current version
	GetVersion() int

	// GetUncommittedEvents returns events that haven't been persisted
	GetUncommittedEvents() []DomainEvent

	// ClearUncommittedEvents clears the uncommitted events
	ClearUncommittedEvents()

	// IncrementVersion increments the aggregate version
	IncrementVersion()
}

// BaseAggregateRoot provides common aggregate root functionality
type BaseAggregateRoot struct {
	ID                string        `json:"id"`
	Version           int           `json:"version"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	uncommittedEvents []DomainEvent `json:"-"`
}

// NewBaseAggregateRoot creates a new base aggregate root
func NewBaseAggregateRoot() BaseAggregateRoot {
	now := time.Now()
	return BaseAggregateRoot{
		ID:                uuid.New().String(),
		Version:           0,
		CreatedAt:         now,
		UpdatedAt:         now,
		uncommittedEvents: make([]DomainEvent, 0),
	}
}

// NewBaseAggregateRootWithID creates a new base aggregate root with specific ID
func NewBaseAggregateRootWithID(id string) BaseAggregateRoot {
	now := time.Now()
	return BaseAggregateRoot{
		ID:                id,
		Version:           0,
		CreatedAt:         now,
		UpdatedAt:         now,
		uncommittedEvents: make([]DomainEvent, 0),
	}
}

// GetID returns the aggregate ID
func (a *BaseAggregateRoot) GetID() string {
	return a.ID
}

// GetVersion returns the current version
func (a *BaseAggregateRoot) GetVersion() int {
	return a.Version
}

// GetUncommittedEvents returns events that haven't been persisted
func (a *BaseAggregateRoot) GetUncommittedEvents() []DomainEvent {
	return a.uncommittedEvents
}

// ClearUncommittedEvents clears the uncommitted events
func (a *BaseAggregateRoot) ClearUncommittedEvents() {
	a.uncommittedEvents = make([]DomainEvent, 0)
}

// IncrementVersion increments the aggregate version
func (a *BaseAggregateRoot) IncrementVersion() {
	a.Version++
	a.UpdatedAt = time.Now()
}

// AddEvent adds a domain event to the uncommitted events
func (a *BaseAggregateRoot) AddEvent(event DomainEvent) {
	a.uncommittedEvents = append(a.uncommittedEvents, event)
}

// ApplyEvent applies an event to the aggregate (for event sourcing)
func (a *BaseAggregateRoot) ApplyEvent(event DomainEvent) {
	a.IncrementVersion()
}

// HasUncommittedEvents checks if there are uncommitted events
func (a *BaseAggregateRoot) HasUncommittedEvents() bool {
	return len(a.uncommittedEvents) > 0
}

// GetCreatedAt returns when the aggregate was created
func (a *BaseAggregateRoot) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// GetUpdatedAt returns when the aggregate was last updated
func (a *BaseAggregateRoot) GetUpdatedAt() time.Time {
	return a.UpdatedAt
}

// MarkAsDeleted marks the aggregate as deleted (for soft delete)
type SoftDeletable interface {
	MarkAsDeleted()
	IsDeleted() bool
	GetDeletedAt() *time.Time
}

// SoftDeleteableAggregate extends BaseAggregateRoot with soft delete capability
type SoftDeleteableAggregate struct {
	BaseAggregateRoot
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewSoftDeleteableAggregate creates a new soft deleteable aggregate
func NewSoftDeleteableAggregate() SoftDeleteableAggregate {
	return SoftDeleteableAggregate{
		BaseAggregateRoot: NewBaseAggregateRoot(),
		DeletedAt:         nil,
	}
}

// MarkAsDeleted marks the aggregate as deleted
func (a *SoftDeleteableAggregate) MarkAsDeleted() {
	now := time.Now()
	a.DeletedAt = &now
	a.IncrementVersion()
}

// IsDeleted checks if the aggregate is deleted
func (a *SoftDeleteableAggregate) IsDeleted() bool {
	return a.DeletedAt != nil
}

// GetDeletedAt returns when the aggregate was deleted
func (a *SoftDeleteableAggregate) GetDeletedAt() *time.Time {
	return a.DeletedAt
}
