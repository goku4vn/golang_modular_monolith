package user

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register user module on package import
func init() {
	registry.RegisterModule("user", func() domain.Module {
		return NewUserModule()
	})
}

// UserModule implements the Module interface
type UserModule struct {
	name string

	// Dependencies
	eventBus domain.EventBus
}

// NewUserModule creates a new user module
func NewUserModule() *UserModule {
	return &UserModule{
		name: "user",
	}
}

// Name returns the module name
func (m *UserModule) Name() string {
	return m.name
}

// Initialize initializes the user module with dependencies
func (m *UserModule) Initialize(deps domain.ModuleDependencies) error {
	log.Printf("🔧 Initializing %s module...", m.name)

	// Store event bus
	m.eventBus = deps.EventBus

	// TODO: Initialize user-specific dependencies
	// - User repositories
	// - User domain services
	// - User command/query handlers
	// - User HTTP handlers

	log.Printf("✅ %s module initialized successfully (skeleton)", m.name)
	return nil
}

// RegisterRoutes registers HTTP routes for the user module
func (m *UserModule) RegisterRoutes(router *gin.RouterGroup) {
	log.Printf("🌐 Registering routes for %s module", m.name)

	// TODO: Register user routes
	userGroup := router.Group("/users")
	{
		userGroup.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "User module is working!",
				"module":  m.name,
				"status":  "skeleton",
			})
		})
	}
}

// Health checks if the user module is healthy
func (m *UserModule) Health(ctx context.Context) error {
	// TODO: Add real health checks
	// - Database connectivity
	// - External service dependencies

	return nil
}

// Start starts the user module (optional lifecycle method)
func (m *UserModule) Start(ctx context.Context) error {
	log.Printf("🚀 Starting %s module", m.name)

	// TODO: Start user-specific services
	// - Background workers
	// - Event handlers

	log.Printf("✅ %s module started successfully (skeleton)", m.name)
	return nil
}

// Stop stops the user module (optional lifecycle method)
func (m *UserModule) Stop(ctx context.Context) error {
	log.Printf("🛑 Stopping %s module", m.name)

	// TODO: Cleanup user resources

	log.Printf("✅ %s module stopped successfully", m.name)
	return nil
}
