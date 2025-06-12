# Getting Started

Hướng dẫn thiết lập và chạy **Module-Based Auto-Registration Architecture** từ đầu.

## Prerequisites

- **Docker & Docker Compose**: Để chạy PostgreSQL và development environment
- **Go 1.21+**: Để build và run application
- **Make**: Để chạy các commands tiện lợi

## Quick Start

### 1. Clone Repository
```bash
git clone <repository-url>
cd modular-monolith
```

### 2. Start Development Environment
```bash
# Start PostgreSQL container và application với module auto-registration
make docker-dev
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

### 3. Configure Modules
Update modules `config/modules.yaml`:
```yaml
modules:
  customer: true    # ✅ Enable customer module
  order: true       # ✅ Enable order module
  user: false       # ❌ Disable user module (not loaded)
```

### 4. Create Databases (Auto-Discovery)
```bash
# Create databases for enabled modules (auto-discovery)
make create-databases
```

**Auto-Discovery Output:**
```
🗄️ Database Creation Script (Module-Based)
🔍 Auto-discovering enabled modules from config...
📋 Enabled modules: customer order
🚫 Skipping user module (disabled in config)
✅ Database modular_monolith_customer created successfully
✅ Database modular_monolith_order created successfully
```

### 5. Run Migrations (Auto-Discovery)
```bash
# Run database migrations for enabled modules
make migrate-up
```

### 6. Verify Setup
```bash
# Check health endpoint with module information
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "databases": ["customer", "order"],
  "modules": ["customer", "order"],
  "service": "modular-monolith",
  "version": "2.0.0"
}
```

### 7. Test Module Endpoints
```bash
# Test customer module
curl http://localhost:8080/api/v1/customers

# Test order module
curl http://localhost:8080/api/v1/orders

# Test disabled module (should return 404)
curl http://localhost:8080/api/v1/users
```

## Development Workflow

### Daily Development
1. **Start containers**: `make docker-dev` (auto-registers modules)
2. **Check loaded modules**: `docker logs tmm-dev | grep -E "(📦|🔧)"`
3. **Make code changes**: Files auto-reload in container
4. **Add new modules**: Update `config/modules.yaml` → `make create-databases`
5. **Run migrations**: `make migrate-up` when adding new migrations

### Adding New Module
```bash
# 1. Create module with auto-registration
# internal/modules/new_module/module.go
func init() {
    registry.RegisterModule("new_module", func() domain.Module {
        return NewNewModule()
    })
}

# 2. Add to centralized import
echo '_ "golang_modular_monolith/internal/modules/new_module"' >> internal/modules/modules.go

# 3. Enable in config
echo "  new_module: true" >> config/modules.yaml

# 4. Restart to trigger registration
docker restart tmm-dev

# 5. Create database
make create-databases

# 6. Create migration
make migrate-create MODULE=new_module NAME=initial_schema

# 7. Run migration
make migrate-up MODULE=new_module

# 8. Test module
curl http://localhost:8080/api/v1/new_module
```

### Stopping Development
```bash
# Stop all containers
make docker-down

# Clean up (removes volumes)
make docker-clean
```

## Module Management

### List Modules
```bash
# List all registered and enabled modules
make list-modules

# Output:
🔧 Registered modules: customer order user new_module
📦 Enabled modules: customer order new_module
🚫 Disabled modules: user
```

### Module Health Check
```bash
# Check health of all enabled modules
curl http://localhost:8080/health | jq .modules
```

### Enable/Disable Modules
```yaml
# config/modules.yaml
modules:
  customer: true     # ✅ Enabled - will be loaded
  order: true        # ✅ Enabled - will be loaded
  user: false        # ❌ Disabled - will be skipped
  analytics: true    # ✅ Enabled - will be loaded
```

After changing configuration:
```bash
# Restart to apply changes
docker restart tmm-dev

# Verify changes
docker logs tmm-dev | grep -E "(📦|🚫)"
```

## Next Steps

- [Module Configuration](module-configuration.md) - Cấu hình modules chi tiết với auto-registration
- [Database Management](database-management.md) - Quản lý databases và migrations với auto-discovery
- [Project Structure](project-structure.md) - Hiểu cấu trúc source code module-based
- [Commands Reference](commands.md) - Tất cả commands có sẵn cho module management

## Troubleshooting

### Common Issues

**1. Module not loading despite being enabled**
```bash
# Check if module is registered
docker logs tmm-dev | grep "🔧 Registered module: your_module"

# Check if module is imported
grep "your_module" internal/modules/modules.go

# Check module configuration
grep "your_module" config/modules.yaml
```

**2. Database connection failed**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check database manager logs
docker logs tmm-dev | grep "Database"

# Recreate databases
make create-databases
```

**3. Module not auto-discovered**
```bash
# Check module configuration syntax
yamllint config/modules.yaml

# Check auto-discovery logs
docker logs tmm-dev | grep "Auto-discovering"

# Verify module is enabled
grep -A 5 "modules:" config/modules.yaml
```

**4. Port already in use**
```bash
# Stop existing containers
make docker-down

# Or change ports in docker-compose.dev.yml
```

**5. Hot reload not working**
```bash
# Restart development container
docker restart tmm-dev

# Check if modules re-register
docker logs tmm-dev | grep "🔧 Registered"
```

**6. Module registration failed**
```bash
# Check if module has init() function
grep -r "func init()" internal/modules/*/module.go

# Check if module implements Module interface
grep -r "Module interface" internal/modules/*/module.go

# Check module initialization logs
docker logs tmm-dev | grep -E "(✅ Initialized|❌ Failed)"
``` 