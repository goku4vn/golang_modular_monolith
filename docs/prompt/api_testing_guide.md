# API Testing Best Practices cho Golang Hexagonal + CQRS

## 🎯 1. Structured Response Format

### Chuẩn hóa Response Structure

```go
// internal/shared/infrastructure/http/response.go
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Meta      *Meta       `json:"meta,omitempty"`
    RequestID string      `json:"request_id"`
    Timestamp time.Time   `json:"timestamp"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    Limit      int `json:"limit,omitempty"`
    Total      int `json:"total,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

// Helper functions
func SuccessResponse(data interface{}, meta *Meta) APIResponse {
    return APIResponse{
        Success:   true,
        Data:      data,
        Meta:      meta,
        RequestID: generateRequestID(),
        Timestamp: time.Now(),
    }
}

func ErrorResponse(code, message string, details interface{}) APIResponse {
    return APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
            Details: details,
        },
        RequestID: generateRequestID(),
        Timestamp: time.Now(),
    }
}
```

### Áp dụng vào HTTP Handlers

```go
// internal/modules/user/infrastructure/http/user_handler.go
func (h *UserHTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var cmd commands.CreateUserCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        response := ErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error())
        writeJSONResponse(w, http.StatusBadRequest, response)
        return
    }
    
    // Validation
    if err := validate.Struct(cmd); err != nil {
        response := ErrorResponse("VALIDATION_ERROR", "Validation failed", formatValidationErrors(err))
        writeJSONResponse(w, http.StatusBadRequest, response)
        return
    }
    
    result, err := h.commandBus.Execute(r.Context(), cmd)
    if err != nil {
        response := ErrorResponse("CREATE_FAILED", "Failed to create user", err.Error())
        writeJSONResponse(w, http.StatusInternalServerError, response)
        return
    }
    
    response := SuccessResponse(result, nil)
    writeJSONResponse(w, http.StatusCreated, response)
}
```

## 🔧 2. Environment-based Configuration

### Config Structure

```go
// internal/shared/infrastructure/config/config.go
type Config struct {
    Environment string `mapstructure:"ENVIRONMENT" default:"development"`
    Port        string `mapstructure:"PORT" default:"8080"`
    Database    DatabaseConfig `mapstructure:",squash"`
    Redis       RedisConfig    `mapstructure:",squash"`
    Testing     TestingConfig  `mapstructure:",squash"`
}

type TestingConfig struct {
    EnableTestRoutes bool `mapstructure:"ENABLE_TEST_ROUTES" default:"false"`
    SeedData        bool `mapstructure:"SEED_TEST_DATA" default:"false"`
    ResetDB         bool `mapstructure:"RESET_DB_ON_START" default:"false"`
}

func Load() *Config {
    viper.SetConfigFile(".env")
    viper.ReadInConfig()
    viper.AutomaticEnv()
    
    var config Config
    viper.Unmarshal(&config)
    
    return &config
}
```

### Environment Files

```bash
# .env.development
ENVIRONMENT=development
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/myapp_dev?sslmode=disable
REDIS_URL=redis://localhost:6379
ENABLE_TEST_ROUTES=true
SEED_TEST_DATA=true
RESET_DB_ON_START=false

# .env.testing
ENVIRONMENT=testing
PORT=8081
DATABASE_URL=postgres://user:pass@localhost:5432/myapp_test?sslmode=disable
REDIS_URL=redis://localhost:6379/1
ENABLE_TEST_ROUTES=true
SEED_TEST_DATA=true
RESET_DB_ON_START=true
```

## 🎭 3. Test Routes & Data Seeding

### Test Routes cho Development

```go
// internal/shared/infrastructure/http/test_routes.go
func RegisterTestRoutes(router chi.Router, db *sql.DB, config *Config) {
    if !config.Testing.EnableTestRoutes {
        return
    }
    
    router.Route("/test", func(r chi.Router) {
        r.Post("/reset-db", resetDatabaseHandler(db))
        r.Post("/seed-data", seedDataHandler(db))
        r.Get("/health", healthCheckHandler)
        r.Get("/users/fake", createFakeUsersHandler)
        r.Delete("/users/all", deleteAllUsersHandler)
    })
}

func resetDatabaseHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Truncate all tables
        tables := []string{"users", "orders", "products"}
        for _, table := range tables {
            db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
        }
        
        response := SuccessResponse("Database reset completed", nil)
        writeJSONResponse(w, http.StatusOK, response)
    }
}

func seedDataHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        seeder := NewSeeder(db)
        if err := seeder.SeedAll(); err != nil {
            response := ErrorResponse("SEED_FAILED", "Failed to seed data", err.Error())
            writeJSONResponse(w, http.StatusInternalServerError, response)
            return
        }
        
        response := SuccessResponse("Data seeded successfully", nil)
        writeJSONResponse(w, http.StatusOK, response)
    }
}
```

### Data Seeder

```go
// internal/shared/infrastructure/database/seeder.go
type Seeder struct {
    db *sql.DB
}

func NewSeeder(db *sql.DB) *Seeder {
    return &Seeder{db: db}
}

func (s *Seeder) SeedAll() error {
    if err := s.SeedUsers(); err != nil {
        return err
    }
    if err := s.SeedProducts(); err != nil {
        return err
    }
    return s.SeedOrders()
}

func (s *Seeder) SeedUsers() error {
    users := []map[string]interface{}{
        {"id": "user-1", "email": "john@example.com", "name": "John Doe"},
        {"id": "user-2", "email": "jane@example.com", "name": "Jane Smith"},
        {"id": "admin-1", "email": "admin@example.com", "name": "Admin User"},
    }
    
    for _, user := range users {
        _, err := s.db.Exec(`
            INSERT INTO user_module.users (id, email, name, created_at) 
            VALUES ($1, $2, $3, NOW()) ON CONFLICT (id) DO NOTHING`,
            user["id"], user["email"], user["name"])
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 📝 4. Request/Response Logging

### Request Logger Middleware

```go
// internal/shared/infrastructure/http/middleware/logger.go
func RequestLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            requestID := generateRequestID()
            
            // Add request ID to context
            ctx := context.WithValue(r.Context(), "request_id", requestID)
            r = r.WithContext(ctx)
            
            // Log request
            logger.Info("HTTP Request",
                zap.String("request_id", requestID),
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.String("query", r.URL.RawQuery),
                zap.String("user_agent", r.UserAgent()),
                zap.String("remote_addr", r.RemoteAddr),
            )
            
            // Capture response
            wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            next.ServeHTTP(wrapped, r)
            
            // Log response
            duration := time.Since(start)
            logger.Info("HTTP Response",
                zap.String("request_id", requestID),
                zap.Int("status_code", wrapped.statusCode),
                zap.Duration("duration", duration),
            )
        })
    }
}
```

## 🔍 5. Validation & Error Handling

### Validation Middleware

```go
// internal/shared/infrastructure/http/middleware/validation.go
func ValidationMiddleware() func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip validation for GET requests
            if r.Method == http.MethodGet {
                next.ServeHTTP(w, r)
                return
            }
            
            // Validate Content-Type
            if r.Header.Get("Content-Type") != "application/json" {
                response := ErrorResponse("INVALID_CONTENT_TYPE", 
                    "Content-Type must be application/json", nil)
                writeJSONResponse(w, http.StatusBadRequest, response)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Validation error formatter
func formatValidationErrors(err error) map[string]string {
    errors := make(map[string]string)
    
    for _, err := range err.(validator.ValidationErrors) {
        field := strings.ToLower(err.Field())
        switch err.Tag() {
        case "required":
            errors[field] = "This field is required"
        case "email":
            errors[field] = "Invalid email format"
        case "min":
            errors[field] = fmt.Sprintf("Minimum length is %s", err.Param())
        case "max":
            errors[field] = fmt.Sprintf("Maximum length is %s", err.Param())
        default:
            errors[field] = "Invalid value"
        }
    }
    
    return errors
}
```

## 🚀 6. Postman Collection Generator

### Auto-generate Postman Collection

```go
// tools/postman/generator.go
type PostmanCollection struct {
    Info PostmanInfo `json:"info"`
    Item []PostmanItem `json:"item"`
}

type PostmanInfo struct {
    Name   string `json:"name"`
    Schema string `json:"schema"`
}

type PostmanItem struct {
    Name    string        `json:"name"`
    Request PostmanRequest `json:"request"`
}

func GeneratePostmanCollection(modules []Module) *PostmanCollection {
    collection := &PostmanCollection{
        Info: PostmanInfo{
            Name:   "My API Collection",
            Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        },
    }
    
    // Add test routes
    collection.Item = append(collection.Item, PostmanItem{
        Name: "Reset Database",
        Request: PostmanRequest{
            Method: "POST",
            URL:    "{{base_url}}/test/reset-db",
            Header: []PostmanHeader{
                {Key: "Content-Type", Value: "application/json"},
            },
        },
    })
    
    // Generate for each module
    for _, module := range modules {
        collection.Item = append(collection.Item, generateModuleRequests(module)...)
    }
    
    return collection
}
```

## 📊 7. Testing Workflow

### Development Testing Flow

```bash
# 1. Start application với test routes
ENVIRONMENT=development ENABLE_TEST_ROUTES=true go run cmd/server/main.go

# 2. Reset database
curl -X POST http://localhost:8080/test/reset-db

# 3. Seed test data
curl -X POST http://localhost:8080/test/seed-data

# 4. Test các endpoints
curl -X GET http://localhost:8080/users
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User"}'
```

### Postman Environment Variables

```json
{
  "name": "Development",
  "values": [
    {
      "key": "base_url",
      "value": "http://localhost:8080",
      "enabled": true
    },
    {
      "key": "test_user_id",
      "value": "user-1",
      "enabled": true
    },
    {
      "key": "auth_token",
      "value": "",
      "enabled": true
    }
  ]
}
```

## 🧪 8. Pre-request Scripts

### Postman Pre-request Script

```javascript
// Pre-request script để auto-generate data
pm.sendRequest({
    url: pm.environment.get("base_url") + "/test/seed-data",
    method: 'POST',
    header: {
        'Content-Type': 'application/json'
    }
}, function (err, response) {
    if (response.code === 200) {
        console.log("Test data seeded successfully");
    }
});

// Generate random data
pm.environment.set("random_email", `test${Math.floor(Math.random() * 1000)}@example.com`);
pm.environment.set("random_name", `Test User ${Math.floor(Math.random() * 100)}`);
```

### Test Scripts

```javascript
// Test script để validate response
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has correct structure", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('success');
    pm.expect(jsonData).to.have.property('data');
    pm.expect(jsonData).to.have.property('request_id');
    pm.expect(jsonData).to.have.property('timestamp');
});

pm.test("Response time is less than 500ms", function () {
    pm.expect(pm.response.responseTime).to.be.below(500);
});

// Extract data for next requests
if (pm.response.code === 201) {
    const jsonData = pm.response.json();
    if (jsonData.data && jsonData.data.id) {
        pm.environment.set("created_user_id", jsonData.data.id);
    }
}
```

## 🔧 9. Makefile Commands

```makefile
# Development helpers
.PHONY: dev test-api reset-db seed-data

dev:
	ENVIRONMENT=development ENABLE_TEST_ROUTES=true go run cmd/server/main.go

test-api:
	@echo "Testing API endpoints..."
	curl -s http://localhost:8080/test/health | jq
	curl -s -X POST http://localhost:8080/test/reset-db | jq
	curl -s -X POST http://localhost:8080/test/seed-data | jq

reset-db:
	curl -X POST http://localhost:8080/test/reset-db

seed-data:
	curl -X POST http://localhost:8080/test/seed-data

generate-postman:
	go run tools/postman/main.go > postman_collection.json
	@echo "Postman collection generated: postman_collection.json"

# Quick test all endpoints
test-all:
	@echo "Testing all endpoints..."
	@./scripts/test_endpoints.sh
```

## 🎯 10. Tips tiết kiệm thời gian

### A. Sử dụng Environment Variables
- `{{base_url}}` thay vì hardcode URL
- `{{auth_token}}` cho authentication
- `{{user_id}}` cho dynamic IDs

### B. Pre-request automation
- Auto reset DB trước khi test
- Auto seed data cần thiết
- Generate random test data

### C. Response validation
- Validate structure consistency
- Check performance thresholds
- Extract IDs cho requests tiếp theo

### D. Collection organization
- Group theo modules
- Order theo workflow logic
- Include cleanup requests

### E. Scripts automation
```bash
#!/bin/bash
# scripts/test_endpoints.sh
BASE_URL="http://localhost:8080"

echo "🔄 Resetting database..."
curl -s -X POST $BASE_URL/test/reset-db | jq '.success'

echo "🌱 Seeding data..."
curl -s -X POST $BASE_URL/test/seed-data | jq '.success'

echo "👥 Testing Users module..."
curl -s -X GET $BASE_URL/users | jq '.data | length'
curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User"}' | jq '.success'

echo "✅ All tests completed!"
```

Với setup này, bạn có thể test API một cách có tổ chức và tiết kiệm thời gian bằng cách:
- Tự động reset/seed data
- Validate responses consistently  
- Sử dụng variables để reuse
- Có logging chi tiết để debug
- Generate Postman collection tự động