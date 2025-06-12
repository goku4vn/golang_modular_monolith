package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang_modular_monolith/internal/shared/infrastructure/config"
	"golang_modular_monolith/internal/shared/infrastructure/database"
	"golang_modular_monolith/internal/shared/infrastructure/migration"
)

func main() {
	var (
		module = flag.String("module", "", "Module name or 'all' for all enabled modules")
		action = flag.String("action", "up", "Migration action (up, down, version, reset, create)")
		name   = flag.String("name", "", "Migration name for create action")
	)
	flag.Parse()

	// Load configuration to get available modules
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get available modules from configuration
	availableModules := getAvailableModules(cfg)
	if len(availableModules) == 0 {
		log.Fatal("No enabled modules found in configuration")
	}

	if *module == "" {
		fmt.Println("Usage: go run cmd/migrate/main.go -module=<module> -action=<action> [options]")
		fmt.Printf("Available modules: %v, all\n", availableModules)
		fmt.Println("Actions: up, down, version, reset, create")
		fmt.Println("Options:")
		fmt.Println("  -version=<version>  Target version for migrate")
		fmt.Println("  -name=<name>        Migration name for create action")
		os.Exit(1)
	}

	// Validate module
	if *module != "all" && !isValidModule(*module, availableModules) {
		log.Fatalf("Invalid module: %s. Available modules: %v", *module, availableModules)
	}

	// Create migration manager
	migrationManager := migration.NewMigrationManager()
	defer migrationManager.Close()

	// Register modules based on input
	if err := registerModules(migrationManager, cfg, *module, availableModules); err != nil {
		log.Fatalf("Failed to register modules: %v", err)
	}

	// Execute action
	switch *action {
	case "up":
		if err := executeUp(migrationManager, *module); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		if err := executeDown(migrationManager, *module); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
	case "version":
		if err := executeVersion(migrationManager, *module); err != nil {
			log.Fatalf("Migration version failed: %v", err)
		}
	case "reset":
		if err := executeReset(migrationManager, *module); err != nil {
			log.Fatalf("Migration reset failed: %v", err)
		}
	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create action")
		}
		if err := executeCreate(cfg, *module, *name, availableModules); err != nil {
			log.Fatalf("Migration create failed: %v", err)
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}

	fmt.Println("Migration completed successfully!")
}

// getAvailableModules extracts enabled modules from configuration
func getAvailableModules(cfg *config.Config) []string {
	var modules []string

	// First try to get from modules config (preferred)
	if cfg.Modules != nil {
		for moduleName, moduleConfig := range cfg.Modules.Modules {
			if moduleConfig.Enabled {
				modules = append(modules, moduleName)
			}
		}
		if len(modules) > 0 {
			return modules
		}
	}

	// Fallback: Get modules from database config (legacy)
	for moduleName := range cfg.Databases {
		modules = append(modules, moduleName)
	}

	return modules
}

// isValidModule checks if the given module is in the available modules list
func isValidModule(module string, availableModules []string) bool {
	for _, available := range availableModules {
		if module == available {
			return true
		}
	}
	return false
}

func registerModules(migrationManager *migration.MigrationManager, cfg *config.Config, module string, availableModules []string) error {
	if module == "all" {
		// Register all available modules
		for _, moduleName := range availableModules {
			if err := registerModule(migrationManager, cfg, moduleName); err != nil {
				return fmt.Errorf("failed to register module %s: %w", moduleName, err)
			}
		}
		return nil
	}

	// Register specific module
	return registerModule(migrationManager, cfg, module)
}

func registerModule(migrationManager *migration.MigrationManager, cfg *config.Config, moduleName string) error {
	// Try to get database config from databases first (legacy)
	dbConfig, exists := cfg.Databases[moduleName]

	// If not found in databases, try to get from modules config
	if !exists && cfg.Modules != nil {
		if moduleConfig, moduleExists := cfg.Modules.Modules[moduleName]; moduleExists && moduleConfig.Enabled {
			// Convert ModuleDatabaseConfig to DatabaseConfig
			dbConfig = config.DatabaseConfig{
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

			exists = true
		}
	}

	if !exists {
		return fmt.Errorf("database configuration not found for module: %s", moduleName)
	}

	// Initialize database manager
	manager := database.GetGlobalManager()

	// Convert config.DatabaseConfig to database.DatabaseConfig
	databaseConfig := &database.DatabaseConfig{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Name:     dbConfig.Name,
		SSLMode:  dbConfig.SSLMode,
	}

	// Register database
	manager.RegisterDatabase(moduleName, databaseConfig)

	// Get database connection
	db, err := manager.GetConnection(moduleName)
	if err != nil {
		return fmt.Errorf("failed to connect to %s database: %w", moduleName, err)
	}

	// Determine migration path - try to get from modules config first
	migrationPath := fmt.Sprintf("internal/modules/%s/migrations", moduleName)
	if cfg.Modules != nil {
		if moduleConfig, moduleExists := cfg.Modules.Modules[moduleName]; moduleExists {
			if moduleConfig.Migration.Path != "" {
				migrationPath = moduleConfig.Migration.Path
			}
		}
	}

	log.Printf("📦 Registering migration for module: %s (path: %s)", moduleName, migrationPath)
	return migrationManager.RegisterModule(moduleName, db, migrationPath)
}

func executeUp(migrationManager *migration.MigrationManager, module string) error {
	if module == "all" {
		return migrationManager.MigrateAllUp()
	}
	return migrationManager.MigrateUp(module)
}

func executeDown(migrationManager *migration.MigrationManager, module string) error {
	if module == "all" {
		return migrationManager.MigrateAllDown()
	}
	return migrationManager.MigrateDown(module)
}

func executeVersion(migrationManager *migration.MigrationManager, module string) error {
	if module == "all" {
		modules := migrationManager.GetRegisteredModules()
		for _, mod := range modules {
			version, dirty, err := migrationManager.GetVersion(mod)
			if err != nil {
				return err
			}
			fmt.Printf("Module %s: version=%d, dirty=%t\n", mod, version, dirty)
		}
		return nil
	}

	version, dirty, err := migrationManager.GetVersion(module)
	if err != nil {
		return err
	}
	fmt.Printf("Module %s: version=%d, dirty=%t\n", module, version, dirty)
	return nil
}

func executeReset(migrationManager *migration.MigrationManager, module string) error {
	if module == "all" {
		modules := migrationManager.GetRegisteredModules()
		for _, mod := range modules {
			if err := migrationManager.Reset(mod); err != nil {
				return err
			}
		}
		return nil
	}
	return migrationManager.Reset(module)
}

func executeCreate(cfg *config.Config, module, name string, availableModules []string) error {
	if module == "all" {
		return fmt.Errorf("cannot create migration for 'all' modules, specify a specific module")
	}

	if !isValidModule(module, availableModules) {
		return fmt.Errorf("invalid module: %s. Available modules: %v", module, availableModules)
	}

	// Determine migration path dynamically
	migrationsPath := fmt.Sprintf("internal/modules/%s/migrations", module)

	// Use migrate CLI to create migration files
	return createMigrationFiles(migrationsPath, name)
}

func createMigrationFiles(migrationsPath, name string) error {
	// This would typically use the migrate CLI or create files manually
	// For now, we'll create a simple implementation
	fmt.Printf("Creating migration files for: %s in %s\n", name, migrationsPath)
	fmt.Printf("Run: migrate create -ext sql -dir %s -seq %s\n", migrationsPath, name)
	return nil
}
