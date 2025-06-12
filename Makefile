.PHONY: help build run test clean migrate-up migrate-down docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  build             - Build the application"
	@echo "  run               - Run the application"
	@echo "  run-dev           - Run the application with hot reload (development)"
	@echo "  dev               - Start full development environment (local)"
	@echo "  docker-dev        - Start full development environment (Docker)"
	@echo "  test              - Run tests"
	@echo "  clean             - Clean build artifacts"
	@echo ""
	@echo "Migration Commands (Dynamic Module Support):"
	@echo "  migrate           - Show available modules and migration help"
	@echo "  migrate-up        - Run migrations up for all modules"
	@echo "  migrate-down      - Run migrations down for all modules"
	@echo "  migrate-version   - Show migration version for all modules"
	@echo "  migrate-reset     - Reset all module databases"
	@echo ""
	@echo "Docker Development Commands:"
	@echo "  docker-dev-build  - Build development Docker image"
	@echo "  docker-dev-up     - Start Docker development environment"
	@echo "  docker-dev-down   - Stop Docker development environment"
	@echo "  docker-dev-logs   - Show Docker development logs"
	@echo "  docker-dev-shell  - Access application container shell"
	@echo "  docker-dev-clean  - Clean Docker development environment"
	@echo ""
	@echo "Local Development Commands:"
	@echo "  docker-up         - Start PostgreSQL with Docker"
	@echo "  docker-down       - Stop PostgreSQL Docker container"
	@echo ""
	@echo "Database Management:"
	@echo "  create-databases  - Create databases for enabled modules"
	@echo ""
	@echo "Migration Examples:"
	@echo "  ./scripts/migrate.sh -m customer -a up      # Migrate customer module up"
	@echo "  ./scripts/migrate.sh -m all -a version      # Show all module versions"
	@echo "  ./scripts/migrate.sh -m customer -a create -n add_email  # Create new migration"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

# Run the application
run:
	@echo "Running application..."
	go run cmd/api/main.go

# Run the application with hot reload (development)
run-dev:
	@echo "Starting development server with hot reload..."
	@echo "Press Ctrl+C to stop"
	air

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf tmp/
	rm -f build-errors.log

# Dynamic migration commands using new migration tool
migrate:
	@echo "Available modules and migration help:"
	@./scripts/migrate.sh

migrate-up:
	@echo "Running migrations up for all modules..."
	@./scripts/migrate.sh -m all -a up

migrate-down:
	@echo "Running migrations down for all modules..."
	@./scripts/migrate.sh -m all -a down

migrate-version:
	@echo "Checking migration versions for all modules..."
	@./scripts/migrate.sh -m all -a version

migrate-reset:
	@echo "Resetting all module databases..."
	@./scripts/migrate.sh -m all -a reset

# Database management
create-databases:
	@echo "Creating databases for enabled modules..."
	@./scripts/create-databases.sh

# Docker commands for PostgreSQL (multiple databases)
docker-up:
	@echo "Starting PostgreSQL with Docker..."
	docker run --name postgres-modular-monolith \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=postgres \
		-p 5433:5432 \
		-d postgres:15-alpine
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	@echo "Creating module databases..."
	@docker exec postgres-modular-monolith psql -U postgres -c "CREATE DATABASE modular_monolith_customer;" || true
	@docker exec postgres-modular-monolith psql -U postgres -c "CREATE DATABASE modular_monolith_order;" || true
	@docker exec postgres-modular-monolith psql -U postgres -c "CREATE DATABASE modular_monolith_product;" || true
	@echo "Module databases created successfully!"

docker-down:
	@echo "Stopping PostgreSQL Docker container..."
	docker stop postgres-modular-monolith || true
	docker rm postgres-modular-monolith || true

# Development setup
dev-setup: docker-up
	@echo "Development environment is ready!"
	@echo "You can now run: make migrate-all-up && make run"

# Start full development environment (Local)
dev:
	@echo "Starting full development environment (Local)..."
	./scripts/dev.sh

# Start full development environment (Docker)
docker-dev:
	@echo "Starting full development environment (Docker)..."
	./scripts/docker-dev.sh

# Docker development commands
docker-dev-build:
	@echo "Building development Docker image..."
	docker compose -f docker/docker-compose.dev.yml build

docker-dev-up:
	@echo "Starting Docker development environment..."
	docker compose -f docker/docker-compose.dev.yml up -d postgres redis vault
	@echo "Waiting for services to be ready..."
	@sleep 15
	@echo "Running migrations..."
	docker compose -f docker/docker-compose.dev.yml run --rm migrate
	@echo "Starting application with hot reload..."
	docker compose -f docker/docker-compose.dev.yml up app

docker-dev-down:
	@echo "Stopping Docker development environment..."
	docker compose -f docker/docker-compose.dev.yml down

docker-dev-logs:
	@echo "Showing Docker development logs..."
	docker compose -f docker/docker-compose.dev.yml logs -f

docker-dev-shell:
	@echo "Accessing application container shell..."
	docker compose -f docker/docker-compose.dev.yml exec app sh

docker-dev-migrate:
	@echo "Running migrations in Docker..."
	docker compose -f docker/docker-compose.dev.yml run --rm migrate

docker-dev-clean:
	@echo "Cleaning Docker development environment..."
	docker compose -f docker/docker-compose.dev.yml down -v
	docker system prune -f

# Vault development commands
vault-dev:
	@echo "Starting Vault development environment..."
	docker compose -f docker/docker-compose.dev.yml up -d vault
	@echo "Waiting for Vault to be ready..."
	@sleep 10
	@echo "Initializing Vault with sample secrets..."
	docker compose -f docker/docker-compose.dev.yml --profile vault-init run --rm vault-init
	@echo "Vault is ready! UI available at: http://localhost:8200/ui"
	@echo "Login token: dev-root-token"

vault-dev-with-app:
	@echo "Starting full development environment with Vault..."
	@echo "Step 1: Starting Vault..."
	docker compose -f docker-compose.dev.yml up -d vault
	@echo "Waiting for Vault to be ready..."
	@sleep 10
	@echo "Step 2: Initializing Vault..."
	docker compose -f docker-compose.dev.yml --profile vault-init run --rm vault-init
	@echo "Step 3: Starting PostgreSQL..."
	docker compose -f docker-compose.dev.yml up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "Step 4: Running migrations..."
	docker compose -f docker-compose.dev.yml run --rm migrate
	@echo "Step 5: Starting application with Vault enabled..."
	@VAULT_ENABLED=true docker compose -f docker-compose.dev.yml up app

vault-ui:
	@echo "Opening Vault UI..."
	@echo "URL: http://localhost:8200/ui"
	@echo "Token: dev-root-token"
	@open http://localhost:8200/ui || xdg-open http://localhost:8200/ui || echo "Please open http://localhost:8200/ui manually"

vault-status:
	@echo "Checking Vault status..."
	@docker compose -f docker-compose.dev.yml exec vault vault status || echo "Vault is not running"

vault-secrets:
	@echo "Listing Vault secrets..."
	@docker compose -f docker-compose.dev.yml exec vault vault kv list kv/ || echo "No secrets found or Vault is not running"

vault-get-secret:
	@echo "Getting application secrets from Vault..."
	@echo "📱 App secrets:"
	@docker compose -f docker-compose.dev.yml exec vault sh -c "VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=dev-root-token vault kv get kv/app" || echo "App secrets not found"
	@echo ""
	@echo "👤 Customer module secrets:"
	@docker compose -f docker-compose.dev.yml exec vault sh -c "VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=dev-root-token vault kv get kv/modules/customer" || echo "Customer secrets not found"
	@echo ""
	@echo "📦 Order module secrets:"
	@docker compose -f docker-compose.dev.yml exec vault sh -c "VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=dev-root-token vault kv get kv/modules/order" || echo "Order secrets not found"
	@echo ""
	@echo "🛍️ Product module secrets:"
	@docker compose -f docker-compose.dev.yml exec vault sh -c "VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=dev-root-token vault kv get kv/modules/product" || echo "Product secrets not found"

vault-clean:
	@echo "Cleaning Vault data..."
	docker compose -f docker-compose.dev.yml stop vault vault-init
	docker volume rm modular-monolith_vault-data modular-monolith_vault-logs || true

# Module-specific commands
customer-dev:
	@echo "Customer module development commands:"
	@echo "  make migrate-customer-up      - Run customer migrations"
	@echo "  make migrate-customer-down    - Rollback customer migrations"
	@echo "  make migrate-customer-version - Check customer migration version"
	@echo "  make migrate-customer-reset   - Reset customer database"
	@echo "  make migrate-create-customer  - Create new customer migration"

order-dev:
	@echo "Order module development commands:"
	@echo "  make migrate-order-up      - Run order migrations"
	@echo "  make migrate-order-down    - Rollback order migrations"
	@echo "  make migrate-order-version - Check order migration version"
	@echo "  make migrate-order-reset   - Reset order database"
	@echo "  make migrate-create-order  - Create new order migration"

# Show all available commands
help-modules:
	@echo "Module-specific migration commands:"
	@echo "  Customer module:"
	@echo "    migrate-customer-up      - Run customer migrations"
	@echo "    migrate-customer-down    - Rollback customer migrations"
	@echo "    migrate-customer-version - Check customer migration version"
	@echo "    migrate-customer-reset   - Reset customer database"
	@echo "    migrate-create-customer  - Create new customer migration"
	@echo ""
	@echo "  Order module:"
	@echo "    migrate-order-up      - Run order migrations"
	@echo "    migrate-order-down    - Rollback order migrations"
	@echo "    migrate-order-version - Check order migration version"
	@echo "    migrate-order-reset   - Reset order database"
	@echo "    migrate-create-order  - Create new order migration"
	@echo ""
	@echo "  All modules:"
	@echo "    migrate-all-up      - Run all migrations"
	@echo "    migrate-all-down    - Rollback all migrations"
	@echo "    migrate-all-version - Check all migration versions"
	@echo "    migrate-all-reset   - Reset all databases" 