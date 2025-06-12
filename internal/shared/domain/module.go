package domain

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Module represents a business module in the system
type Module interface {
	// Name returns the module name
	Name() string

	// Initialize initializes the module with dependencies
	Initialize(deps ModuleDependencies) error

	// RegisterRoutes registers HTTP routes for the module
	RegisterRoutes(router *gin.RouterGroup)

	// Health checks if the module is healthy
	Health(ctx context.Context) error

	// Start starts the module (optional lifecycle method)
	Start(ctx context.Context) error

	// Stop stops the module (optional lifecycle method)
	Stop(ctx context.Context) error
}

// ModuleDependencies contains shared dependencies for modules
type ModuleDependencies struct {
	EventBus EventBus
	Config   interface{} // Module-specific config
}

// ModuleRegistry manages module registration and lifecycle
type ModuleRegistry struct {
	modules map[string]Module
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]Module),
	}
}

// Register registers a module
func (r *ModuleRegistry) Register(module Module) {
	r.modules[module.Name()] = module
}

// GetModule returns a module by name
func (r *ModuleRegistry) GetModule(name string) (Module, bool) {
	module, exists := r.modules[name]
	return module, exists
}

// GetAllModules returns all registered modules
func (r *ModuleRegistry) GetAllModules() map[string]Module {
	return r.modules
}

// GetModuleNames returns all registered module names
func (r *ModuleRegistry) GetModuleNames() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

// InitializeAll initializes all registered modules
func (r *ModuleRegistry) InitializeAll(deps ModuleDependencies) error {
	for name, module := range r.modules {
		if err := module.Initialize(deps); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", name, err)
		}
	}
	return nil
}

// RegisterAllRoutes registers routes for all modules
func (r *ModuleRegistry) RegisterAllRoutes(router *gin.RouterGroup) {
	for _, module := range r.modules {
		module.RegisterRoutes(router)
	}
}

// StartAll starts all modules
func (r *ModuleRegistry) StartAll(ctx context.Context) error {
	for name, module := range r.modules {
		if err := module.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", name, err)
		}
	}
	return nil
}

// StopAll stops all modules
func (r *ModuleRegistry) StopAll(ctx context.Context) error {
	for name, module := range r.modules {
		if err := module.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop module %s: %w", name, err)
		}
	}
	return nil
}

// HealthCheckAll checks health of all modules
func (r *ModuleRegistry) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for name, module := range r.modules {
		results[name] = module.Health(ctx)
	}
	return results
}
