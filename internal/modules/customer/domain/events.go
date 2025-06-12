package domain

import (
	"golang_modular_monolith/internal/shared/domain"
)

// Customer domain event types
const (
	CustomerCreatedEventType       = "customer.created"
	CustomerNameUpdatedEventType   = "customer.name_updated"
	CustomerEmailChangedEventType  = "customer.email_changed"
	CustomerStatusChangedEventType = "customer.status_changed"
	CustomerDeletedEventType       = "customer.deleted"
)

// CustomerCreatedEvent represents the event when a customer is created
type CustomerCreatedEvent struct {
	domain.BaseDomainEvent
	CustomerID string `json:"customer_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}

// NewCustomerCreatedEvent creates a new customer created event
func NewCustomerCreatedEvent(customer *Customer) CustomerCreatedEvent {
	eventData := map[string]interface{}{
		"customer_id": customer.GetID(),
		"name":        customer.Name,
		"email":       customer.Email.Value,
		"status":      customer.Status,
	}

	return CustomerCreatedEvent{
		BaseDomainEvent: domain.NewBaseDomainEvent(
			customer.GetID(),
			"customer",
			CustomerCreatedEventType,
			eventData,
		),
		CustomerID: customer.GetID(),
		Name:       customer.Name,
		Email:      customer.Email.Value,
		Status:     string(customer.Status),
	}
}

// CustomerNameUpdatedEvent represents the event when customer's name is updated
type CustomerNameUpdatedEvent struct {
	domain.BaseDomainEvent
	CustomerID string `json:"customer_id"`
	OldName    string `json:"old_name"`
	NewName    string `json:"new_name"`
}

// NewCustomerNameUpdatedEvent creates a new customer name updated event
func NewCustomerNameUpdatedEvent(customer *Customer, oldName string) CustomerNameUpdatedEvent {
	eventData := map[string]interface{}{
		"customer_id": customer.GetID(),
		"old_name":    oldName,
		"new_name":    customer.Name,
	}

	return CustomerNameUpdatedEvent{
		BaseDomainEvent: domain.NewBaseDomainEvent(
			customer.GetID(),
			"customer",
			CustomerNameUpdatedEventType,
			eventData,
		),
		CustomerID: customer.GetID(),
		OldName:    oldName,
		NewName:    customer.Name,
	}
}

// CustomerEmailChangedEvent represents the event when customer's email is changed
type CustomerEmailChangedEvent struct {
	domain.BaseDomainEvent
	CustomerID string `json:"customer_id"`
	OldEmail   string `json:"old_email"`
	NewEmail   string `json:"new_email"`
}

// NewCustomerEmailChangedEvent creates a new customer email changed event
func NewCustomerEmailChangedEvent(customer *Customer, oldEmail, newEmail Email) CustomerEmailChangedEvent {
	eventData := map[string]interface{}{
		"customer_id": customer.GetID(),
		"old_email":   oldEmail.Value,
		"new_email":   newEmail.Value,
	}

	return CustomerEmailChangedEvent{
		BaseDomainEvent: domain.NewBaseDomainEvent(
			customer.GetID(),
			"customer",
			CustomerEmailChangedEventType,
			eventData,
		),
		CustomerID: customer.GetID(),
		OldEmail:   oldEmail.Value,
		NewEmail:   newEmail.Value,
	}
}

// CustomerStatusChangedEvent represents the event when customer's status is changed
type CustomerStatusChangedEvent struct {
	domain.BaseDomainEvent
	CustomerID string `json:"customer_id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
}

// NewCustomerStatusChangedEvent creates a new customer status changed event
func NewCustomerStatusChangedEvent(customer *Customer, oldStatus, newStatus CustomerStatus) CustomerStatusChangedEvent {
	eventData := map[string]interface{}{
		"customer_id": customer.GetID(),
		"old_status":  oldStatus,
		"new_status":  newStatus,
	}

	return CustomerStatusChangedEvent{
		BaseDomainEvent: domain.NewBaseDomainEvent(
			customer.GetID(),
			"customer",
			CustomerStatusChangedEventType,
			eventData,
		),
		CustomerID: customer.GetID(),
		OldStatus:  string(oldStatus),
		NewStatus:  string(newStatus),
	}
}

// CustomerDeletedEvent represents the event when customer is deleted
type CustomerDeletedEvent struct {
	domain.BaseDomainEvent
	CustomerID string `json:"customer_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}

// NewCustomerDeletedEvent creates a new customer deleted event
func NewCustomerDeletedEvent(customer *Customer) CustomerDeletedEvent {
	eventData := map[string]interface{}{
		"customer_id": customer.GetID(),
		"name":        customer.Name,
		"email":       customer.Email.Value,
	}

	return CustomerDeletedEvent{
		BaseDomainEvent: domain.NewBaseDomainEvent(
			customer.GetID(),
			"customer",
			CustomerDeletedEventType,
			eventData,
		),
		CustomerID: customer.GetID(),
		Name:       customer.Name,
		Email:      customer.Email.Value,
	}
}
