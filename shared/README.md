# ShopSphere Shared Package

This package contains shared domain models, utilities, and common functionality used across all ShopSphere microservices.

## Overview

The shared package provides:
- **Domain Models**: Common data structures for User, Product, Order, Cart, Review, Notification, Payment, and Events
- **Error Handling**: Standardized error structures and utilities
- **Database Utilities**: Connection pooling and database management
- **Logging**: Structured logging with correlation IDs and tracing
- **Validation**: Input sanitization and business rule validation
- **Circuit Breaker**: Resilience patterns for service communication

## Domain Models

### Core Models

#### User (`models/user.go`)
- User account management with roles and status
- Address management for shipping and billing
- Support for customer, admin, and moderator roles

#### Product (`models/product.go`)
- Product catalog with SKU, pricing, and inventory
- Category hierarchy support
- Rich product attributes and dimensions

#### Order (`models/order.go`)
- Order lifecycle management with status tracking
- Order items with pricing and quantity
- Payment method integration

#### Cart (`models/cart.go`)
- Shopping cart with session management
- Cart item management with automatic total calculation
- Expiration and abandonment tracking

#### Review (`models/review.go`)
- Product reviews with rating system (1-5 stars)
- Review moderation and verification
- Aggregated review summaries

#### Notification (`models/notification.go`)
- Multi-channel notifications (email, SMS, push)
- Template-based messaging with variables
- Retry mechanisms and delivery tracking

#### Payment (`models/payment.go`)
- Payment processing with multiple gateways
- Payment method storage and management
- Refund processing and tracking

### Event Models (`models/events.go`)

Domain events for event-driven architecture:
- User events (registered, updated, deleted)
- Product events (created, updated, inventory changes)
- Order events (created, confirmed, shipped, delivered)
- Payment events (processed, failed, refunded)
- Cart events (item added/removed, abandoned)
- Review events (created, updated)

## Utilities

### Error Handling (`utils/errors.go`)

Standardized error handling with:
- Application-specific error codes
- HTTP status code mapping
- Error response structures
- Contextual error information

```go
// Create application errors
err := NewValidationError("Invalid email format")
err := NewNotFoundError("User")
err := NewInternalError("Database connection failed", cause)

// Get HTTP status code
statusCode := err.HTTPStatusCode()
```

### Structured Logging (`utils/logger.go`)

JSON-structured logging with:
- Correlation IDs and trace IDs
- Context-aware logging
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- HTTP middleware for request logging

```go
// Context-aware logging
ctx = WithTraceID(ctx)
ctx = WithUserID(ctx, userID)

Logger.Info(ctx, "User created", map[string]interface{}{
    "user_id": userID,
    "email": email,
})

// HTTP middleware
router.Use(LogMiddleware("user-service"))
```

### Database Utilities (`utils/database.go`)

Database connection management with:
- Connection pooling configuration
- Health checks and monitoring
- PostgreSQL and Redis support
- Connection statistics

```go
// Create database connection
config := DefaultDatabaseConfig()
config.Host = "localhost"
config.Database = "shopsphere"

db, err := NewPostgresConnection(config)
if err != nil {
    log.Fatal(err)
}

// Health check
if err := db.Health(ctx); err != nil {
    log.Error("Database health check failed", err)
}
```

### Validation (`utils/validation.go`)

Comprehensive input validation with:
- Field-level validation rules
- Business rule validation
- Input sanitization
- Pre-built validators for common use cases

```go
// Manual validation
v := NewValidator()
v.Required("email", email).Email("email", email)
v.Required("password", password).Password("password", password)

if v.HasErrors() {
    return v.Errors()
}

// Pre-built validators
errors := ValidateUserRegistration(email, username, firstName, lastName, password)
errors := ValidateProductCreation(sku, name, description, price)
```

### Circuit Breaker (`utils/circuit_breaker.go`)

Resilience patterns for service communication:
- Configurable failure thresholds
- Automatic recovery mechanisms
- State monitoring and callbacks
- Context-aware execution

```go
// Create circuit breaker
config := DefaultCircuitBreakerConfig("user-service")
cb := NewCircuitBreaker(config)

// Execute with circuit breaker
result, err := cb.Execute(func() (interface{}, error) {
    return userService.GetUser(userID)
})

// Execute with context
result, err := cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
    return userService.GetUserWithContext(ctx, userID)
})
```

## Usage Examples

### Service Integration

```go
package main

import (
    "context"
    "github.com/shopsphere/shared/models"
    "github.com/shopsphere/shared/utils"
)

func main() {
    // Initialize logger
    utils.Logger.SetServiceName("user-service")
    utils.Logger.SetLevel(utils.LogLevelInfo)
    
    // Create database connection
    dbConfig := utils.DefaultDatabaseConfig()
    dbConfig.Database = "users"
    db, err := utils.NewPostgresConnection(dbConfig)
    if err != nil {
        utils.Logger.Fatal(context.Background(), "Failed to connect to database", err)
    }
    
    // Create circuit breaker for external service
    cbConfig := utils.DefaultCircuitBreakerConfig("payment-service")
    cb := utils.NewCircuitBreaker(cbConfig)
    
    // Use domain models
    user := models.NewUser("john@example.com", "johndoe", "John", "Doe")
    
    // Validate user data
    errors := utils.ValidateUserRegistration(
        user.Email, user.Username, user.FirstName, user.LastName, "password123")
    if errors.HasErrors() {
        utils.Logger.Error(context.Background(), "Validation failed", nil, 
            map[string]interface{}{"errors": errors})
        return
    }
    
    // Log with context
    ctx := utils.WithTraceID(context.Background())
    ctx = utils.WithUserID(ctx, user.ID)
    utils.Logger.Info(ctx, "User registration started")
}
```

### Event Publishing

```go
// Create and publish domain event
eventData := models.UserRegisteredData{
    UserID:    user.ID,
    Email:     user.Email,
    Username:  user.Username,
    FirstName: user.FirstName,
    LastName:  user.LastName,
}

metadata := models.EventMetadata{
    UserID:      user.ID,
    ServiceName: "user-service",
    TraceID:     utils.GetTraceID(ctx),
}

event, err := models.NewDomainEvent(
    models.EventUserRegistered,
    user.ID,
    eventData,
    metadata,
)

if err != nil {
    utils.Logger.Error(ctx, "Failed to create domain event", err)
    return
}

// Publish event to message queue
publisher.Publish(event)
```

## Testing

The package includes comprehensive tests for all utilities:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestValidator_Email ./utils
```

## Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/shopspring/decimal` - Decimal arithmetic for financial calculations
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/golang-jwt/jwt/v5` - JWT token handling

## Configuration

### Environment Variables

The shared package respects the following environment variables:

- `LOG_LEVEL` - Set logging level (DEBUG, INFO, WARN, ERROR, FATAL)
- `SERVICE_NAME` - Service name for logging and tracing
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_NAME` - Database name
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password

### Best Practices

1. **Error Handling**: Always use the standardized error types and include context
2. **Logging**: Use structured logging with correlation IDs for traceability
3. **Validation**: Validate all input data using the provided validators
4. **Database**: Use connection pooling and health checks
5. **Circuit Breakers**: Implement circuit breakers for external service calls
6. **Events**: Use domain events for loose coupling between services

## Contributing

When adding new functionality to the shared package:

1. Follow the existing patterns and conventions
2. Add comprehensive tests for new functionality
3. Update this README with usage examples
4. Ensure backward compatibility
5. Add appropriate validation and error handling