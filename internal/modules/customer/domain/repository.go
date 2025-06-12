package domain

import (
	"context"
)

// CustomerRepository defines the interface for customer persistence
type CustomerRepository interface {
	// Save saves a customer (create or update)
	Save(ctx context.Context, customer *Customer) error

	// GetByID retrieves a customer by ID
	GetByID(ctx context.Context, id string) (*Customer, error)

	// GetByEmail retrieves a customer by email
	GetByEmail(ctx context.Context, email string) (*Customer, error)

	// Delete soft deletes a customer
	Delete(ctx context.Context, id string) error

	// Exists checks if a customer exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByEmail checks if a customer exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// CustomerQueryRepository defines the interface for customer queries (read-side CQRS)
type CustomerQueryRepository interface {
	// GetByID retrieves a customer view by ID
	GetByID(ctx context.Context, id string) (*CustomerView, error)

	// GetByEmail retrieves a customer view by email
	GetByEmail(ctx context.Context, email string) (*CustomerView, error)

	// List retrieves customers with pagination and filtering
	List(ctx context.Context, params ListCustomersParams) (*CustomerListResult, error)

	// Search searches customers by various criteria
	Search(ctx context.Context, params SearchCustomersParams) (*CustomerListResult, error)

	// Count returns the total number of customers matching criteria
	Count(ctx context.Context, params CountCustomersParams) (int64, error)
}

// CustomerView represents a read-model for customer queries
type CustomerView struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Status    CustomerStatus `json:"status"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

// ListCustomersParams represents parameters for listing customers
type ListCustomersParams struct {
	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // id, email, name, created_at, updated_at
	SortOrder string `json:"sort_order"` // asc, desc

	// Filtering
	Status         *CustomerStatus `json:"status,omitempty"`
	IncludeDeleted bool            `json:"include_deleted"`

	// Date filtering
	CreatedAfter  *string `json:"created_after,omitempty"`
	CreatedBefore *string `json:"created_before,omitempty"`
	UpdatedAfter  *string `json:"updated_after,omitempty"`
	UpdatedBefore *string `json:"updated_before,omitempty"`
}

// SearchCustomersParams represents parameters for searching customers
type SearchCustomersParams struct {
	ListCustomersParams

	// Search criteria
	Query     string `json:"query"`      // Search in name, email
	Email     string `json:"email"`      // Exact email match
	FirstName string `json:"first_name"` // Partial first name match (for compatibility)
	LastName  string `json:"last_name"`  // Partial last name match (for compatibility)
}

// CountCustomersParams represents parameters for counting customers
type CountCustomersParams struct {
	Status         *CustomerStatus `json:"status,omitempty"`
	IncludeDeleted bool            `json:"include_deleted"`
	CreatedAfter   *string         `json:"created_after,omitempty"`
	CreatedBefore  *string         `json:"created_before,omitempty"`
}

// CustomerListResult represents the result of a customer list query
type CustomerListResult struct {
	Customers  []CustomerView   `json:"customers"`
	Pagination PaginationResult `json:"pagination"`
}

// PaginationResult represents pagination information
type PaginationResult struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginationResult creates a new pagination result
func NewPaginationResult(page, limit int, total int64) PaginationResult {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationResult{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// Validate validates the list parameters
func (p *ListCustomersParams) Validate() error {
	if p.Page <= 0 {
		p.Page = 1
	}

	if p.Limit <= 0 {
		p.Limit = 20
	}

	// Maximum limit
	if p.Limit > 100 {
		p.Limit = 100
	}

	// Valid sort fields
	validSortFields := map[string]bool{
		"id":         true,
		"email":      true,
		"name":       true,
		"created_at": true,
		"updated_at": true,
	}

	if p.SortBy != "" && !validSortFields[p.SortBy] {
		p.SortBy = "created_at"
	}

	if p.SortBy == "" {
		p.SortBy = "created_at"
	}

	if p.SortOrder != "asc" && p.SortOrder != "desc" {
		p.SortOrder = "desc"
	}

	return nil
}

// Validate validates the search parameters
func (p *SearchCustomersParams) Validate() error {
	return p.ListCustomersParams.Validate()
}

// GetOffset calculates the offset for pagination
func (p *ListCustomersParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// CustomerDomainService defines domain services for customer
type CustomerDomainService interface {
	// IsEmailUnique checks if email is unique
	IsEmailUnique(ctx context.Context, email string, excludeCustomerID ...string) (bool, error)

	// CanDeleteCustomer checks if customer can be deleted
	CanDeleteCustomer(ctx context.Context, customerID string) (bool, error)
}
