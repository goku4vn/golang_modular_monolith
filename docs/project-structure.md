# Project Structure

Mô tả chi tiết cấu trúc source code của Modular Monolith với **Module-Based Auto-Registration Architecture**.

## Overview

Modular Monolith sử dụng **Domain-Driven Design (DDD)** với **Clean Architecture** và **Module Auto-Registration**:
- **Modules**: Tách biệt theo business domain với auto-registration
- **Layers**: Presentation → Application → Domain → Infrastructure
- **Shared**: Common utilities và infrastructure
- **Configuration**: Flexible module configuration system
- **Auto-Discovery**: Modules tự đăng ký và load based on config

## Root Directory Structure

```
modular-monolith/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── modules/           # Business modules + centralized import
│   └── shared/            # Shared components
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
│   └── main.go           # Application entry point (module-based)
├── migrate/              # Database migration tool
│   └── main.go           # Migration CLI
└── tools/                # Development tools
    └── list-modules.go   # Module listing utility
```

### Entry Points
- **`cmd/api/main.go`**: Main HTTP API server với module auto-loading
- **`cmd/migrate/main.go`**: Database migration CLI tool
- **`cmd/tools/`**: Development and maintenance tools

## Internal Directory (`internal/`)

```
internal/
├── modules/              # Business modules + centralized management
│   ├── modules.go       # ✨ Centralized module import & registration
│   ├── customer/        # Customer domain module
│   ├── order/           # Order domain module
│   └── user/            # User domain module
└── shared/              # Shared components
    ├── domain/          # Shared domain logic + Module interface
    ├── infrastructure/  # Shared infrastructure + Module registry
    └── application/     # Shared application logic
```

### Module Centralized Management

**`internal/modules/modules.go`** - Centralized module import:
```go
package modules

import (
    // Import all modules to trigger auto-registration via init() functions
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    _ "golang_modular_monolith/internal/modules/user"
)

// InitializeAllModules ensures all modules are imported and registered
func InitializeAllModules() {
    // This function exists to ensure this package is imported
    // and all module init() functions are called
}
```

## Module Structure (Auto-Registration)

Mỗi module tuân theo **Clean Architecture** với **Auto-Registration**:

```
internal/modules/{module}/
├── module.go             # ✨ Module implementation + auto-registration
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
│   ├── command_handlers/ # Command handlers
│   ├── query_handlers/  # Query handlers
│   ├── services/        # Application services
│   └── dto/             # Data transfer objects
├── infrastructure/      # Infrastructure layer
│   ├── database/        # Database implementations
│   ├── http/            # HTTP handlers + route registration
│   ├── persistence/     # Repository implementations
│   └── external/        # External service clients
└── presentation/        # Presentation layer
    ├── http/            # HTTP controllers
    ├── grpc/            # gRPC handlers
    └── graphql/         # GraphQL resolvers
```

### Example: Customer Module with Auto-Registration

```
internal/modules/customer/
├── module.go                      # ✨ CustomerModule + auto-registration
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
│   ├── command_handlers/
│   │   ├── create_customer.go    # Create customer command handler
│   │   └── update_customer.go    # Update customer command handler
│   ├── query_handlers/
│   │   ├── get_customer.go       # Get customer query handler
│   │   └── list_customers.go     # List customers query handler
│   └── services/
│       └── customer_app_service.go # Application service
├── infrastructure/
│   ├── database/
│   │   └── connection.go         # Database connection helper
│   ├── persistence/
│   │   └── customer_repository.go # PostgreSQL implementation
│   └── http/
│       ├── handlers/             # HTTP handlers
│       └── routes.go             # HTTP route registration
└── presentation/
    └── http/
        └── customer_handler.go   # HTTP handlers
```

**Module Implementation Example:**
```go
// internal/modules/customer/module.go
package customer

import (
    "golang_modular_monolith/internal/shared/domain"
    "golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register customer module on package import
func init() {
    registry.RegisterModule("customer", func() domain.Module {
        return NewCustomerModule()
    })
}

type CustomerModule struct {
    name     string
    handler  *handlers.CustomerHandler
    eventBus domain.EventBus
}

func (m *CustomerModule) Name() string { return m.name }
func (m *CustomerModule) Initialize(deps domain.ModuleDependencies) error { /* ... */ }
func (m *CustomerModule) RegisterRoutes(router *gin.RouterGroup) { /* ... */ }
func (m *CustomerModule) Health(ctx context.Context) error { /* ... */ }
func (m *CustomerModule) Start(ctx context.Context) error { /* ... */ }
func (m *CustomerModule) Stop(ctx context.Context) error { /* ... */ }
```

## Shared Components (`internal/shared/`)

```
internal/shared/
├── domain/                       # Shared domain logic
│   ├── module.go                # ✨ Module interface + ModuleRegistry
│   ├── events/                  # Domain event system
│   ├── errors/                  # Common error types
│   └── values/                  # Shared value objects
├── infrastructure/              # Shared infrastructure
│   ├── config/                  # Configuration management
│   │   ├── modules.go          # Module configuration
│   │   └── database.go         # Database configuration
│   ├── database/               # Database utilities
│   │   ├── manager.go          # Database manager (global)
│   │   └── migration.go        # Migration utilities
│   ├── registry/               # ✨ Module management
│   │   └── module_manager.go   # Unified module factory + loader
│   ├── eventbus/               # Event bus implementation
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

### Module Interface & Registry

**`internal/shared/domain/module.go`**:
```go
type Module interface {
    Name() string
    Initialize(deps ModuleDependencies) error
    RegisterRoutes(router *gin.RouterGroup)
    Health(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}

type ModuleRegistry struct {
    modules map[string]Module
}
```

**`internal/shared/infrastructure/registry/module_manager.go`**:
```go
type ModuleManager struct {
    registry *domain.ModuleRegistry
    creators map[string]ModuleCreator
}

// Unified functionality: Factory + Loader + Registry
func (m *ModuleManager) RegisterModule(name string, creator ModuleCreator)
func (m *ModuleManager) CreateModule(name string) (domain.Module, error)
func (m *ModuleManager) LoadEnabledModules(cfg *config.Config) error
func (m *ModuleManager) GetRegistry() *domain.ModuleRegistry
```

## Configuration Directory (`config/`)

```
config/
├── modules.yaml              # Module configuration (enable/disable)
├── app.yaml                  # Application configuration
└── environments/             # Environment-specific configs
    ├── development.yaml
    ├── staging.yaml
    └── production.yaml
```

### Module Configuration Example
```yaml
# config/modules.yaml
modules:
  customer: true    # ✅ Enabled - will be loaded
  order: true       # ✅ Enabled - will be loaded
  user: false       # ❌ Disabled - will be skipped
```

## Application Entry Point (Module-Based)

### Main Application Flow
```go
// cmd/api/main.go - Module-based architecture
func main() {
    // 1. Initialize all modules (triggers auto-registration)
    modules.InitializeAllModules()

    // 2. Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // 3. Initialize database manager
    if err := initDatabases(cfg); err != nil {
        log.Fatalf("Failed to initialize databases: %v", err)
    }

    // 4. Initialize event bus
    eventBus := eventbus.NewInMemoryEventBus()

    // 5. Load enabled modules dynamically
    moduleRegistry, err := initModules(cfg, eventBus)
    if err != nil {
        log.Fatalf("Failed to initialize modules: %v", err)
    }

    // 6. Initialize router with dynamic route registration
    router := initRouter(cfg, moduleRegistry)

    // 7. Start all modules
    ctx := context.Background()
    if err := moduleRegistry.StartAll(ctx); err != nil {
        log.Fatalf("Failed to start modules: %v", err)
    }

    // 8. Start server
    if err := router.Run(cfg.GetServerAddress()); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func initModules(cfg *config.Config, eventBus domain.EventBus) (*domain.ModuleRegistry, error) {
    // Get global module manager
    manager := registry.GetGlobalManager()

    // Load enabled modules from configuration
    if err := manager.LoadEnabledModules(cfg); err != nil {
        return nil, err
    }

    // Get module registry
    moduleRegistry := manager.GetRegistry()

    // Initialize all modules with dependencies
    deps := domain.ModuleDependencies{
        EventBus: eventBus,
        Config:   cfg,
    }

    if err := moduleRegistry.InitializeAll(deps); err != nil {
        return nil, err
    }

    return moduleRegistry, nil
}

func initRouter(cfg *config.Config, moduleRegistry *domain.ModuleRegistry) *gin.Engine {
    router := gin.New()
    
    // Add middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())

    // Add health check with module health
    router.GET("/health", healthCheckHandler(cfg, moduleRegistry))

    // API routes - Dynamic registration for all enabled modules
    api := router.Group("/api/v1")
    {
        moduleRegistry.RegisterAllRoutes(api)
    }

    return router
}
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
**Purpose**: Use cases, command/query handlers
**Dependencies**: Domain layer only

```go
// Example: Create customer command handler
type CreateCustomerHandler struct {
    repository    domain.CustomerRepository
    domainService domain.CustomerDomainService
    eventBus      domain.EventBus
}

func (h *CreateCustomerHandler) Handle(ctx context.Context, cmd CreateCustomerCommand) error {
    customer := domain.NewCustomer(cmd.Email, cmd.Name)
    
    if err := h.repository.Save(ctx, customer); err != nil {
        return err
    }
    
    // Publish domain event
    event := domain.NewCustomerCreatedEvent(customer.ID, customer.Email)
    h.eventBus.Publish(event)
    
    return nil
}
```

### 3. Infrastructure Layer
**Location**: `internal/modules/{module}/infrastructure/`
**Purpose**: External concerns (database, HTTP, etc.)
**Dependencies**: Domain và Application layers

```go
// Example: PostgreSQL repository using Database Manager
func NewPostgreSQLCustomerRepositoryFromManager() (*PostgreSQLCustomerRepository, error) {
    db, err := customerdb.GetCustomerDB()
    if err != nil {
        return nil, fmt.Errorf("failed to get customer database: %w", err)
    }

    return &PostgreSQLCustomerRepository{db: db}, nil
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
    
    if err := h.createHandler.Handle(c.Request.Context(), cmd); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, gin.H{"message": "Customer created"})
}
```

## Module Lifecycle (Auto-Registration)

### 1. Module Auto-Registration
```go
// internal/modules/customer/module.go
func init() {
    registry.RegisterModule("customer", func() domain.Module {
        return NewCustomerModule()
    })
}
```

### 2. Module Configuration
```yaml
# config/modules.yaml
modules:
  customer: true    # Enable customer module
  order: true       # Enable order module
  user: false       # Disable user module
```

### 3. Dynamic Module Loading
```go
// Automatic loading based on configuration
manager := registry.GetGlobalManager()
err := manager.LoadEnabledModules(cfg)  // Only loads enabled modules
```

### 4. Module Lifecycle Management
```go
// Full lifecycle management
moduleRegistry.InitializeAll(deps)  // Initialize all enabled modules
moduleRegistry.StartAll(ctx)        // Start all modules
// ... application runs ...
moduleRegistry.StopAll(ctx)         // Stop all modules (on shutdown)
```

## Dependency Flow

```
Presentation Layer (HTTP/gRPC)
        ↓
Application Layer (Command/Query Handlers)
        ↓
Domain Layer (Business Logic)
        ↑
Infrastructure Layer (Database/External)
        ↑
Shared Infrastructure (Database Manager, Event Bus, Module Registry)
```

### Dependency Rules
1. **Domain**: No dependencies (pure business logic)
2. **Application**: Depends on Domain only
3. **Infrastructure**: Depends on Domain và Application
4. **Presentation**: Depends on Application only
5. **Modules**: Communicate via events, not direct imports

## Best Practices

### 1. Module Independence
- Modules không được import trực tiếp từ nhau
- Communication qua events hoặc shared interfaces
- Mỗi module có database riêng
- Auto-registration via init() functions

### 2. Clean Architecture
- Dependency inversion principle
- Interface segregation
- Single responsibility per layer
- Module interface compliance

### 3. Configuration Management
- Environment variables cho sensitive data
- YAML files cho structure configuration
- Module-level enable/disable configuration
- Config-driven module loading

### 4. Testing Structure
```
internal/modules/customer/
├── module_test.go                 # Module lifecycle tests
├── domain/
│   └── entities/
│       ├── customer.go
│       └── customer_test.go      # Unit tests
├── application/
│   └── command_handlers/
│       ├── create_customer.go
│       └── create_customer_test.go # Use case tests
└── infrastructure/
    └── persistence/
        ├── customer_repository.go
        └── customer_repository_test.go # Integration tests
```

## Adding New Modules

### 1. Create Module Structure
```bash
mkdir -p internal/modules/new_module/{domain,application,infrastructure,presentation}
mkdir -p internal/modules/new_module/{entities,repositories,services}
mkdir -p internal/modules/new_module/migrations
```

### 2. Implement Module Interface
```go
// internal/modules/new_module/module.go
package new_module

import (
    "golang_modular_monolith/internal/shared/domain"
    "golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register module
func init() {
    registry.RegisterModule("new_module", func() domain.Module {
        return NewNewModule()
    })
}

type NewModule struct {
    name string
    // ... other fields
}

// Implement all Module interface methods
func (m *NewModule) Name() string { return m.name }
func (m *NewModule) Initialize(deps domain.ModuleDependencies) error { /* ... */ }
func (m *NewModule) RegisterRoutes(router *gin.RouterGroup) { /* ... */ }
func (m *NewModule) Health(ctx context.Context) error { /* ... */ }
func (m *NewModule) Start(ctx context.Context) error { /* ... */ }
func (m *NewModule) Stop(ctx context.Context) error { /* ... */ }
```

### 3. Add to Centralized Import
```go
// internal/modules/modules.go
import (
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    _ "golang_modular_monolith/internal/modules/user"
    _ "golang_modular_monolith/internal/modules/new_module"  // ✨ Add here
)
```

### 4. Enable in Configuration
```yaml
# config/modules.yaml
modules:
  customer: true
  order: true
  user: false
  new_module: true  # ✨ Enable new module
```

### 5. Create Database và Migrations
```bash
make create-databases
make migrate-create MODULE=new_module NAME=initial_schema
```

**No need to modify main.go!** 🎉

Cấu trúc này đảm bảo:
- **Modularity**: Modules độc lập với auto-registration
- **Scalability**: Dễ dàng thêm modules mới mà không sửa main.go
- **Maintainability**: Code được tổ chức rõ ràng theo layers
- **Testability**: Mỗi layer có thể test độc lập
- **Configuration-Driven**: Enable/disable modules qua config
- **Auto-Discovery**: Modules tự đăng ký và load dynamically 