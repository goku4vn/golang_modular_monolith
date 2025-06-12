package modules

// This file imports all available modules to trigger their auto-registration
// Add new modules here when they are created

import (
	// Import all modules to trigger auto-registration via init() functions
	_ "golang_modular_monolith/internal/modules/customer"
	_ "golang_modular_monolith/internal/modules/order"
	_ "golang_modular_monolith/internal/modules/user"
)

// InitializeAllModules is called to ensure all modules are imported and registered
// This function doesn't need to do anything - the imports above trigger init() functions
func InitializeAllModules() {
	// This function exists to ensure this package is imported
	// and all module init() functions are called
}
