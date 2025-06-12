package domain

import (
	"errors"
	"fmt"
)

// Common domain errors
var (
	// ErrNotFound indicates that a resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates that a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates forbidden access
	ErrForbidden = errors.New("forbidden")

	// ErrConcurrencyConflict indicates a concurrency conflict
	ErrConcurrencyConflict = errors.New("concurrency conflict")

	// ErrInvalidState indicates invalid aggregate state
	ErrInvalidState = errors.New("invalid state")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Cause   error  `json:"-"`
}

// Error implements the error interface
func (e DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
	}
}

// NewDomainErrorWithField creates a new domain error with field
func NewDomainErrorWithField(code, message, field string) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// NewDomainErrorWithCause creates a new domain error with cause
func NewDomainErrorWithCause(code, message string, cause error) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Common domain error codes
const (
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeAlreadyExists       = "ALREADY_EXISTS"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeValidationFailed    = "VALIDATION_FAILED"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeConcurrencyConflict = "CONCURRENCY_CONFLICT"
	ErrCodeInvalidState        = "INVALID_STATE"
	ErrCodeBusinessRule        = "BUSINESS_RULE_VIOLATION"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrorWithValue creates a new validation error with value
func NewValidationErrorWithValue(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("validation failed for %d fields", len(e))
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, NewValidationError(field, message))
}

// AddWithValue adds a validation error with value
func (e *ValidationErrors) AddWithValue(field, message string, value interface{}) {
	*e = append(*e, NewValidationErrorWithValue(field, message, value))
}

// HasErrors checks if there are any validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// BusinessRuleError represents a business rule violation
type BusinessRuleError struct {
	Rule    string                 `json:"rule"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e BusinessRuleError) Error() string {
	return fmt.Sprintf("business rule violation: %s - %s", e.Rule, e.Message)
}

// NewBusinessRuleError creates a new business rule error
func NewBusinessRuleError(rule, message string) BusinessRuleError {
	return BusinessRuleError{
		Rule:    rule,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewBusinessRuleErrorWithContext creates a new business rule error with context
func NewBusinessRuleErrorWithContext(rule, message string, context map[string]interface{}) BusinessRuleError {
	return BusinessRuleError{
		Rule:    rule,
		Message: message,
		Context: context,
	}
}

// AddContext adds context to the business rule error
func (e *BusinessRuleError) AddContext(key string, value interface{}) {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDomainError checks if error is a domain error
func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}
