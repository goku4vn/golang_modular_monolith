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
	@echo "Docker Development Commands:"
	@echo "  docker-dev-build  - Build development Docker image"
	@echo "  docker-dev-up     - Start Docker development environment"
	@echo "  docker-dev-down   - Stop Docker development environment"
	@echo "  docker-dev-logs   - Show Docker development logs"
	@echo "  docker-dev-shell  - Access application container shell"
	@echo "  docker-dev-migrate- Run migrations in Docker"
	@echo "  docker-dev-clean  - Clean Docker development environment"
	@echo ""
	@echo "Local Development Commands:"
	@echo "  migrate-up        - Run database migrations up"
	@echo "  migrate-down      - Run database migrations down"
	@echo "  docker-up         - Start PostgreSQL with Docker"
	@echo "  docker-down       - Stop PostgreSQL Docker container"

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

# Database migrations using golang-migrate/migrate
migrate-customer-up:
	@echo "Running customer migrations up..."
	@go run cmd/migrate/main.go -module=customer -action=up

migrate-customer-down:
	@echo "Running customer migrations down..."
	@go run cmd/migrate/main.go -module=customer -action=down

migrate-customer-version:
	@echo "Checking customer migration version..."
	@go run cmd/migrate/main.go -module=customer -action=version

migrate-customer-reset:
	@echo "Resetting customer database..."
	@go run cmd/migrate/main.go -module=customer -action=reset

migrate-order-up:
	@echo "Running order migrations up..."
	@go run cmd/migrate/main.go -module=order -action=up

migrate-order-down:
	@echo "Running order migrations down..."
	@go run cmd/migrate/main.go -module=order -action=down

migrate-order-version:
	@echo "Checking order migration version..."
	@go run cmd/migrate/main.go -module=order -action=version

migrate-order-reset:
	@echo "Resetting order database..."
	@go run cmd/migrate/main.go -module=order -action=reset

migrate-all-up:
	@echo "Running all migrations up..."
	@go run cmd/migrate/main.go -module=all -action=up

migrate-all-down:
	@echo "Running all migrations down..."
	@go run cmd/migrate/main.go -module=all -action=down

migrate-all-version:
	@echo "Checking all migration versions..."
	@go run cmd/migrate/main.go -module=all -action=version

migrate-all-reset:
	@echo "Resetting all databases..."
	@go run cmd/migrate/main.go -module=all -action=reset

# Create new migrations
migrate-create-customer:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/modules/customer/migrations -seq $$name

migrate-create-order:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/modules/order/migrations -seq $$name

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
	docker compose -f docker-compose.dev.yml build

docker-dev-up:
	@echo "Starting Docker development environment..."
	docker compose -f docker-compose.dev.yml up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "Running migrations..."
	docker compose -f docker-compose.dev.yml run --rm migrate
	@echo "Starting application with hot reload..."
	docker compose -f docker-compose.dev.yml up app

docker-dev-down:
	@echo "Stopping Docker development environment..."
	docker compose -f docker-compose.dev.yml down

docker-dev-logs:
	@echo "Showing Docker development logs..."
	docker compose -f docker-compose.dev.yml logs -f

docker-dev-shell:
	@echo "Accessing application container shell..."
	docker compose -f docker-compose.dev.yml exec app sh

docker-dev-migrate:
	@echo "Running migrations in Docker..."
	docker compose -f docker-compose.dev.yml run --rm migrate

docker-dev-clean:
	@echo "Cleaning Docker development environment..."
	docker compose -f docker-compose.dev.yml down -v
	docker system prune -f

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