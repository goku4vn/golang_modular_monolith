package commandhandlers

import (
	"context"
	"fmt"

	"golang_modular_monolith/internal/modules/customer/application/commands"
	"golang_modular_monolith/internal/modules/customer/domain"
	shareddomain "golang_modular_monolith/internal/shared/domain"
)

// CreateCustomerHandler handles CreateCustomerCommand
type CreateCustomerHandler struct {
	repo      domain.CustomerRepository
	domainSvc domain.CustomerDomainService
	eventBus  shareddomain.EventBus
}

// NewCreateCustomerHandler creates a new CreateCustomerHandler
func NewCreateCustomerHandler(
	repo domain.CustomerRepository,
	domainSvc domain.CustomerDomainService,
	eventBus shareddomain.EventBus,
) *CreateCustomerHandler {
	return &CreateCustomerHandler{
		repo:      repo,
		domainSvc: domainSvc,
		eventBus:  eventBus,
	}
}

// Handle handles the CreateCustomerCommand
func (h *CreateCustomerHandler) Handle(ctx context.Context, cmd *commands.CreateCustomerCommand) (*commands.CreateCustomerResult, error) {
	// Validate command
	if cmd.Name == "" {
		return nil, shareddomain.NewDomainError(
			shareddomain.ErrCodeInvalidInput,
			"name is required",
		)
	}
	if cmd.Email == "" {
		return nil, shareddomain.NewDomainError(
			shareddomain.ErrCodeInvalidInput,
			"email is required",
		)
	}

	// Check if email is unique
	isUnique, err := h.domainSvc.IsEmailUnique(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}

	if !isUnique {
		return nil, shareddomain.NewDomainError(
			shareddomain.ErrCodeAlreadyExists,
			"customer with this email already exists",
		)
	}

	// Create customer
	customer, err := domain.NewCustomer(cmd.Name, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Save to repository
	if err := h.repo.Save(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to save customer: %w", err)
	}

	// Publish domain events
	if err := h.publishEvents(ctx, customer); err != nil {
		// Log error but don't fail the operation
		// In a real application, you might want to use outbox pattern or similar
		fmt.Printf("Warning: failed to publish events for customer %s: %v\n", customer.GetID(), err)
	}

	return &commands.CreateCustomerResult{
		CustomerID: customer.GetID(),
		Name:       customer.Name,
		Email:      customer.Email.Value,
		Status:     string(customer.Status),
	}, nil
}

// publishEvents publishes domain events
func (h *CreateCustomerHandler) publishEvents(ctx context.Context, customer *domain.Customer) error {
	events := customer.GetUncommittedEvents()
	for _, event := range events {
		if err := h.eventBus.Publish(event); err != nil {
			return fmt.Errorf("failed to publish event %T: %w", event, err)
		}
	}
	return nil
}
