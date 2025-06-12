package queries

import "golang_modular_monolith/internal/modules/customer/domain"

// GetCustomerQuery represents a query to get a customer by ID
type GetCustomerQuery struct {
	ID string `json:"id"`
}

// GetCustomerResult represents the result of GetCustomerQuery
type GetCustomerResult struct {
	Customer domain.CustomerView `json:"customer"`
}

// ListCustomersQuery represents a query to list customers with pagination
type ListCustomersQuery struct {
	Page           int                    `json:"page"`
	Limit          int                    `json:"limit"`
	Status         *domain.CustomerStatus `json:"status,omitempty"`
	IncludeDeleted bool                   `json:"include_deleted"`
	SortBy         string                 `json:"sort_by"`
	SortOrder      string                 `json:"sort_order"`
	CreatedAfter   *string                `json:"created_after,omitempty"`
	CreatedBefore  *string                `json:"created_before,omitempty"`
	UpdatedAfter   *string                `json:"updated_after,omitempty"`
	UpdatedBefore  *string                `json:"updated_before,omitempty"`
}

// ListCustomersResult represents the result of ListCustomersQuery
type ListCustomersResult struct {
	domain.CustomerListResult
}

// SearchCustomersQuery represents a query to search customers
type SearchCustomersQuery struct {
	Query     string                 `json:"query"`
	Email     string                 `json:"email"`
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Page      int                    `json:"page"`
	Limit     int                    `json:"limit"`
	Status    *domain.CustomerStatus `json:"status,omitempty"`
	SortBy    string                 `json:"sort_by"`
	SortOrder string                 `json:"sort_order"`
}

// SearchCustomersResult represents the result of SearchCustomersQuery
type SearchCustomersResult struct {
	domain.CustomerListResult
}
