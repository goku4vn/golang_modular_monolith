# Modular Monolith

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-blue?style=for-the-badge)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)

A flexible, modular monolith architecture built with Go, featuring dynamic module configuration and clean architecture principles.

## 🚀 Quick Start

```bash
# 1. Start development environment
make docker-dev

# 2. Configure modules
echo "modules:
  customer: true
  order: true" > config/modules.yaml

# 3. Create databases
make create-databases

# 4. Run migrations
make migrate-up

# 5. Test API
curl http://localhost:8080/health
```

## ✨ Key Features

- **🎛️ Flexible Module Configuration**: Enable/disable modules with simple `true/false`
- **🗄️ Manual Database Management**: App controls database lifecycle, not containers
- **🔧 98% Verbosity Reduction**: `customer: true` instead of 50+ lines of config
- **🏗️ Clean Architecture**: Domain-driven design with clear layer separation
- **🚫 Perfect Disable Logic**: Disabled modules are completely excluded
- **🔄 Hot Reload**: Development with instant code reloading
- **📦 Modular Design**: Independent modules with their own databases

## 📚 Documentation

### 🎯 Getting Started
- **[Getting Started Guide](docs/getting-started.md)** - Setup project từ đầu
  - Prerequisites và installation
  - Quick start workflow
  - Development environment setup
  - Troubleshooting common issues

### ⚙️ Configuration
- **[Module Configuration](docs/module-configuration.md)** - Cấu hình modules chi tiết
  - Simple boolean configuration (`customer: true`)
  - Complex configuration overrides
  - Environment-specific settings
  - Module states và dependencies
  - Migration from verbose configs

### 🗄️ Database Management
- **[Database Management](docs/database-management.md)** - Quản lý databases và migrations
  - Manual database creation workflow
  - Migration commands và best practices
  - Database per module architecture
  - Environment-specific database setup
  - Backup và restore procedures

### 🏗️ Architecture
- **[Project Structure](docs/project-structure.md)** - Cấu trúc source code
  - Clean architecture layers
  - Module structure và organization
  - Dependency flow và rules
  - Adding new modules
  - Best practices

### 📋 Commands
- **[Commands Reference](docs/commands.md)** - Tất cả commands có sẵn
  - Make commands
  - Docker commands
  - Database commands
  - API testing commands
  - Troubleshooting commands

## 🎛️ Module Configuration Examples

### Simple Configuration (Recommended)
```yaml
# config/modules.yaml
modules:
  customer: true     # Enable with defaults
  order: true        # Enable with defaults
  analytics: false   # Completely disable
```

### Advanced Configuration
```yaml
# config/modules.yaml
modules:
  customer: true                    # Simple enable
  order:                           # Complex configuration
    enabled: true
    migration:
      enabled: false               # Module enabled but no database
  user: false                      # Completely disabled
  analytics:                       # Environment-specific
    enabled: ${ANALYTICS_ENABLED:false}
```

### Configuration Results
- **`customer: true`** → ✅ Module loaded, database created
- **`order: { migration: { enabled: false } }`** → ✅ Module loaded, no database
- **`user: false`** → 🚫 Module completely excluded
- **Logs**: `🚫 Module user explicitly disabled in central config`

## 🗄️ Database Architecture

### Database Per Module
```
PostgreSQL Instance
├── modular_monolith_customer    # Customer module
├── modular_monolith_order       # Order module  
├── modular_monolith_analytics   # Analytics module
└── modular_monolith_reporting   # Reporting module
```

### Manual Database Creation
```bash
# App controls database lifecycle
make create-databases

# Output:
# 🗄️ Database Creation Script
# ✅ PostgreSQL connection successful
# 📋 Enabled modules: customer order
# ✅ Database modular_monolith_customer created
# ✅ Database modular_monolith_order created
```

## 🏗️ Architecture Overview

### Clean Architecture Layers
```
┌─────────────────────────────────────┐
│         Presentation Layer          │  ← HTTP/gRPC/GraphQL
│            (Controllers)            │
├─────────────────────────────────────┤
│         Application Layer           │  ← Use Cases/Commands/Queries
│          (Business Logic)           │
├─────────────────────────────────────┤
│           Domain Layer              │  ← Entities/Domain Services
│        (Core Business Rules)        │
├─────────────────────────────────────┤
│        Infrastructure Layer         │  ← Database/External Services
│     (Technical Implementation)      │
└─────────────────────────────────────┘
```

### Module Structure
```
internal/modules/customer/
├── module.yaml              # Module configuration
├── migrations/              # Database migrations
├── domain/                  # Business logic
├── application/             # Use cases
├── infrastructure/          # Database/HTTP
└── presentation/            # Controllers
```

## 🚀 Development Workflow

### Daily Development
```bash
# Start development
make docker-dev

# Make code changes (auto-reload)
# Edit files in internal/modules/...

# Add new module
echo "  new_module: true" >> config/modules.yaml
make create-databases

# Create migration
make migrate-create MODULE=new_module NAME=initial_schema

# Run migration
make migrate-up MODULE=new_module
```

### Adding New Features
```bash
# 1. Create module structure
mkdir -p internal/modules/feature/{domain,application,infrastructure,presentation}

# 2. Add module config
echo "enabled: true" > internal/modules/feature/module.yaml

# 3. Enable in central config
echo "  feature: true" >> config/modules.yaml

# 4. Create database and migrations
make create-databases
make migrate-create MODULE=feature NAME=initial_schema
make migrate-up MODULE=feature
```

## 🔧 Environment Configuration

### Development
```bash
export ENVIRONMENT=development
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433
export DATABASE_PREFIX=dev_modular_monolith
```

### Production
```bash
export ENVIRONMENT=production
export POSTGRES_HOST=prod-postgres.example.com
export POSTGRES_PORT=5432
export DATABASE_PREFIX=modular_monolith
```

### Module-Specific Overrides
```bash
# Override specific module settings
export CUSTOMER_DATABASE_HOST=custom-host
export ORDER_MIGRATION_ENABLED=false
export ANALYTICS_ENABLED=true
```

## 📊 System Status

### Health Check
```bash
curl -s http://localhost:8080/health | jq .
```

```json
{
  "status": "healthy",
  "service": "modular-monolith",
  "version": "2.0.0",
  "environment": "development",
  "databases": ["customer", "order"],
  "timestamp": "2025-06-12",
  "message": "🔥 Viper + Docker hot reload is working perfectly!"
}
```

### Module Status
```bash
# Check loaded modules
docker logs modular-monolith-dev | grep "📦 Loaded"
# Output: 📦 Loaded configuration for 2 modules: [customer order]

# Check disabled modules  
docker logs modular-monolith-dev | grep "🚫"
# Output: 🚫 Module user explicitly disabled in central config
```

## 🛠️ Available Commands

### Essential Commands
```bash
make docker-dev        # Start development environment
make create-databases  # Create databases for enabled modules
make migrate-up        # Run all migrations
make migrate-down      # Rollback migrations
make docker-clean      # Clean environment (removes data!)
```

### Database Commands
```bash
make migrate-create MODULE=customer NAME=add_field
make migrate-status MODULE=customer
make migrate-up MODULE=customer
make migrate-down MODULE=customer VERSION=1
```

### Development Commands
```bash
make build            # Build application
make test             # Run tests
make lint             # Run linter
curl http://localhost:8080/health  # Test API
```

## 🔍 Troubleshooting

### Common Issues

**Module not loading despite enabled**
```bash
# Check configuration
cat config/modules.yaml

# Check logs for disable messages
docker logs modular-monolith-dev | grep "🚫"
```

**Database connection failed**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Recreate databases
make create-databases
```

**Hot reload not working**
```bash
# Restart application container
docker restart modular-monolith-dev
```

## 🎯 Key Achievements

- ✅ **98% Configuration Reduction**: `customer: true` vs 50+ lines
- ✅ **Perfect Module Disable**: `user: false` completely excludes module
- ✅ **Clean Database Management**: App controls lifecycle, not containers
- ✅ **Flexible Architecture**: Support simple and complex configurations
- ✅ **Production Ready**: Environment-specific configurations
- ✅ **Developer Experience**: Hot reload, colored output, clear commands

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/amazing-feature`
3. **Add module configuration**: Update `config/modules.yaml`
4. **Create databases**: `make create-databases`
5. **Add migrations**: `make migrate-create MODULE=feature NAME=initial`
6. **Test changes**: `make test && curl http://localhost:8080/health`
7. **Commit changes**: `git commit -m 'Add amazing feature'`
8. **Push to branch**: `git push origin feature/amazing-feature`
9. **Open Pull Request**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Clean Architecture** principles by Robert C. Martin
- **Domain-Driven Design** concepts by Eric Evans
- **Modular Monolith** patterns for scalable architecture
- **Go community** for excellent tooling and libraries

---

**Built with ❤️ using Go, Docker, PostgreSQL, and Clean Architecture principles.**

For detailed documentation, see the [docs/](docs/) directory.
