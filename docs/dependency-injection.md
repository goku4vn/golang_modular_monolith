# Dependency Injection Architecture

Mô tả chi tiết cơ chế Dependency Injection (DI) trong TMM Modular Monolith và cách dependencies được quản lý thực tế.

## 🎯 Overview

TMM đã phát triển từ **Simple Manual Dependency Injection** sang **Module-Based Auto-Registration System** để:
- **Modularity**: Mỗi module tự quản lý dependencies
- **Auto-Discovery**: Modules tự đăng ký, không cần hardcode
- **Scalability**: Dễ dàng thêm modules mới mà không sửa main.go
- **Testability**: Easy mocking và unit testing
- **Configuration-Driven**: Enable/disable modules qua config

## 🏗️ Current DI Architecture (Module-Based)

### **Module Auto-Registration Flow**
```
┌─────────────────────────────────────┐
│         main.go Entry Point         │  ← Application startup
│                                     │
├─────────────────────────────────────┤
│      modules.InitializeAllModules() │  ← Trigger auto-registration
│                                     │
├─────────────────────────────────────┤
│       Module Auto-Registration      │  ← init() functions in modules
│      (ModuleManager Factory)        │
├─────────────────────────────────────┤
│      Config-Driven Module Loading   │  ← Load enabled modules only
│                                     │
├─────────────────────────────────────┤
│      Module Lifecycle Management    │  ← Initialize → Start → Stop
│                                     │
├─────────────────────────────────────┤
│      Database Manager Pattern       │  ← Global database manager
│     (Connection Management)         │
├─────────────────────────────────────┤
│     Repository Factory Pattern      │  ← FromManager() constructors
│    (Database-backed Repositories)   │
├─────────────────────────────────────┤
│       Application Services          │  ← Command/Query handlers
│        (Business Logic)             │
├─────────────────────────────────────┤
│         HTTP Handlers               │  ← Gin HTTP handlers
│      (Presentation Layer)           │
└─────────────────────────────────────┘
```

## 📦 Module-Based Implementation

### **1. Module Interface Definition**
```go
// internal/shared/domain/module.go
type Module interface {
    // Name returns the module name
    Name() string

    // Initialize initializes the module with dependencies
    Initialize(deps ModuleDependencies) error

    // RegisterRoutes registers HTTP routes for the module
    RegisterRoutes(router *gin.RouterGroup)

    // Health checks if the module is healthy
    Health(ctx context.Context) error

    // Start starts the module (optional lifecycle method)
    Start(ctx context.Context) error

    // Stop stops the module (optional lifecycle method)
    Stop(ctx context.Context) error
}

type ModuleDependencies struct {
    EventBus EventBus
    Config   interface{} // Module-specific config
}
```

### **2. Module Auto-Registration System**
```go
// internal/shared/infrastructure/registry/module_manager.go
type ModuleManager struct {
    registry *domain.ModuleRegistry
    creators map[string]ModuleCreator
}

type ModuleCreator func() domain.Module

// Global manager instance
var globalManager = NewModuleManager()

// RegisterModule registers a module creator globally
func RegisterModule(name string, creator ModuleCreator) {
    globalManager.RegisterModule(name, creator)
}
```

### **3. Customer Module Implementation**
```go
// internal/modules/customer/module.go
package customer

import (
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

func (m *CustomerModule) Initialize(deps domain.ModuleDependencies) error {
    // Create repositories using factory pattern
    customerRepo, err := persistence.NewPostgreSQLCustomerRepositoryFromManager()
    if err != nil {
        return fmt.Errorf("failed to create customer repository: %w", err)
    }

    // Create domain services
    customerDomainService := persistence.NewCustomerDomainService(customerRepo)

    // Create command/query handlers
    createCustomerHandler := commandhandlers.NewCreateCustomerHandler(
        customerRepo, customerDomainService, m.eventBus)

    // Create HTTP handlers
    m.handler = handlers.NewCustomerHandler(
        createCustomerHandler, getCustomerHandler, 
        listCustomersHandler, searchCustomersHandler)

    return nil
}

func (m *CustomerModule) RegisterRoutes(router *gin.RouterGroup) {
    customerhttp.RegisterCustomerRoutes(router, m.handler)
}
```

### **4. Centralized Module Import**
```go
// internal/modules/modules.go
package modules

import (
    // Import all modules to trigger auto-registration via init() functions
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    _ "golang_modular_monolith/internal/modules/user"
)

func InitializeAllModules() {
    // This function exists to ensure this package is imported
    // and all module init() functions are called
}
```

### **5. Clean Main Application Entry Point**
```go
// cmd/api/main.go - Refactored implementation
func main() {
    // Initialize all modules (triggers auto-registration)
    modules.InitializeAllModules()

    // Load configuration using Viper
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize database manager with config
    if err := initDatabases(cfg); err != nil {
        log.Fatalf("Failed to initialize databases: %v", err)
    }

    // Initialize event bus
    eventBus := eventbus.NewInMemoryEventBus()

    // Load enabled modules
    moduleRegistry, err := initModules(cfg, eventBus)
    if err != nil {
        log.Fatalf("Failed to initialize modules: %v", err)
    }

    // Initialize router with module registry
    router := initRouter(cfg, moduleRegistry)

    // Start modules
    ctx := context.Background()
    if err := moduleRegistry.StartAll(ctx); err != nil {
        log.Fatalf("Failed to start modules: %v", err)
    }

    // Start server
    if err := router.Run(cfg.GetServerAddress()); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

## 🔧 Module Loading and Lifecycle

### **Module Loading Process**
```go
// cmd/api/main.go - Module initialization
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
```

### **Configuration-Driven Loading**
```yaml
# config/modules.yaml
modules:
  customer: true    # ✅ Loaded and registered
  order: true       # ✅ Loaded and registered  
  user: false       # ⏭️ Skipped
```

### **Module Lifecycle Management**
```go
// Module lifecycle in main.go
moduleRegistry.InitializeAll(deps)  // Initialize all enabled modules
moduleRegistry.StartAll(ctx)        // Start all modules
// ... application runs ...
moduleRegistry.StopAll(ctx)         // Stop all modules (on shutdown)
```

## 🌐 HTTP Layer Integration

### **Dynamic Route Registration**
```go
// cmd/api/main.go - Router initialization
func initRouter(cfg *config.Config, moduleRegistry *domain.ModuleRegistry) *gin.Engine {
    router := gin.New()
    
    // Add middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(corsMiddleware())

    // Add health check with module health
    router.GET("/health", healthCheckHandler(cfg, moduleRegistry))

    // API routes - Dynamic registration
    api := router.Group("/api/v1")
    {
        // Register routes for all enabled modules
        moduleRegistry.RegisterAllRoutes(api)
    }

    return router
}
```

### **Enhanced Health Check**
```go
func healthCheckHandler(cfg *config.Config, moduleRegistry *domain.ModuleRegistry) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check module health
        ctx := context.Background()
        moduleHealth := moduleRegistry.HealthCheckAll(ctx)

        // Determine overall status
        status := "healthy"
        for _, err := range moduleHealth {
            if err != nil {
                status = "unhealthy"
                break
            }
        }

        response := gin.H{
            "status":        status,
            "modules":       moduleRegistry.GetModuleNames(),
            "module_health": moduleHealth,
            "message":       "🚀 Modular system with dynamic module loading!",
        }

        if status == "healthy" {
            c.JSON(200, response)
        } else {
            c.JSON(503, response)
        }
    }
}
```

## 🔄 Event Bus Integration

### **Current Event Bus Implementation**
```go
// internal/shared/infrastructure/eventbus/in_memory_event_bus.go
type InMemoryEventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
    return &InMemoryEventBus{
        handlers: make(map[string][]EventHandler),
    }
}

func (b *InMemoryEventBus) Publish(event domain.DomainEvent) error {
    eventType := reflect.TypeOf(event).String()
    
    b.mu.RLock()
    handlers := b.handlers[eventType]
    b.mu.RUnlock()

    for _, handler := range handlers {
        if err := handler(event); err != nil {
            log.Printf("Error handling event %s: %v", eventType, err)
        }
    }
    return nil
}
```

### **Event Bus Usage trong Command Handlers**
```go
// internal/modules/customer/application/command_handlers/create_customer.go
type CreateCustomerHandler struct {
    repository    domain.CustomerRepository
    domainService domain.CustomerDomainService
    eventBus      domain.EventBus
}

func (h *CreateCustomerHandler) Handle(ctx context.Context, cmd CreateCustomerCommand) (*CreateCustomerResult, error) {
    // Business logic...
    
    // Publish domain events
    for _, event := range customer.GetUncommittedEvents() {
        if err := h.eventBus.Publish(event); err != nil {
            return nil, fmt.Errorf("failed to publish event: %w", err)
        }
    }
    
    return result, nil
}
```

## 📊 Current DI Best Practices

### **1. Constructor Injection Pattern**
```go
// Good: All dependencies via constructor
func NewCreateCustomerHandler(
    repository domain.CustomerRepository,
    domainService domain.CustomerDomainService,
    eventBus domain.EventBus,
) *CreateCustomerHandler {
    return &CreateCustomerHandler{
        repository:    repository,
        domainService: domainService,
        eventBus:      eventBus,
    }
}
```

### **2. Interface-Based Dependencies**
```go
// Domain interfaces for loose coupling
type CustomerRepository interface {
    Save(ctx context.Context, customer *Customer) error
    GetByID(ctx context.Context, id string) (*Customer, error)
    GetByEmail(ctx context.Context, email string) (*Customer, error)
    // ...
}

type CustomerDomainService interface {
    IsEmailUnique(ctx context.Context, email string, excludeCustomerID ...string) (bool, error)
    CanDeleteCustomer(ctx context.Context, customerID string) (bool, error)
}
```

### **3. Factory Pattern cho Repositories**
```go
// Repository factory với error handling
func NewPostgreSQLCustomerRepositoryFromManager() (*PostgreSQLCustomerRepository, error) {
    db, err := customerdb.GetCustomerDB()
    if err != nil {
        return nil, fmt.Errorf("failed to get customer database: %w", err)
    }

    return &PostgreSQLCustomerRepository{db: db}, nil
}
```

### **4. Error Handling trong DI**
```go
func initDependencies(eventBus domain.EventBus) (*Dependencies, error) {
    customerRepo, err := persistence.NewPostgreSQLCustomerRepositoryFromManager()
    if err != nil {
        return nil, fmt.Errorf("failed to create customer repository: %w", err)
    }
    
    // Continue with other dependencies...
    
    return &Dependencies{
        CustomerHandler: customerHandler,
    }, nil
}
```

## 🧪 Testing với Module-Based DI

### **Unit Testing Module Components**
```go
// internal/modules/customer/module_test.go
func TestCustomerModule_Initialize(t *testing.T) {
    // Arrange
    module := NewCustomerModule()
    mockEventBus := &mocks.EventBus{}
    
    deps := domain.ModuleDependencies{
        EventBus: mockEventBus,
        Config:   &config.Config{},
    }
    
    // Act
    err := module.Initialize(deps)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, module.handler)
}

func TestCustomerModule_Health(t *testing.T) {
    // Arrange
    module := NewCustomerModule()
    ctx := context.Background()
    
    // Initialize module first
    deps := domain.ModuleDependencies{
        EventBus: &mocks.EventBus{},
        Config:   &config.Config{},
    }
    module.Initialize(deps)
    
    // Act
    err := module.Health(ctx)
    
    // Assert
    assert.NoError(t, err)
}
```

### **Integration Testing với Module Registry**
```go
func TestModuleRegistry_Integration(t *testing.T) {
    // Setup test configuration
    cfg := &config.Config{
        Modules: &config.ModulesConfig{
            Modules: map[string]config.ModuleConfig{
                "customer": {Enabled: true},
                "order":    {Enabled: false},
            },
        },
    }
    
    // Initialize module manager
    manager := registry.GetGlobalManager()
    
    // Register test modules
    manager.RegisterModule("customer", func() domain.Module {
        return customer.NewCustomerModule()
    })
    
    // Load enabled modules
    err := manager.LoadEnabledModules(cfg)
    require.NoError(t, err)
    
    // Verify only enabled modules are loaded
    moduleRegistry := manager.GetRegistry()
    moduleNames := moduleRegistry.GetModuleNames()
    
    assert.Contains(t, moduleNames, "customer")
    assert.NotContains(t, moduleNames, "order")
}
```

### **Testing Module Auto-Registration**
```go
func TestModuleAutoRegistration(t *testing.T) {
    // Reset global manager for test
    registry.ResetGlobalManager() // Test helper function
    
    // Import module package to trigger auto-registration
    _ = customer.NewCustomerModule() // This triggers init()
    
    // Verify module is registered
    manager := registry.GetGlobalManager()
    assert.True(t, manager.HasModule("customer"))
    
    // Test module creation
    module, err := manager.CreateModule("customer")
    assert.NoError(t, err)
    assert.Equal(t, "customer", module.Name())
}
```

## 🚀 Evolution Path: Module-Based Architecture

### **Current State: Module-Based Auto-Registration**
```go
// Current: Module-based with auto-registration
type ModuleManager struct {
    registry *domain.ModuleRegistry
    creators map[string]ModuleCreator
}

// Modules auto-register via init() functions
func init() {
    registry.RegisterModule("customer", func() domain.Module {
        return NewCustomerModule()
    })
}
```

### **Next Evolution: Plugin-Based Architecture**
```go
// Future: Plugin-based modules with hot-reload
type PluginManager struct {
    plugins map[string]*Plugin
    loader  *PluginLoader
}

type Plugin struct {
    Name     string
    Version  string
    Module   domain.Module
    Metadata PluginMetadata
}

func (pm *PluginManager) LoadPlugin(path string) error {
    plugin, err := pm.loader.Load(path)
    if err != nil {
        return err
    }
    
    pm.plugins[plugin.Name] = plugin
    return nil
}
```

### **Advanced Evolution: Microservices Migration**
```go
// Future: Gradual migration to microservices
type ServiceRegistry struct {
    localModules  map[string]domain.Module
    remoteServices map[string]ServiceClient
}

func (sr *ServiceRegistry) GetService(name string) (Service, error) {
    // Try local module first
    if module, exists := sr.localModules[name]; exists {
        return &LocalService{module: module}, nil
    }
    
    // Fall back to remote service
    if client, exists := sr.remoteServices[name]; exists {
        return &RemoteService{client: client}, nil
    }
    
    return nil, fmt.Errorf("service %s not found", name)
}
```

## 📈 Performance Considerations

### **1. Module Initialization Performance**
```go
// Parallel module initialization for better performance
func (r *ModuleRegistry) InitializeAll(deps ModuleDependencies) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(r.modules))
    
    for name, module := range r.modules {
        wg.Add(1)
        go func(name string, module Module) {
            defer wg.Done()
            if err := module.Initialize(deps); err != nil {
                errChan <- fmt.Errorf("failed to initialize module %s: %w", name, err)
            }
        }(name, module)
    }
    
    wg.Wait()
    close(errChan)
    
    // Check for errors
    for err := range errChan {
        return err
    }
    
    return nil
}
```

### **2. Lazy Module Loading**
```go
// Load modules only when first accessed
type LazyModuleRegistry struct {
    modules map[string]Module
    loaders map[string]func() Module
    mu      sync.RWMutex
}

func (r *LazyModuleRegistry) GetModule(name string) Module {
    r.mu.RLock()
    if module, exists := r.modules[name]; exists {
        r.mu.RUnlock()
        return module
    }
    r.mu.RUnlock()
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Double-check pattern
    if module, exists := r.modules[name]; exists {
        return module
    }
    
    // Load module
    if loader, exists := r.loaders[name]; exists {
        module := loader()
        r.modules[name] = module
        return module
    }
    
    return nil
}
```

### **3. Memory Efficiency**
```go
// Module registry with memory-efficient storage
type ModuleRegistry struct {
    modules map[string]Module
    // Use sync.Pool for temporary objects
    modulePool sync.Pool
}

func (r *ModuleRegistry) getModuleFromPool() *ModuleWrapper {
    if wrapper := r.modulePool.Get(); wrapper != nil {
        return wrapper.(*ModuleWrapper)
    }
    return &ModuleWrapper{}
}

func (r *ModuleRegistry) putModuleToPool(wrapper *ModuleWrapper) {
    wrapper.Reset()
    r.modulePool.Put(wrapper)
}
```

## 🔧 Configuration-Driven Module Management

### **Advanced Module Configuration**
```yaml
# config/modules.yaml
modules:
  customer:
    enabled: true
    priority: 1          # Load order
    dependencies: []     # Module dependencies
    features:
      events_enabled: true
      caching_enabled: false
    resources:
      max_memory: "100MB"
      max_cpu: "0.5"
  
  order:
    enabled: true
    priority: 2
    dependencies: ["customer"]  # Depends on customer module
    features:
      events_enabled: true
      caching_enabled: true
```

### **Dependency-Aware Module Loading**
```go
func (m *ModuleManager) LoadEnabledModulesWithDependencies(cfg *config.Config) error {
    // Build dependency graph
    graph := m.buildDependencyGraph(cfg)
    
    // Topological sort for load order
    loadOrder, err := graph.TopologicalSort()
    if err != nil {
        return fmt.Errorf("circular dependency detected: %w", err)
    }
    
    // Load modules in dependency order
    for _, moduleName := range loadOrder {
        if m.isModuleEnabled(cfg, moduleName) {
            if err := m.loadModule(moduleName); err != nil {
                return fmt.Errorf("failed to load module %s: %w", moduleName, err)
            }
        }
    }
    
    return nil
}
```

### **Runtime Module Management**
```go
// Enable/disable modules at runtime
func (m *ModuleManager) EnableModule(name string) error {
    module, err := m.CreateModule(name)
    if err != nil {
        return err
    }
    
    // Initialize with current dependencies
    deps := m.getCurrentDependencies()
    if err := module.Initialize(deps); err != nil {
        return err
    }
    
    // Start module
    ctx := context.Background()
    if err := module.Start(ctx); err != nil {
        return err
    }
    
    // Register with registry
    m.registry.Register(module)
    
    log.Printf("✅ Module %s enabled at runtime", name)
    return nil
}

func (m *ModuleManager) DisableModule(name string) error {
    module, exists := m.registry.GetModule(name)
    if !exists {
        return fmt.Errorf("module %s not found", name)
    }
    
    // Stop module gracefully
    ctx := context.Background()
    if err := module.Stop(ctx); err != nil {
        log.Printf("⚠️ Error stopping module %s: %v", name, err)
    }
    
    // Remove from registry
    m.registry.Unregister(name)
    
    log.Printf("🛑 Module %s disabled at runtime", name)
    return nil
}
```

## 🎯 Module DI Best Practices

### **1. Module Interface Compliance**
```go
// Ensure all modules implement the full interface
var _ domain.Module = (*CustomerModule)(nil)
var _ domain.Module = (*OrderModule)(nil)
var _ domain.Module = (*UserModule)(nil)
```

### **2. Graceful Error Handling**
```go
func (m *CustomerModule) Initialize(deps domain.ModuleDependencies) error {
    // Validate dependencies
    if deps.EventBus == nil {
        return fmt.Errorf("event bus is required for customer module")
    }
    
    // Initialize with proper error wrapping
    customerRepo, err := persistence.NewPostgreSQLCustomerRepositoryFromManager()
    if err != nil {
        return fmt.Errorf("failed to create customer repository: %w", err)
    }
    
    // Store dependencies
    m.eventBus = deps.EventBus
    m.customerRepo = customerRepo
    
    return nil
}
```

### **3. Resource Cleanup**
```go
func (m *CustomerModule) Stop(ctx context.Context) error {
    log.Printf("🛑 Stopping %s module", m.name)
    
    // Cleanup resources with timeout
    cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Stop background workers
    if m.backgroundWorker != nil {
        if err := m.backgroundWorker.Stop(cleanupCtx); err != nil {
            log.Printf("⚠️ Error stopping background worker: %v", err)
        }
    }
    
    // Close connections
    if m.cache != nil {
        m.cache.Close()
    }
    
    log.Printf("✅ %s module stopped successfully", m.name)
    return nil
}
```

## 🎯 DI Anti-Patterns to Avoid

### **1. Module Circular Dependencies**
```go
// Bad: Circular dependency between modules
type CustomerModule struct {
    orderModule *OrderModule  // ❌ Direct module dependency
}

type OrderModule struct {
    customerModule *CustomerModule  // ❌ Circular dependency
}

// Good: Use event bus for cross-module communication
type CustomerModule struct {
    eventBus domain.EventBus  // ✅ Communicate via events
}

func (m *CustomerModule) handleCustomerCreated(customer *domain.Customer) {
    event := domain.NewCustomerCreatedEvent(customer.ID, customer.Email)
    m.eventBus.Publish(event)  // ✅ Publish event instead of direct call
}
```

### **2. Global Module State**
```go
// Bad: Global module instances
var globalCustomerModule *CustomerModule  // ❌ Global state

// Good: Module registry management
func GetCustomerModule() *CustomerModule {
    registry := GetGlobalManager().GetRegistry()
    module, exists := registry.GetModule("customer")
    if !exists {
        return nil
    }
    return module.(*CustomerModule)  // ✅ Managed by registry
}
```

### **3. Hardcoded Module Dependencies**
```go
// Bad: Hardcoded module loading
func main() {
    customerModule := customer.NewCustomerModule()  // ❌ Hardcoded
    orderModule := order.NewOrderModule()          // ❌ Hardcoded
    userModule := user.NewUserModule()             // ❌ Hardcoded
}

// Good: Configuration-driven loading
func main() {
    modules.InitializeAllModules()  // ✅ Auto-registration
    
    manager := registry.GetGlobalManager()
    err := manager.LoadEnabledModules(cfg)  // ✅ Config-driven
    if err != nil {
        log.Fatal(err)
    }
}
```

## 📋 Adding New Modules

### **Step-by-Step Module Addition**
```go
// 1. Create module structure
internal/modules/order/
├── domain/
├── application/
├── infrastructure/
└── module.yaml

// 2. Add to Dependencies struct
type Dependencies struct {
    CustomerHandler *handlers.CustomerHandler
    OrderHandler    *handlers.OrderHandler  // Add new handler
}

// 3. Update initDependencies
func initDependencies(eventBus domain.EventBus) (*Dependencies, error) {
    // Customer dependencies...
    
    // Add order dependencies
    orderRepo, err := orderpersistence.NewPostgreSQLOrderRepositoryFromManager()
    if err != nil {
        return nil, err
    }
    
    orderHandler := orderhandlers.NewOrderHandler(orderRepo, eventBus)
    
    return &Dependencies{
        CustomerHandler: customerHandler,
        OrderHandler:    orderHandler,
    }, nil
}

// 4. Register routes
func initRouter(cfg *config.Config, deps *Dependencies) *gin.Engine {
    // ...
    api := router.Group("/api/v1")
    {
        customerhttp.RegisterCustomerRoutes(api, deps.CustomerHandler)
        orderhttp.RegisterOrderRoutes(api, deps.OrderHandler) // Add new routes
    }
    return router
}
```

## 🏆 Current Architecture Benefits

### **Simplicity Benefits**
- **Easy to Understand**: Straightforward dependency flow
- **Fast Development**: Quick to add new features
- **Minimal Boilerplate**: No complex DI framework setup
- **Debugging Friendly**: Clear call stack, easy to trace

### **Performance Benefits**
- **Fast Startup**: Minimal initialization overhead
- **Low Memory**: Simple structures, no heavy containers
- **Efficient**: Direct method calls, no reflection

### **Maintainability Benefits**
- **Explicit Dependencies**: All dependencies visible in constructors
- **Type Safety**: Compile-time dependency checking
- **Testable**: Easy to mock and unit test
- **Refactorable**: IDE can track all usages

**TMM's current DI architecture provides the right balance of simplicity and functionality for a modular monolith!** 🏗️✨ 