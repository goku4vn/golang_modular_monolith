package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	commandhandlers "golang_modular_monolith/internal/modules/customer/application/command_handlers"
	queryhandlers "golang_modular_monolith/internal/modules/customer/application/query_handlers"
	customerdb "golang_modular_monolith/internal/modules/customer/infrastructure/database"
	customerhttp "golang_modular_monolith/internal/modules/customer/infrastructure/http"
	"golang_modular_monolith/internal/modules/customer/infrastructure/http/handlers"
	"golang_modular_monolith/internal/modules/customer/infrastructure/persistence"
	orderdb "golang_modular_monolith/internal/modules/order/infrastructure/database"

	// productdb "golang_modular_monolith/internal/modules/product/infrastructure/database"
	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/database"
	"golang_modular_monolith/internal/shared/infrastructure/eventbus"
)

func main() {
	// Initialize database manager and register all databases
	if err := initDatabases(); err != nil {
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
	router := initRouter(dependencies)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabases initializes all module databases
func initDatabases() error {
	log.Println("Initializing databases...")

	// Register customer database
	if err := customerdb.RegisterCustomerDatabase(); err != nil {
		return err
	}
	log.Println("Customer database registered")

	// Register order database
	if err := orderdb.RegisterOrderDatabase(); err != nil {
		return err
	}
	log.Println("Order database registered")

	// Register product database (TODO: Implement when product module is ready)
	// if err := productdb.RegisterProductDatabase(); err != nil {
	// 	return err
	// }
	// log.Println("Product database registered")

	// Test connections by getting them
	manager := database.GetGlobalManager()
	databases := manager.GetRegisteredDatabases()

	for _, dbName := range databases {
		_, err := manager.GetConnection(dbName)
		if err != nil {
			return err
		}
		log.Printf("Database connection verified for: %s", dbName)
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
func initRouter(deps *Dependencies) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Add health check
	router.GET("/health", healthCheck)

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

// healthCheck handles health check requests
func healthCheck(c *gin.Context) {
	manager := database.GetGlobalManager()
	databases := manager.GetRegisteredDatabases()

	c.JSON(200, gin.H{
		"status":    "healthy",
		"service":   "modular-monolith",
		"databases": databases,
		"message":   "🔥 Docker + .env hot reload is working perfectly!",
		"timestamp": "2025-06-12",
	})
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
