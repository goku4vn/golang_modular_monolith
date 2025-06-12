# Module Configuration

Hướng dẫn chi tiết về cách cấu hình modules trong **Module-Based Auto-Registration Architecture**.

## Overview

Modular Monolith sử dụng hệ thống cấu hình linh hoạt với **Auto-Registration** cho phép:
- **Bật/tắt modules** dễ dàng qua configuration
- **Auto-discovery** modules từ registered creators
- **Config-driven loading** - chỉ load modules được enable
- **Zero hardcoding** - không cần sửa main.go khi thêm module mới

## Module Auto-Registration Flow

```
1. Module init() → Auto-register creator function
2. modules.InitializeAllModules() → Import all modules
3. ModuleManager.LoadEnabledModules() → Load based on config
4. ModuleRegistry.InitializeAll() → Initialize enabled modules
5. ModuleRegistry.RegisterAllRoutes() → Register routes dynamically
```

## Configuration Files

### 1. Central Configuration: `config/modules.yaml`
File chính để điều khiển modules (config-driven loading):
```yaml
modules:
  customer: true                    # ✅ Will be loaded and initialized
  order: true                       # ✅ Will be loaded and initialized  
  user: false                       # ❌ Will be skipped completely
  product:                          # 🔧 Custom configuration
    enabled: true
    database:
      host: "custom-host"
    migration:
      enabled: false
```

### 2. Module-Level Configuration: `internal/modules/{module}/module.yaml`
Mỗi module có file config riêng với cấu hình mặc định:
```yaml
enabled: true
database:
  host: "${POSTGRES_HOST:localhost}"
  port: "${POSTGRES_PORT:5432}"
  name: "modular_monolith_customer"
migration:
  enabled: true
  path: "./migrations"
```

## Module Auto-Registration System

### 1. Module Registration (Auto)
```go
// internal/modules/customer/module.go
package customer

import (
    "golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register on package import
func init() {
    registry.RegisterModule("customer", func() domain.Module {
        return NewCustomerModule()
    })
}
```

### 2. Centralized Import
```go
// internal/modules/modules.go
package modules

import (
    // Import all modules to trigger auto-registration
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    _ "golang_modular_monolith/internal/modules/user"
)

func InitializeAllModules() {
    // Ensures all modules are imported and registered
}
```

### 3. Config-Driven Loading
```go
// cmd/api/main.go
func main() {
    // 1. Trigger auto-registration
    modules.InitializeAllModules()
    
    // 2. Load only enabled modules
    manager := registry.GetGlobalManager()
    err := manager.LoadEnabledModules(cfg)  // Only loads enabled modules
    
    // 3. Get registry with loaded modules
    moduleRegistry := manager.GetRegistry()
}
```

## Configuration Formats

### 1. Simple Boolean (Recommended)
```yaml
modules:
  customer: true     # ✅ Enable with default config
  user: false        # ❌ Completely disable (not loaded)
```

### 2. Array Format
```yaml
modules: [customer, order, product]  # ✅ Enable all with defaults
```

### 3. Mixed Format (Most Flexible)
```yaml
modules:
  customer: true                     # ✅ Simple enable
  order:                            # 🔧 Complex override
    enabled: true
    migration:
      enabled: false                # Disable migrations only
  user: false                       # ❌ Disable completely
```

## Module States in Auto-Registration

### ✅ Registered & Enabled (`module: true`)
```
1. Module registered via init() function
2. Module creator stored in ModuleManager
3. Module loaded during LoadEnabledModules()
4. Module initialized with dependencies
5. Routes registered dynamically
6. Module started and ready
```

### 🔧 Registered & Enabled with Custom Config
```yaml
order:
  enabled: true
  migration:
    enabled: false    # Module enabled but no database
```
```
1. Module registered via init() function
2. Module loaded with custom configuration
3. Module initialized but skips database setup
4. Routes registered normally
```

### 🚫 Registered but Disabled (`module: false`)
```
1. Module registered via init() function
2. Module creator stored but NOT loaded
3. Module skipped during LoadEnabledModules()
4. No initialization, no routes, no resources
```

### ❓ Not Registered (Missing import)
```
1. Module not imported in modules.go
2. init() function never called
3. Module creator not registered
4. Module unavailable even if enabled in config
```

## Configuration Override Priority

1. **Environment Variables** (Highest)
   ```bash
   export CUSTOMER_DATABASE_HOST=custom-host
   ```

2. **Central Config** (`config/modules.yaml`)
   ```yaml
   modules:
     customer:
       database:
         host: "override-host"
   ```

3. **Module-Level Config** (Lowest)
   ```yaml
   # internal/modules/customer/module.yaml
   database:
     host: "default-host"
   ```

## Common Use Cases

### 1. Development Environment
```yaml
modules:
  customer: true      # Core module
  order: true         # Core module
  analytics: false    # Skip heavy modules (not loaded)
  reporting: false    # Skip heavy modules (not loaded)
```

### 2. Testing Environment
```yaml
modules:
  customer: true
  order:
    enabled: true
    migration:
      enabled: false  # Use test fixtures instead
  user: true
```

### 3. Production Environment
```yaml
modules:
  customer: true
  order: true
  user: true
  analytics: true
  reporting: true
```

### 4. Feature Flags
```yaml
modules:
  customer: true
  order: true
  new_feature: false  # Disable until ready (not loaded)
```

## Environment-Specific Configuration

### Using Environment Variables
```bash
# .env.development
ANALYTICS_ENABLED=false
REPORTING_ENABLED=false

# .env.production  
ANALYTICS_ENABLED=true
REPORTING_ENABLED=true
```

### Using Different Config Files
```bash
# Development
cp config/modules.dev.yaml config/modules.yaml

# Production
cp config/modules.prod.yaml config/modules.yaml
```

## Module Dependencies in Auto-Registration

### Handling Dependencies
```yaml
modules:
  user: true          # Required by order
  order: true         # Depends on user
  # customer: false   # ❌ This would break order module
```

### Best Practices
- **Document dependencies** in module README
- **Validate dependencies** in module Initialize() method
- **Graceful degradation** when optional modules disabled
- **Dependency injection** via ModuleDependencies

### Example: Dependency Validation
```go
// internal/modules/order/module.go
func (m *OrderModule) Initialize(deps domain.ModuleDependencies) error {
    // Check if required modules are loaded
    registry := deps.ModuleRegistry
    if !registry.IsModuleLoaded("user") {
        return fmt.Errorf("order module requires user module to be enabled")
    }
    
    // Initialize with dependencies
    return m.initializeWithDependencies(deps)
}
```

## Validation and Debugging

### Check Current Configuration
```bash
# View loaded modules
docker logs tmm-dev | grep "📦 Loaded"

# View registered modules
docker logs tmm-dev | grep "🔧 Registered"

# View skipped modules
docker logs tmm-dev | grep "🚫 Skipped"

# View databases
curl http://localhost:8080/health | jq .databases
```

### Module Loading Logs
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
```

### Common Configuration Errors

**1. Module registered but not loaded**
```yaml
# ❌ Module registered via init() but disabled
modules:
  user: false

# ✅ Check logs for:
# 🚫 Skipped module: user (enabled: false)
```

**2. Module not registered (missing import)**
```go
// ❌ Missing import in modules.go
import (
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    // Missing: _ "golang_modular_monolith/internal/modules/user"
)

// ✅ Add missing import
import (
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    _ "golang_modular_monolith/internal/modules/user"
)
```

**3. Database not created**
```yaml
# Check if migration is disabled
modules:
  order:
    enabled: true
    migration:
      enabled: false  # This prevents database creation
```

**4. Override not working**
```yaml
# ❌ Wrong nesting
modules:
  customer:
    database:
      host: "wrong"

# ✅ Correct nesting (check module.yaml structure)
modules:
  customer:
    enabled: true
    database:
      host: "correct"
```

## Advanced Configuration

### Custom Module Configuration
```yaml
modules:
  custom_module:
    enabled: true
    custom_setting: "value"
    database:
      pool_size: 20
```

### Conditional Configuration
```yaml
modules:
  analytics: ${ANALYTICS_ENABLED:false}
  reporting: ${REPORTING_ENABLED:false}
```

### Module Groups
```yaml
modules:
  # Core modules (always enabled)
  customer: true
  order: true
  
  # Optional modules (environment-dependent)
  analytics: ${ENABLE_ANALYTICS:false}
  reporting: ${ENABLE_REPORTING:false}
  
  # Feature flags (development)
  new_checkout: ${FEATURE_NEW_CHECKOUT:false}
```

## Adding New Modules

### 1. Create Module with Auto-Registration
```go
// internal/modules/new_module/module.go
package new_module

import (
    "golang_modular_monolith/internal/shared/infrastructure/registry"
)

// Auto-register on import
func init() {
    registry.RegisterModule("new_module", func() domain.Module {
        return NewNewModule()
    })
}

type NewModule struct {
    name string
}

// Implement Module interface
func (m *NewModule) Name() string { return m.name }
func (m *NewModule) Initialize(deps domain.ModuleDependencies) error { /* ... */ }
func (m *NewModule) RegisterRoutes(router *gin.RouterGroup) { /* ... */ }
func (m *NewModule) Health(ctx context.Context) error { /* ... */ }
func (m *NewModule) Start(ctx context.Context) error { /* ... */ }
func (m *NewModule) Stop(ctx context.Context) error { /* ... */ }
```

### 2. Add to Centralized Import
```go
// internal/modules/modules.go
import (
    _ "golang_modular_monolith/internal/modules/customer"
    _ "golang_modular_monolith/internal/modules/order"
    _ "golang_modular_monolith/internal/modules/user"
    _ "golang_modular_monolith/internal/modules/new_module"  // ✨ Add here
)
```

### 3. Enable in Configuration
```yaml
# config/modules.yaml
modules:
  customer: true
  order: true
  user: false
  new_module: true  # ✨ Enable new module
```

### 4. No Changes to main.go Required! 🎉
The module will be automatically:
- Registered via init() function
- Loaded if enabled in config
- Initialized with dependencies
- Routes registered dynamically
- Started with other modules

## Migration Guide

### From Old Hardcoded System
```go
// ❌ Old hardcoded approach (main.go)
import customerhttp "golang_modular_monolith/internal/modules/customer/infrastructure/http"

type Dependencies struct {
    CustomerHandler *handlers.CustomerHandler
}

customerhttp.RegisterCustomerRoutes(api, deps.CustomerHandler)

// ✅ New auto-registration approach (main.go)
import "golang_modular_monolith/internal/modules"

func main() {
    modules.InitializeAllModules()  // Trigger auto-registration
    manager := registry.GetGlobalManager()
    manager.LoadEnabledModules(cfg)  // Load based on config
    moduleRegistry := manager.GetRegistry()
    moduleRegistry.RegisterAllRoutes(api)  // Dynamic registration
}
```

### From Old Configuration
```yaml
# ❌ Old verbose format (50+ lines per module)
modules:
  customer:
    enabled: true
    database:
      host: "localhost"
      port: 5432
      name: "modular_monolith_customer"
    migration:
      enabled: true
      path: "./migrations"
    # ... 40+ more lines

# ✅ New simple format (1 line per module)
modules:
  customer: true
```

### Gradual Migration Steps
1. **Implement Module interface** for existing modules
2. **Add auto-registration** via init() functions
3. **Add to centralized import** in modules.go
4. **Simplify configuration** to boolean values
5. **Remove hardcoded imports** from main.go
6. **Test each module** thoroughly

## Best Practices

### 1. Module Design
- **Implement Module interface** completely
- **Validate dependencies** in Initialize()
- **Handle graceful shutdown** in Stop()
- **Provide health checks** in Health()

### 2. Configuration
- **Use simple boolean** for most modules
- **Document dependencies** clearly
- **Use environment variables** for sensitive data
- **Test different configurations** thoroughly

### 3. Auto-Registration
- **Always add init() function** for new modules
- **Import in modules.go** immediately
- **Test registration** before deployment
- **Monitor loading logs** for issues

### 4. Debugging
- **Check registration logs** first
- **Verify configuration syntax** 
- **Test module isolation** individually
- **Monitor resource usage** per module