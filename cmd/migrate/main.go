package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"golang_modular_monolith/internal/shared/infrastructure/database"
	"golang_modular_monolith/internal/shared/infrastructure/migration"
)

func main() {
	var (
		module  = flag.String("module", "", "Module name (customer, order, product, all)")
		action  = flag.String("action", "up", "Migration action (up, down, version, reset, create)")
		version = flag.String("version", "", "Target version for migrate to specific version")
		name    = flag.String("name", "", "Migration name for create action")
	)
	flag.Parse()

	if *module == "" {
		fmt.Println("Usage: go run cmd/migrate/main.go -module=<module> -action=<action> [options]")
		fmt.Println("Modules: customer, order, product, all")
		fmt.Println("Actions: up, down, version, reset, create")
		fmt.Println("Options:")
		fmt.Println("  -version=<version>  Target version for migrate")
		fmt.Println("  -name=<name>        Migration name for create action")
		os.Exit(1)
	}

	// Create migration manager
	migrationManager := migration.NewMigrationManager()
	defer migrationManager.Close()

	// Register modules based on input
	if err := registerModules(migrationManager, *module); err != nil {
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
		if err := executeVersion(migrationManager, *module, *version); err != nil {
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
		if err := executeCreate(*module, *name); err != nil {
			log.Fatalf("Migration create failed: %v", err)
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}

	fmt.Println("Migration completed successfully!")
}

func registerModules(migrationManager *migration.MigrationManager, module string) error {
	switch module {
	case "customer":
		return registerCustomerModule(migrationManager)
	case "order":
		return registerOrderModule(migrationManager)
	case "product":
		// TODO: Implement when product module is ready
		return fmt.Errorf("product module not implemented yet")
	case "all":
		if err := registerCustomerModule(migrationManager); err != nil {
			return err
		}
		if err := registerOrderModule(migrationManager); err != nil {
			return err
		}
		// TODO: Add product module when ready
		return nil
	default:
		return fmt.Errorf("unknown module: %s", module)
	}
}

func registerCustomerModule(migrationManager *migration.MigrationManager) error {
	// Initialize database manager
	manager := database.GetGlobalManager()

	// Register customer database
	config := database.LoadConfigFromEnv("CUSTOMER_DATABASE")
	if config.Name == "" {
		config.Name = "modular_monolith_customer"
	}
	manager.RegisterDatabase("customer", config)

	// Get database connection
	db, err := manager.GetConnection("customer")
	if err != nil {
		return fmt.Errorf("failed to connect to customer database: %w", err)
	}

	return migrationManager.RegisterModule("customer", db, "internal/modules/customer/migrations")
}

func registerOrderModule(migrationManager *migration.MigrationManager) error {
	// Initialize database manager
	manager := database.GetGlobalManager()

	// Register order database
	config := database.LoadConfigFromEnv("ORDER_DATABASE")
	if config.Name == "" {
		config.Name = "modular_monolith_order"
	}
	manager.RegisterDatabase("order", config)

	// Get database connection
	db, err := manager.GetConnection("order")
	if err != nil {
		return fmt.Errorf("failed to connect to order database: %w", err)
	}

	return migrationManager.RegisterModule("order", db, "internal/modules/order/migrations")
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

func executeVersion(migrationManager *migration.MigrationManager, module, versionStr string) error {
	if versionStr == "" {
		// Show current version
		if module == "all" {
			modules := migrationManager.GetRegisteredModules()
			for _, mod := range modules {
				version, dirty, err := migrationManager.GetVersion(mod)
				if err != nil {
					return err
				}
				fmt.Printf("Module %s: version=%d, dirty=%t\n", mod, version, dirty)
			}
		} else {
			version, dirty, err := migrationManager.GetVersion(module)
			if err != nil {
				return err
			}
			fmt.Printf("Module %s: version=%d, dirty=%t\n", module, version, dirty)
		}
		return nil
	}

	// Migrate to specific version
	version, err := strconv.ParseUint(versionStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid version: %s", versionStr)
	}

	if module == "all" {
		modules := migrationManager.GetRegisteredModules()
		for _, mod := range modules {
			if err := migrationManager.MigrateToVersion(mod, uint(version)); err != nil {
				return err
			}
		}
		return nil
	}

	return migrationManager.MigrateToVersion(module, uint(version))
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

func executeCreate(module, name string) error {
	if module == "all" {
		return fmt.Errorf("cannot create migration for 'all' modules, specify a specific module")
	}

	var migrationsPath string
	switch module {
	case "customer":
		migrationsPath = "internal/modules/customer/migrations"
	case "order":
		migrationsPath = "internal/modules/order/migrations"
	case "product":
		migrationsPath = "internal/modules/product/migrations"
	default:
		return fmt.Errorf("unknown module: %s", module)
	}

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
