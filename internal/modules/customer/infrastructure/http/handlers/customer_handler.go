package handlers

import (
	"errors"
	"net/http"
	"strconv"

	commandhandlers "golang_modular_monolith/internal/modules/customer/application/command_handlers"
	"golang_modular_monolith/internal/modules/customer/application/commands"
	"golang_modular_monolith/internal/modules/customer/application/queries"
	queryhandlers "golang_modular_monolith/internal/modules/customer/application/query_handlers"
	"golang_modular_monolith/internal/modules/customer/domain"
	shareddomain "golang_modular_monolith/internal/shared/domain"

	"github.com/gin-gonic/gin"
)

// CustomerHandler handles HTTP requests for customer operations
type CustomerHandler struct {
	createCustomerHandler  *commandhandlers.CreateCustomerHandler
	getCustomerHandler     *queryhandlers.GetCustomerHandler
	listCustomersHandler   *queryhandlers.ListCustomersHandler
	searchCustomersHandler *queryhandlers.SearchCustomersHandler
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(
	createCustomerHandler *commandhandlers.CreateCustomerHandler,
	getCustomerHandler *queryhandlers.GetCustomerHandler,
	listCustomersHandler *queryhandlers.ListCustomersHandler,
	searchCustomersHandler *queryhandlers.SearchCustomersHandler,
) *CustomerHandler {
	return &CustomerHandler{
		createCustomerHandler:  createCustomerHandler,
		getCustomerHandler:     getCustomerHandler,
		listCustomersHandler:   listCustomersHandler,
		searchCustomersHandler: searchCustomersHandler,
	}
}

// CreateCustomerRequest represents the request body for creating a customer
type CreateCustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// CreateCustomer handles POST /customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, shareddomain.NewDomainError(
			shareddomain.ErrCodeInvalidInput,
			"Invalid request body: "+err.Error(),
		))
		return
	}

	cmd := &commands.CreateCustomerCommand{
		Name:  req.Name,
		Email: req.Email,
	}

	result, err := h.createCustomerHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetCustomer handles GET /customers/:id
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.handleError(c, shareddomain.NewDomainError(
			shareddomain.ErrCodeInvalidInput,
			"Customer ID is required",
		))
		return
	}

	query := &queries.GetCustomerQuery{
		ID: id,
	}

	result, err := h.getCustomerHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Customer,
	})
}

// ListCustomers handles GET /customers
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// Parse query parameters
	query := &queries.ListCustomersQuery{
		Page:           h.getIntParam(c, "page", 1),
		Limit:          h.getIntParam(c, "limit", 20),
		SortBy:         h.getStringParam(c, "sort_by", "created_at"),
		SortOrder:      h.getStringParam(c, "sort_order", "desc"),
		IncludeDeleted: h.getBoolParam(c, "include_deleted", false),
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := domain.CustomerStatus(statusStr)
		query.Status = &status
	}

	// Parse date filters
	if createdAfter := c.Query("created_after"); createdAfter != "" {
		query.CreatedAfter = &createdAfter
	}
	if createdBefore := c.Query("created_before"); createdBefore != "" {
		query.CreatedBefore = &createdBefore
	}
	if updatedAfter := c.Query("updated_after"); updatedAfter != "" {
		query.UpdatedAfter = &updatedAfter
	}
	if updatedBefore := c.Query("updated_before"); updatedBefore != "" {
		query.UpdatedBefore = &updatedBefore
	}

	result, err := h.listCustomersHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       result.Customers,
		"pagination": result.Pagination,
	})
}

// SearchCustomers handles GET /customers/search
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	query := &queries.SearchCustomersQuery{
		Query:     c.Query("q"),
		Email:     c.Query("email"),
		FirstName: c.Query("first_name"),
		LastName:  c.Query("last_name"),
		Page:      h.getIntParam(c, "page", 1),
		Limit:     h.getIntParam(c, "limit", 20),
		SortBy:    h.getStringParam(c, "sort_by", "created_at"),
		SortOrder: h.getStringParam(c, "sort_order", "desc"),
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := domain.CustomerStatus(statusStr)
		query.Status = &status
	}

	result, err := h.searchCustomersHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       result.Customers,
		"pagination": result.Pagination,
	})
}

// Helper methods

// getIntParam gets an integer parameter with default value
func (h *CustomerHandler) getIntParam(c *gin.Context, key string, defaultValue int) int {
	if str := c.Query(key); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultValue
}

// getStringParam gets a string parameter with default value
func (h *CustomerHandler) getStringParam(c *gin.Context, key string, defaultValue string) string {
	if val := c.Query(key); val != "" {
		return val
	}
	return defaultValue
}

// getBoolParam gets a boolean parameter with default value
func (h *CustomerHandler) getBoolParam(c *gin.Context, key string, defaultValue bool) bool {
	if str := c.Query(key); str != "" {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return defaultValue
}

// handleError handles errors and returns appropriate HTTP responses
func (h *CustomerHandler) handleError(c *gin.Context, err error) {
	var domainErr *shareddomain.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case shareddomain.ErrCodeNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    domainErr.Code,
					"message": domainErr.Message,
				},
			})
		case shareddomain.ErrCodeAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    domainErr.Code,
					"message": domainErr.Message,
				},
			})
		case shareddomain.ErrCodeInvalidInput, shareddomain.ErrCodeValidationFailed:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    domainErr.Code,
					"message": domainErr.Message,
					"field":   domainErr.Field,
				},
			})
		case shareddomain.ErrCodeUnauthorized:
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    domainErr.Code,
					"message": domainErr.Message,
				},
			})
		case shareddomain.ErrCodeForbidden:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    domainErr.Code,
					"message": domainErr.Message,
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "An internal error occurred",
				},
			})
		}
		return
	}

	// Handle standard errors
	if shareddomain.IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Resource not found",
			},
		})
		return
	}

	// Generic error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An internal error occurred",
		},
	})
}
