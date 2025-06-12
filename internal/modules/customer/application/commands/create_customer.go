package commands

import (
	"golang_modular_monolith/internal/shared/application"
)

// CreateCustomerCommand represents a command to create a new customer
type CreateCustomerCommand struct {
	application.BaseCommand
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// NewCreateCustomerCommand creates a new create customer command
func NewCreateCustomerCommand(name, email string) CreateCustomerCommand {
	return CreateCustomerCommand{
		BaseCommand: application.NewBaseCommand("create_customer"),
		Name:        name,
		Email:       email,
	}
}

// CreateCustomerResult represents the result of creating a customer
type CreateCustomerResult struct {
	CustomerID string `json:"customer_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}
