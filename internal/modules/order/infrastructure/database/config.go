package database

import (
	"golang_modular_monolith/internal/shared/infrastructure/database"

	"gorm.io/gorm"
)

const (
	// OrderDatabaseName is the identifier for order database
	OrderDatabaseName = "order"
)

// InitOrderDatabase initializes order database configuration
func InitOrderDatabase() *database.DatabaseConfig {
	// Load configuration from environment variables with ORDER prefix
	config := database.LoadConfigFromEnv("ORDER_DATABASE")

	// Set default database name if not provided
	if config.Name == "" {
		config.Name = "modular_monolith_order"
	}

	return config
}

// RegisterOrderDatabase registers order database with the global manager
func RegisterOrderDatabase() error {
	manager := database.GetGlobalManager()
	config := InitOrderDatabase()

	manager.RegisterDatabase(OrderDatabaseName, config)
	return nil
}

// GetOrderDB returns the order database connection
func GetOrderDB() (*gorm.DB, error) {
	manager := database.GetGlobalManager()
	return manager.GetConnection(OrderDatabaseName)
}
