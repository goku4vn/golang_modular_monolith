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

### 3. Database Migration Setup

#### **golang-migrate** (Khuyến nghị chính)

```bash
# Cài đặt CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Hoặc với Docker
docker run --rm -v $(pwd)/migrations:/migrations \
  migrate/migrate:v4.17.0 -path=/migrations -database="postgres://..." up
```

#### **Project Structure**

```
migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_create_users.up.sql
├── 000002_create_users.down.sql
├── 000003_create_orders.up.sql
└── 000003_create_orders.down.sql

internal/
├── infrastructure/
│   └── database/
│       ├── connection.go
│       ├── migrator.go        # Migration wrapper
│       └── seeder.go          # Data seeding
```

#### **Migration Files**

```sql
-- migrations/000001_init_schema.up.sql
-- Create schemas per module
CREATE SCHEMA IF NOT EXISTS user_module;
CREATE SCHEMA IF NOT EXISTS order_module;
CREATE SCHEMA IF NOT EXISTS product_module;

-- Shared tables
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    version INTEGER NOT NULL,
    occurred_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(aggregate_id, version)
);

CREATE INDEX idx_events_aggregate ON events(aggregate_id);
CREATE INDEX idx_events_type ON events(aggregate_type, event_type);
CREATE INDEX idx_events_occurred ON events(occurred_at);
```

```sql
-- migrations/000002_create_users.up.sql
CREATE TABLE user_module.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON user_module.users(email);
CREATE INDEX idx_users_status ON user_module.users(status);
```

```sql
-- migrations/000002_create_users.down.sql
DROP TABLE IF EXISTS user_module.users;
```

#### **Go Integration**

```go
// internal/infrastructure/database/migrator.go
package database

import (
    "database/sql"
    "fmt"
    
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
    db *sql.DB
    m  *migrate.Migrate
}

func NewMigrator(db *sql.DB, migrationsPath string) (*Migrator, error) {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to create postgres driver: %w", err)
    }
    
    m, err := migrate.NewWithDatabaseInstance(
        fmt.Sprintf("file://%s", migrationsPath),
        "postgres", driver,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create migrator: %w", err)
    }
    
    return &Migrator{db: db, m: m}, nil
}

func (m *Migrator) Up() error {
    if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    return nil
}

func (m *Migrator) Down() error {
    return m.m.Down()
}

func (m *Migrator) Version() (uint, bool, error) {
    return m.m.Version()
}

func (m *Migrator) Close() error {
    sourceErr, dbErr := m.m.Close()
    if sourceErr != nil {
        return sourceErr
    }
    return dbErr
}
```

#### **Main Application Integration**

```go
// cmd/server/main.go
func main() {
    cfg := config.Load()
    
    // Database connection
    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Run migrations
    migrator, err := database.NewMigrator(db, "./migrations")
    if err != nil {
        log.Fatal("Failed to create migrator:", err)
    }
    defer migrator.Close()
    
    if err := migrator.Up(); err != nil {
        log.Fatal("Failed to run migrations:", err)
    }
    
    log.Println("Migrations completed successfully")
    
    // Continue with application setup...
}
```

#### **Makefile Commands**

```makefile
# Makefile
DB_URL=postgres://user:pass@localhost:5432/mydb?sslmode=disable

.PHONY: migrate-up migrate-down migrate-create migrate-force migrate-version

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $name

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path migrations -database "$(DB_URL)" force $version

migrate-version:
	migrate -path migrations -database "$(DB_URL)" version

migrate-drop:
	migrate -path migrations -database "$(DB_URL)" drop -f

# Development helpers
db-reset: migrate-drop migrate-up
	@echo "Database reset completed"

db-seed:
	go run cmd/seeder/main.go
```

#### **Docker Compose Integration**

```yaml
# docker-compose.yml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
  
  migrate:
    image: migrate/migrate:v4.17.0
    profiles: ["tools"]
    volumes:
      - ./migrations:/migrations
    command: [
      "-path", "/migrations",
      "-database", "postgres://user:pass@postgres:5432/myapp?sslmode=disable",
      "up"
    ]
    depends_on:
      - postgres

volumes:
  postgres_data:
```

#### **Usage Examples**

```bash
# Development workflow
make migrate-create    # Create new migration
make migrate-up        # Apply migrations
make migrate-down      # Rollback last migration
make migrate-version   # Check current version

# Production deployment
docker-compose --profile tools run --rm migrate

# CI/CD pipeline
migrate -path migrations -database "$DATABASE_URL" up
```

#### **Advanced Features**

```go
// Conditional migrations
func (m *Migrator) MigrateToVersion(version uint) error {
    return m.m.Migrate(version)
}

// Migration with transaction
func (m *Migrator) SafeUp() error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return tx.Commit()
}
```

## ⚡ CRUD Generation Framework

### 1. Code Generator Structure

```
tools/
├── generator/
│   ├── main.go              # CLI tool
│   ├── templates/           # Go templates
│   │   ├── domain/
│   │   │   ├── entity.go.tmpl
│   │   │   ├── repository.go.tmpl
│   │   │   └── events.go.tmpl
│   │   ├── application/
│   │   │   ├── commands.go.tmpl
│   │   │   ├── queries.go.tmpl
│   │   │   └── handlers.go.tmpl
│   │   ├── infrastructure/
│   │   │   ├── repository.go.tmpl
│   │   │   ├── query_repo.go.tmpl
│   │   │   └── http_handler.go.tmpl
│   │   ├── migration.sql.tmpl
│   │   └── module.go.tmpl
│   └── config/
│       └── entity_config.yaml
```

### 2. Entity Configuration

```yaml
# tools/generator/config/product_config.yaml
entity:
  name: Product
  table: products
  schema: product_module
  
fields:
  - name: id
    type: UUID
    primary: true
    generate: true
  - name: name
    type: string
    required: true
    validate: "min=2,max=100"
  - name: price
    type: decimal
    required: true
    validate: "gt=0"
  - name: category_id
    type: UUID
    required: true
    foreign_key: categories.id
  - name: description
    type: text
    required: false
  - name: status
    type: enum
    values: [active, inactive, draft]
    default: draft
  - name: created_at
    type: timestamp
    auto: true
  - name: updated_at
    type: timestamp
    auto: true

operations:
  create: true
  update: true
  delete: true  # soft delete
  list: true
  get: true
  
indexes:
  - fields: [category_id]
  - fields: [status, created_at]
  
events:
  - ProductCreated
  - ProductUpdated
  - ProductDeleted
```

### 3. Generator CLI Tool

```go
// tools/generator/main.go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "text/template"
    
    "gopkg.in/yaml.v2"
)

type EntityConfig struct {
    Entity     EntityInfo            `yaml:"entity"`
    Fields     []FieldInfo          `yaml:"fields"`
    Operations map[string]bool      `yaml:"operations"`
    Indexes    []IndexInfo          `yaml:"indexes"`
    Events     []string             `yaml:"events"`
}

type EntityInfo struct {
    Name   string `yaml:"name"`
    Table  string `yaml:"table"`
    Schema string `yaml:"schema"`
}

type FieldInfo struct {
    Name       string   `yaml:"name"`
    Type       string   `yaml:"type"`
    Primary    bool     `yaml:"primary"`
    Required   bool     `yaml:"required"`
    Generate   bool     `yaml:"generate"`
    Validate   string   `yaml:"validate"`
    ForeignKey string   `yaml:"foreign_key"`
    Values     []string `yaml:"values"`
    Default    string   `yaml:"default"`
    Auto       bool     `yaml:"auto"`
}

func main() {
    var (
        configPath = flag.String("config", "", "Path to entity config file")
        outputDir  = flag.String("output", "./internal/modules", "Output directory")
    )
    flag.Parse()
    
    if *configPath == "" {
        log.Fatal("Config path is required")
    }
    
    config, err := loadConfig(*configPath)
    if err != nil {
        log.Fatal(err)
    }
    
    generator := NewGenerator(*outputDir)
    if err := generator.Generate(config); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Generated CRUD for %s successfully!\n", config.Entity.Name)
}

type Generator struct {
    outputDir     string
    templates     map[string]*template.Template
}

func NewGenerator(outputDir string) *Generator {
    return &Generator{
        outputDir: outputDir,
        templates: loadTemplates(),
    }
}

func (g *Generator) Generate(config *EntityConfig) error {
    moduleName := strings.ToLower(config.Entity.Name)
    moduleDir := filepath.Join(g.outputDir, moduleName)
    
    // Create directory structure
    dirs := []string{
        "domain", "application/commands", "application/queries", 
        "application/handlers", "infrastructure/persistence", 
        "infrastructure/http",
    }
    
    for _, dir := range dirs {
        if err := os.MkdirAll(filepath.Join(moduleDir, dir), 0755); err != nil {
            return err
        }
    }
    
    // Generate files
    files := map[string]string{
        "domain/entity.go.tmpl":                    "domain/" + moduleName + ".go",
        "domain/repository.go.tmpl":                "domain/repository.go",
        "domain/events.go.tmpl":                    "domain/events.go",
        "application/commands.go.tmpl":             "application/commands/" + moduleName + "_commands.go",
        "application/queries.go.tmpl":              "application/queries/" + moduleName + "_queries.go",
        "application/handlers.go.tmpl":             "application/handlers/" + moduleName + "_handlers.go",
        "infrastructure/repository.go.tmpl":       "infrastructure/persistence/" + moduleName + "_repository.go",
        "infrastructure/query_repo.go.tmpl":       "infrastructure/persistence/" + moduleName + "_query_repository.go",
        "infrastructure/http_handler.go.tmpl":     "infrastructure/http/" + moduleName + "_handler.go",
        "module.go.tmpl":                          "module.go",
        "migration.sql.tmpl":                      "../../migrations/" + time.Now().Format("20060102150405") + "_create_" + config.Entity.Table + ".up.sql",
    }
    
    for tmplFile, outputFile := range files {
        if err := g.generateFile(tmplFile, outputFile, moduleDir, config); err != nil {
            return fmt.Errorf("failed to generate %s: %w", outputFile, err)
        }
    }
    
    return nil
}
```

### 4. Entity Template

```go
// tools/generator/templates/domain/entity.go.tmpl
package domain

import (
    "time"
    "github.com/google/uuid"
    "your-app/internal/shared/domain"
)

// {{.Entity.Name}} represents the {{.Entity.Name}} aggregate root
type {{.Entity.Name}} struct {
    {{range .Fields}}
    {{if eq .Type "UUID"}}{{.Name}} uuid.UUID `json:"{{snake .Name}}"`{{end}}
    {{if eq .Type "string"}}{{.Name}} string `json:"{{snake .Name}}"{{if .Validate}} validate:"{{.Validate}}"{{end}}`{{end}}
    {{if eq .Type "decimal"}}{{.Name}} float64 `json:"{{snake .Name}}"{{if .Validate}} validate:"{{.Validate}}"{{end}}`{{end}}
    {{if eq .Type "timestamp"}}{{.Name}} time.Time `json:"{{snake .Name}}"`{{end}}
    {{if eq .Type "enum"}}{{.Name}} {{.Name}}Status `json:"{{snake .Name}}"`{{end}}
    {{end}}
    
    events []domain.DomainEvent
}

{{range .Fields}}
{{if eq .Type "enum"}}
type {{.Name}}Status string

const (
    {{range .Values}}{{title .}}{{$.Name}}Status {{$.Name}}Status = "{{.}}"
    {{end}}
)
{{end}}
{{end}}

// New{{.Entity.Name}} creates a new {{.Entity.Name}}
func New{{.Entity.Name}}({{range $i, $field := .Fields}}{{if and (not .Primary) (not .Auto) (.Required)}}{{if $i}}, {{end}}{{camel .Name}} {{goType .Type}}{{end}}{{end}}) (*{{.Entity.Name}}, error) {
    {{range .Fields}}{{if and .Primary .Generate}}{{camel .Name}} := uuid.New(){{end}}{{end}}
    {{range .Fields}}{{if and .Auto (eq .Name "created_at")}}{{camel .Name}} := time.Now(){{end}}{{end}}
    {{range .Fields}}{{if and .Auto (eq .Name "updated_at")}}{{camel .Name}} := time.Now(){{end}}{{end}}
    
    {{camel .Entity.Name}} := &{{.Entity.Name}}{
        {{range .Fields}}{{title .Name}}: {{camel .Name}},
        {{end}}
    }
    
    {{camel .Entity.Name}}.addEvent({{.Entity.Name}}Created{
        {{.Entity.Name}}ID: {{camel .Entity.Name}}.ID,
        OccurredAt: time.Now(),
    })
    
    return {{camel .Entity.Name}}, nil
}

{{if .Operations.update}}
// Update{{.Entity.Name}} updates the {{.Entity.Name}}
func ({{lower (substr .Entity.Name 0 1)}} *{{.Entity.Name}}) Update({{range $i, $field := .Fields}}{{if and (not .Primary) (not .Auto) (not eq .Name "created_at")}}{{if $i}}, {{end}}{{camel .Name}} {{goType .Type}}{{end}}{{end}}) {
    {{range .Fields}}{{if and (not .Primary) (not .Auto) (not eq .Name "created_at")}}{{lower (substr $.Entity.Name 0 1)}}.{{title .Name}} = {{camel .Name}}
    {{end}}{{end}}
    {{lower (substr .Entity.Name 0 1)}}.UpdatedAt = time.Now()
    
    {{lower (substr .Entity.Name 0 1)}}.addEvent({{.Entity.Name}}Updated{
        {{.Entity.Name}}ID: {{lower (substr .Entity.Name 0 1)}}.ID,
        OccurredAt: time.Now(),
    })
}
{{end}}

{{if .Operations.delete}}
// Delete marks the {{.Entity.Name}} as deleted
func ({{lower (substr .Entity.Name 0 1)}} *{{.Entity.Name}}) Delete() {
    {{lower (substr .Entity.Name 0 1)}}.Status = Inactive{{.Entity.Name}}Status
    {{lower (substr .Entity.Name 0 1)}}.UpdatedAt = time.Now()
    
    {{lower (substr .Entity.Name 0 1)}}.addEvent({{.Entity.Name}}Deleted{
        {{.Entity.Name}}ID: {{lower (substr .Entity.Name 0 1)}}.ID,
        OccurredAt: time.Now(),
    })
}
{{end}}

// Events returns domain events
func ({{lower (substr .Entity.Name 0 1)}} *{{.Entity.Name}}) Events() []domain.DomainEvent {
    return {{lower (substr .Entity.Name 0 1)}}.events
}

// ClearEvents clears domain events
func ({{lower (substr .Entity.Name 0 1)}} *{{.Entity.Name}}) ClearEvents() {
    {{lower (substr .Entity.Name 0 1)}}.events = nil
}

func ({{lower (substr .Entity.Name 0 1)}} *{{.Entity.Name}}) addEvent(event domain.DomainEvent) {
    {{lower (substr .Entity.Name 0 1)}}.events = append({{lower (substr .Entity.Name 0 1)}}.events, event)
}
```

### 5. Repository Template

```go
// tools/generator/templates/infrastructure/repository.go.tmpl
package persistence

import (
    "context"
    "database/sql"
    "fmt"
    
    "your-app/internal/modules/{{lower .Entity.Name}}/domain"
)

type PostgreSQL{{.Entity.Name}}Repository struct {
    db *sql.DB
}

func NewPostgreSQL{{.Entity.Name}}Repository(db *sql.DB) *PostgreSQL{{.Entity.Name}}Repository {
    return &PostgreSQL{{.Entity.Name}}Repository{db: db}
}

{{if .Operations.create}}
func (r *PostgreSQL{{.Entity.Name}}Repository) Save(ctx context.Context, {{camel .Entity.Name}} *domain.{{.Entity.Name}}) error {
    query := `
        INSERT INTO {{.Entity.Schema}}.{{.Entity.Table}} (
            {{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}}
        ) VALUES (
            {{range $i, $field := .Fields}}{{if $i}}, {{end}}${{add $i 1}}{{end}}
        ) ON CONFLICT (id) DO UPDATE SET
            {{range $i, $field := .Fields}}{{if and (not .Primary) (not eq .Name "created_at")}}{{if ne $i 1}}, {{end}}{{.Name}} = EXCLUDED.{{.Name}}{{end}}{{end}}
    `
    
    _, err := r.db.ExecContext(ctx, query,
        {{range .Fields}}{{camel .Entity.Name}}.{{title .Name}},
        {{end}}
    )
    
    return err
}
{{end}}

{{if .Operations.get}}
func (r *PostgreSQL{{.Entity.Name}}Repository) GetByID(ctx context.Context, id string) (*domain.{{.Entity.Name}}, error) {
    query := `
        SELECT {{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}}
        FROM {{.Entity.Schema}}.{{.Entity.Table}}
        WHERE id = $1 AND status != 'deleted'
    `
    
    var {{camel .Entity.Name}} domain.{{.Entity.Name}}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        {{range .Fields}}&{{camel .Entity.Name}}.{{title .Name}},
        {{end}}
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, domain.ErrNotFound
        }
        return nil, err
    }
    
    return &{{camel .Entity.Name}}, nil
}
{{end}}

{{if .Operations.delete}}
func (r *PostgreSQL{{.Entity.Name}}Repository) Delete(ctx context.Context, id string) error {
    query := `UPDATE {{.Entity.Schema}}.{{.Entity.Table}} SET status = 'deleted', updated_at = NOW() WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}
{{end}}
```

### 6. HTTP Handler Template

```go
// tools/generator/templates/infrastructure/http_handler.go.tmpl
package http

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "your-app/internal/modules/{{lower .Entity.Name}}/application/commands"
    "your-app/internal/modules/{{lower .Entity.Name}}/application/queries"
    "your-app/internal/shared/application"
)

type {{.Entity.Name}}HTTPHandler struct {
    commandBus application.CommandBus
    queryBus   application.QueryBus
}

func New{{.Entity.Name}}HTTPHandler(commandBus application.CommandBus, queryBus application.QueryBus) *{{.Entity.Name}}HTTPHandler {
    return &{{.Entity.Name}}HTTPHandler{
        commandBus: commandBus,
        queryBus:   queryBus,
    }
}

{{if .Operations.create}}
func (h *{{.Entity.Name}}HTTPHandler) Create{{.Entity.Name}}(w http.ResponseWriter, r *http.Request) {
    var cmd commands.Create{{.Entity.Name}}Command
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if err := h.commandBus.Execute(r.Context(), cmd); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}
{{end}}

{{if .Operations.get}}
func (h *{{.Entity.Name}}HTTPHandler) Get{{.Entity.Name}}(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    query := queries.Get{{.Entity.Name}}Query{ID: id}
    result, err := h.queryBus.Execute(r.Context(), query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
{{end}}

{{if .Operations.list}}
func (h *{{.Entity.Name}}HTTPHandler) List{{.Entity.Name}}s(w http.ResponseWriter, r *http.Request) {
    query := queries.List{{.Entity.Name}}sQuery{
        Page:  getIntParam(r, "page", 1),
        Limit: getIntParam(r, "limit", 20),
    }
    
    result, err := h.queryBus.Execute(r.Context(), query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
{{end}}

// RegisterRoutes registers HTTP routes
func (h *{{.Entity.Name}}HTTPHandler) RegisterRoutes(router chi.Router) {
    router.Route("/{{kebab .Entity.Name}}s", func(r chi.Router) {
        {{if .Operations.create}}r.Post("/", h.Create{{.Entity.Name}}){{end}}
        {{if .Operations.list}}r.Get("/", h.List{{.Entity.Name}}s){{end}}
        {{if .Operations.get}}r.Get("/{id}", h.Get{{.Entity.Name}}){{end}}
        {{if .Operations.update}}r.Put("/{id}", h.Update{{.Entity.Name}}){{end}}
        {{if .Operations.delete}}r.Delete("/{id}", h.Delete{{.Entity.Name}}){{end}}
    })
}
```

### 7. Usage Example

```bash
# 1. Create entity config
cat > tools/generator/config/product.yaml << EOF
entity:
  name: Product
  table: products  
  schema: product_module

fields:
  - name: id
    type: UUID
    primary: true
    generate: true
  - name: name
    type: string
    required: true
    validate: "min=2,max=100"
  - name: price
    type: decimal
    required: true
    validate: "gt=0"
  # ... other fields
EOF

# 2. Generate CRUD
go run tools/generator/main.go -config=tools/generator/config/product.yaml

# 3. Generated structure:
# internal/modules/product/
# ├── domain/
# ├── application/
# ├── infrastructure/
# └── module.go

# 4. Run migration
migrate -path migrations -database "postgres://..." up

# 5. Test endpoints
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"iPhone","price":999.99}'
```

### 8. Advanced Features

```go
// Enhanced generator with:
// - Validation rules
// - Relationships (foreign keys)  
// - Pagination
// - Filtering
// - Sorting
// - Soft delete
// - Audit fields
// - Search functionality
// - Bulk operations
```

## 💡 Key Benefits

1. **Minimal Setup**: Chỉ cần PostgreSQL + Redis
2. **Type Safety**: Golang's strong typing
3. **Testable**: Clear separation of concerns
4. **Scalable**: Easy transition to microservices
5. **Maintainable**: Clean architecture principles
6. **Performance**: Compiled binary, efficient resource usage
7. **⚡ Rapid Development**: Generate full CRUD in seconds
8. **🔧 Customizable**: Template-based generation
9. **📋 Consistent**: Same patterns across all entities

## 🔄 Evolutionary Path

1. **Start**: Single module, simple CRUD
2. **Grow**: Add more modules, implement CQRS
3. **Scale**: Add event sourcing if needed
4. **Distribute**: Extract to microservices when team grows

Kiến trúc này cho phép team nhỏ phát triển nhanh trong giai đoạn đầu, đồng thời dễ dàng scale khi cần thiết.