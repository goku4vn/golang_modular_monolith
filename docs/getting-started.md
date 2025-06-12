# Getting Started

Hướng dẫn thiết lập và chạy Modular Monolith từ đầu.

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
# Start PostgreSQL container
make docker-dev
```

### 3. Configure Modules
Tạo file `config/modules.yaml`:
```yaml
modules:
  customer: true    # Enable customer module
  order: true       # Enable order module
```

### 4. Create Databases
```bash
# Create databases for enabled modules
make create-databases
```

### 5. Run Migrations
```bash
# Run database migrations
make migrate-up
```

### 6. Start Application
Application sẽ tự động start trong Docker container và hot reload khi có thay đổi code.

### 7. Verify Setup
```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "databases": ["customer", "order"],
  "service": "modular-monolith",
  "version": "2.0.0"
}
```

## Development Workflow

### Daily Development
1. **Start containers**: `make docker-dev`
2. **Make code changes**: Files auto-reload in container
3. **Add new modules**: Update `config/modules.yaml` → `make create-databases`
4. **Run migrations**: `make migrate-up` when adding new migrations

### Stopping Development
```bash
# Stop all containers
make docker-down

# Clean up (removes volumes)
make docker-clean
```

## Next Steps

- [Module Configuration](module-configuration.md) - Cấu hình modules chi tiết
- [Database Management](database-management.md) - Quản lý databases và migrations
- [Project Structure](project-structure.md) - Hiểu cấu trúc source code
- [Commands Reference](commands.md) - Tất cả commands có sẵn

## Troubleshooting

### Common Issues

**1. Database connection failed**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Recreate databases
make create-databases
```

**2. Module not loading**
```bash
# Check module configuration
cat config/modules.yaml

# Check logs
docker logs modular-monolith-dev
```

**3. Port already in use**
```bash
# Stop existing containers
make docker-down

# Or change ports in docker-compose.dev.yml
```

**4. Hot reload not working**
```bash
# Restart development container
docker restart modular-monolith-dev
``` 