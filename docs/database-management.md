# Database Management

Hướng dẫn quản lý databases và migrations trong Modular Monolith.

## Overview

Modular Monolith sử dụng **manual database management** approach:
- **App controls database lifecycle** (không phải PostgreSQL container)
- **Module-based database creation** (mỗi module có database riêng)
- **Manual migration execution** (developer control)

## Database Architecture

### Database Per Module
```
PostgreSQL Instance
├── modular_monolith_customer    # Customer module database
├── modular_monolith_order       # Order module database  
├── modular_monolith_user        # User module database
└── modular_monolith_analytics   # Analytics module database
```

### Naming Convention
- **Pattern**: `{DATABASE_PREFIX}_{module_name}`
- **Default prefix**: `modular_monolith`
- **Example**: `modular_monolith_customer`

## Database Creation

### 1. Automatic Creation Script
```bash
# Create databases for all enabled modules
make create-databases

# Or run script directly
./scripts/create-databases.sh
```

### 2. Manual Creation
```bash
# Connect to PostgreSQL
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres

# Create database manually
CREATE DATABASE modular_monolith_customer;
CREATE DATABASE modular_monolith_order;
```

### 3. Environment Variables
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

### How It Works
1. **Parse config**: Read `config/modules.yaml`
2. **Filter enabled modules**: Only modules with `enabled: true`
3. **Check migration settings**: Skip if `migration.enabled: false`
4. **Connect to PostgreSQL**: Validate connection
5. **Create databases**: Only if not exists
6. **Report results**: Colored output with summary

### Example Output
```bash
🗄️ Database Creation Script
================================
🔍 Checking PostgreSQL connection...
✅ PostgreSQL connection successful
🔍 Discovering enabled modules...
📋 Enabled modules: customer order
📦 Checking database: modular_monolith_customer
🔨 Creating database: modular_monolith_customer
✅ Database modular_monolith_customer created successfully
📦 Checking database: modular_monolith_order
✅ Database modular_monolith_order already exists
🎉 Database creation completed!
```

## Migration Management

### Migration Structure
```
internal/modules/{module}/migrations/
├── 001_create_users_table.up.sql      # Up migration
├── 001_create_users_table.down.sql    # Down migration
├── 002_add_email_index.up.sql
└── 002_add_email_index.down.sql
```

### Migration Commands

#### Run All Migrations
```bash
# Migrate all modules up
make migrate-up

# Migrate specific module
make migrate-up MODULE=customer
```

#### Rollback Migrations
```bash
# Rollback all modules
make migrate-down

# Rollback specific module  
make migrate-down MODULE=customer

# Rollback to specific version
make migrate-down MODULE=customer VERSION=1
```

#### Migration Status
```bash
# Check migration status
make migrate-status

# Check specific module
make migrate-status MODULE=customer
```

#### Create New Migration
```bash
# Create new migration for module
make migrate-create MODULE=customer NAME=add_phone_column
```

### Migration Tool Usage

#### Direct Usage
```bash
# Run migration tool directly
go run cmd/migrate/main.go -action=up -module=customer

# Available actions: up, down, status, create
# Available modules: customer, order, user (based on config)
```

#### Migration Tool Options
```bash
-action string    # Migration action (up/down/status/create)
-module string    # Target module name  
-version int      # Target version (for down migrations)
-name string      # Migration name (for create action)
-config string    # Config file path (default: config/modules.yaml)
```

## Database Configuration

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

### Environment Override
```bash
# Override database settings per module
export CUSTOMER_DATABASE_HOST=custom-host
export CUSTOMER_DATABASE_PORT=5433
export ORDER_DATABASE_NAME=custom_order_db
```

## Common Scenarios

### 1. Adding New Module
```bash
# 1. Add module to config
echo "  new_module: true" >> config/modules.yaml

# 2. Create database
make create-databases

# 3. Create initial migration
make migrate-create MODULE=new_module NAME=initial_schema

# 4. Run migration
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

# Create test databases
make create-databases

# Run migrations
make migrate-up
```

## Troubleshooting

### Common Issues

**1. Database creation failed**
```bash
# Check PostgreSQL connection
docker ps | grep postgres
PGPASSWORD=postgres pg_isready -h localhost -p 5433 -U postgres

# Check logs
docker logs modular-monolith-postgres-dev
```

**2. Migration failed**
```bash
# Check migration files syntax
cat internal/modules/customer/migrations/001_*.sql

# Check database connection
make migrate-status MODULE=customer

# Manual migration
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -d modular_monolith_customer -f migration.sql
```

**3. Module database not found**
```bash
# Check if database exists
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "\l" | grep modular_monolith

# Recreate database
make create-databases

# Check module configuration
cat config/modules.yaml
```

**4. Wrong database connection**
```bash
# App connects to container (port 5432)
# Script connects to localhost (port 5433)

# Create database in container
docker exec modular-monolith-postgres-dev createdb -U postgres modular_monolith_customer
```

## Best Practices

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

### 3. Database Lifecycle
```bash
# ✅ Good workflow
1. Update config/modules.yaml
2. make create-databases  
3. make migrate-create
4. make migrate-up

# ❌ Bad workflow  
1. Manual database creation
2. Skip migration files
3. Direct SQL execution
```

### 4. Environment Management
```bash
# ✅ Good - Use environment variables
export POSTGRES_HOST=prod-host

# ❌ Bad - Hardcode in config
database:
  host: "prod-host"  # Don't do this
```

## Advanced Usage

### Custom Database Names
```yaml
# Override default naming
modules:
  customer:
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

### Database Backup/Restore
```bash
# Backup all module databases
for db in $(PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -lqt | cut -d \| -f 1 | grep modular_monolith); do
    PGPASSWORD=postgres pg_dump -h localhost -p 5433 -U postgres $db > backup_$db.sql
done

# Restore database
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres modular_monolith_customer < backup_modular_monolith_customer.sql
``` 