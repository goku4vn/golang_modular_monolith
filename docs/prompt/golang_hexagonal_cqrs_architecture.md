# Golang Hexagonal + CQRS Architecture - Modular Monolith

## 🏗️ Tổng quan kiến trúc

```
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── shared/           # Shared kernel
│   │   ├── domain/
│   │   ├── infrastructure/
│   │   └── application/
│   └── modules/          # Business modules
│       ├── user/
│       ├── order/
│       └── product/
├── pkg/                  # Public packages
├── config/
├── migrations/
└── docker/
```

## 🔧 Tech Stack Tối Ưu

### Core Framework & Libraries
- **HTTP Framework**: Chi hoặc Gin (nhẹ, hiệu năng cao)
- **Database**: PostgreSQL + GORM/Sqlx
- **Message Queue**: Redis Streams (đơn giản, không cần thêm service)
- **Config**: Viper
- **Logging**: Zap
- **Validation**: Go-playground/validator
- **Testing**: Testify + Testcontainers

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Database Migration**: Migrate
- **Monitoring**: Prometheus + Grafana (optional)

## 📁 Chi tiết cấu trúc Module

### 1. Shared Kernel (`internal/shared/`)

```go
// internal/shared/domain/
├── event.go              # Domain events interface
├── aggregate.go          # Base aggregate root
├── repository.go         # Repository interfaces
├── value_objects.go      # Common value objects
└── errors.go            # Domain errors

// internal/shared/infrastructure/
├── database/
│   ├── connection.go
│   └── transaction.go
├── messaging/
│   ├── event_bus.go
│   └── redis_streams.go
├── logging/
│   └── logger.go
└── config/
    └── config.go

// internal/shared/application/
├── command.go           # Command interfaces
├── query.go            # Query interfaces
├── handler.go          # Handler interfaces
└── middleware.go       # Common middleware
```

### 2. Business Module Structure (`internal/modules/user/`)

```go
user/
├── domain/
│   ├── user.go          # User aggregate
│   ├── repository.go    # User repository interface
│   ├── events.go        # User domain events
│   └── services.go      # Domain services
├── application/
│   ├── commands/
│   │   ├── create_user.go
│   │   └── update_user.go
│   ├── queries/
│   │   ├── get_user.go
│   │   └── list_users.go
│   └── handlers/
│       ├── command_handlers.go
│       └── query_handlers.go
├── infrastructure/
│   ├── persistence/
│   │   ├── user_repository.go
│   │   └── user_query_repository.go
│   └── http/
│       └── user_handler.go
└── module.go           # Module registration
```

## 🎯 Implementation Best Practices

### 1. Domain Layer (Hexagon Core)

```go
// internal/modules/user/domain/user.go
type User struct {
    id       UserID
    email    Email
    name     string
    events   []shared.DomainEvent
}

func (u *User) ChangeEmail(newEmail Email) error {
    if u.email == newEmail {
        return nil
    }
    
    u.email = newEmail
    u.addEvent(UserEmailChanged{
        UserID:   u.id,
        NewEmail: newEmail,
        OccurredAt: time.Now(),
    })
    
    return nil
}

func (u *User) addEvent(event shared.DomainEvent) {
    u.events = append(u.events, event)
}
```

### 2. Application Layer (CQRS)

```go
// Commands
type CreateUserCommand struct {
    Email string `validate:"required,email"`
    Name  string `validate:"required,min=2"`
}

type CreateUserHandler struct {
    userRepo domain.UserRepository
    eventBus shared.EventBus
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) error {
    user, err := domain.NewUser(cmd.Email, cmd.Name)
    if err != nil {
        return err
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        return err
    }
    
    // Publish events
    return h.eventBus.PublishAll(ctx, user.Events())
}

// Queries
type GetUserQuery struct {
    UserID string
}

type GetUserHandler struct {
    queryRepo UserQueryRepository
}

func (h *GetUserHandler) Handle(ctx context.Context, query GetUserQuery) (*UserView, error) {
    return h.queryRepo.GetByID(ctx, query.UserID)
}
```

### 3. Infrastructure Layer (Adapters)

```go
// PostgreSQL Repository
type PostgreSQLUserRepository struct {
    db *sql.DB
}

func (r *PostgreSQLUserRepository) Save(ctx context.Context, user *domain.User) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Save user
    _, err = tx.ExecContext(ctx, 
        "INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET email=$2, name=$3",
        user.ID(), user.Email(), user.Name())
    
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

// Redis Event Bus
type RedisEventBus struct {
    client *redis.Client
}

func (e *RedisEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
    data, _ := json.Marshal(event)
    return e.client.XAdd(ctx, &redis.XAddArgs{
        Stream: fmt.Sprintf("events:%s", event.AggregateType()),
        Values: map[string]interface{}{
            "type": event.EventType(),
            "data": data,
        },
    }).Err()
}
```

### 4. HTTP Adapter

```go
type UserHTTPHandler struct {
    commandBus shared.CommandBus
    queryBus   shared.QueryBus
}

func (h *UserHTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var cmd CreateUserCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if err := h.commandBus.Execute(r.Context(), cmd); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
}
```

## 🔌 Module Registration & Dependency Injection

```go
// internal/modules/user/module.go
type Module struct {
    userRepo      domain.UserRepository
    queryRepo     UserQueryRepository
    commandBus    shared.CommandBus
    queryBus      shared.QueryBus
}

func NewModule(db *sql.DB, eventBus shared.EventBus, commandBus shared.CommandBus, queryBus shared.QueryBus) *Module {
    userRepo := NewPostgreSQLUserRepository(db)
    queryRepo := NewUserQueryRepository(db)
    
    // Register handlers
    commandBus.RegisterHandler(&CreateUserHandler{userRepo, eventBus})
    commandBus.RegisterHandler(&UpdateUserHandler{userRepo, eventBus})
    
    queryBus.RegisterHandler(&GetUserHandler{queryRepo})
    queryBus.RegisterHandler(&ListUsersHandler{queryRepo})
    
    return &Module{
        userRepo:   userRepo,
        queryRepo:  queryRepo,
        commandBus: commandBus,
        queryBus:   queryBus,
    }
}

func (m *Module) RegisterRoutes(router chi.Router) {
    handler := &UserHTTPHandler{m.commandBus, m.queryBus}
    
    router.Route("/users", func(r chi.Router) {
        r.Post("/", handler.CreateUser)
        r.Get("/{id}", handler.GetUser)
        r.Put("/{id}", handler.UpdateUser)
    })
}
```

## 🚀 Main Application Bootstrap

```go
// cmd/server/main.go
func main() {
    // Load config
    cfg := config.Load()
    
    // Setup infrastructure
    db := setupDatabase(cfg)
    eventBus := setupEventBus(cfg)
    commandBus := shared.NewCommandBus()
    queryBus := shared.NewQueryBus()
    
    // Register modules
    userModule := user.NewModule(db, eventBus, commandBus, queryBus)
    orderModule := order.NewModule(db, eventBus, commandBus, queryBus)
    
    // Setup HTTP router
    router := chi.NewRouter()
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)
    
    // Register module routes
    userModule.RegisterRoutes(router)
    orderModule.RegisterRoutes(router)
    
    // Start server
    log.Printf("Server starting on port %s", cfg.Port)
    http.ListenAndServe(":"+cfg.Port, router)
}
```

## 📊 Migration Strategy to Microservices

### Phase 1: Modular Monolith (Current)
- Modules communicate through in-process event bus
- Shared database với schema separation
- Single deployment unit

### Phase 2: Distributed Monolith
- Replace in-process event bus với message queue (RabbitMQ/Kafka)
- Separate databases per module
- Keep single deployment

### Phase 3: Microservices
- Extract modules thành independent services
- API Gateway (như Kong/Traefik)
- Service discovery (Consul/etcd)

## 🛠️ Development Workflow

### 1. Adding New Feature
```bash
# 1. Create domain entity
internal/modules/user/domain/user.go

# 2. Create command/query
internal/modules/user/application/commands/

# 3. Create handler
internal/modules/user/application/handlers/

# 4. Create repository implementation
internal/modules/user/infrastructure/persistence/

# 5. Register in module
internal/modules/user/module.go
```

### 2. Testing Strategy
- **Unit Tests**: Domain logic và handlers
- **Integration Tests**: Repository implementations
- **End-to-End Tests**: HTTP endpoints với testcontainers

### 3. Database Schema
```sql
-- Per module schema
CREATE SCHEMA user_module;
CREATE SCHEMA order_module;

-- Event store for CQRS
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    version INTEGER NOT NULL,
    occurred_at TIMESTAMP DEFAULT NOW()
);
```

## 💡 Key Benefits

1. **Minimal Setup**: Chỉ cần PostgreSQL + Redis
2. **Type Safety**: Golang's strong typing
3. **Testable**: Clear separation of concerns
4. **Scalable**: Easy transition to microservices
5. **Maintainable**: Clean architecture principles
6. **Performance**: Compiled binary, efficient resource usage

## 🔄 Evolutionary Path

1. **Start**: Single module, simple CRUD
2. **Grow**: Add more modules, implement CQRS
3. **Scale**: Add event sourcing if needed
4. **Distribute**: Extract to microservices when team grows

Kiến trúc này cho phép team nhỏ phát triển nhanh trong giai đoạn đầu, đồng thời dễ dàng scale khi cần thiết.