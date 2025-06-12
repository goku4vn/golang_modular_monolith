package application

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Command represents a command in CQRS pattern
type Command interface {
	// CommandName returns the name of the command
	CommandName() string
}

// CommandHandler handles commands
type CommandHandler[T Command] interface {
	Handle(ctx context.Context, cmd T) error
}

// CommandBus represents the command bus interface
type CommandBus interface {
	// Execute executes a command
	Execute(ctx context.Context, cmd Command) error

	// RegisterHandler registers a command handler
	RegisterHandler(cmdType reflect.Type, handler interface{}) error

	// RegisterHandlerFunc registers a command handler function
	RegisterHandlerFunc(cmdType reflect.Type, handlerFunc interface{}) error
}

// InMemoryCommandBus is an in-memory implementation of CommandBus
type InMemoryCommandBus struct {
	handlers map[reflect.Type]interface{}
	mutex    sync.RWMutex
}

// NewInMemoryCommandBus creates a new in-memory command bus
func NewInMemoryCommandBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{
		handlers: make(map[reflect.Type]interface{}),
	}
}

// Execute executes a command
func (bus *InMemoryCommandBus) Execute(ctx context.Context, cmd Command) error {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	cmdType := reflect.TypeOf(cmd)
	handler, exists := bus.handlers[cmdType]
	if !exists {
		return fmt.Errorf("no handler registered for command %s", cmdType.Name())
	}

	// Use reflection to call the handler
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	// Check if it's a method (Handle)
	if handlerValue.Kind() == reflect.Ptr {
		method := handlerValue.MethodByName("Handle")
		if !method.IsValid() {
			return fmt.Errorf("handler for command %s does not have Handle method", cmdType.Name())
		}

		// Call Handle method
		results := method.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
		})

		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		return nil
	}

	// Check if it's a function
	if handlerType.Kind() == reflect.Func {
		results := handlerValue.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
		})

		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		return nil
	}

	return fmt.Errorf("invalid handler type for command %s", cmdType.Name())
}

// RegisterHandler registers a command handler
func (bus *InMemoryCommandBus) RegisterHandler(cmdType reflect.Type, handler interface{}) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if _, exists := bus.handlers[cmdType]; exists {
		return fmt.Errorf("handler already registered for command %s", cmdType.Name())
	}

	bus.handlers[cmdType] = handler
	return nil
}

// RegisterHandlerFunc registers a command handler function
func (bus *InMemoryCommandBus) RegisterHandlerFunc(cmdType reflect.Type, handlerFunc interface{}) error {
	return bus.RegisterHandler(cmdType, handlerFunc)
}

// Helper function to register handler with type inference
func RegisterCommandHandler[T Command](bus CommandBus, handler CommandHandler[T]) error {
	var cmd T
	cmdType := reflect.TypeOf(cmd)

	// Remove pointer if it's a pointer type
	if cmdType.Kind() == reflect.Ptr {
		cmdType = cmdType.Elem()
	}

	return bus.RegisterHandler(cmdType, handler)
}

// Helper function to register handler function with type inference
func RegisterCommandHandlerFunc[T Command](bus CommandBus, handlerFunc func(context.Context, T) error) error {
	var cmd T
	cmdType := reflect.TypeOf(cmd)

	// Remove pointer if it's a pointer type
	if cmdType.Kind() == reflect.Ptr {
		cmdType = cmdType.Elem()
	}

	return bus.RegisterHandlerFunc(cmdType, handlerFunc)
}

// BaseCommand provides a base implementation for commands
type BaseCommand struct {
	name string
}

// NewBaseCommand creates a new base command
func NewBaseCommand(name string) BaseCommand {
	return BaseCommand{name: name}
}

// CommandName returns the name of the command
func (c BaseCommand) CommandName() string {
	return c.name
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    interface{}            `json:"data,omitempty"`
	Errors  []string               `json:"errors,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// NewSuccessResult creates a successful command result
func NewSuccessResult(message string, data interface{}) CommandResult {
	return CommandResult{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    make(map[string]interface{}),
	}
}

// NewErrorResult creates an error command result
func NewErrorResult(message string, errors ...string) CommandResult {
	return CommandResult{
		Success: false,
		Message: message,
		Errors:  errors,
		Meta:    make(map[string]interface{}),
	}
}

// AddMeta adds metadata to the command result
func (r *CommandResult) AddMeta(key string, value interface{}) {
	if r.Meta == nil {
		r.Meta = make(map[string]interface{})
	}
	r.Meta[key] = value
}

// CommandMiddleware represents middleware for command processing
type CommandMiddleware interface {
	Execute(ctx context.Context, cmd Command, next func(context.Context, Command) error) error
}

// CommandMiddlewareFunc is a function type that implements CommandMiddleware
type CommandMiddlewareFunc func(ctx context.Context, cmd Command, next func(context.Context, Command) error) error

// Execute implements CommandMiddleware interface
func (f CommandMiddlewareFunc) Execute(ctx context.Context, cmd Command, next func(context.Context, Command) error) error {
	return f(ctx, cmd, next)
}

// MiddlewareCommandBus wraps a command bus with middleware support
type MiddlewareCommandBus struct {
	bus         CommandBus
	middlewares []CommandMiddleware
}

// NewMiddlewareCommandBus creates a new middleware command bus
func NewMiddlewareCommandBus(bus CommandBus) *MiddlewareCommandBus {
	return &MiddlewareCommandBus{
		bus:         bus,
		middlewares: make([]CommandMiddleware, 0),
	}
}

// Use adds middleware to the command bus
func (bus *MiddlewareCommandBus) Use(middleware CommandMiddleware) {
	bus.middlewares = append(bus.middlewares, middleware)
}

// Execute executes a command with middleware
func (bus *MiddlewareCommandBus) Execute(ctx context.Context, cmd Command) error {
	return bus.executeWithMiddleware(ctx, cmd, 0)
}

func (bus *MiddlewareCommandBus) executeWithMiddleware(ctx context.Context, cmd Command, index int) error {
	if index >= len(bus.middlewares) {
		return bus.bus.Execute(ctx, cmd)
	}

	middleware := bus.middlewares[index]
	return middleware.Execute(ctx, cmd, func(ctx context.Context, cmd Command) error {
		return bus.executeWithMiddleware(ctx, cmd, index+1)
	})
}

// RegisterHandler registers a command handler
func (bus *MiddlewareCommandBus) RegisterHandler(cmdType reflect.Type, handler interface{}) error {
	return bus.bus.RegisterHandler(cmdType, handler)
}

// RegisterHandlerFunc registers a command handler function
func (bus *MiddlewareCommandBus) RegisterHandlerFunc(cmdType reflect.Type, handlerFunc interface{}) error {
	return bus.bus.RegisterHandlerFunc(cmdType, handlerFunc)
}
