package customer

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	commandhandlers "golang_modular_monolith/internal/modules/customer/application/command_handlers"
	queryhandlers "golang_modular_monolith/internal/modules/customer/application/query_handlers"
	customerhttp "golang_modular_monolith/internal/modules/customer/infrastructure/http"
	"golang_modular_monolith/internal/modules/customer/infrastructure/http/handlers"
	"golang_modular_monolith/internal/modules/customer/infrastructure/persistence"

	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register customer module on package import
func init() {
	registry.RegisterModule("customer", func() domain.Module {
		return NewCustomerModule()
	})
}

// CustomerModule implements the Module interface
type CustomerModule struct {
	name    string
	handler *handlers.CustomerHandler

	// Dependencies
	eventBus domain.EventBus
}

// NewCustomerModule creates a new customer module
func NewCustomerModule() *CustomerModule {
	return &CustomerModule{
		name: "customer",
	}
}

// Name returns the module name
func (m *CustomerModule) Name() string {
	return m.name
}

// Initialize initializes the customer module with dependencies
func (m *CustomerModule) Initialize(deps domain.ModuleDependencies) error {
	log.Printf("🔧 Initializing %s module...", m.name)

	// Store event bus
	m.eventBus = deps.EventBus

	// Create repositories using factory pattern
	customerRepo, err := persistence.NewPostgreSQLCustomerRepositoryFromManager()
	if err != nil {
		return fmt.Errorf("failed to create customer repository: %w", err)
	}

	customerQueryRepo, err := persistence.NewPostgreSQLCustomerQueryRepositoryFromManager()
	if err != nil {
		return fmt.Errorf("failed to create customer query repository: %w", err)
	}

	// Create domain services
	customerDomainService := persistence.NewCustomerDomainService(customerRepo)

	// Create command handlers
	createCustomerHandler := commandhandlers.NewCreateCustomerHandler(
		customerRepo,
		customerDomainService,
		m.eventBus,
	)

	// Create query handlers
	getCustomerHandler := queryhandlers.NewGetCustomerHandler(customerQueryRepo)
	listCustomersHandler := queryhandlers.NewListCustomersHandler(customerQueryRepo)
	searchCustomersHandler := queryhandlers.NewSearchCustomersHandler(customerQueryRepo)

	// Create HTTP handlers
	m.handler = handlers.NewCustomerHandler(
		createCustomerHandler,
		getCustomerHandler,
		listCustomersHandler,
		searchCustomersHandler,
	)

	log.Printf("✅ %s module initialized successfully", m.name)
	return nil
}

// RegisterRoutes registers HTTP routes for the customer module
func (m *CustomerModule) RegisterRoutes(router *gin.RouterGroup) {
	log.Printf("🌐 Registering routes for %s module", m.name)
	customerhttp.RegisterCustomerRoutes(router, m.handler)
}

// Health checks if the customer module is healthy
func (m *CustomerModule) Health(ctx context.Context) error {
	// Check if handler is initialized
	if m.handler == nil {
		return fmt.Errorf("customer handler not initialized")
	}

	// Could add more health checks here:
	// - Database connectivity
	// - External service dependencies
	// - Cache connectivity

	return nil
}

// Start starts the customer module (optional lifecycle method)
func (m *CustomerModule) Start(ctx context.Context) error {
	log.Printf("🚀 Starting %s module", m.name)

	// Register event handlers if needed
	if err := m.registerEventHandlers(); err != nil {
		return fmt.Errorf("failed to register event handlers: %w", err)
	}

	log.Printf("✅ %s module started successfully", m.name)
	return nil
}

// Stop stops the customer module (optional lifecycle method)
func (m *CustomerModule) Stop(ctx context.Context) error {
	log.Printf("🛑 Stopping %s module", m.name)

	// Cleanup resources if needed
	// - Close connections
	// - Unregister event handlers
	// - Stop background workers

	log.Printf("✅ %s module stopped successfully", m.name)
	return nil
}

// registerEventHandlers registers event handlers for cross-module communication
func (m *CustomerModule) registerEventHandlers() error {
	// Example: Register handlers for events from other modules
	// m.eventBus.SubscribeToEventType("order.created", m.handleOrderCreated)

	return nil
}

// GetHandler returns the HTTP handler (for backward compatibility)
func (m *CustomerModule) GetHandler() *handlers.CustomerHandler {
	return m.handler
}
