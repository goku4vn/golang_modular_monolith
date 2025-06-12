# 🏗️ Modular Monolith - Golang Hexagonal + CQRS Architecture

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal-blue?style=for-the-badge)](https://alistair.cockburn.us/hexagonal-architecture/)
[![Pattern](https://img.shields.io/badge/Pattern-CQRS-green?style=for-the-badge)](https://martinfowler.com/bliki/CQRS.html)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)

> **Enterprise-grade Modular Monolith** được xây dựng bằng Golang với **Hexagonal Architecture** và **CQRS Pattern**. Dự án này cung cấp foundation mạnh mẽ để phát triển ứng dụng có thể scale từ monolith sang microservices một cách tự nhiên.

## 🎯 Mục tiêu dự án

- ✅ **Clean Architecture**: Tách biệt rõ ràng business logic và infrastructure
- ✅ **CQRS Pattern**: Tối ưu hóa read/write operations
- ✅ **Domain-Driven Design**: Tập trung vào business domain
- ✅ **Event-Driven**: Loose coupling giữa các modules
- ✅ **Scalability**: Dễ dàng chuyển đổi sang microservices
- ✅ **Code Generation**: Rapid development với CRUD generator

## 🏗️ Kiến trúc tổng quan

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Layer (Chi/Gin)                     │
├─────────────────────────────────────────────────────────────┤
│                   Application Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Commands  │  │   Queries   │  │  Handlers   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                     Domain Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Aggregates  │  │   Events    │  │  Services   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                 Infrastructure Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ PostgreSQL  │  │    Redis    │  │   Logger    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Cấu trúc dự án

```
├── cmd/
│   └── server/                 # Application entry point
├── internal/
│   ├── shared/                 # Shared kernel
│   │   ├── domain/            # Common domain objects
│   │   ├── infrastructure/    # Shared infrastructure
│   │   └── application/       # Common application services
│   └── modules/               # Business modules
│       ├── user/              # User management module
│       ├── order/             # Order management module
│       └── product/           # Product catalog module
├── tools/
│   └── generator/             # CRUD code generator
├── config/                    # Configuration files
├── migrations/                # Database migrations
├── docker/                    # Docker configuration
└── pkg/                       # Public packages
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 15+**
- **Redis 7+**
- **Docker & Docker Compose** (optional)

### 1. Clone và Setup

```bash
git clone https://github.com/goku4vn/golang_modular_monolith.git
cd golang_modular_monolith

# Install dependencies
go mod tidy

# Copy configuration
cp config/config.example.yaml config/config.yaml
```

### 2. Database Setup

```bash
# Start PostgreSQL & Redis với Docker
docker-compose up -d postgres redis

# Run migrations
make migrate-up
```

### 3. Run Application

```bash
# Development mode
make run-dev

# Production mode
make build && ./bin/server
```

### 4. Test API

```bash
# Health check
curl http://localhost:8080/health

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# Get user
curl http://localhost:8080/api/v1/users/{user-id}
```

## 🛠️ Development

### Code Generation

Generate complete CRUD operations cho entity mới:

```bash
# 1. Create entity config
cat > tools/generator/config/category.yaml << EOF
entity:
  name: Category
  table: categories
  schema: product_module
fields:
  - name: id
    type: UUID
    primary: true
  - name: name
    type: string
    required: true
    validate: "min=2,max=100"
  - name: description
    type: text
operations:
  create: true
  update: true
  delete: true
  list: true
  get: true
EOF

# 2. Generate code
go run tools/generator/main.go -config=tools/generator/config/category.yaml

# 3. Run migration
make migrate-up
```

### Available Commands

```bash
# Development
make run-dev          # Run with hot reload
make test             # Run all tests
make test-unit        # Run unit tests only
make test-integration # Run integration tests
make lint             # Run linter

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make migrate-create   # Create new migration

# Build & Deploy
make build            # Build binary
make docker-build     # Build Docker image
make docker-run       # Run with Docker Compose

# Code Generation
make generate-user    # Generate User CRUD
make generate-order   # Generate Order CRUD
make generate-product # Generate Product CRUD
```

## 📚 Architecture Patterns

### 1. Hexagonal Architecture (Ports & Adapters)

```go
// Domain Layer - Business Logic
type User struct {
    ID    UserID
    Email Email
    Name  string
}

// Port - Interface
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id UserID) (*User, error)
}

// Adapter - Implementation
type PostgreSQLUserRepository struct {
    db *sql.DB
}
```

### 2. CQRS Pattern

```go
// Command - Write Operation
type CreateUserCommand struct {
    Email string `validate:"required,email"`
    Name  string `validate:"required,min=2"`
}

// Query - Read Operation
type GetUserQuery struct {
    UserID string
}

// Separate Handlers
type CreateUserHandler struct { /* ... */ }
type GetUserHandler struct { /* ... */ }
```

### 3. Domain Events

```go
// Domain Event
type UserCreated struct {
    UserID     string
    Email      string
    OccurredAt time.Time
}

// Event Handler
func (h *EmailNotificationHandler) Handle(event UserCreated) error {
    return h.emailService.SendWelcomeEmail(event.Email)
}
```

## 🔧 Tech Stack

### Core Framework
- **HTTP Router**: Chi (lightweight, fast)
- **Database**: PostgreSQL + GORM
- **Cache**: Redis
- **Config**: Viper
- **Logging**: Zap
- **Validation**: Go Playground Validator

### Development Tools
- **Testing**: Testify + Testcontainers
- **Migration**: Golang-migrate
- **Linting**: GolangCI-lint
- **Documentation**: Swagger/OpenAPI

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Monitoring**: Prometheus + Grafana
- **Tracing**: Jaeger (optional)

## 📈 Scalability Roadmap

### Phase 1: Modular Monolith (Current)
- ✅ Modules communicate via in-process events
- ✅ Shared database với schema separation
- ✅ Single deployment unit

### Phase 2: Distributed Monolith
- 🔄 Replace in-process events với message queue
- 🔄 Separate databases per module
- 🔄 Keep single deployment

### Phase 3: Microservices
- ⏳ Extract modules thành independent services
- ⏳ API Gateway
- ⏳ Service discovery

## 🧪 Testing Strategy

```bash
# Unit Tests - Domain Logic
make test-unit

# Integration Tests - Repository Layer
make test-integration

# E2E Tests - HTTP Endpoints
make test-e2e

# Load Tests
make test-load
```

### Test Coverage

- **Domain Layer**: 95%+ coverage
- **Application Layer**: 90%+ coverage
- **Infrastructure Layer**: 80%+ coverage

## 📊 Monitoring & Observability

### Metrics
- **HTTP Metrics**: Request count, duration, errors
- **Database Metrics**: Connection pool, query performance
- **Business Metrics**: Users created, orders processed

### Logging
- **Structured Logging**: JSON format với Zap
- **Correlation IDs**: Request tracing
- **Log Levels**: Debug, Info, Warn, Error

### Health Checks
```bash
curl http://localhost:8080/health
{
  "status": "healthy",
  "timestamp": "2024-06-11T19:58:00Z",
  "checks": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

## 🤝 Contributing

1. **Fork** repository
2. **Create** feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** changes (`git commit -m 'Add amazing feature'`)
4. **Push** branch (`git push origin feature/amazing-feature`)
5. **Open** Pull Request

### Code Standards

- **Go fmt**: Automatic formatting
- **Linting**: Pass golangci-lint checks
- **Testing**: Maintain test coverage > 80%
- **Documentation**: Update README cho breaking changes

## 📝 License

Dự án này được phân phối dưới **MIT License**. Xem file [LICENSE](LICENSE) để biết thêm chi tiết.

## 👥 Authors

- **Goku4VN** - *Initial work* - [@goku4vn](https://github.com/goku4vn)

## 🙏 Acknowledgments

- **Robert C. Martin** - Clean Architecture concepts
- **Eric Evans** - Domain-Driven Design principles
- **Alistair Cockburn** - Hexagonal Architecture pattern
- **Martin Fowler** - CQRS pattern inspiration

---

⭐ **Star this repo** if you find it helpful!

📧 **Contact**: [your-email@example.com](mailto:your-email@example.com)

🔗 **Blog**: [Your Architecture Blog](https://your-blog.com)

---
*Built with ❤️ by Baby (Claude Assistant) cho Daddy*
