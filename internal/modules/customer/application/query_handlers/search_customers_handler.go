package queryhandlers

import (
	"context"
	"fmt"

	"golang_modular_monolith/internal/modules/customer/application/queries"
	"golang_modular_monolith/internal/modules/customer/domain"
)

// SearchCustomersHandler handles SearchCustomersQuery
type SearchCustomersHandler struct {
	queryRepo domain.CustomerQueryRepository
}

// NewSearchCustomersHandler creates a new SearchCustomersHandler
func NewSearchCustomersHandler(queryRepo domain.CustomerQueryRepository) *SearchCustomersHandler {
	return &SearchCustomersHandler{
		queryRepo: queryRepo,
	}
}

// Handle handles the SearchCustomersQuery
func (h *SearchCustomersHandler) Handle(ctx context.Context, query *queries.SearchCustomersQuery) (*queries.SearchCustomersResult, error) {
	// Convert query to domain params
	params := domain.SearchCustomersParams{
		ListCustomersParams: domain.ListCustomersParams{
			Page:      query.Page,
			Limit:     query.Limit,
			Status:    query.Status,
			SortBy:    query.SortBy,
			SortOrder: query.SortOrder,
		},
		Query:     query.Query,
		Email:     query.Email,
		FirstName: query.FirstName,
		LastName:  query.LastName,
	}

	// Search customers from repository
	result, err := h.queryRepo.Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	return &queries.SearchCustomersResult{
		CustomerListResult: *result,
	}, nil
}
