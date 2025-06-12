package queryhandlers

import (
	"context"
	"fmt"

	"golang_modular_monolith/internal/modules/customer/application/queries"
	"golang_modular_monolith/internal/modules/customer/domain"
	shareddomain "golang_modular_monolith/internal/shared/domain"
)

// GetCustomerHandler handles GetCustomerQuery
type GetCustomerHandler struct {
	queryRepo domain.CustomerQueryRepository
}

// NewGetCustomerHandler creates a new GetCustomerHandler
func NewGetCustomerHandler(queryRepo domain.CustomerQueryRepository) *GetCustomerHandler {
	return &GetCustomerHandler{
		queryRepo: queryRepo,
	}
}

// Handle handles the GetCustomerQuery
func (h *GetCustomerHandler) Handle(ctx context.Context, query *queries.GetCustomerQuery) (*queries.GetCustomerResult, error) {
	// Validate query
	if query.ID == "" {
		return nil, shareddomain.NewDomainError(
			shareddomain.ErrCodeInvalidInput,
			"customer ID is required",
		)
	}

	// Get customer from repository
	customer, err := h.queryRepo.GetByID(ctx, query.ID)
	if err != nil {
		if shareddomain.IsNotFoundError(err) {
			return nil, shareddomain.NewDomainError(
				shareddomain.ErrCodeNotFound,
				fmt.Sprintf("customer with ID %s not found", query.ID),
			)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &queries.GetCustomerResult{
		Customer: *customer,
	}, nil
}
