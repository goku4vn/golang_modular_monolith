package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/config"
	"golang_modular_monolith/internal/shared/infrastructure/database"
	"golang_modular_monolith/internal/shared/infrastructure/eventbus"
	"golang_modular_monolith/internal/shared/infrastructure/registry"

	// Import modules package to trigger auto-registration of all modules
	"golang_modular_monolith/internal/modules"
)

func main() {
	// Initialize all modules (triggers auto-registration)
	modules.InitializeAllModules()

	// Load configuration using Viper
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("🔧 Configuration loaded successfully")
	log.Printf("📱 App: %s v%s (%s)", cfg.App.Name, cfg.App.Version, cfg.App.Environment)
	log.Printf("🌐 Server: %s", cfg.GetServerAddress())
	log.Printf("🗄️ Databases: %v", cfg.GetAvailableDatabases())

	// Initialize database manager with Viper config
	if err := initDatabases(cfg); err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}

	// Initialize event bus
	eventBus := eventbus.NewInMemoryEventBus()

	// Load enabled modules
	moduleRegistry, err := initModules(cfg, eventBus)
	if err != nil {
		log.Fatalf("Failed to initialize modules: %v", err)
	}

	// Initialize Gin router
	router := initRouter(cfg, moduleRegistry)

	// Start modules
	ctx := context.Background()
	if err := moduleRegistry.StartAll(ctx); err != nil {
		log.Fatalf("Failed to start modules: %v", err)
	}

	// Start server
	log.Printf("Starting server on port %s", cfg.App.Port)
	if err := router.Run(cfg.GetServerAddress()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabases initializes all module databases using Viper config
func initDatabases(cfg *config.Config) error {
	log.Println("Initializing databases...")

	// Initialize database manager with Viper config
	manager := database.InitializeWithConfig(cfg)

	// Verify all database connections
	for _, dbName := range cfg.GetAvailableDatabases() {
		if err := manager.VerifyConnection(dbName); err != nil {
			return err
		}
	}

	return nil
}

// initModules loads and initializes all enabled modules
func initModules(cfg *config.Config, eventBus domain.EventBus) (*domain.ModuleRegistry, error) {
	log.Println("🔧 Initializing modules...")

	// Get global module manager
	manager := registry.GetGlobalManager()

	// Load enabled modules from configuration
	if err := manager.LoadEnabledModules(cfg); err != nil {
		return nil, err
	}

	// Get module registry
	moduleRegistry := manager.GetRegistry()

	// Initialize all modules with dependencies
	deps := domain.ModuleDependencies{
		EventBus: eventBus,
		Config:   cfg, // Pass full config, modules can extract what they need
	}

	if err := moduleRegistry.InitializeAll(deps); err != nil {
		return nil, err
	}

	log.Printf("✅ Modules initialized successfully: %v", moduleRegistry.GetModuleNames())
	return moduleRegistry, nil
}

// initRouter initializes Gin router with all routes
func initRouter(cfg *config.Config, moduleRegistry *domain.ModuleRegistry) *gin.Engine {
	// Set Gin mode from config
	gin.SetMode(cfg.App.GinMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Add health check
	router.GET("/health", healthCheckHandler(cfg, moduleRegistry))

	// API routes
	api := router.Group("/api/v1")
	{
		// Register routes for all modules
		moduleRegistry.RegisterAllRoutes(api)
	}

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// healthCheckHandler returns a health check handler with config and modules
func healthCheckHandler(cfg *config.Config, moduleRegistry *domain.ModuleRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		manager := database.GetGlobalManager()
		databases := manager.GetRegisteredDatabases()

		// Check module health
		ctx := context.Background()
		moduleHealth := moduleRegistry.HealthCheckAll(ctx)

		// Determine overall status
		status := "healthy"
		for _, err := range moduleHealth {
			if err != nil {
				status = "unhealthy"
				break
			}
		}

		response := gin.H{
			"status":      status,
			"service":     cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
			"databases":   databases,
			"modules":     moduleRegistry.GetModuleNames(),
			"module_health": func() map[string]string {
				health := make(map[string]string)
				for name, err := range moduleHealth {
					if err != nil {
						health[name] = err.Error()
					} else {
						health[name] = "healthy"
					}
				}
				return health
			}(),
			"message":   "🚀 Modular system with dynamic module loading!",
			"timestamp": "2025-06-12",
		}

		if status == "healthy" {
			c.JSON(200, response)
		} else {
			c.JSON(503, response)
		}
	}
}
