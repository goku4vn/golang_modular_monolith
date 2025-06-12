# Commands Reference

Tất cả commands có sẵn trong Modular Monolith project.

## Make Commands

### Development Environment

#### `make docker-dev`
Khởi động development environment với Docker.
```bash
make docker-dev
```
**Chức năng:**
- Start PostgreSQL container
- Start application container với hot reload
- Mount source code để development
- Expose ports: 8080 (API), 5433 (PostgreSQL)

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

### Database Management

#### `make create-databases`
Tạo databases cho tất cả enabled modules.
```bash
make create-databases
```
**Chức năng:**
- Parse `config/modules.yaml`
- Tạo database cho mỗi enabled module
- Skip modules có `migration.enabled: false`
- Báo cáo kết quả với colored output

#### `make migrate-up`
Chạy migrations lên phiên bản mới nhất.
```bash
# Migrate tất cả modules
make migrate-up

# Migrate module cụ thể
make migrate-up MODULE=customer
```

#### `make migrate-down`
Rollback migrations.
```bash
# Rollback tất cả modules (1 step)
make migrate-down

# Rollback module cụ thể
make migrate-down MODULE=customer

# Rollback về version cụ thể
make migrate-down MODULE=customer VERSION=1
```

#### `make migrate-status`
Kiểm tra trạng thái migrations.
```bash
# Status tất cả modules
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

### Build Commands

#### `make build`
Build application binary.
```bash
make build
```
**Output:** `bin/modular-monolith`

#### `make run`
Build và chạy application locally.
```bash
make run
```

#### `make test`
Chạy tất cả tests.
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

### Database Creation Script

#### `./scripts/create-databases.sh`
Chạy database creation script trực tiếp.
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

### Development Script

#### `./scripts/docker-dev.sh`
Setup development environment.
```bash
./scripts/docker-dev.sh
```

## Go Commands

### Migration Tool

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
```

#### Migration Tool Options
```bash
-action string    # Migration action: up, down, status, create
-module string    # Target module name
-version int      # Target version (for down migrations)
-name string      # Migration name (for create action)
-config string    # Config file path (default: config/modules.yaml)
```

### API Server

#### Run API Server
```bash
go run cmd/api/main.go
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
**Output:** Danh sách enabled modules từ config

## Docker Commands

### Container Management

#### View Running Containers
```bash
docker ps
```

#### View Logs
```bash
# Application logs
docker logs tmm-dev

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
# Restart application (for hot reload issues)
docker restart tmm-dev

# Restart PostgreSQL
docker restart tmm-postgres-dev
```

### Database Commands in Container

#### Connect to PostgreSQL
```bash
docker exec -it tmm-postgres-dev psql -U postgres
```

#### List Databases
```bash
docker exec tmm-postgres-dev psql -U postgres -c "\l"
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

### Database Management

#### List Databases
```bash
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "\l"
```

#### Create Database
```bash
PGPASSWORD=postgres createdb -h localhost -p 5433 -U postgres modular_monolith_customer
```

#### Drop Database
```bash
PGPASSWORD=postgres dropdb -h localhost -p 5433 -U postgres modular_monolith_customer
```

#### Backup Database
```bash
PGPASSWORD=postgres pg_dump -h localhost -p 5433 -U postgres modular_monolith_customer > backup.sql
```

#### Restore Database
```bash
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres modular_monolith_customer < backup.sql
```

## API Testing Commands

### Health Check
```bash
curl http://localhost:8080/health
```

### Pretty JSON Output
```bash
curl -s http://localhost:8080/health | jq .
```

### API Endpoints Testing
```bash
# GET request
curl -X GET http://localhost:8080/api/v1/customers

# POST request
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# PUT request
curl -X PUT http://localhost:8080/api/v1/customers/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "email": "jane@example.com"}'

# DELETE request
curl -X DELETE http://localhost:8080/api/v1/customers/1
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
```

## Troubleshooting Commands

### Debug Commands

#### Check Module Configuration
```bash
cat config/modules.yaml
```

#### Check Application Logs
```bash
docker logs tmm-dev | grep -E "(📦|🗄️|🚫|Failed)"
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

### Recovery Commands

#### Reset Development Environment
```bash
# Stop everything
make docker-down

# Clean volumes
make docker-clean

# Restart fresh
make docker-dev

# Recreate databases
make create-databases

# Run migrations
make migrate-up
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
# Recreate databases
make create-databases

# Reset migrations
make migrate-down
make migrate-up

# Check migration status
make migrate-status
```

## Useful Command Combinations

### Complete Setup from Scratch
```bash
# 1. Start environment
make docker-dev

# 2. Create databases
make create-databases

# 3. Run migrations
make migrate-up

# 4. Test API
curl http://localhost:8080/health
```

### Daily Development Workflow
```bash
# Start development
make docker-dev

# Make code changes (auto-reload)
# ...

# Add new migration
make migrate-create MODULE=customer NAME=add_new_field

# Run migration
make migrate-up MODULE=customer

# Test changes
curl http://localhost:8080/api/v1/customers
```

### Production Deployment
```bash
# Build application
make build

# Run tests
make test

# Deploy (example)
./bin/modular-monolith -config=config/production.yaml
```

## Command Aliases

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
``` 