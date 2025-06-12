package order

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register order module on package import
func init() {
	registry.RegisterModule("order", func() domain.Module {
		return NewOrderModule()
	})
}

// OrderModule implements the Module interface
type OrderModule struct {
	name string

	// Dependencies
	eventBus domain.EventBus
}

// NewOrderModule creates a new order module
func NewOrderModule() *OrderModule {
	return &OrderModule{
		name: "order",
	}
}

// Name returns the module name
func (m *OrderModule) Name() string {
	return m.name
}

// Initialize initializes the order module with dependencies
func (m *OrderModule) Initialize(deps domain.ModuleDependencies) error {
	log.Printf("🔧 Initializing %s module...", m.name)

	// Store event bus
	m.eventBus = deps.EventBus

	// TODO: Initialize order-specific dependencies
	// - Order repositories
	// - Order domain services
	// - Order command/query handlers
	// - Order HTTP handlers

	log.Printf("✅ %s module initialized successfully (skeleton)", m.name)
	return nil
}

// RegisterRoutes registers HTTP routes for the order module
func (m *OrderModule) RegisterRoutes(router *gin.RouterGroup) {
	log.Printf("🌐 Registering routes for %s module", m.name)

	// TODO: Register order routes
	orderGroup := router.Group("/orders")
	{
		orderGroup.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Order module is working!",
				"module":  m.name,
				"status":  "skeleton",
			})
		})
	}
}

// Health checks if the order module is healthy
func (m *OrderModule) Health(ctx context.Context) error {
	// TODO: Add real health checks
	// - Database connectivity
	// - External service dependencies

	return nil
}

// Start starts the order module (optional lifecycle method)
func (m *OrderModule) Start(ctx context.Context) error {
	log.Printf("🚀 Starting %s module", m.name)

	// TODO: Start order-specific services
	// - Background workers
	// - Event handlers

	log.Printf("✅ %s module started successfully (skeleton)", m.name)
	return nil
}

// Stop stops the order module (optional lifecycle method)
func (m *OrderModule) Stop(ctx context.Context) error {
	log.Printf("🛑 Stopping %s module", m.name)

	// TODO: Cleanup order resources

	log.Printf("✅ %s module stopped successfully", m.name)
	return nil
}
