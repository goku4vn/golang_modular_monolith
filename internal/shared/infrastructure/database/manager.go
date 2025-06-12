package database

import (
	"fmt"
	"log"
	"os"
	"sync"

	"golang_modular_monolith/internal/shared/infrastructure/config"

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
	appConfig   *config.Config
	mu          sync.RWMutex
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{
		connections: make(map[string]*gorm.DB),
		configs:     make(map[string]*DatabaseConfig),
	}
}

// NewDatabaseManagerWithConfig creates a new database manager with Viper config
func NewDatabaseManagerWithConfig(cfg *config.Config) *DatabaseManager {
	dm := &DatabaseManager{
		connections: make(map[string]*gorm.DB),
		configs:     make(map[string]*DatabaseConfig),
		appConfig:   cfg,
	}

	// Auto-register databases from config
	dm.registerDatabasesFromConfig()

	return dm
}

// registerDatabasesFromConfig registers databases from Viper configuration
func (dm *DatabaseManager) registerDatabasesFromConfig() {
	if dm.appConfig == nil {
		return
	}

	for name, dbConfig := range dm.appConfig.Databases {
		dm.configs[name] = &DatabaseConfig{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Name:     dbConfig.Name,
			User:     dbConfig.User,
			Password: dbConfig.Password,
			SSLMode:  dbConfig.SSLMode,
		}
		log.Printf("%s database registered", name)
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

// VerifyConnection verifies database connection
func (dm *DatabaseManager) VerifyConnection(name string) error {
	db, err := dm.GetConnection(name)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for %s: %w", name, err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database %s: %w", name, err)
	}

	log.Printf("Database connection verified for: %s", name)
	return nil
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

// LoadConfigFromEnv loads database configuration from environment variables (legacy support)
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

// InitializeWithConfig initializes global manager with Viper config
func InitializeWithConfig(cfg *config.Config) *DatabaseManager {
	once.Do(func() {
		globalManager = NewDatabaseManagerWithConfig(cfg)
	})
	return globalManager
}
