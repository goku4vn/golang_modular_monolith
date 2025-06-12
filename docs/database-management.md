# Database Management

Hướng dẫn quản lý databases và migrations trong **Module-Based Auto-Registration Architecture**.

## Overview

Modular Monolith sử dụng **Global Database Manager** với **Module-Based Database Management**:
- **Global Database Manager** quản lý tất cả database connections
- **Module-based database creation** (mỗi module có database riêng)
- **Auto-discovery modules** từ configuration
- **Manual migration execution** (developer control)

## Database Architecture

### Database Per Module with Global Manager
```
Global Database Manager
├── Customer DB Connection → modular_monolith_customer
├── Order DB Connection    → modular_monolith_order  
├── User DB Connection     → modular_monolith_user
└── Analytics DB Connection → modular_monolith_analytics
```

### Module Database Access Pattern
```go
// Old approach - Direct database access
db, err := sql.Open("postgres", connectionString)

// New approach - Via Database Manager
import "golang_modular_monolith/internal/shared/infrastructure/database"

// Get database for specific module
customerDB, err := database.GetCustomerDB()
orderDB, err := database.GetOrderDB()
```

### Naming Convention
- **Pattern**: `{DATABASE_PREFIX}_{module_name}`
- **Default prefix**: `modular_monolith`
- **Example**: `modular_monolith_customer`

## Database Creation (Auto-Discovery)

### 1. Automatic Creation Script (Module-Based)
```bash
# Create databases for all enabled modules (auto-discovery)
make create-databases

# Or run script directly
./scripts/create-databases.sh
```

### 2. How Auto-Discovery Works
```bash
🗄️ Database Creation Script (Module-Based)
================================
🔍 Checking PostgreSQL connection...
✅ PostgreSQL connection successful
🔍 Auto-discovering enabled modules from config...
📋 Enabled modules: customer order (user disabled)
📦 Checking database: modular_monolith_customer
🔨 Creating database: modular_monolith_customer
✅ Database modular_monolith_customer created successfully
📦 Checking database: modular_monolith_order
✅ Database modular_monolith_order already exists
🚫 Skipping user module (disabled in config)
🎉 Database creation completed!
```

### 3. Manual Creation
```bash
# Connect to PostgreSQL
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres

# Create database manually
CREATE DATABASE modular_monolith_customer;
CREATE DATABASE modular_monolith_order;
```

### 4. Environment Variables
```bash
# Customize database settings
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export DATABASE_PREFIX=modular_monolith
```

## Database Creation Script Details

### Script Location
```
scripts/create-databases.sh
```

### How It Works (Module-Based)
1. **Parse config**: Read `config/modules.yaml`
2. **Auto-discover enabled modules**: Only modules with `enabled: true`
3. **Check migration settings**: Skip if `migration.enabled: false`
4. **Connect to PostgreSQL**: Validate connection
5. **Create databases**: Only if not exists
6. **Report results**: Colored output with summary

### Example Output (Auto-Discovery)
```bash
🗄️ Database Creation Script (Module-Based)
================================
🔍 Checking PostgreSQL connection...
✅ PostgreSQL connection successful
🔍 Auto-discovering enabled modules from config...
📋 Enabled modules: customer order
📦 Checking database: modular_monolith_customer
🔨 Creating database: modular_monolith_customer
✅ Database modular_monolith_customer created successfully
📦 Checking database: modular_monolith_order
✅ Database modular_monolith_order already exists
🎉 Database creation completed!
```

## Global Database Manager

### Database Manager Architecture
```go
// internal/shared/infrastructure/database/manager.go
type DatabaseManager struct {
    connections map[string]*sql.DB
    config      *config.Config
}

// Global instance
var globalManager *DatabaseManager

// Module-specific getters
func GetCustomerDB() (*sql.DB, error)
func GetOrderDB() (*sql.DB, error)
func GetUserDB() (*sql.DB, error)
```

### Database Manager Usage in Modules
```go
// internal/modules/customer/infrastructure/persistence/customer_repository.go
package persistence

import (
    customerdb "golang_modular_monolith/internal/shared/infrastructure/database"
)

func NewPostgreSQLCustomerRepository() (*PostgreSQLCustomerRepository, error) {
    // Get database via global manager
    db, err := customerdb.GetCustomerDB()
    if err != nil {
        return nil, fmt.Errorf("failed to get customer database: %w", err)
    }

    return &PostgreSQLCustomerRepository{db: db}, nil
}
```

### Database Manager Initialization
```go
// cmd/api/main.go
func initDatabases(cfg *config.Config) error {
    // Initialize global database manager
    if err := database.InitializeGlobalManager(cfg); err != nil {
        return fmt.Errorf("failed to initialize database manager: %w", err)
    }

    // Create databases for enabled modules
    if err := database.CreateDatabasesForEnabledModules(cfg); err != nil {
        return fmt.Errorf("failed to create databases: %w", err)
    }

    return nil
}
```

## Migration Management (Module-Based)

### Migration Structure
```
internal/modules/{module}/migrations/
├── 001_create_users_table.up.sql      # Up migration
├── 001_create_users_table.down.sql    # Down migration
├── 002_add_email_index.up.sql
└── 002_add_email_index.down.sql
```

### Migration Commands (Auto-Discovery)

#### Run All Migrations (All Enabled Modules)
```bash
# Migrate all enabled modules up (auto-discovery)
make migrate-up

# Migrate specific module
make migrate-up MODULE=customer
```

#### Rollback Migrations
```bash
# Rollback all enabled modules
make migrate-down

# Rollback specific module  
make migrate-down MODULE=customer

# Rollback to specific version
make migrate-down MODULE=customer VERSION=1
```

#### Migration Status (Auto-Discovery)
```bash
# Check migration status for all enabled modules
make migrate-status

# Check specific module
make migrate-status MODULE=customer
```

#### Create New Migration
```bash
# Create new migration for module
make migrate-create MODULE=customer NAME=add_phone_column
```

### Migration Tool Usage (Module-Based)

#### Direct Usage
```bash
# Run migration tool directly
go run cmd/migrate/main.go -action=up -module=customer

# Available actions: up, down, status, create
# Available modules: auto-discovered from config/modules.yaml
```

#### Migration Tool Options
```bash
-action string    # Migration action (up/down/status/create)
-module string    # Target module name (must be enabled in config)
-version int      # Target version (for down migrations)
-name string      # Migration name (for create action)
-config string    # Config file path (default: config/modules.yaml)
```

#### Migration Tool Auto-Discovery
```bash
# Migration tool automatically discovers enabled modules
🔍 Auto-discovering enabled modules from config...
📋 Available modules: customer order
✅ Module 'customer' is enabled and available
🚫 Module 'user' is disabled - skipping
```

## Database Configuration (Module-Based)

### Module Database Config
```yaml
# internal/modules/customer/module.yaml
enabled: true
database:
  host: "${POSTGRES_HOST:localhost}"
  port: "${POSTGRES_PORT:5432}"
  user: "${POSTGRES_USER:postgres}"
  password: "${POSTGRES_PASSWORD:postgres}"
  name: "modular_monolith_customer"
  sslmode: "${POSTGRES_SSLMODE:disable}"
migration:
  enabled: true
  path: "./migrations"
```

### Global Database Configuration
```yaml
# config/app.yaml
database:
  global:
    host: "${POSTGRES_HOST:localhost}"
    port: "${POSTGRES_PORT:5432}"
    user: "${POSTGRES_USER:postgres}"
    password: "${POSTGRES_PASSWORD:postgres}"
    sslmode: "${POSTGRES_SSLMODE:disable}"
    max_connections: 25
    max_idle_connections: 5
```

### Environment Override
```bash
# Override database settings per module
export CUSTOMER_DATABASE_HOST=custom-host
export CUSTOMER_DATABASE_PORT=5433
export ORDER_DATABASE_NAME=custom_order_db

# Global database settings
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
```

## Common Scenarios (Module-Based)

### 1. Adding New Module (Auto-Registration)
```bash
# 1. Create module with auto-registration
# internal/modules/new_module/module.go
func init() {
    registry.RegisterModule("new_module", func() domain.Module {
        return NewNewModule()
    })
}

# 2. Add to centralized import
# internal/modules/modules.go
import _ "golang_modular_monolith/internal/modules/new_module"

# 3. Enable in config
echo "  new_module: true" >> config/modules.yaml

# 4. Create database (auto-discovery)
make create-databases

# 5. Create initial migration
make migrate-create MODULE=new_module NAME=initial_schema

# 6. Run migration
make migrate-up MODULE=new_module
```

### 2. Disabling Module Database
```yaml
# config/modules.yaml
modules:
  analytics:
    enabled: true
    migration:
      enabled: false    # Module enabled but no database
```

### 3. Development vs Production
```bash
# Development - local PostgreSQL
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433

# Production - managed PostgreSQL
export POSTGRES_HOST=prod-postgres.example.com
export POSTGRES_PORT=5432
export POSTGRES_PASSWORD=secure-password
```

### 4. Testing Environment
```bash
# Use test database prefix
export DATABASE_PREFIX=test_modular_monolith

# Create test databases (auto-discovery)
make create-databases

# Run migrations (auto-discovery)
make migrate-up
```

## Module Database Integration

### Repository Pattern with Database Manager
```go
// internal/modules/customer/infrastructure/persistence/customer_repository.go
type PostgreSQLCustomerRepository struct {
    db *sql.DB
}

func NewPostgreSQLCustomerRepository() (*PostgreSQLCustomerRepository, error) {
    // Use global database manager
    db, err := customerdb.GetCustomerDB()
    if err != nil {
        return nil, fmt.Errorf("failed to get customer database: %w", err)
    }

    return &PostgreSQLCustomerRepository{db: db}, nil
}

func (r *PostgreSQLCustomerRepository) Save(ctx context.Context, customer *domain.Customer) error {
    query := `INSERT INTO customers (id, email, name, created_at) VALUES ($1, $2, $3, $4)`
    _, err := r.db.ExecContext(ctx, query, customer.ID, customer.Email, customer.Name, customer.CreatedAt)
    return err
}
```

### Module Database Initialization
```go
// internal/modules/customer/module.go
func (m *CustomerModule) Initialize(deps domain.ModuleDependencies) error {
    // Repository will automatically use global database manager
    customerRepo, err := persistence.NewPostgreSQLCustomerRepository()
    if err != nil {
        return fmt.Errorf("failed to initialize customer repository: %w", err)
    }

    // Initialize other components
    m.customerService = services.NewCustomerService(customerRepo)
    m.handler = handlers.NewCustomerHandler(m.customerService)

    return nil
}
```

## Troubleshooting (Module-Based)

### Common Issues

**1. Database creation failed**
```bash
# Check PostgreSQL connection
docker ps | grep postgres
PGPASSWORD=postgres pg_isready -h localhost -p 5433 -U postgres

# Check logs
docker logs tmm-postgres-dev

# Check module configuration
cat config/modules.yaml
```

**2. Migration failed**
```bash
# Check migration files syntax
cat internal/modules/customer/migrations/001_*.sql

# Check database connection via manager
go run cmd/migrate/main.go -action=status -module=customer

# Manual migration
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -d modular_monolith_customer -f migration.sql
```

**3. Module database not found**
```bash
# Check if database exists
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "\l" | grep modular_monolith

# Check if module is enabled
grep -A 5 "modules:" config/modules.yaml

# Recreate database (auto-discovery)
make create-databases

# Check database manager logs
docker logs tmm-dev | grep "Database"
```

**4. Database Manager connection failed**
```bash
# Check database manager initialization
docker logs tmm-dev | grep "DatabaseManager"

# Check environment variables
env | grep POSTGRES

# Test database connection manually
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "SELECT 1"
```

**5. Module not auto-discovered**
```bash
# Check if module is imported
grep "new_module" internal/modules/modules.go

# Check if module is enabled
grep "new_module" config/modules.yaml

# Check auto-discovery logs
docker logs tmm-dev | grep "Auto-discovering"
```

## Best Practices (Module-Based)

### 1. Migration Naming
```
001_create_users_table.up.sql       # ✅ Good
002_add_email_index.up.sql          # ✅ Good
add_column.sql                      # ❌ Bad - no version
```

### 2. Migration Content
```sql
-- ✅ Good - Reversible
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL
);

-- ❌ Bad - Data loss
DROP TABLE old_users;
```

### 3. Database Lifecycle (Module-Based)
```bash
# ✅ Good workflow
1. Implement Module interface with auto-registration
2. Add to centralized import (modules.go)
3. Update config/modules.yaml
4. make create-databases (auto-discovery)
5. make migrate-create
6. make migrate-up

# ❌ Bad workflow  
1. Manual database creation
2. Skip migration files
3. Direct SQL execution
4. Hardcode database connections
```

### 4. Database Manager Usage
```go
// ✅ Good - Use Database Manager
db, err := customerdb.GetCustomerDB()

// ❌ Bad - Direct connection
db, err := sql.Open("postgres", connectionString)
```

### 5. Environment Management
```bash
# ✅ Good - Use environment variables
export POSTGRES_HOST=prod-host

# ❌ Bad - Hardcode in config
database:
  host: "prod-host"  # Don't do this
```

## Advanced Usage (Module-Based)

### Custom Database Names
```yaml
# Override default naming
modules:
  customer:
    enabled: true
    database:
      name: "custom_customer_db"
```

### Multiple Environments
```bash
# Development
export DATABASE_PREFIX=dev_app
make create-databases

# Staging  
export DATABASE_PREFIX=staging_app
make create-databases

# Production
export DATABASE_PREFIX=prod_app
make create-databases
```

### Database Health Checks
```go
// Health check includes all module databases
func healthCheckHandler(cfg *config.Config, moduleRegistry *domain.ModuleRegistry) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check all module databases via Database Manager
        databases := database.GetAllDatabaseNames()
        
        response := gin.H{
            "status":    "healthy",
            "databases": databases,
            "modules":   moduleRegistry.GetLoadedModuleNames(),
        }
        
        c.JSON(200, response)
    }
}
```

### Database Backup/Restore (Auto-Discovery)
```bash
# Backup all enabled module databases
for module in $(grep -E "^\s+\w+:\s+true" config/modules.yaml | cut -d: -f1 | tr -d ' '); do
    db_name="modular_monolith_$module"
    PGPASSWORD=postgres pg_dump -h localhost -p 5433 -U postgres $db_name > backup_$db_name.sql
done

# Restore database
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres modular_monolith_customer < backup_modular_monolith_customer.sql
``` 