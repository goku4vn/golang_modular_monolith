# Project Structure

Mô tả chi tiết cấu trúc source code của Modular Monolith.

## Overview

Modular Monolith sử dụng **Domain-Driven Design (DDD)** với **Clean Architecture**:
- **Modules**: Tách biệt theo business domain
- **Layers**: Presentation → Application → Domain → Infrastructure
- **Shared**: Common utilities và infrastructure
- **Configuration**: Flexible module configuration system

## Root Directory Structure

```
modular-monolith/
├── cmd/                    # Application entry points
├── internal/               # Private application code
├── config/                 # Configuration files
├── scripts/                # Build and deployment scripts
├── docs/                   # Documentation
├── docker/                 # Docker configurations
├── Makefile               # Build commands
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
└── README.md              # Project overview
```

## Command Directory (`cmd/`)

```
cmd/
├── api/                   # Main API server
│   └── main.go           # Application entry point
├── migrate/              # Database migration tool
│   └── main.go           # Migration CLI
└── tools/                # Development tools
    └── list-modules.go   # Module listing utility
```

### Entry Points
- **`cmd/api/main.go`**: Main HTTP API server
- **`cmd/migrate/main.go`**: Database migration CLI tool
- **`cmd/tools/`**: Development and maintenance tools

## Internal Directory (`internal/`)

```
internal/
├── modules/              # Business modules
│   ├── customer/        # Customer domain module
│   ├── order/           # Order domain module
│   └── user/            # User domain module
└── shared/              # Shared components
    ├── domain/          # Shared domain logic
    ├── infrastructure/  # Shared infrastructure
    └── application/     # Shared application logic
```

## Module Structure

Mỗi module tuân theo **Clean Architecture**:

```
internal/modules/{module}/
├── module.yaml           # Module configuration
├── migrations/           # Database migrations
│   ├── 001_create_table.up.sql
│   └── 001_create_table.down.sql
├── domain/              # Domain layer (business logic)
│   ├── entities/        # Domain entities
│   ├── repositories/    # Repository interfaces
│   ├── services/        # Domain services
│   └── events/          # Domain events
├── application/         # Application layer (use cases)
│   ├── commands/        # Command handlers
│   ├── queries/         # Query handlers
│   ├── services/        # Application services
│   └── dto/             # Data transfer objects
├── infrastructure/      # Infrastructure layer
│   ├── database/        # Database implementations
│   ├── http/            # HTTP handlers
│   ├── repositories/    # Repository implementations
│   └── external/        # External service clients
└── presentation/        # Presentation layer
    ├── http/            # HTTP controllers
    ├── grpc/            # gRPC handlers
    └── graphql/         # GraphQL resolvers
```

### Example: Customer Module

```
internal/modules/customer/
├── module.yaml                    # Configuration
├── migrations/
│   ├── 001_create_customers_table.up.sql
│   └── 001_create_customers_table.down.sql
├── domain/
│   ├── entities/
│   │   └── customer.go           # Customer entity
│   ├── repositories/
│   │   └── customer_repository.go # Repository interface
│   ├── services/
│   │   └── customer_service.go   # Domain service
│   └── events/
│       └── customer_created.go   # Domain event
├── application/
│   ├── commands/
│   │   ├── create_customer.go    # Create customer command
│   │   └── update_customer.go    # Update customer command
│   ├── queries/
│   │   ├── get_customer.go       # Get customer query
│   │   └── list_customers.go     # List customers query
│   └── services/
│       └── customer_app_service.go # Application service
├── infrastructure/
│   ├── database/
│   │   └── config.go             # Database configuration
│   ├── repositories/
│   │   └── postgres_customer_repository.go # PostgreSQL implementation
│   └── http/
│       └── routes.go             # HTTP route registration
└── presentation/
    └── http/
        └── customer_handler.go   # HTTP handlers
```

## Shared Components (`internal/shared/`)

```
internal/shared/
├── domain/                       # Shared domain logic
│   ├── events/                  # Domain event system
│   ├── errors/                  # Common error types
│   └── values/                  # Shared value objects
├── infrastructure/              # Shared infrastructure
│   ├── config/                  # Configuration management
│   │   ├── modules.go          # Module configuration
│   │   └── database.go         # Database configuration
│   ├── database/               # Database utilities
│   │   ├── connection.go       # Connection management
│   │   └── migration.go        # Migration utilities
│   ├── http/                   # HTTP infrastructure
│   │   ├── server.go           # HTTP server
│   │   ├── middleware/         # HTTP middleware
│   │   └── handlers/           # Common handlers
│   ├── logging/                # Logging utilities
│   └── monitoring/             # Monitoring and metrics
└── application/                # Shared application logic
    ├── bus/                    # Command/Query bus
    ├── events/                 # Event handling
    └── services/               # Shared services
```

## Configuration Directory (`config/`)

```
config/
├── modules.yaml              # Module configuration
├── app.yaml                  # Application configuration
└── environments/             # Environment-specific configs
    ├── development.yaml
    ├── staging.yaml
    └── production.yaml
```

### Configuration Files
- **`modules.yaml`**: Module enable/disable và configuration
- **`app.yaml`**: Application-level settings
- **`environments/`**: Environment-specific overrides

## Scripts Directory (`scripts/`)

```
scripts/
├── create-databases.sh       # Database creation script
├── docker-dev.sh            # Development environment setup
├── build.sh                 # Build script
└── deploy.sh                # Deployment script
```

## Docker Directory (`docker/`)

```
docker/
├── postgres/                # PostgreSQL configuration
│   └── Dockerfile          # Clean PostgreSQL image
├── app/                     # Application Docker config
│   └── Dockerfile          # Multi-stage build
└── docker-compose.dev.yml  # Development compose file
```

## Architecture Layers

### 1. Domain Layer
**Location**: `internal/modules/{module}/domain/`
**Purpose**: Core business logic, entities, và domain rules
**Dependencies**: None (pure business logic)

```go
// Example: Customer entity
type Customer struct {
    ID    CustomerID
    Email Email
    Name  string
}

func (c *Customer) ChangeEmail(newEmail Email) error {
    // Business validation logic
    if !newEmail.IsValid() {
        return ErrInvalidEmail
    }
    c.Email = newEmail
    return nil
}
```

### 2. Application Layer
**Location**: `internal/modules/{module}/application/`
**Purpose**: Use cases, commands, queries
**Dependencies**: Domain layer only

```go
// Example: Create customer use case
type CreateCustomerCommand struct {
    Email string
    Name  string
}

func (h *CreateCustomerHandler) Handle(cmd CreateCustomerCommand) error {
    customer := domain.NewCustomer(cmd.Email, cmd.Name)
    return h.repository.Save(customer)
}
```

### 3. Infrastructure Layer
**Location**: `internal/modules/{module}/infrastructure/`
**Purpose**: External concerns (database, HTTP, etc.)
**Dependencies**: Domain và Application layers

```go
// Example: PostgreSQL repository
type PostgresCustomerRepository struct {
    db *sql.DB
}

func (r *PostgresCustomerRepository) Save(customer *domain.Customer) error {
    query := "INSERT INTO customers (id, email, name) VALUES ($1, $2, $3)"
    _, err := r.db.Exec(query, customer.ID, customer.Email, customer.Name)
    return err
}
```

### 4. Presentation Layer
**Location**: `internal/modules/{module}/presentation/`
**Purpose**: API endpoints, controllers
**Dependencies**: Application layer

```go
// Example: HTTP handler
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
    var req CreateCustomerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    cmd := CreateCustomerCommand{
        Email: req.Email,
        Name:  req.Name,
    }
    
    if err := h.commandHandler.Handle(cmd); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, gin.H{"message": "Customer created"})
}
```

## Module Lifecycle

### 1. Module Registration
```go
// internal/modules/{module}/module.go
func RegisterModule(router *gin.Engine, db *sql.DB) {
    // Initialize repositories
    customerRepo := repositories.NewPostgresCustomerRepository(db)
    
    // Initialize services
    customerService := services.NewCustomerService(customerRepo)
    
    // Initialize handlers
    customerHandler := handlers.NewCustomerHandler(customerService)
    
    // Register routes
    customerHandler.RegisterRoutes(router)
}
```

### 2. Module Configuration
```yaml
# internal/modules/customer/module.yaml
enabled: true
database:
  host: "${POSTGRES_HOST:localhost}"
  port: "${POSTGRES_PORT:5432}"
  name: "modular_monolith_customer"
migration:
  enabled: true
  path: "./migrations"
http:
  prefix: "/api/v1/customers"
```

### 3. Module Loading
```go
// cmd/api/main.go
func main() {
    config := loadConfig()
    
    for moduleName, moduleConfig := range config.Modules {
        if moduleConfig.Enabled {
            module := modules.LoadModule(moduleName, moduleConfig)
            module.Register(router, databases[moduleName])
        }
    }
}
```

## Dependency Flow

```
Presentation Layer (HTTP/gRPC)
        ↓
Application Layer (Use Cases)
        ↓
Domain Layer (Business Logic)
        ↑
Infrastructure Layer (Database/External)
```

### Dependency Rules
1. **Domain**: No dependencies (pure business logic)
2. **Application**: Depends on Domain only
3. **Infrastructure**: Depends on Domain và Application
4. **Presentation**: Depends on Application only

## Best Practices

### 1. Module Independence
- Modules không được import trực tiếp từ nhau
- Communication qua events hoặc shared interfaces
- Mỗi module có database riêng

### 2. Clean Architecture
- Dependency inversion principle
- Interface segregation
- Single responsibility per layer

### 3. Configuration Management
- Environment variables cho sensitive data
- YAML files cho structure configuration
- Module-level configuration overrides

### 4. Testing Structure
```
internal/modules/customer/
├── domain/
│   └── entities/
│       ├── customer.go
│       └── customer_test.go      # Unit tests
├── application/
│   └── commands/
│       ├── create_customer.go
│       └── create_customer_test.go # Use case tests
└── infrastructure/
    └── repositories/
        ├── postgres_customer_repository.go
        └── postgres_customer_repository_test.go # Integration tests
```

## Adding New Modules

### 1. Create Module Structure
```bash
mkdir -p internal/modules/new_module/{domain,application,infrastructure,presentation}
mkdir -p internal/modules/new_module/{entities,repositories,services}
mkdir -p internal/modules/new_module/migrations
```

### 2. Create Module Configuration
```yaml
# internal/modules/new_module/module.yaml
enabled: true
database:
  name: "modular_monolith_new_module"
migration:
  enabled: true
```

### 3. Add to Central Configuration
```yaml
# config/modules.yaml
modules:
  new_module: true
```

### 4. Create Database và Migrations
```bash
make create-databases
make migrate-create MODULE=new_module NAME=initial_schema
```

Cấu trúc này đảm bảo:
- **Modularity**: Modules độc lập và có thể tái sử dụng
- **Scalability**: Dễ dàng thêm modules mới
- **Maintainability**: Code được tổ chức rõ ràng theo layers
- **Testability**: Mỗi layer có thể test độc lập 