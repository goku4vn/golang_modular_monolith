package queryhandlers

import (
	"context"
	"fmt"

	"golang_modular_monolith/internal/modules/customer/application/queries"
	"golang_modular_monolith/internal/modules/customer/domain"
)

// ListCustomersHandler handles ListCustomersQuery
type ListCustomersHandler struct {
	queryRepo domain.CustomerQueryRepository
}

// NewListCustomersHandler creates a new ListCustomersHandler
func NewListCustomersHandler(queryRepo domain.CustomerQueryRepository) *ListCustomersHandler {
	return &ListCustomersHandler{
		queryRepo: queryRepo,
	}
}

// Handle handles the ListCustomersQuery
func (h *ListCustomersHandler) Handle(ctx context.Context, query *queries.ListCustomersQuery) (*queries.ListCustomersResult, error) {
	// Convert query to domain params
	params := domain.ListCustomersParams{
		Page:           query.Page,
		Limit:          query.Limit,
		Status:         query.Status,
		IncludeDeleted: query.IncludeDeleted,
		SortBy:         query.SortBy,
		SortOrder:      query.SortOrder,
		CreatedAfter:   query.CreatedAfter,
		CreatedBefore:  query.CreatedBefore,
		UpdatedAfter:   query.UpdatedAfter,
		UpdatedBefore:  query.UpdatedBefore,
	}

	// Get customers from repository
	result, err := h.queryRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	return &queries.ListCustomersResult{
		CustomerListResult: *result,
	}, nil
}
