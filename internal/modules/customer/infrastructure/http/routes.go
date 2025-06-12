package http

import (
	"golang_modular_monolith/internal/modules/customer/infrastructure/http/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterCustomerRoutes registers customer routes
func RegisterCustomerRoutes(router *gin.RouterGroup, customerHandler *handlers.CustomerHandler) {
	// Customer routes
	customers := router.Group("/customers")
	{
		customers.POST("", customerHandler.CreateCustomer)
		customers.GET("", customerHandler.ListCustomers)
		customers.GET("/search", customerHandler.SearchCustomers)
		customers.GET("/:id", customerHandler.GetCustomer)
	}
}
