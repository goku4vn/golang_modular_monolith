package persistence

import (
	"context"
	"errors"
	"fmt"

	"golang_modular_monolith/internal/modules/customer/domain"
	customerdb "golang_modular_monolith/internal/modules/customer/infrastructure/database"
	shareddomain "golang_modular_monolith/internal/shared/domain"

	"gorm.io/gorm"
)

// CustomerModel represents the customer database model
type CustomerModel struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	Name      string `gorm:"type:varchar(255);not null"`
	Email     string `gorm:"type:varchar(255);not null;unique"`
	Status    string `gorm:"type:customer_status;not null;default:active"`
	Version   int    `gorm:"not null;default:0"`
	CreatedAt string `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt string `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for GORM
func (CustomerModel) TableName() string {
	return "customers"
}

// ToEntity converts database model to domain entity
func (m *CustomerModel) ToEntity() (*domain.Customer, error) {
	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email in database: %w", err)
	}

	customer := &domain.Customer{
		BaseAggregateRoot: shareddomain.NewBaseAggregateRootWithID(m.ID),
		Name:              m.Name,
		Email:             email,
		Status:            domain.CustomerStatus(m.Status),
	}

	// Set version from database
	customer.Version = m.Version

	return customer, nil
}

// FromEntity converts domain entity to database model
func (m *CustomerModel) FromEntity(customer *domain.Customer) {
	m.ID = customer.GetID()
	m.Name = customer.Name
	m.Email = customer.Email.Value
	m.Status = string(customer.Status)
	m.Version = customer.GetVersion()
}

// PostgreSQLCustomerRepository implements CustomerRepository using PostgreSQL
type PostgreSQLCustomerRepository struct {
	db *gorm.DB
}

// NewPostgreSQLCustomerRepository creates a new PostgreSQL customer repository
func NewPostgreSQLCustomerRepository(db *gorm.DB) *PostgreSQLCustomerRepository {
	return &PostgreSQLCustomerRepository{
		db: db,
	}
}

// NewPostgreSQLCustomerRepositoryFromManager creates repository using database manager
func NewPostgreSQLCustomerRepositoryFromManager() (*PostgreSQLCustomerRepository, error) {
	db, err := customerdb.GetCustomerDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get customer database: %w", err)
	}

	return &PostgreSQLCustomerRepository{
		db: db,
	}, nil
}

// Save saves a customer (create or update)
func (r *PostgreSQLCustomerRepository) Save(ctx context.Context, customer *domain.Customer) error {
	model := &CustomerModel{}
	model.FromEntity(customer)

	// Use optimistic locking with version
	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		// Check for unique constraint violation (email)
		if isUniqueViolationError(result.Error) {
			return shareddomain.NewDomainErrorWithCause(
				shareddomain.ErrCodeAlreadyExists,
				"customer with this email already exists",
				result.Error,
			)
		}
		return fmt.Errorf("failed to save customer: %w", result.Error)
	}

	// Clear uncommitted events after successful save
	customer.ClearUncommittedEvents()

	return nil
}

// GetByID retrieves a customer by ID
func (r *PostgreSQLCustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	var model CustomerModel
	result := r.db.WithContext(ctx).Where("id = ? AND status != ?", id, domain.CustomerStatusDeleted).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shareddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by ID: %w", result.Error)
	}

	return model.ToEntity()
}

// GetByEmail retrieves a customer by email
func (r *PostgreSQLCustomerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	var model CustomerModel
	result := r.db.WithContext(ctx).Where("email = ? AND status != ?", email, domain.CustomerStatusDeleted).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, shareddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", result.Error)
	}

	return model.ToEntity()
}

// Delete soft deletes a customer
func (r *PostgreSQLCustomerRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&CustomerModel{}).
		Where("id = ? AND status != ?", id, domain.CustomerStatusDeleted).
		Update("status", domain.CustomerStatusDeleted)

	if result.Error != nil {
		return fmt.Errorf("failed to delete customer: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shareddomain.ErrNotFound
	}

	return nil
}

// Exists checks if a customer exists by ID
func (r *PostgreSQLCustomerRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&CustomerModel{}).
		Where("id = ? AND status != ?", id, domain.CustomerStatusDeleted).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check customer existence: %w", result.Error)
	}

	return count > 0, nil
}

// ExistsByEmail checks if a customer exists by email
func (r *PostgreSQLCustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&CustomerModel{}).
		Where("email = ? AND status != ?", email, domain.CustomerStatusDeleted).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check customer existence by email: %w", result.Error)
	}

	return count > 0, nil
}

// isUniqueViolationError checks if the error is a unique constraint violation
func isUniqueViolationError(err error) bool {
	// Check for PostgreSQL unique violation error
	// Error code 23505 is unique_violation in PostgreSQL
	return err != nil && (
	// GORM may wrap the error, so check the string content
	fmt.Sprintf("%v", err) == "ERROR: duplicate key value violates unique constraint" ||
		fmt.Sprintf("%v", err) == "UNIQUE constraint failed")
}

// CustomerDomainServiceImpl implements CustomerDomainService
type CustomerDomainServiceImpl struct {
	repo domain.CustomerRepository
}

// NewCustomerDomainService creates a new customer domain service
func NewCustomerDomainService(repo domain.CustomerRepository) *CustomerDomainServiceImpl {
	return &CustomerDomainServiceImpl{
		repo: repo,
	}
}

// IsEmailUnique checks if email is unique
func (s *CustomerDomainServiceImpl) IsEmailUnique(ctx context.Context, email string, excludeCustomerID ...string) (bool, error) {
	customer, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, shareddomain.ErrNotFound) {
			return true, nil // Email is unique
		}
		return false, err
	}

	// If we found a customer, check if it's the excluded one
	if len(excludeCustomerID) > 0 && customer.GetID() == excludeCustomerID[0] {
		return true, nil // Same customer, email is unique for others
	}

	return false, nil // Email is not unique
}

// CanDeleteCustomer checks if customer can be deleted
func (s *CustomerDomainServiceImpl) CanDeleteCustomer(ctx context.Context, customerID string) (bool, error) {
	exists, err := s.repo.Exists(ctx, customerID)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, shareddomain.ErrNotFound
	}

	// Add business rules for deletion here
	// For example: check if customer has active orders, etc.
	// For now, all existing customers can be deleted
	return true, nil
}
