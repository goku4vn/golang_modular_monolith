# Commands Reference

Tất cả commands có sẵn trong **Module-Based Auto-Registration Architecture**.

## Make Commands

### Development Environment

#### `make docker-dev`
Khởi động development environment với Docker và module auto-registration.
```bash
make docker-dev
```
**Chức năng:**
- Start PostgreSQL container
- Start application container với hot reload
- Mount source code để development
- Expose ports: 8080 (API), 5433 (PostgreSQL)
- **Auto-register modules** via init() functions
- **Load enabled modules** từ config/modules.yaml

#### `make docker-down`
Dừng tất cả containers.
```bash
make docker-down
```

#### `make docker-clean`
Dừng containers và xóa volumes.
```bash
make docker-clean
```
**⚠️ Cảnh báo:** Sẽ xóa tất cả data trong PostgreSQL!

### Database Management (Auto-Discovery)

#### `make create-databases`
Tạo databases cho tất cả enabled modules (auto-discovery).
```bash
make create-databases
```
**Chức năng:**
- **Auto-discover enabled modules** từ `config/modules.yaml`
- Tạo database cho mỗi enabled module
- Skip modules có `enabled: false`
- Skip modules có `migration.enabled: false`
- Báo cáo kết quả với colored output

**Example Output:**
```
🗄️ Database Creation Script (Module-Based)
🔍 Auto-discovering enabled modules from config...
📋 Enabled modules: customer order
🚫 Skipping user module (disabled in config)
✅ Database modular_monolith_customer created successfully
✅ Database modular_monolith_order already exists
```

#### `make migrate-up`
Chạy migrations lên phiên bản mới nhất (auto-discovery).
```bash
# Migrate tất cả enabled modules (auto-discovery)
make migrate-up

# Migrate module cụ thể
make migrate-up MODULE=customer
```

#### `make migrate-down`
Rollback migrations.
```bash
# Rollback tất cả enabled modules (1 step)
make migrate-down

# Rollback module cụ thể
make migrate-down MODULE=customer

# Rollback về version cụ thể
make migrate-down MODULE=customer VERSION=1
```

#### `make migrate-status`
Kiểm tra trạng thái migrations (auto-discovery).
```bash
# Status tất cả enabled modules (auto-discovery)
make migrate-status

# Status module cụ thể
make migrate-status MODULE=customer
```

#### `make migrate-create`
Tạo migration file mới.
```bash
make migrate-create MODULE=customer NAME=add_phone_column
```
**Output:**
- `internal/modules/customer/migrations/002_add_phone_column.up.sql`
- `internal/modules/customer/migrations/002_add_phone_column.down.sql`

### Module Management

#### `make list-modules`
Liệt kê tất cả modules (registered và enabled).
```bash
make list-modules
```
**Output:**
```
🔧 Registered modules: customer order user
📦 Enabled modules: customer order
🚫 Disabled modules: user
```

#### `make module-health`
Kiểm tra health của tất cả enabled modules.
```bash
make module-health
```

### Build Commands

#### `make build`
Build application binary với module auto-registration.
```bash
make build
```
**Output:** `bin/modular-monolith`

#### `make run`
Build và chạy application locally với module auto-loading.
```bash
make run
```

#### `make test`
Chạy tất cả tests (bao gồm module tests).
```bash
make test
```

#### `make test-coverage`
Chạy tests với coverage report.
```bash
make test-coverage
```

### Code Quality

#### `make lint`
Chạy linter (golangci-lint).
```bash
make lint
```

#### `make fmt`
Format code với gofmt.
```bash
make fmt
```

#### `make vet`
Chạy go vet.
```bash
make vet
```

## Direct Script Commands

### Database Creation Script (Auto-Discovery)

#### `./scripts/create-databases.sh`
Chạy database creation script với auto-discovery.
```bash
./scripts/create-databases.sh
```

**Environment Variables:**
```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export DATABASE_PREFIX=modular_monolith
```

**Auto-Discovery Process:**
1. Parse `config/modules.yaml`
2. Filter enabled modules (`enabled: true`)
3. Check migration settings (`migration.enabled`)
4. Create databases for discovered modules

### Development Script

#### `./scripts/docker-dev.sh`
Setup development environment với module support.
```bash
./scripts/docker-dev.sh
```

## Go Commands

### Migration Tool (Module-Based)

#### Basic Usage
```bash
go run cmd/migrate/main.go -action=up -module=customer
```

#### Available Actions
```bash
# Migrate up
go run cmd/migrate/main.go -action=up -module=customer

# Migrate down
go run cmd/migrate/main.go -action=down -module=customer

# Check status
go run cmd/migrate/main.go -action=status -module=customer

# Create new migration
go run cmd/migrate/main.go -action=create -module=customer -name=add_index

# Auto-discovery all enabled modules
go run cmd/migrate/main.go -action=up  # No module specified = all enabled
```

#### Migration Tool Options
```bash
-action string    # Migration action: up, down, status, create
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

### API Server (Module-Based)

#### Run API Server
```bash
go run cmd/api/main.go
```

**Module Loading Process:**
```
🔧 Registered module: customer
🔧 Registered module: order
🔧 Registered module: user
📦 Loaded module: customer (enabled: true)
📦 Loaded module: order (enabled: true)
🚫 Skipped module: user (enabled: false)
✅ Initialized module: customer
✅ Initialized module: order
🚀 Started module: customer
🚀 Started module: order
🌐 Server started on :8080
```

#### With Custom Config
```bash
go run cmd/api/main.go -config=config/production.yaml
```

### Development Tools

#### List Modules Tool
```bash
go run cmd/tools/list-modules.go
```
**Output:** 
```
🔧 Registered Modules:
  - customer (enabled: true)
  - order (enabled: true)
  - user (enabled: false)

📦 Loaded Modules:
  - customer
  - order

🚫 Disabled Modules:
  - user
```

#### Module Health Check Tool
```bash
go run cmd/tools/module-health.go
```

## Docker Commands

### Container Management

#### View Running Containers
```bash
docker ps
```

#### View Logs (Module-Based)
```bash
# Application logs (includes module loading logs)
docker logs tmm-dev

# Filter module-specific logs
docker logs tmm-dev | grep -E "(📦|🔧|🚫|✅|🚀)"

# PostgreSQL logs
docker logs tmm-postgres-dev

# Follow logs
docker logs -f tmm-dev
```

#### Execute Commands in Container
```bash
# Access application container
docker exec -it tmm-dev sh

# Access PostgreSQL container
docker exec -it tmm-postgres-dev psql -U postgres
```

#### Restart Containers
```bash
# Restart application (triggers module re-registration)
docker restart tmm-dev

# Restart PostgreSQL
docker restart tmm-postgres-dev
```

### Database Commands in Container

#### Connect to PostgreSQL
```bash
docker exec -it tmm-postgres-dev psql -U postgres
```

#### List Databases (Module-Based)
```bash
# List all databases
docker exec tmm-postgres-dev psql -U postgres -c "\l"

# List only module databases
docker exec tmm-postgres-dev psql -U postgres -c "\l" | grep modular_monolith
```

#### Create Database Manually
```bash
docker exec tmm-postgres-dev createdb -U postgres modular_monolith_customer
```

#### Drop Database
```bash
docker exec tmm-postgres-dev dropdb -U postgres modular_monolith_customer
```

## PostgreSQL Commands

### Connection Commands

#### Connect via psql
```bash
# Local connection (port 5433)
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres

# Container connection (port 5432)
PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres
```

#### Check Connection
```bash
PGPASSWORD=postgres pg_isready -h localhost -p 5433 -U postgres
```

### Database Management (Module-Based)

#### List Module Databases
```bash
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "\l" | grep modular_monolith
```

#### Create Database
```bash
PGPASSWORD=postgres createdb -h localhost -p 5433 -U postgres modular_monolith_customer
```

#### Drop Database
```bash
PGPASSWORD=postgres dropdb -h localhost -p 5433 -U postgres modular_monolith_customer
```

#### Backup All Module Databases
```bash
# Backup all enabled module databases
for module in $(grep -E "^\s+\w+:\s+true" config/modules.yaml | cut -d: -f1 | tr -d ' '); do
    db_name="modular_monolith_$module"
    PGPASSWORD=postgres pg_dump -h localhost -p 5433 -U postgres $db_name > backup_$db_name.sql
done
```

#### Restore Database
```bash
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres modular_monolith_customer < backup_modular_monolith_customer.sql
```

## API Testing Commands (Module-Based)

### Health Check (Module-Aware)
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "databases": ["customer", "order"],
  "modules": ["customer", "order"],
  "service": "modular-monolith",
  "version": "2.0.0"
}
```

### Pretty JSON Output
```bash
curl -s http://localhost:8080/health | jq .
```

### Module-Specific API Testing
```bash
# Customer module endpoints
curl -X GET http://localhost:8080/api/v1/customers
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Order module endpoints
curl -X GET http://localhost:8080/api/v1/orders
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "1", "amount": 100.00}'

# User module endpoints (if enabled)
curl -X GET http://localhost:8080/api/v1/users
```

### Test Module Availability
```bash
# Test if module is loaded and responding
curl -I http://localhost:8080/api/v1/customers  # Should return 200 if customer module loaded
curl -I http://localhost:8080/api/v1/users     # Should return 404 if user module disabled
```

## Environment Commands

### Environment Variables

#### Set Development Environment
```bash
export ENVIRONMENT=development
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433
export DATABASE_PREFIX=dev_modular_monolith
```

#### Set Production Environment
```bash
export ENVIRONMENT=production
export POSTGRES_HOST=prod-postgres.example.com
export POSTGRES_PORT=5432
export DATABASE_PREFIX=modular_monolith
```

#### Module-Specific Environment
```bash
# Override customer module database
export CUSTOMER_DATABASE_HOST=custom-host
export CUSTOMER_DATABASE_PORT=5433

# Disable order module migration
export ORDER_MIGRATION_ENABLED=false

# Module enable/disable via environment
export ANALYTICS_ENABLED=false
export REPORTING_ENABLED=true
```

## Troubleshooting Commands (Module-Based)

### Debug Commands

#### Check Module Configuration
```bash
cat config/modules.yaml
```

#### Check Module Registration Logs
```bash
docker logs tmm-dev | grep "🔧 Registered"
```

#### Check Module Loading Logs
```bash
docker logs tmm-dev | grep -E "(📦 Loaded|🚫 Skipped)"
```

#### Check Module Initialization Logs
```bash
docker logs tmm-dev | grep -E "(✅ Initialized|❌ Failed)"
```

#### Check Application Logs
```bash
docker logs tmm-dev | grep -E "(📦|🗄️|🚫|Failed|Error)"
```

#### Check Database Connections
```bash
# Test PostgreSQL connection
PGPASSWORD=postgres pg_isready -h localhost -p 5433 -U postgres

# List active connections
docker exec tmm-postgres-dev psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

#### Check Port Usage
```bash
# Check if ports are in use
lsof -i :8080  # API port
lsof -i :5433  # PostgreSQL port
```

#### Debug Module Registration
```bash
# Check if modules are imported
grep -r "_ \"golang_modular_monolith/internal/modules" internal/modules/modules.go

# Check if modules have init() functions
grep -r "func init()" internal/modules/*/module.go
```

### Recovery Commands

#### Reset Development Environment
```bash
# Stop everything
make docker-down

# Clean volumes
make docker-clean

# Restart fresh
make docker-dev

# Recreate databases (auto-discovery)
make create-databases

# Run migrations (auto-discovery)
make migrate-up
```

#### Fix Module Registration Issues
```bash
# Restart application to trigger re-registration
docker restart tmm-dev

# Check module imports
cat internal/modules/modules.go

# Verify module configuration
cat config/modules.yaml
```

#### Fix Hot Reload Issues
```bash
# Restart application container
docker restart tmm-dev

# Or trigger reload manually
docker exec tmm-dev touch /app/cmd/api/main.go
```

#### Fix Database Issues
```bash
# Recreate databases (auto-discovery)
make create-databases

# Reset migrations
make migrate-down
make migrate-up

# Check migration status
make migrate-status
```

#### Fix Module Loading Issues
```bash
# Check module configuration syntax
yamllint config/modules.yaml

# Verify module is registered
docker logs tmm-dev | grep "🔧 Registered module: your_module"

# Verify module is enabled
grep "your_module" config/modules.yaml

# Check module initialization
docker logs tmm-dev | grep "your_module"
```

## Useful Command Combinations

### Complete Setup from Scratch (Module-Based)
```bash
# 1. Start environment
make docker-dev

# 2. Create databases (auto-discovery)
make create-databases

# 3. Run migrations (auto-discovery)
make migrate-up

# 4. Test API with module health
curl http://localhost:8080/health

# 5. List loaded modules
make list-modules
```

### Daily Development Workflow (Module-Based)
```bash
# Start development
make docker-dev

# Check loaded modules
docker logs tmm-dev | grep -E "(📦|🔧)"

# Make code changes (auto-reload)
# ...

# Add new migration
make migrate-create MODULE=customer NAME=add_new_field

# Run migration
make migrate-up MODULE=customer

# Test changes
curl http://localhost:8080/api/v1/customers
```

### Adding New Module Workflow
```bash
# 1. Create module with auto-registration
# (implement Module interface + init() function)

# 2. Add to centralized import
echo '_ "golang_modular_monolith/internal/modules/new_module"' >> internal/modules/modules.go

# 3. Enable in config
echo "  new_module: true" >> config/modules.yaml

# 4. Restart to trigger registration
docker restart tmm-dev

# 5. Create database
make create-databases

# 6. Create initial migration
make migrate-create MODULE=new_module NAME=initial_schema

# 7. Run migration
make migrate-up MODULE=new_module

# 8. Test module
curl http://localhost:8080/api/v1/new_module
```

### Production Deployment (Module-Based)
```bash
# Build application
make build

# Run tests (including module tests)
make test

# Deploy with module auto-registration
./bin/modular-monolith -config=config/production.yaml
```

## Command Aliases (Module-Based)

Để tiện lợi, có thể tạo aliases:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias tmm-dev='make docker-dev'
alias tmm-down='make docker-down'
alias tmm-clean='make docker-clean'
alias tmm-db='make create-databases'
alias tmm-up='make migrate-up'
alias tmm-down='make migrate-down'
alias tmm-status='make migrate-status'
alias tmm-health='curl -s http://localhost:8080/health | jq .'
alias tmm-modules='make list-modules'
alias tmm-logs='docker logs tmm-dev | grep -E "(📦|🔧|🚫|✅|🚀)"'
``` 