package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang_modular_monolith/internal/modules/customer/domain"
	customerdb "golang_modular_monolith/internal/modules/customer/infrastructure/database"
	shareddomain "golang_modular_monolith/internal/shared/domain"

	"gorm.io/gorm"
)

// PostgreSQLCustomerQueryRepository implements CustomerQueryRepository using PostgreSQL
type PostgreSQLCustomerQueryRepository struct {
	db *gorm.DB
}

// NewPostgreSQLCustomerQueryRepository creates a new PostgreSQL customer query repository
func NewPostgreSQLCustomerQueryRepository(db *gorm.DB) *PostgreSQLCustomerQueryRepository {
	return &PostgreSQLCustomerQueryRepository{
		db: db,
	}
}

// NewPostgreSQLCustomerQueryRepositoryFromManager creates repository using database manager
func NewPostgreSQLCustomerQueryRepositoryFromManager() (*PostgreSQLCustomerQueryRepository, error) {
	db, err := customerdb.GetCustomerDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get customer database: %w", err)
	}

	return &PostgreSQLCustomerQueryRepository{
		db: db,
	}, nil
}

// toCustomerView converts CustomerModel to CustomerView
func (r *PostgreSQLCustomerQueryRepository) toCustomerView(model *CustomerModel) *domain.CustomerView {
	return &domain.CustomerView{
		ID:        model.ID,
		Email:     model.Email,
		Name:      model.Name,
		Status:    domain.CustomerStatus(model.Status),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// GetByID retrieves a customer view by ID
func (r *PostgreSQLCustomerQueryRepository) GetByID(ctx context.Context, id string) (*domain.CustomerView, error) {
	var model CustomerModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shareddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by ID: %w", result.Error)
	}

	return r.toCustomerView(&model), nil
}

// GetByEmail retrieves a customer view by email
func (r *PostgreSQLCustomerQueryRepository) GetByEmail(ctx context.Context, email string) (*domain.CustomerView, error) {
	var model CustomerModel
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shareddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", result.Error)
	}

	return r.toCustomerView(&model), nil
}

// List retrieves customers with pagination and filtering
func (r *PostgreSQLCustomerQueryRepository) List(ctx context.Context, params domain.ListCustomersParams) (*domain.CustomerListResult, error) {
	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&CustomerModel{})

	// Apply filters
	query = r.applyListFilters(query, params)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Apply pagination and sorting
	query = query.Offset(params.GetOffset()).Limit(params.Limit)
	query = query.Order(fmt.Sprintf("%s %s", params.SortBy, params.SortOrder))

	// Execute query
	var models []CustomerModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	// Convert to views
	customers := make([]domain.CustomerView, len(models))
	for i, model := range models {
		customers[i] = *r.toCustomerView(&model)
	}

	return &domain.CustomerListResult{
		Customers:  customers,
		Pagination: domain.NewPaginationResult(params.Page, params.Limit, total),
	}, nil
}

// Search searches customers by various criteria
func (r *PostgreSQLCustomerQueryRepository) Search(ctx context.Context, params domain.SearchCustomersParams) (*domain.CustomerListResult, error) {
	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&CustomerModel{})

	// Apply filters
	query = r.applyListFilters(query, params.ListCustomersParams)

	// Apply search criteria
	query = r.applySearchFilters(query, params)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Apply pagination and sorting
	query = query.Offset(params.GetOffset()).Limit(params.Limit)
	query = query.Order(fmt.Sprintf("%s %s", params.SortBy, params.SortOrder))

	// Execute query
	var models []CustomerModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	// Convert to views
	customers := make([]domain.CustomerView, len(models))
	for i, model := range models {
		customers[i] = *r.toCustomerView(&model)
	}

	return &domain.CustomerListResult{
		Customers:  customers,
		Pagination: domain.NewPaginationResult(params.Page, params.Limit, total),
	}, nil
}

// Count returns the total number of customers matching criteria
func (r *PostgreSQLCustomerQueryRepository) Count(ctx context.Context, params domain.CountCustomersParams) (int64, error) {
	query := r.db.WithContext(ctx).Model(&CustomerModel{})

	// Apply filters
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if !params.IncludeDeleted {
		query = query.Where("status != ?", domain.CustomerStatusDeleted)
	}

	if params.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *params.CreatedAfter)
	}

	if params.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *params.CreatedBefore)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return count, nil
}

// applyListFilters applies common list filters to the query
func (r *PostgreSQLCustomerQueryRepository) applyListFilters(query *gorm.DB, params domain.ListCustomersParams) *gorm.DB {
	// Status filter
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	// Include deleted filter
	if !params.IncludeDeleted {
		query = query.Where("status != ?", domain.CustomerStatusDeleted)
	}

	// Date filters
	if params.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *params.CreatedAfter)
	}

	if params.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *params.CreatedBefore)
	}

	if params.UpdatedAfter != nil {
		query = query.Where("updated_at >= ?", *params.UpdatedAfter)
	}

	if params.UpdatedBefore != nil {
		query = query.Where("updated_at <= ?", *params.UpdatedBefore)
	}

	return query
}

// applySearchFilters applies search-specific filters to the query
func (r *PostgreSQLCustomerQueryRepository) applySearchFilters(query *gorm.DB, params domain.SearchCustomersParams) *gorm.DB {
	// General search query (search in name and email)
	if params.Query != "" {
		searchTerm := "%" + strings.ToLower(params.Query) + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(email) LIKE ?)", searchTerm, searchTerm)
	}

	// Specific field searches
	if params.Email != "" {
		query = query.Where("email = ?", params.Email)
	}

	if params.FirstName != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(params.FirstName)+"%")
	}

	if params.LastName != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(params.LastName)+"%")
	}

	return query
}
