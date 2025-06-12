package main

import (
	"log"

	"github.com/gin-gonic/gin"

	commandhandlers "golang_modular_monolith/internal/modules/customer/application/command_handlers"
	queryhandlers "golang_modular_monolith/internal/modules/customer/application/query_handlers"
	customerhttp "golang_modular_monolith/internal/modules/customer/infrastructure/http"
	"golang_modular_monolith/internal/modules/customer/infrastructure/http/handlers"
	"golang_modular_monolith/internal/modules/customer/infrastructure/persistence"

	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/config"
	"golang_modular_monolith/internal/shared/infrastructure/database"
	"golang_modular_monolith/internal/shared/infrastructure/eventbus"
)

func main() {
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

	// Initialize dependencies
	dependencies, err := initDependencies(eventBus)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Initialize Gin router
	router := initRouter(cfg, dependencies)

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

// Dependencies holds all application dependencies
type Dependencies struct {
	CustomerHandler *handlers.CustomerHandler
}

// initDependencies initializes all application dependencies
func initDependencies(eventBus domain.EventBus) (*Dependencies, error) {
	// Customer repositories using database manager
	customerRepo, err := persistence.NewPostgreSQLCustomerRepositoryFromManager()
	if err != nil {
		return nil, err
	}

	customerQueryRepo, err := persistence.NewPostgreSQLCustomerQueryRepositoryFromManager()
	if err != nil {
		return nil, err
	}

	// Domain services
	customerDomainService := persistence.NewCustomerDomainService(customerRepo)

	// Command handlers
	createCustomerHandler := commandhandlers.NewCreateCustomerHandler(
		customerRepo,
		customerDomainService,
		eventBus,
	)

	// Query handlers
	getCustomerHandler := queryhandlers.NewGetCustomerHandler(customerQueryRepo)
	listCustomersHandler := queryhandlers.NewListCustomersHandler(customerQueryRepo)
	searchCustomersHandler := queryhandlers.NewSearchCustomersHandler(customerQueryRepo)

	// HTTP handlers
	customerHandler := handlers.NewCustomerHandler(
		createCustomerHandler,
		getCustomerHandler,
		listCustomersHandler,
		searchCustomersHandler,
	)

	return &Dependencies{
		CustomerHandler: customerHandler,
	}, nil
}

// initRouter initializes Gin router with all routes
func initRouter(cfg *config.Config, deps *Dependencies) *gin.Engine {
	// Set Gin mode from config
	gin.SetMode(cfg.App.GinMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Add health check
	router.GET("/health", healthCheckHandler(cfg))

	// API routes
	api := router.Group("/api/v1")
	{
		// Register customer routes
		customerhttp.RegisterCustomerRoutes(api, deps.CustomerHandler)
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

// healthCheckHandler returns a health check handler with config
func healthCheckHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		manager := database.GetGlobalManager()
		databases := manager.GetRegisteredDatabases()

		c.JSON(200, gin.H{
			"status":      "healthy",
			"service":     cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
			"databases":   databases,
			"message":     "🔥 Viper + Docker hot reload is working perfectly!",
			"timestamp":   "2025-06-12",
		})
	}
}
