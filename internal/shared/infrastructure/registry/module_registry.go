package registry

import (
	"fmt"
	"log"
	"sync"

	"golang_modular_monolith/internal/shared/infrastructure/config"
)

// ModuleRegistry manages the registration and lifecycle of modules
type ModuleRegistry struct {
	modules       map[string]*ModuleInfo
	modulesConfig *config.ModulesConfig
	mu            sync.RWMutex
}

// ModuleInfo contains information about a registered module
type ModuleInfo struct {
	Name    string
	Enabled bool
	Config  config.ModuleConfig
	Loaded  bool
	Error   error
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry(modulesConfig *config.ModulesConfig) *ModuleRegistry {
	return &ModuleRegistry{
		modules:       make(map[string]*ModuleInfo),
		modulesConfig: modulesConfig,
	}
}

// RegisterModule registers a module with the registry
func (mr *ModuleRegistry) RegisterModule(name string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Check if module exists in configuration
	moduleConfig, exists := mr.modulesConfig.Modules[name]
	if !exists {
		return fmt.Errorf("module %s not found in configuration", name)
	}

	// Create module info
	moduleInfo := &ModuleInfo{
		Name:    name,
		Enabled: moduleConfig.Enabled,
		Config:  moduleConfig,
		Loaded:  false,
	}

	mr.modules[name] = moduleInfo
	log.Printf("📦 Module registered: %s (enabled: %v)", name, moduleConfig.Enabled)

	return nil
}

// GetModule returns module information
func (mr *ModuleRegistry) GetModule(name string) (*ModuleInfo, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not registered", name)
	}

	return module, nil
}

// GetEnabledModules returns all enabled modules
func (mr *ModuleRegistry) GetEnabledModules() []*ModuleInfo {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var enabled []*ModuleInfo
	for _, module := range mr.modules {
		if module.Enabled {
			enabled = append(enabled, module)
		}
	}

	return enabled
}

// GetAllModules returns all registered modules
func (mr *ModuleRegistry) GetAllModules() []*ModuleInfo {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var all []*ModuleInfo
	for _, module := range mr.modules {
		all = append(all, module)
	}

	return all
}

// GetEnabledModuleNames returns names of enabled modules
func (mr *ModuleRegistry) GetEnabledModuleNames() []string {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var names []string
	for _, module := range mr.modules {
		if module.Enabled {
			names = append(names, module.Name)
		}
	}

	return names
}

// GetAllModuleNames returns names of all registered modules
func (mr *ModuleRegistry) GetAllModuleNames() []string {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var names []string
	for name := range mr.modules {
		names = append(names, name)
	}

	return names
}

// IsModuleEnabled checks if a module is enabled
func (mr *ModuleRegistry) IsModuleEnabled(name string) bool {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return false
	}

	return module.Enabled
}

// IsModuleLoaded checks if a module is loaded
func (mr *ModuleRegistry) IsModuleLoaded(name string) bool {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return false
	}

	return module.Loaded
}

// MarkModuleLoaded marks a module as loaded
func (mr *ModuleRegistry) MarkModuleLoaded(name string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	module, exists := mr.modules[name]
	if !exists {
		return fmt.Errorf("module %s not registered", name)
	}

	module.Loaded = true
	log.Printf("✅ Module loaded: %s", name)

	return nil
}

// MarkModuleError marks a module as having an error
func (mr *ModuleRegistry) MarkModuleError(name string, err error) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	module, exists := mr.modules[name]
	if !exists {
		return fmt.Errorf("module %s not registered", name)
	}

	module.Error = err
	log.Printf("❌ Module error: %s - %v", name, err)

	return nil
}

// GetModuleDatabaseConfig returns database configuration for a module
func (mr *ModuleRegistry) GetModuleDatabaseConfig(name string) (*config.ModuleDatabaseConfig, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not registered", name)
	}

	return &module.Config.Database, nil
}

// GetModuleMigrationPath returns migration path for a module
func (mr *ModuleRegistry) GetModuleMigrationPath(name string) (string, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return "", fmt.Errorf("module %s not registered", name)
	}

	return module.Config.Migration.Path, nil
}

// GetModuleVaultPath returns Vault path for a module
func (mr *ModuleRegistry) GetModuleVaultPath(name string) (string, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return "", fmt.Errorf("module %s not registered", name)
	}

	return module.Config.Vault.Path, nil
}

// GetModuleHTTPPrefix returns HTTP prefix for a module
func (mr *ModuleRegistry) GetModuleHTTPPrefix(name string) (string, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return "", fmt.Errorf("module %s not registered", name)
	}

	return module.Config.HTTP.Prefix, nil
}

// IsModuleHTTPEnabled checks if HTTP is enabled for a module
func (mr *ModuleRegistry) IsModuleHTTPEnabled(name string) bool {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return false
	}

	return module.Config.HTTP.Enabled
}

// IsModuleMigrationEnabled checks if migration is enabled for a module
func (mr *ModuleRegistry) IsModuleMigrationEnabled(name string) bool {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return false
	}

	return module.Config.Migration.Enabled
}

// IsModuleVaultEnabled checks if Vault is enabled for a module
func (mr *ModuleRegistry) IsModuleVaultEnabled(name string) bool {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	module, exists := mr.modules[name]
	if !exists {
		return false
	}

	return module.Config.Vault.Enabled
}

// GetModulesConfig returns the modules configuration
func (mr *ModuleRegistry) GetModulesConfig() *config.ModulesConfig {
	return mr.modulesConfig
}

// PrintModuleStatus prints the status of all modules
func (mr *ModuleRegistry) PrintModuleStatus() {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	log.Println("📋 Module Status:")
	for name, module := range mr.modules {
		status := "❌ Disabled"
		if module.Enabled {
			if module.Loaded {
				status = "✅ Loaded"
			} else if module.Error != nil {
				status = fmt.Sprintf("⚠️ Error: %v", module.Error)
			} else {
				status = "🔄 Enabled (not loaded)"
			}
		}
		log.Printf("  - %s: %s", name, status)
	}
}

// InitializeModules registers all modules from configuration
func (mr *ModuleRegistry) InitializeModules() error {
	log.Println("🚀 Initializing modules from configuration...")

	for moduleName := range mr.modulesConfig.Modules {
		if err := mr.RegisterModule(moduleName); err != nil {
			return fmt.Errorf("failed to register module %s: %w", moduleName, err)
		}
	}

	mr.PrintModuleStatus()
	return nil
}
