package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App       AppConfig                 `mapstructure:"app"`
	Databases map[string]DatabaseConfig `mapstructure:"databases"`
	Modules   *ModulesConfig            `mapstructure:"modules"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Port        string `mapstructure:"port"`
	GinMode     string `mapstructure:"gin_mode"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

// LoadConfig loads configuration from environment variables, Vault, and config files
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("No config file found, using environment variables and defaults")
	}

	// Load modules configuration
	modulesConfig, err := LoadModulesConfig()
	if err != nil {
		log.Printf("⚠️ Failed to load modules config: %v", err)
		// Create default modules config
		modulesConfig = createDefaultModulesConfig()
	}

	// Load secrets from Vault (highest priority)
	if err := loadFromVault(modulesConfig); err != nil {
		log.Printf("⚠️ Failed to load secrets from Vault: %v", err)
		// Don't fail completely, continue with other config sources
	}

	// Load environment-specific configurations
	loadDatabaseConfigs()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Set modules config
	config.Modules = modulesConfig

	// Convert modules config to database config
	if err := convertModulesConfigToDatabaseConfig(&config, modulesConfig); err != nil {
		log.Printf("⚠️ Failed to convert modules config to database config: %v", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "modular-monolith")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.gin_mode", "debug")

	// Set dynamic database defaults based on modules configuration
	setDynamicDatabaseDefaults()
}

// setDynamicDatabaseDefaults sets database defaults based on modules configuration
func setDynamicDatabaseDefaults() {
	// Get modules from config - if empty, no defaults will be set
	modules := getAvailableModuleNames()
	if len(modules) == 0 {
		log.Println("⚠️ No modules found in config, skipping database defaults")
		return
	}

	log.Printf("🔧 Setting database defaults for modules: %v", modules)
	// Get database prefix from modules config or use default
	databasePrefix := getDatabasePrefix()

	// Set defaults for each module
	for _, module := range modules {
		viper.SetDefault(fmt.Sprintf("databases.%s.host", module), "localhost")
		viper.SetDefault(fmt.Sprintf("databases.%s.port", module), "5432")
		viper.SetDefault(fmt.Sprintf("databases.%s.user", module), "postgres")
		viper.SetDefault(fmt.Sprintf("databases.%s.password", module), "postgres")
		viper.SetDefault(fmt.Sprintf("databases.%s.name", module), fmt.Sprintf("%s_%s", databasePrefix, module))
		viper.SetDefault(fmt.Sprintf("databases.%s.sslmode", module), "disable")
	}
}

// loadDatabaseConfigs loads database configurations from environment variables
func loadDatabaseConfigs() {
	// Get modules from loaded config - if empty, no database configs will be loaded
	modules := getAvailableModuleNames()
	if len(modules) == 0 {
		log.Println("⚠️ No modules found in config, skipping database config loading")
		return
	}

	log.Printf("🔧 Loading database configs for modules: %v", modules)
	for _, module := range modules {
		prefix := strings.ToUpper(module) + "_DATABASE_"

		// Map environment variables to viper keys
		envMappings := map[string]string{
			prefix + "HOST":     fmt.Sprintf("databases.%s.host", module),
			prefix + "PORT":     fmt.Sprintf("databases.%s.port", module),
			prefix + "USER":     fmt.Sprintf("databases.%s.user", module),
			prefix + "PASSWORD": fmt.Sprintf("databases.%s.password", module),
			prefix + "NAME":     fmt.Sprintf("databases.%s.name", module),
			prefix + "SSLMODE":  fmt.Sprintf("databases.%s.sslmode", module),
		}

		for envKey, viperKey := range envMappings {
			if value := viper.GetString(envKey); value != "" {
				viper.Set(viperKey, value)
			}
		}
	}

	// Also handle generic app environment variables
	appEnvMappings := map[string]string{
		"GIN_MODE":    "app.gin_mode",
		"PORT":        "app.port",
		"APP_VERSION": "app.version",
	}

	for envKey, viperKey := range appEnvMappings {
		if value := viper.GetString(envKey); value != "" {
			viper.Set(viperKey, value)
		}
	}
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	// Validate app config
	if config.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if config.App.Port == "" {
		return fmt.Errorf("app port is required")
	}

	// Validate database configs
	for name, dbConfig := range config.Databases {
		if dbConfig.Host == "" {
			return fmt.Errorf("database %s host is required", name)
		}
		if dbConfig.Port == "" {
			return fmt.Errorf("database %s port is required", name)
		}
		if dbConfig.User == "" {
			return fmt.Errorf("database %s user is required", name)
		}
		if dbConfig.Name == "" {
			return fmt.Errorf("database %s name is required", name)
		}
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN(module string) string {
	db, exists := c.Databases[module]
	if !exists {
		return ""
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode)
}

// GetAvailableDatabases returns list of configured databases
func (c *Config) GetAvailableDatabases() []string {
	var databases []string
	for name := range c.Databases {
		databases = append(databases, name)
	}
	return databases
}

// LoadModulesConfig loads modules configuration from both module-level and central configs
func LoadModulesConfig() (*ModulesConfig, error) {
	// Try to load with module-level support first
	if config, err := LoadModulesConfigWithModuleLevelSupport(); err == nil {
		return config, nil
	}

	// Fallback to original central-only loading
	v := viper.New()
	v.SetConfigName("modules")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Enable environment variable support for modules config
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read modules config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading modules config file: %w", err)
	}

	var modulesConfig ModulesConfig
	if err := v.Unmarshal(&modulesConfig); err != nil {
		return nil, fmt.Errorf("error unmarshaling modules config: %w", err)
	}

	log.Printf("📦 Loaded modules configuration: %d modules defined", len(modulesConfig.Modules))
	return &modulesConfig, nil
}

// createDefaultModulesConfig creates a default modules configuration
// This is used as fallback when modules.yaml cannot be loaded
func createDefaultModulesConfig() *ModulesConfig {
	log.Println("⚠️ Creating fallback modules configuration (modules.yaml not available)")

	// Try to load from modules.yaml first, even in fallback mode
	if config, err := loadModulesConfigWithoutEnv(); err == nil {
		log.Println("✅ Successfully loaded modules.yaml as fallback")
		return config
	}

	// If modules.yaml is completely unavailable, create minimal config
	// This is the absolute last resort fallback with only essential modules
	log.Println("⚠️ modules.yaml unavailable, using minimal emergency fallback config")
	return createEmergencyFallbackConfig()
}

// createEmergencyFallbackConfig creates an empty configuration
// This should only be used when modules.yaml is completely unavailable
func createEmergencyFallbackConfig() *ModulesConfig {
	log.Println("📦 Creating empty modules configuration - no modules will be loaded")

	return &ModulesConfig{
		Modules: make(map[string]ModuleConfig), // Empty modules map
		Global: GlobalConfig{
			Database: DatabaseGlobalConfig{
				DefaultMaxOpenConns:    25,
				DefaultMaxIdleConns:    5,
				DefaultConnMaxLifetime: "5m",
				HealthCheckInterval:    "30s",
				ConnectionTimeout:      "10s",
				DatabasePrefix:         "modular_monolith", // Default prefix
			},
		},
	}
}

// loadModulesConfigWithoutEnv loads modules config without environment variable expansion
// This is used for fallback scenarios where we just want the basic structure
func loadModulesConfigWithoutEnv() (*ModulesConfig, error) {
	v := viper.New()
	v.SetConfigName("modules")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Don't enable environment variable support for fallback mode
	// This prevents issues when env vars are not available

	// Read modules config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading modules config file: %w", err)
	}

	var modulesConfig ModulesConfig
	if err := v.Unmarshal(&modulesConfig); err != nil {
		return nil, fmt.Errorf("error unmarshaling modules config: %w", err)
	}

	log.Printf("📦 Loaded fallback modules configuration: %d modules defined", len(modulesConfig.Modules))
	return &modulesConfig, nil
}

// getAvailableModuleNames returns module names from modules.yaml if available
func getAvailableModuleNames() []string {
	if config, err := loadModulesConfigWithoutEnv(); err == nil {
		var names []string
		for name := range config.Modules {
			names = append(names, name)
		}
		return names
	}
	return []string{} // Return empty slice if config not available
}

// getDatabasePrefix returns database prefix from modules config or default
func getDatabasePrefix() string {
	if config, err := loadModulesConfigWithoutEnv(); err == nil {
		return config.Global.Database.GetDatabasePrefix()
	}
	return "modular_monolith" // Default fallback
}

// loadFromVault loads secrets from HashiCorp Vault
func loadFromVault(modulesConfig *ModulesConfig) error {
	vaultClient, err := NewVaultClient()
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	if !vaultClient.IsEnabled() {
		return nil // Vault is disabled, skip loading
	}

	return vaultClient.LoadSecrets(modulesConfig)
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return ":" + c.App.Port
}

// convertModulesConfigToDatabaseConfig converts modules configuration to database configuration
func convertModulesConfigToDatabaseConfig(config *Config, modulesConfig *ModulesConfig) error {
	if config.Databases == nil {
		config.Databases = make(map[string]DatabaseConfig)
	}

	for moduleName, moduleConfig := range modulesConfig.Modules {
		if moduleConfig.Enabled {
			// Convert ModuleDatabaseConfig to DatabaseConfig
			dbConfig := DatabaseConfig{
				Host:     moduleConfig.Database.Host,
				Port:     moduleConfig.Database.Port,
				User:     moduleConfig.Database.User,
				Password: moduleConfig.Database.Password,
				Name:     moduleConfig.Database.Name,
				SSLMode:  moduleConfig.Database.SSLMode,
			}

			// Set defaults if empty
			if dbConfig.Host == "" {
				dbConfig.Host = "postgres"
			}
			if dbConfig.Port == "" {
				dbConfig.Port = "5432"
			}
			if dbConfig.User == "" {
				dbConfig.User = "postgres"
			}
			if dbConfig.Password == "" {
				dbConfig.Password = "postgres"
			}
			if dbConfig.Name == "" {
				dbConfig.Name = fmt.Sprintf("modular_monolith_%s", moduleName)
			}
			if dbConfig.SSLMode == "" {
				dbConfig.SSLMode = "disable"
			}

			config.Databases[moduleName] = dbConfig
			log.Printf("🔧 Converted database config for module: %s", moduleName)
		}
	}

	return nil
}
