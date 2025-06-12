package registry

import (
	"fmt"
	"log"

	"golang_modular_monolith/internal/shared/domain"
	"golang_modular_monolith/internal/shared/infrastructure/config"
)

// ModuleCreator is a function that creates a module
type ModuleCreator func() domain.Module

// ModuleManager manages module factory and loading
type ModuleManager struct {
	registry *domain.ModuleRegistry
	creators map[string]ModuleCreator
}

// NewModuleManager creates a new module manager
func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		registry: domain.NewModuleRegistry(),
		creators: make(map[string]ModuleCreator),
	}
}

// RegisterModule registers a module creator
func (m *ModuleManager) RegisterModule(name string, creator ModuleCreator) {
	m.creators[name] = creator
	log.Printf("📦 Registered module creator: %s", name)
}

// CreateModule creates a module by name
func (m *ModuleManager) CreateModule(name string) (domain.Module, error) {
	creator, exists := m.creators[name]
	if !exists {
		return nil, fmt.Errorf("unknown module: %s", name)
	}

	module := creator()
	log.Printf("🏗️ Created module: %s", name)
	return module, nil
}

// GetAvailableModules returns list of available modules
func (m *ModuleManager) GetAvailableModules() []string {
	modules := make([]string, 0, len(m.creators))
	for name := range m.creators {
		modules = append(modules, name)
	}
	return modules
}

// HasModule checks if a module is available
func (m *ModuleManager) HasModule(name string) bool {
	_, exists := m.creators[name]
	return exists
}

// LoadEnabledModules loads all enabled modules from configuration
func (m *ModuleManager) LoadEnabledModules(cfg *config.Config) error {
	log.Println("🔧 Loading enabled modules...")

	if cfg.Modules == nil {
		log.Println("⚠️ No modules configuration found")
		return nil
	}

	// Get all available modules
	availableModules := m.GetAvailableModules()
	log.Printf("📋 Available modules: %v", availableModules)

	// Load each enabled module
	for _, moduleName := range availableModules {
		if m.isModuleEnabled(cfg, moduleName) {
			log.Printf("📦 Loading %s module...", moduleName)

			// Create module
			module, err := m.CreateModule(moduleName)
			if err != nil {
				log.Printf("❌ Failed to create %s module: %v", moduleName, err)
				continue
			}

			// Register module
			m.registry.Register(module)
			log.Printf("✅ %s module registered", moduleName)
		} else {
			log.Printf("⏭️ %s module disabled in config", moduleName)
		}
	}

	loadedModules := m.registry.GetModuleNames()
	log.Printf("✅ Loaded modules: %v", loadedModules)

	return nil
}

// GetRegistry returns the module registry
func (m *ModuleManager) GetRegistry() *domain.ModuleRegistry {
	return m.registry
}

// isModuleEnabled checks if a module is enabled in configuration
func (m *ModuleManager) isModuleEnabled(cfg *config.Config, moduleName string) bool {
	if cfg.Modules == nil {
		return false
	}

	return cfg.Modules.IsModuleEnabled(moduleName)
}

// Global manager instance
var globalManager = NewModuleManager()

// RegisterModule registers a module creator globally
func RegisterModule(name string, creator ModuleCreator) {
	globalManager.RegisterModule(name, creator)
}

// GetGlobalManager returns the global module manager
func GetGlobalManager() *ModuleManager {
	return globalManager
}
