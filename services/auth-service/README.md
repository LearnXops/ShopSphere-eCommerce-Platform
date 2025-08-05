# Auth Service

The Auth Service handles user authentication and authorization for the ShopSphere eCommerce platform. It provides JWT-based authentication with secure token management, password hashing, and role-based access control (RBAC).

## Features

- **JWT Token Management**: Secure access and refresh token generation and validation
- **Password Security**: bcrypt password hashing with configurable cost
- **Token Refresh**: Secure refresh token mechanism with session tracking
- **Role-Based Access Control**: RBAC middleware for different user roles
- **Session Management**: Database-backed session storage with cleanup
- **Comprehensive Testing**: Unit and integration tests with >90% coverage

## API Endpoints

### Authentication Endpoints

- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - User logout
- `POST /auth/validate` - Validate access token

### Protected Endpoints

- `GET /auth/me` - Get current user information (requires authentication)

## Configuration

The service can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | Server port |
| `DATABASE_URL` | `postgres://user:password@localhost/shopsphere_auth?sslmode=disable` | PostgreSQL connection string |
| `JWT_ACCESS_SECRET` | `your-super-secret-access-key-change-in-production` | JWT access token secret |
| `JWT_REFRESH_SECRET` | `your-super-secret-refresh-key-change-in-production` | JWT refresh token secret |
| `JWT_ISSUER` | `shopsphere-auth` | JWT issuer |
| `JWT_ACCESS_TTL` | `15m` | Access token TTL |
| `JWT_REFRESH_TTL` | `168h` | Refresh token TTL (7 days) |

## Password Requirements

- Minimum 8 characters
- Maximum 128 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character

## Token Structure

### Access Token Claims
```json
{
  "user_id": "uuid",
  "role": "customer|admin|moderator",
  "status": "active|suspended|deleted|pending",
  "iss": "shopsphere-auth",
  "sub": "user_id",
  "aud": ["shopsphere-api"],
  "exp": 1234567890,
  "nbf": 1234567890,
  "iat": 1234567890
}
```

### Refresh Token Claims
```json
{
  "user_id": "uuid",
  "token_hash": "secure_hash",
  "iss": "shopsphere-auth",
  "sub": "user_id",
  "aud": ["shopsphere-refresh"],
  "exp": 1234567890,
  "nbf": 1234567890,
  "iat": 1234567890
}
```

## Usage Examples

### Login
```bash
curl -X POST http://localhost:8001/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### Access Protected Endpoint
```bash
curl -X GET http://localhost:8001/auth/me \
  -H "Authorization: Bearer <access_token>"
```

### Refresh Token
```bash
curl -X POST http://localhost:8001/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -v -run TestIntegration
```

### Building
```bash
go build -o bin/auth-service .
```

### Running Locally
```bash
# Set environment variables
export DATABASE_URL="postgres://user:password@localhost/shopsphere_auth?sslmode=disable"
export JWT_ACCESS_SECRET="your-secret-key"
export JWT_REFRESH_SECRET="your-refresh-secret"

# Run the service
./bin/auth-service
```

## Security Considerations

1. **Secrets Management**: Use secure secret management in production
2. **HTTPS Only**: Always use HTTPS in production
3. **Token Storage**: Store refresh tokens securely on client side
4. **Session Cleanup**: Expired sessions are automatically cleaned up every hour
5. **Password Hashing**: Uses bcrypt with cost 12 for secure password storage
6. **Token Validation**: All tokens are validated for expiry and signature

## Architecture

The service follows clean architecture principles:

- **Handlers**: HTTP request/response handling
- **Service**: Business logic and orchestration
- **Repository**: Data access layer
- **JWT**: Token management utilities
- **Auth**: Password and security utilities
- **Middleware**: RBAC and authentication middleware

## Dependencies

- `gorilla/mux`: HTTP routing
- `golang-jwt/jwt/v5`: JWT token handling
- `golang.org/x/crypto`: Password hashing
- `lib/pq`: PostgreSQL driver
- `google/uuid`: UUID generation