# ShopSphere eCommerce Platform

A cloud-native, microservices-based eCommerce platform built with Go, Kubernetes, and modern DevOps practices.

## Architecture

This is a monorepo containing all microservices for the ShopSphere platform:

- **services/**: Individual microservices
- **shared/**: Common libraries and utilities
- **deployments/**: Kubernetes manifests and Helm charts
- **scripts/**: Development and deployment scripts
- **docs/**: Documentation and specifications

## Services

- `auth-service`: JWT token management and authorization
- `user-service`: User registration, authentication, and profile management
- `product-service`: Product catalog and inventory management
- `cart-service`: Shopping cart operations and session management
- `order-service`: Order processing and status management
- `payment-service`: Payment processing and gateway integration
- `shipping-service`: Shipping calculations and tracking
- `review-service`: Product reviews and ratings
- `notification-service`: Email, SMS, and push notifications
- `admin-service`: Administrative operations and reporting
- `search-service`: Full-text search and filtering
- `recommendation-service`: ML-based product recommendations

## Development

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- kubectl
- Helm 3.x

### Local Development

```bash
# Start all services with dependencies
docker-compose up -d

# Run a specific service
cd services/user-service
go run main.go

# Run tests
make test

# Build all services
make build
```

## Deployment

The platform uses GitOps with ArgoCD for deployment management. See `deployments/` directory for Kubernetes manifests and Helm charts.# ShopSphere-eCommerce-Platform
