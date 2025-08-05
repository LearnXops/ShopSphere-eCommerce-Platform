# User Service

The User Service is a microservice responsible for user management in the ShopSphere eCommerce platform. It handles user registration, authentication, profile management, and administrative operations.

## Features

- **User Registration**: Register new users with email verification
- **Profile Management**: CRUD operations for user profiles
- **Password Management**: Password changes and reset functionality
- **User Status Management**: Admin operations for user status (active, suspended, deleted)
- **User Search and Filtering**: Admin operations for user discovery
- **Email Verification**: Secure email verification process
- **Password Reset**: Secure password reset with tokens

## API Endpoints

### Public Endpoints

- `POST /api/v1/users/register` - Register a new user
- `POST /api/v1/users/verify-email` - Verify user email address
- `POST /api/v1/users/password-reset/request` - Request password reset
- `POST /api/v1/users/password-reset/confirm` - Confirm password reset

### User Endpoints (Authentication Required)

- `GET /api/v1/users/{id}` - Get user profile
- `PUT /api/v1/users/{id}` - Update user profile
- `PUT /api/v1/users/{id}/password` - Change user password
- `DELETE /api/v1/users/{id}` - Delete user account (soft delete)

### Admin Endpoints (Admin Role Required)

- `GET /api/v1/users` - List users with pagination and filtering
- `GET /api/v1/users/search` - Search users
- `PUT /api/v1/users/{id}/status` - Update user status

### Health Check

- `GET /health` - Service health check

## Request/Response Examples

### User Registration

**Request:**
```bash
curl -X POST http://localhost:8002/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "password": "SecurePassword123!"
  }'
```

**Response:**
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "role": "customer",
    "status": "pending",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "message": "User registered successfully. Please check your email for verification."
}
```

### Email Verification

**Request:**
```bash
curl -X POST http://localhost:8002/api/v1/users/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "token": "verification_token_here"
  }'
```

**Response:**
```json
{
  "message": "Email verified successfully"
}
```

### Update User Profile

**Request:**
```bash
curl -X PUT http://localhost:8002/api/v1/users/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "phone": "+1234567890"
  }'
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "username": "johndoe",
  "first_name": "Jane",
  "last_name": "Doe",
  "phone": "+1234567890",
  "role": "customer",
  "status": "active",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:30:00Z"
}
```

### List Users (Admin)

**Request:**
```bash
curl "http://localhost:8002/api/v1/users?limit=10&offset=0&status=active"
```

**Response:**
```json
{
  "users": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "username": "johndoe",
      "first_name": "John",
      "last_name": "Doe",
      "role": "customer",
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

## Environment Variables

- `PORT` - Service port (default: 8002)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: shopsphere)
- `DB_PASSWORD` - Database password (default: shopsphere123)
- `DB_NAME` - Database name (default: user_service)

## Database Schema

The service uses PostgreSQL with the following main tables:

- `users` - User account information
- `password_reset_tokens` - Password reset tokens
- `email_verification_tokens` - Email verification tokens
- `user_sessions` - User session tracking
- `addresses` - User addresses

## Running the Service

### Prerequisites

- Go 1.21+
- PostgreSQL database
- Database migrations applied

### Local Development

1. Set up the database:
```bash
# Run database setup script
../../scripts/setup-databases.sh
```

2. Run the service:
```bash
go run .
```

3. The service will be available at `http://localhost:8002`

### Using Docker

```bash
docker build -t user-service .
docker run -p 8002:8002 -e DB_HOST=host.docker.internal user-service
```

## Testing

### Unit Tests

Run unit tests for the service layer:
```bash
go test ./internal/service -v
```

### Integration Tests

Run integration tests with test containers:
```bash
go test -v -timeout 60s
```

### Manual Testing

Use the provided test script:
```bash
./test_service.sh
```

## Architecture

The service follows a layered architecture:

```
├── main.go                 # Application entry point
├── internal/
│   ├── handlers/          # HTTP handlers
│   ├── service/           # Business logic
│   └── repository/        # Data access layer
├── integration_test.go    # Integration tests
└── README.md
```

### Layers

1. **Handler Layer** (`internal/handlers/`): HTTP request/response handling
2. **Service Layer** (`internal/service/`): Business logic and validation
3. **Repository Layer** (`internal/repository/`): Database operations

## Security Features

- **Password Hashing**: Uses bcrypt for secure password storage
- **Token-based Operations**: Secure tokens for email verification and password reset
- **Input Validation**: Comprehensive input validation and sanitization
- **SQL Injection Protection**: Parameterized queries
- **Rate Limiting**: (Implemented at API Gateway level)

## Error Handling

The service uses structured error handling with standardized error codes:

- `VALIDATION_ERROR` - Input validation failures
- `AUTHENTICATION_ERROR` - Authentication failures
- `AUTHORIZATION_ERROR` - Permission denied
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource conflict (duplicate email/username)
- `INTERNAL_ERROR` - Server errors

## Logging

The service uses structured logging with:

- Request/response logging
- Error logging with stack traces
- Performance metrics
- Correlation IDs for request tracing

## Monitoring

Health check endpoint provides:
- Service status
- Database connectivity
- Basic service metrics

## Development

### Adding New Features

1. Define the interface in `internal/repository/interfaces.go`
2. Implement repository methods in `internal/repository/user_repository.go`
3. Add business logic in `internal/service/user_service.go`
4. Create HTTP handlers in `internal/handlers/user_handlers.go`
5. Add routes in `main.go`
6. Write tests for all layers

### Code Style

- Follow Go conventions
- Use structured logging
- Implement comprehensive error handling
- Write unit tests for business logic
- Write integration tests for API endpoints

## Dependencies

- `github.com/gorilla/mux` - HTTP router
- `github.com/lib/pq` - PostgreSQL driver
- `golang.org/x/crypto` - Password hashing
- `github.com/testcontainers/testcontainers-go` - Integration testing
- `github.com/shopsphere/shared` - Shared utilities and models