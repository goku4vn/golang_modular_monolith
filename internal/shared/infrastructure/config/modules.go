package config

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ModulesConfig represents the complete module configuration
type ModulesConfig struct {
	Modules map[string]ModuleConfig `yaml:"modules" mapstructure:"modules"`
	Global  GlobalConfig            `yaml:"global" mapstructure:"global"`
}

// FlexibleModulesConfig represents flexible module configuration that supports both simple and complex formats
type FlexibleModulesConfig struct {
	Modules interface{}  `yaml:"modules" mapstructure:"modules"`
	Global  GlobalConfig `yaml:"global" mapstructure:"global"`
}

// ModuleConfig represents configuration for a single module
type ModuleConfig struct {
	Enabled   bool                 `yaml:"enabled" mapstructure:"enabled"`
	Database  ModuleDatabaseConfig `yaml:"database" mapstructure:"database"`
	Migration MigrationConfig      `yaml:"migration" mapstructure:"migration"`
	Vault     ModuleVaultConfig    `yaml:"vault" mapstructure:"vault"`
	HTTP      HTTPConfig           `yaml:"http" mapstructure:"http"`
	Features  FeatureConfig        `yaml:"features" mapstructure:"features"`
	// Module-specific metadata
	Module ModuleMetadata `yaml:"module" mapstructure:"module"`
	// Custom module-specific settings (stored as map for flexibility)
	Custom map[string]interface{} `yaml:",inline" mapstructure:",remain"`
}

// ModuleMetadata represents metadata about a module
type ModuleMetadata struct {
	Name        string `yaml:"name" mapstructure:"name"`
	Version     string `yaml:"version" mapstructure:"version"`
	Description string `yaml:"description" mapstructure:"description"`
}

// ModuleDatabaseConfig represents database configuration for a module
type ModuleDatabaseConfig struct {
	Host            string `yaml:"host" mapstructure:"host"`
	Port            string `yaml:"port" mapstructure:"port"`
	User            string `yaml:"user" mapstructure:"user"`
	Password        string `yaml:"password" mapstructure:"password"`
	Name            string `yaml:"name" mapstructure:"name"`
	SSLMode         string `yaml:"sslmode" mapstructure:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
}

// MigrationConfig represents migration configuration for a module
type MigrationConfig struct {
	Path    string `yaml:"path" mapstructure:"path"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

// ModuleVaultConfig represents Vault configuration for a module
type ModuleVaultConfig struct {
	Path    string `yaml:"path" mapstructure:"path"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

// HTTPConfig represents HTTP configuration for a module
type HTTPConfig struct {
	Prefix     string   `yaml:"prefix" mapstructure:"prefix"`
	Enabled    bool     `yaml:"enabled" mapstructure:"enabled"`
	Middleware []string `yaml:"middleware" mapstructure:"middleware"`
}

// FeatureConfig represents feature flags for a module
type FeatureConfig struct {
	EventsEnabled  bool `yaml:"events_enabled" mapstructure:"events_enabled"`
	CachingEnabled bool `yaml:"caching_enabled" mapstructure:"caching_enabled"`
}

// GlobalConfig represents global configuration settings
type GlobalConfig struct {
	Database DatabaseGlobalConfig `yaml:"database" mapstructure:"database"`
	Vault    VaultGlobalConfig    `yaml:"vault" mapstructure:"vault"`
	HTTP     HTTPGlobalConfig     `yaml:"http" mapstructure:"http"`
	Features FeatureGlobalConfig  `yaml:"features" mapstructure:"features"`
}

// DatabaseGlobalConfig represents global database settings
type DatabaseGlobalConfig struct {
	DefaultMaxOpenConns    int    `yaml:"default_max_open_conns" mapstructure:"default_max_open_conns"`
	DefaultMaxIdleConns    int    `yaml:"default_max_idle_conns" mapstructure:"default_max_idle_conns"`
	DefaultConnMaxLifetime string `yaml:"default_conn_max_lifetime" mapstructure:"default_conn_max_lifetime"`
	HealthCheckInterval    string `yaml:"health_check_interval" mapstructure:"health_check_interval"`
	ConnectionTimeout      string `yaml:"connection_timeout" mapstructure:"connection_timeout"`
	DatabasePrefix         string `yaml:"database_prefix" mapstructure:"database_prefix"`
}

// VaultGlobalConfig represents global Vault settings
type VaultGlobalConfig struct {
	MountPath  string `yaml:"mount_path" mapstructure:"mount_path"`
	SecretPath string `yaml:"secret_path" mapstructure:"secret_path"`
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
}

// HTTPGlobalConfig represents global HTTP settings
type HTTPGlobalConfig struct {
	DefaultMiddleware []string        `yaml:"default_middleware" mapstructure:"default_middleware"`
	RateLimiting      RateLimitConfig `yaml:"rate_limiting" mapstructure:"rate_limiting"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled" mapstructure:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute" mapstructure:"requests_per_minute"`
}

// FeatureGlobalConfig represents global feature flags
type FeatureGlobalConfig struct {
	EventsEnabled  bool `yaml:"events_enabled" mapstructure:"events_enabled"`
	MetricsEnabled bool `yaml:"metrics_enabled" mapstructure:"metrics_enabled"`
	TracingEnabled bool `yaml:"tracing_enabled" mapstructure:"tracing_enabled"`
}

// LoadModulesConfigWithModuleLevelSupport loads module configurations from both module-level and central configs
func LoadModulesConfigWithModuleLevelSupport() (*ModulesConfig, error) {
	// 1. Load module-level configs first (as defaults)
	moduleConfigs, err := loadModuleLevelConfigs()
	if err != nil {
		log.Printf("⚠️ Failed to load module-level configs: %v", err)
		moduleConfigs = make(map[string]ModuleConfig)
	}

	// 2. Try to load central config with flexible format first
	centralConfigWithDisabled, err := loadCentralModulesConfigFlexible()
	if err != nil {
		log.Printf("⚠️ Failed to load central modules config: %v", err)
		// If no central config, use only module configs
		return &ModulesConfig{
			Modules: moduleConfigs,
			Global:  getDefaultGlobalConfig(),
		}, nil
	}

	// 3. Merge configs (central overrides module)
	finalConfig := mergeModuleConfigsWithDisabled(moduleConfigs, centralConfigWithDisabled)

	log.Printf("📦 Loaded configuration for %d modules: %v",
		len(finalConfig.Modules), finalConfig.GetModuleNames())

	return finalConfig, nil
}

// loadModuleLevelConfigs scans for module.yaml files in module directories
func loadModuleLevelConfigs() (map[string]ModuleConfig, error) {
	configs := make(map[string]ModuleConfig)

	// Scan internal/modules directory
	modulesDir := "internal/modules"
	if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
		return configs, nil // No modules directory
	}

	err := filepath.WalkDir(modulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Look for module.yaml files
		if d.Name() == "module.yaml" && !d.IsDir() {
			// Extract module name from path
			moduleName := extractModuleNameFromPath(path)
			if moduleName == "" {
				return nil
			}

			// Load module config
			config, err := loadSingleModuleConfig(path)
			if err != nil {
				log.Printf("⚠️ Failed to load config for module %s: %v", moduleName, err)
				return nil // Continue with other modules
			}

			// Set module name if not specified
			if config.Module.Name == "" {
				config.Module.Name = moduleName
			}

			configs[moduleName] = *config
			log.Printf("📦 Loaded module config: %s (v%s)", moduleName, config.Module.Version)
		}

		return nil
	})

	return configs, err
}

// extractModuleNameFromPath extracts module name from file path
func extractModuleNameFromPath(path string) string {
	// path format: internal/modules/{module_name}/module.yaml
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) >= 3 && parts[0] == "internal" && parts[1] == "modules" {
		return parts[2]
	}
	return ""
}

// loadSingleModuleConfig loads configuration from a single module.yaml file
func loadSingleModuleConfig(configPath string) (*ModuleConfig, error) {
	// Create a new viper instance for this module
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Enable environment variable substitution
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Expand environment variables in the config
	expandedConfig := make(map[string]interface{})
	if err := v.Unmarshal(&expandedConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Convert back to YAML and parse again to handle env var expansion
	yamlData, err := yaml.Marshal(expandedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal expanded config: %w", err)
	}

	// Expand environment variables in YAML content
	expandedYaml := os.ExpandEnv(string(yamlData))

	// Parse the final config
	var config ModuleConfig
	if err := yaml.Unmarshal([]byte(expandedYaml), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal final config: %w", err)
	}

	return &config, nil
}

// loadCentralModulesConfigFlexible loads central config with support for flexible module format
func loadCentralModulesConfigFlexible() (*ModulesConfigWithDisabled, error) {
	v := viper.New()
	v.SetConfigName("modules")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Enable environment variable support
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read modules config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading modules config file: %w", err)
	}

	// First try to unmarshal as flexible config
	var flexConfig FlexibleModulesConfig
	if err := v.Unmarshal(&flexConfig); err != nil {
		return nil, fmt.Errorf("error unmarshaling flexible modules config: %w", err)
	}

	// Process flexible modules format
	return processFlexibleModulesConfig(&flexConfig)
}

// ModulesConfigWithDisabled extends ModulesConfig to track disabled modules
type ModulesConfigWithDisabled struct {
	*ModulesConfig
	DisabledModules map[string]bool
}

// processFlexibleModulesConfig converts flexible format to standard ModulesConfig
func processFlexibleModulesConfig(flexConfig *FlexibleModulesConfig) (*ModulesConfigWithDisabled, error) {
	result := &ModulesConfigWithDisabled{
		ModulesConfig: &ModulesConfig{
			Modules: make(map[string]ModuleConfig),
			Global:  flexConfig.Global,
		},
		DisabledModules: make(map[string]bool),
	}

	// Handle different module formats
	switch modules := flexConfig.Modules.(type) {
	case map[string]interface{}:
		// Object format: { customer: true, order: { enabled: true, ... } }
		for name, value := range modules {
			moduleConfig, isDisabled, err := processModuleValue(name, value)
			if err != nil {
				log.Printf("⚠️ Failed to process module %s: %v", name, err)
				continue
			}
			if isDisabled {
				// Track explicitly disabled modules
				result.DisabledModules[name] = true
			} else if moduleConfig != nil {
				result.ModulesConfig.Modules[name] = *moduleConfig
			}
		}
	case []interface{}:
		// Array format: [customer, order]
		for _, item := range modules {
			if name, ok := item.(string); ok {
				moduleConfig, err := loadModuleLevelConfigByName(name)
				if err != nil {
					log.Printf("⚠️ Failed to load module-level config for %s: %v", name, err)
					continue
				}
				if moduleConfig != nil {
					moduleConfig.Enabled = true
					result.ModulesConfig.Modules[name] = *moduleConfig
				}
			}
		}
	case nil:
		// No modules defined
		log.Println("📦 No modules defined in central config")
	default:
		return nil, fmt.Errorf("unsupported modules format: %T", modules)
	}

	return result, nil
}

// processModuleValue processes a single module value (bool, string, or object)
// Returns (config, isExplicitlyDisabled, error)
func processModuleValue(name string, value interface{}) (*ModuleConfig, bool, error) {
	switch v := value.(type) {
	case bool:
		if !v {
			// Module explicitly disabled
			return nil, true, nil
		}
		// Module enabled - load from module-level config and force enable
		config, err := loadModuleLevelConfigByName(name)
		if err != nil {
			return nil, false, err
		}
		if config != nil {
			config.Enabled = true // Force enable when explicitly set to true in central config
		}
		return config, false, nil

	case string:
		if v == "true" || v == "enabled" {
			// Module enabled - load from module-level config and force enable
			config, err := loadModuleLevelConfigByName(name)
			if err != nil {
				return nil, false, err
			}
			if config != nil {
				config.Enabled = true // Force enable when explicitly set to enabled in central config
			}
			return config, false, nil
		}
		// Module disabled or invalid value
		return nil, true, nil

	case map[string]interface{}:
		// Complex object - parse as ModuleConfig
		config, err := parseComplexModuleConfig(name, v)
		return config, false, err

	default:
		return nil, false, fmt.Errorf("unsupported module value type: %T", v)
	}
}

// loadModuleLevelConfigByName loads module-level config for a specific module
func loadModuleLevelConfigByName(moduleName string) (*ModuleConfig, error) {
	configPath := fmt.Sprintf("internal/modules/%s/module.yaml", moduleName)

	// Check if module config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if module.yaml doesn't exist
		log.Printf("📦 Creating default config for module: %s (no module.yaml found)", moduleName)
		return createDefaultModuleConfig(moduleName), nil
	}

	return loadSingleModuleConfig(configPath)
}

// parseComplexModuleConfig parses complex module configuration object
func parseComplexModuleConfig(name string, configMap map[string]interface{}) (*ModuleConfig, error) {
	// First load module-level config as base
	baseConfig, err := loadModuleLevelConfigByName(name)
	if err != nil {
		log.Printf("⚠️ Failed to load module-level config for %s, using defaults: %v", name, err)
		baseConfig = createDefaultModuleConfig(name)
	}

	// Convert map to YAML and then unmarshal to ModuleConfig
	yamlData, err := yaml.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config map: %w", err)
	}

	var overrideConfig ModuleConfig
	if err := yaml.Unmarshal(yamlData, &overrideConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal override config: %w", err)
	}

	// Merge base config with override
	merged := mergeModuleConfig(*baseConfig, overrideConfig)
	return &merged, nil
}

// createDefaultModuleConfig creates a default module configuration
func createDefaultModuleConfig(moduleName string) *ModuleConfig {
	return &ModuleConfig{
		Enabled: true,
		Database: ModuleDatabaseConfig{
			Host:            "postgres",
			Port:            "5432",
			User:            "postgres",
			Password:        "postgres",
			Name:            fmt.Sprintf("modular_monolith_%s", moduleName),
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: "5m",
		},
		Migration: MigrationConfig{
			Path:    fmt.Sprintf("internal/modules/%s/migrations", moduleName),
			Enabled: true,
		},
		Vault: ModuleVaultConfig{
			Path:    fmt.Sprintf("modules/%s", moduleName),
			Enabled: false,
		},
		HTTP: HTTPConfig{
			Prefix:  fmt.Sprintf("/api/v1/%ss", moduleName),
			Enabled: true,
		},
		Features: FeatureConfig{
			EventsEnabled:  true,
			CachingEnabled: false,
		},
		Module: ModuleMetadata{
			Name:        moduleName,
			Version:     "1.0.0",
			Description: fmt.Sprintf("%s module", strings.Title(moduleName)),
		},
	}
}

// mergeModuleConfigs merges module-level configs with central config
func mergeModuleConfigs(moduleConfigs map[string]ModuleConfig, centralConfig *ModulesConfig) *ModulesConfig {
	result := &ModulesConfig{
		Modules: make(map[string]ModuleConfig),
		Global:  centralConfig.Global,
	}

	// First, process central config to determine which modules should be included
	centralModuleNames := make(map[string]bool)
	for name := range centralConfig.Modules {
		centralModuleNames[name] = true
	}

	// Add modules from central config first (these take priority)
	for name, centralModuleConfig := range centralConfig.Modules {
		if moduleConfig, exists := moduleConfigs[name]; exists {
			// Merge module config with central config (central takes priority)
			merged := mergeModuleConfig(moduleConfig, centralModuleConfig)
			result.Modules[name] = merged
		} else {
			// Add new module from central config only
			result.Modules[name] = centralModuleConfig
		}
	}

	// Add module-level configs that are NOT mentioned in central config
	// (This allows module-level configs to work independently when not overridden)
	for name, config := range moduleConfigs {
		if !centralModuleNames[name] {
			// Module not mentioned in central config, use module-level config as-is
			result.Modules[name] = config
		}
	}

	// Note: Modules that are explicitly disabled in central config (user: false)
	// are not added to result.Modules because processFlexibleModulesConfig()
	// already filtered them out

	return result
}

// mergeModuleConfigsWithDisabled merges module-level configs with central config, respecting disabled modules
func mergeModuleConfigsWithDisabled(moduleConfigs map[string]ModuleConfig, centralConfigWithDisabled *ModulesConfigWithDisabled) *ModulesConfig {
	result := &ModulesConfig{
		Modules: make(map[string]ModuleConfig),
		Global:  centralConfigWithDisabled.ModulesConfig.Global,
	}

	// First, add modules from central config (these take priority)
	for name, centralModuleConfig := range centralConfigWithDisabled.ModulesConfig.Modules {
		if moduleConfig, exists := moduleConfigs[name]; exists {
			// Merge module config with central config (central takes priority)
			merged := mergeModuleConfig(moduleConfig, centralModuleConfig)
			result.Modules[name] = merged
		} else {
			// Add new module from central config only
			result.Modules[name] = centralModuleConfig
		}
	}

	// Add module-level configs that are NOT mentioned in central config AND NOT explicitly disabled
	for name, config := range moduleConfigs {
		// Skip if module is in central config (already processed above)
		if _, inCentral := centralConfigWithDisabled.ModulesConfig.Modules[name]; inCentral {
			continue
		}
		// Skip if module is explicitly disabled in central config
		if centralConfigWithDisabled.DisabledModules[name] {
			log.Printf("🚫 Module %s explicitly disabled in central config", name)
			continue
		}
		// Module not mentioned in central config and not disabled, use module-level config as-is
		result.Modules[name] = config
	}

	return result
}

// mergeModuleConfig merges two module configs (second takes priority)
func mergeModuleConfig(base, override ModuleConfig) ModuleConfig {
	// Simple merge - override takes priority for non-zero values
	// This is a basic implementation, can be enhanced for more sophisticated merging

	result := base // Start with base

	// Override enabled status
	if override.Enabled != base.Enabled {
		result.Enabled = override.Enabled
	}

	// Merge database config
	if override.Database.Host != "" {
		result.Database.Host = override.Database.Host
	}
	if override.Database.Port != "" {
		result.Database.Port = override.Database.Port
	}
	if override.Database.User != "" {
		result.Database.User = override.Database.User
	}
	if override.Database.Password != "" {
		result.Database.Password = override.Database.Password
	}
	if override.Database.Name != "" {
		result.Database.Name = override.Database.Name
	}
	if override.Database.SSLMode != "" {
		result.Database.SSLMode = override.Database.SSLMode
	}
	if override.Database.MaxOpenConns != 0 {
		result.Database.MaxOpenConns = override.Database.MaxOpenConns
	}
	if override.Database.MaxIdleConns != 0 {
		result.Database.MaxIdleConns = override.Database.MaxIdleConns
	}
	if override.Database.ConnMaxLifetime != "" {
		result.Database.ConnMaxLifetime = override.Database.ConnMaxLifetime
	}

	// Merge other configs similarly...
	if override.Migration.Path != "" {
		result.Migration.Path = override.Migration.Path
	}
	if override.Migration.Enabled != base.Migration.Enabled {
		result.Migration.Enabled = override.Migration.Enabled
	}

	if override.Vault.Path != "" {
		result.Vault.Path = override.Vault.Path
	}
	if override.Vault.Enabled != base.Vault.Enabled {
		result.Vault.Enabled = override.Vault.Enabled
	}

	if override.HTTP.Prefix != "" {
		result.HTTP.Prefix = override.HTTP.Prefix
	}
	if override.HTTP.Enabled != base.HTTP.Enabled {
		result.HTTP.Enabled = override.HTTP.Enabled
	}
	if len(override.HTTP.Middleware) > 0 {
		result.HTTP.Middleware = override.HTTP.Middleware
	}

	// Merge features
	if override.Features.EventsEnabled != base.Features.EventsEnabled {
		result.Features.EventsEnabled = override.Features.EventsEnabled
	}
	if override.Features.CachingEnabled != base.Features.CachingEnabled {
		result.Features.CachingEnabled = override.Features.CachingEnabled
	}

	// Merge metadata
	if override.Module.Name != "" {
		result.Module.Name = override.Module.Name
	}
	if override.Module.Version != "" {
		result.Module.Version = override.Module.Version
	}
	if override.Module.Description != "" {
		result.Module.Description = override.Module.Description
	}

	// Merge custom fields
	if len(override.Custom) > 0 {
		if result.Custom == nil {
			result.Custom = make(map[string]interface{})
		}
		for k, v := range override.Custom {
			result.Custom[k] = v
		}
	}

	return result
}

// getDefaultGlobalConfig returns default global configuration
func getDefaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Database: DatabaseGlobalConfig{
			DefaultMaxOpenConns:    25,
			DefaultMaxIdleConns:    5,
			DefaultConnMaxLifetime: "5m",
			HealthCheckInterval:    "30s",
			ConnectionTimeout:      "10s",
			DatabasePrefix:         "modular_monolith",
		},
		Vault: VaultGlobalConfig{
			MountPath:  "secret",
			SecretPath: "modular-monolith",
			Enabled:    false,
		},
		HTTP: HTTPGlobalConfig{
			DefaultMiddleware: []string{"cors", "logging", "recovery"},
			RateLimiting: RateLimitConfig{
				Enabled:           false,
				RequestsPerMinute: 100,
			},
		},
		Features: FeatureGlobalConfig{
			EventsEnabled:  true,
			MetricsEnabled: true,
			TracingEnabled: false,
		},
	}
}

// GetEnabledModules returns a list of enabled module names
func (mc *ModulesConfig) GetEnabledModules() []string {
	var enabled []string
	for name, config := range mc.Modules {
		if config.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// GetModuleNames returns all module names (enabled and disabled)
func (mc *ModulesConfig) GetModuleNames() []string {
	var names []string
	for name := range mc.Modules {
		names = append(names, name)
	}
	return names
}

// GetModuleDatabaseConfig returns database configuration for a specific module
func (mc *ModulesConfig) GetModuleDatabaseConfig(moduleName string) (*ModuleDatabaseConfig, error) {
	module, exists := mc.Modules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module %s not found", moduleName)
	}
	return &module.Database, nil
}

// GetModuleMigrationPath returns migration path for a specific module
func (mc *ModulesConfig) GetModuleMigrationPath(moduleName string) (string, error) {
	module, exists := mc.Modules[moduleName]
	if !exists {
		return "", fmt.Errorf("module %s not found", moduleName)
	}
	return module.Migration.Path, nil
}

// GetModuleVaultPath returns Vault path for a specific module
func (mc *ModulesConfig) GetModuleVaultPath(moduleName string) (string, error) {
	module, exists := mc.Modules[moduleName]
	if !exists {
		return "", fmt.Errorf("module %s not found", moduleName)
	}
	return module.Vault.Path, nil
}

// IsModuleEnabled checks if a module is enabled
func (mc *ModulesConfig) IsModuleEnabled(moduleName string) bool {
	module, exists := mc.Modules[moduleName]
	if !exists {
		return false
	}
	return module.Enabled
}

// GetConnMaxLifetimeDuration parses and returns connection max lifetime as duration
func (dc *ModuleDatabaseConfig) GetConnMaxLifetimeDuration() (time.Duration, error) {
	if dc.ConnMaxLifetime == "" {
		return 5 * time.Minute, nil // default
	}
	return time.ParseDuration(dc.ConnMaxLifetime)
}

// GetHealthCheckIntervalDuration parses and returns health check interval as duration
func (dgc *DatabaseGlobalConfig) GetHealthCheckIntervalDuration() (time.Duration, error) {
	if dgc.HealthCheckInterval == "" {
		return 30 * time.Second, nil // default
	}
	return time.ParseDuration(dgc.HealthCheckInterval)
}

// GetConnectionTimeoutDuration parses and returns connection timeout as duration
func (dgc *DatabaseGlobalConfig) GetConnectionTimeoutDuration() (time.Duration, error) {
	if dgc.ConnectionTimeout == "" {
		return 10 * time.Second, nil // default
	}
	return time.ParseDuration(dgc.ConnectionTimeout)
}

// GetDatabasePrefix returns the database prefix, with default fallback
func (dgc *DatabaseGlobalConfig) GetDatabasePrefix() string {
	if dgc.DatabasePrefix == "" {
		return "modular_monolith" // default fallback
	}
	return dgc.DatabasePrefix
}
