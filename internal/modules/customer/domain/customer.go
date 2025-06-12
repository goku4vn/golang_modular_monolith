package domain

import (
	"regexp"
	"strings"

	"golang_modular_monolith/internal/shared/domain"
)

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusDeleted  CustomerStatus = "deleted"
)

// Customer represents the customer aggregate root
type Customer struct {
	domain.BaseAggregateRoot
	Name   string         `json:"name"`
	Email  Email          `json:"email"`
	Status CustomerStatus `json:"status"`
}

// Email represents customer email value object
type Email struct {
	Value string `json:"value"`
}

// NewEmail creates a new email value object
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return Email{}, domain.NewValidationError("email", "email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return Email{}, domain.NewValidationError("email", "invalid email format")
	}

	return Email{Value: email}, nil
}

// String returns the email as string
func (e Email) String() string {
	return e.Value
}

// IsEmpty checks if email is empty
func (e Email) IsEmpty() bool {
	return e.Value == ""
}

// NewCustomer creates a new customer
func NewCustomer(name, email string) (*Customer, error) {
	// Validate input
	var validationErrors domain.ValidationErrors

	name = strings.TrimSpace(name)
	if name == "" {
		validationErrors.Add("name", "name is required")
	}

	customerEmail, err := NewEmail(email)
	if err != nil {
		if validationErr, ok := err.(domain.ValidationError); ok {
			validationErrors = append(validationErrors, validationErr)
		} else {
			return nil, err
		}
	}

	if validationErrors.HasErrors() {
		return nil, validationErrors
	}

	// Create customer
	customer := &Customer{
		BaseAggregateRoot: domain.NewBaseAggregateRoot(),
		Name:              name,
		Email:             customerEmail,
		Status:            CustomerStatusActive,
	}

	// Add domain event
	customer.AddEvent(NewCustomerCreatedEvent(customer))

	return customer, nil
}

// UpdateName updates customer's name
func (c *Customer) UpdateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.NewValidationError("name", "name is required")
	}

	// Check if anything changed
	if c.Name == name {
		return nil
	}

	oldName := c.Name
	c.Name = name
	c.IncrementVersion()

	// Add domain event
	c.AddEvent(NewCustomerNameUpdatedEvent(c, oldName))

	return nil
}

// ChangeEmail changes customer's email
func (c *Customer) ChangeEmail(newEmail string) error {
	email, err := NewEmail(newEmail)
	if err != nil {
		return err
	}

	// Check if email is the same
	if c.Email.Value == email.Value {
		return nil
	}

	oldEmail := c.Email
	c.Email = email
	c.IncrementVersion()

	// Add domain event
	c.AddEvent(NewCustomerEmailChangedEvent(c, oldEmail, email))

	return nil
}

// Activate activates the customer
func (c *Customer) Activate() error {
	if c.Status == CustomerStatusActive {
		return nil
	}

	if c.Status == CustomerStatusDeleted {
		return domain.NewBusinessRuleError("customer_deleted", "cannot activate deleted customer")
	}

	oldStatus := c.Status
	c.Status = CustomerStatusActive
	c.IncrementVersion()

	// Add domain event
	c.AddEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusActive))

	return nil
}

// Deactivate deactivates the customer
func (c *Customer) Deactivate() error {
	if c.Status == CustomerStatusInactive {
		return nil
	}

	if c.Status == CustomerStatusDeleted {
		return domain.NewBusinessRuleError("customer_deleted", "cannot deactivate deleted customer")
	}

	oldStatus := c.Status
	c.Status = CustomerStatusInactive
	c.IncrementVersion()

	// Add domain event
	c.AddEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusInactive))

	return nil
}

// Delete marks the customer as deleted
func (c *Customer) Delete() error {
	if c.Status == CustomerStatusDeleted {
		return nil
	}

	c.Status = CustomerStatusDeleted
	c.IncrementVersion()

	// Add domain event
	c.AddEvent(NewCustomerDeletedEvent(c))

	return nil
}

// IsDeleted checks if customer is deleted
func (c *Customer) IsDeleted() bool {
	return c.Status == CustomerStatusDeleted
}

// IsActive checks if customer is active
func (c *Customer) IsActive() bool {
	return c.Status == CustomerStatusActive
}

// ValidateForCreation validates customer data for creation
func (c *Customer) ValidateForCreation() error {
	var validationErrors domain.ValidationErrors

	if c.Email.IsEmpty() {
		validationErrors.Add("email", "email is required")
	}

	if strings.TrimSpace(c.Name) == "" {
		validationErrors.Add("name", "name is required")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
