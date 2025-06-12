package database

import (
	"golang_modular_monolith/internal/shared/infrastructure/database"

	"gorm.io/gorm"
)

const (
	// CustomerDatabaseName is the identifier for customer database
	CustomerDatabaseName = "customer"
)

// InitCustomerDatabase initializes customer database configuration
func InitCustomerDatabase() *database.DatabaseConfig {
	// Load configuration from environment variables with CUSTOMER prefix
	config := database.LoadConfigFromEnv("CUSTOMER_DATABASE")

	// Set default database name if not provided
	if config.Name == "" {
		config.Name = "modular_monolith_customer"
	}

	return config
}

// RegisterCustomerDatabase registers customer database with the global manager
func RegisterCustomerDatabase() error {
	manager := database.GetGlobalManager()
	config := InitCustomerDatabase()

	manager.RegisterDatabase(CustomerDatabaseName, config)
	return nil
}

// GetCustomerDB returns the customer database connection
func GetCustomerDB() (*gorm.DB, error) {
	manager := database.GetGlobalManager()
	return manager.GetConnection(CustomerDatabaseName)
}
