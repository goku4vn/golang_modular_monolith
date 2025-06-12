package migration

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// MigrationManager manages database migrations for modules
type MigrationManager struct {
	migrators map[string]*migrate.Migrate
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrators: make(map[string]*migrate.Migrate),
	}
}

// RegisterModule registers a module's migration path with its database
func (mm *MigrationManager) RegisterModule(moduleName string, db *gorm.DB, migrationsPath string) error {
	// Get underlying sql.DB from GORM
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver for %s: %w", moduleName, err)
	}

	// Get absolute path for migrations
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", migrationsPath, err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance for %s: %w", moduleName, err)
	}

	mm.migrators[moduleName] = m
	log.Printf("Migration registered for module: %s (path: %s)", moduleName, migrationsPath)
	return nil
}

// MigrateUp runs all up migrations for a module
func (mm *MigrationManager) MigrateUp(moduleName string) error {
	migrator, exists := mm.migrators[moduleName]
	if !exists {
		return fmt.Errorf("no migrator found for module: %s", moduleName)
	}

	err := migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate up for %s: %w", moduleName, err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("No migrations to apply for module: %s", moduleName)
	} else {
		log.Printf("Successfully migrated up for module: %s", moduleName)
	}

	return nil
}

// MigrateDown runs one down migration for a module
func (mm *MigrationManager) MigrateDown(moduleName string) error {
	migrator, exists := mm.migrators[moduleName]
	if !exists {
		return fmt.Errorf("no migrator found for module: %s", moduleName)
	}

	err := migrator.Steps(-1)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate down for %s: %w", moduleName, err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("No migrations to rollback for module: %s", moduleName)
	} else {
		log.Printf("Successfully migrated down for module: %s", moduleName)
	}

	return nil
}

// MigrateToVersion migrates to a specific version for a module
func (mm *MigrationManager) MigrateToVersion(moduleName string, version uint) error {
	migrator, exists := mm.migrators[moduleName]
	if !exists {
		return fmt.Errorf("no migrator found for module: %s", moduleName)
	}

	err := migrator.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d for %s: %w", version, moduleName, err)
	}

	log.Printf("Successfully migrated to version %d for module: %s", version, moduleName)
	return nil
}

// GetVersion returns the current migration version for a module
func (mm *MigrationManager) GetVersion(moduleName string) (uint, bool, error) {
	migrator, exists := mm.migrators[moduleName]
	if !exists {
		return 0, false, fmt.Errorf("no migrator found for module: %s", moduleName)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get version for %s: %w", moduleName, err)
	}

	return version, dirty, nil
}

// Reset drops all tables and re-runs all migrations for a module
func (mm *MigrationManager) Reset(moduleName string) error {
	migrator, exists := mm.migrators[moduleName]
	if !exists {
		return fmt.Errorf("no migrator found for module: %s", moduleName)
	}

	// Drop all tables
	err := migrator.Drop()
	if err != nil {
		return fmt.Errorf("failed to drop tables for %s: %w", moduleName, err)
	}

	// Run all migrations
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate up after reset for %s: %w", moduleName, err)
	}

	log.Printf("Successfully reset and migrated module: %s", moduleName)
	return nil
}

// MigrateAllUp runs up migrations for all registered modules
func (mm *MigrationManager) MigrateAllUp() error {
	for moduleName := range mm.migrators {
		if err := mm.MigrateUp(moduleName); err != nil {
			return err
		}
	}
	return nil
}

// MigrateAllDown runs down migrations for all registered modules
func (mm *MigrationManager) MigrateAllDown() error {
	for moduleName := range mm.migrators {
		if err := mm.MigrateDown(moduleName); err != nil {
			return err
		}
	}
	return nil
}

// GetRegisteredModules returns list of registered module names
func (mm *MigrationManager) GetRegisteredModules() []string {
	modules := make([]string, 0, len(mm.migrators))
	for moduleName := range mm.migrators {
		modules = append(modules, moduleName)
	}
	return modules
}

// Close closes all migrators
func (mm *MigrationManager) Close() error {
	for moduleName, migrator := range mm.migrators {
		if sourceErr, dbErr := migrator.Close(); sourceErr != nil || dbErr != nil {
			log.Printf("Error closing migrator for %s: source=%v, db=%v", moduleName, sourceErr, dbErr)
		}
	}
	return nil
}
