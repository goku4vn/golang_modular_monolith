package database

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds configuration for a single database
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	URL      string // Alternative to individual fields
}

// DatabaseManager manages multiple database connections
type DatabaseManager struct {
	connections map[string]*gorm.DB
	configs     map[string]*DatabaseConfig
	mu          sync.RWMutex
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{
		connections: make(map[string]*gorm.DB),
		configs:     make(map[string]*DatabaseConfig),
	}
}

// RegisterDatabase registers a database configuration
func (dm *DatabaseManager) RegisterDatabase(name string, config *DatabaseConfig) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.configs[name] = config
}

// GetConnection returns a database connection by name
func (dm *DatabaseManager) GetConnection(name string) (*gorm.DB, error) {
	dm.mu.RLock()
	if conn, exists := dm.connections[name]; exists {
		dm.mu.RUnlock()
		return conn, nil
	}
	dm.mu.RUnlock()

	// Create new connection
	return dm.createConnection(name)
}

// createConnection creates a new database connection
func (dm *DatabaseManager) createConnection(name string) (*gorm.DB, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check again in case another goroutine created it
	if conn, exists := dm.connections[name]; exists {
		return conn, nil
	}

	config, exists := dm.configs[name]
	if !exists {
		return nil, fmt.Errorf("database configuration not found for: %s", name)
	}

	dsn := dm.buildDSN(config)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", name, err)
	}

	dm.connections[name] = db
	log.Printf("Database connection established for: %s", name)

	return db, nil
}

// buildDSN builds database connection string
func (dm *DatabaseManager) buildDSN(config *DatabaseConfig) string {
	if config.URL != "" {
		return config.URL
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Name,
		config.SSLMode,
	)
}

// CloseAll closes all database connections
func (dm *DatabaseManager) CloseAll() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for name, db := range dm.connections {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing database %s: %v", name, err)
			} else {
				log.Printf("Database connection closed for: %s", name)
			}
		}
	}

	dm.connections = make(map[string]*gorm.DB)
	return nil
}

// GetRegisteredDatabases returns list of registered database names
func (dm *DatabaseManager) GetRegisteredDatabases() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	names := make([]string, 0, len(dm.configs))
	for name := range dm.configs {
		names = append(names, name)
	}
	return names
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv(prefix string) *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv(prefix+"_HOST", "localhost"),
		Port:     getEnv(prefix+"_PORT", "5432"),
		Name:     getEnv(prefix+"_NAME", ""),
		User:     getEnv(prefix+"_USER", "postgres"),
		Password: getEnv(prefix+"_PASSWORD", "postgres"),
		SSLMode:  getEnv(prefix+"_SSL_MODE", "disable"),
		URL:      getEnv(prefix+"_URL", ""),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Global database manager instance
var globalManager *DatabaseManager
var once sync.Once

// GetGlobalManager returns the global database manager instance
func GetGlobalManager() *DatabaseManager {
	once.Do(func() {
		globalManager = NewDatabaseManager()
	})
	return globalManager
}
