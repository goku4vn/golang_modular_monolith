# Module Configuration

Hướng dẫn chi tiết về cách cấu hình modules trong Modular Monolith.

## Overview

Modular Monolith sử dụng hệ thống cấu hình linh hoạt cho phép:
- **Bật/tắt modules** dễ dàng
- **Override cấu hình** cho từng module
- **Giảm 98% verbosity** so với cấu hình truyền thống

## Configuration Files

### 1. Central Configuration: `config/modules.yaml`
File chính để điều khiển modules:
```yaml
modules:
  customer: true                    # Simple enable
  order: true                       # Simple enable  
  user: false                       # Completely disable
  product:                          # Complex configuration
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

## Configuration Formats

### 1. Simple Boolean (Recommended)
```yaml
modules:
  customer: true     # Enable with default config
  user: false        # Completely disable
```

### 2. Array Format
```yaml
modules: [customer, order, product]  # Enable all with defaults
```

### 3. Mixed Format (Most Flexible)
```yaml
modules:
  customer: true                     # Simple enable
  order:                            # Complex override
    migration:
      enabled: false                # Disable migrations only
  user: false                       # Disable completely
```

## Module States

### ✅ Enabled (`module: true`)
- Module được load và khởi tạo
- Database được tạo (nếu migration enabled)
- Routes và handlers được đăng ký
- Module hoạt động đầy đủ

### 🔧 Enabled with Custom Config
```yaml
order:
  enabled: true
  migration:
    enabled: false    # Module enabled but no database
```
- Module được load nhưng không tạo database
- Useful cho modules không cần database

### 🚫 Disabled (`module: false`)
- Module hoàn toàn không được load
- Không tạo database
- Không đăng ký routes
- Tiết kiệm resources

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
  analytics: false    # Skip heavy modules
  reporting: false    # Skip heavy modules
```

### 2. Testing Environment
```yaml
modules:
  customer: true
  order:
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
  new_feature: false  # Disable until ready
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

## Module Dependencies

### Handling Dependencies
```yaml
modules:
  user: true          # Required by order
  order: true         # Depends on user
  # customer: false   # ❌ This would break order module
```

### Best Practices
- **Document dependencies** in module README
- **Validate dependencies** in module initialization
- **Graceful degradation** when optional modules disabled

## Validation and Debugging

### Check Current Configuration
```bash
# View loaded modules
docker logs modular-monolith-dev | grep "📦 Loaded"

# View databases
curl http://localhost:8080/health | jq .databases
```

### Common Configuration Errors

**1. Module still loading despite `false`**
```yaml
# ❌ Wrong
modules:
  user: false

# ✅ Correct - check logs for:
# 🚫 Module user explicitly disabled in central config
```

**2. Database not created**
```yaml
# Check if migration is disabled
modules:
  order:
    migration:
      enabled: false  # This prevents database creation
```

**3. Override not working**
```yaml
# ❌ Wrong nesting
modules:
  customer:
    database:
      host: "wrong"

# ✅ Correct nesting (check module.yaml structure)
modules:
  customer:
    database:
      host: "correct"
```

## Advanced Configuration

### Custom Module Paths
```yaml
modules:
  custom_module:
    enabled: true
    path: "./custom/modules/custom_module"
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
  # Core modules
  customer: true
  order: true
  
  # Optional modules  
  analytics: ${ENABLE_ANALYTICS:false}
  reporting: ${ENABLE_REPORTING:false}
  
  # Feature flags
  new_checkout: ${FEATURE_NEW_CHECKOUT:false}
```

## Migration Guide

### From Old Configuration
```yaml
# ❌ Old verbose format (50+ lines)
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

# ✅ New simple format (1 line)
modules:
  customer: true
```

### Gradual Migration
1. **Keep existing config** - Backward compatible
2. **Simplify one module** at a time
3. **Test each change** thoroughly
4. **Remove verbose config** when confident 