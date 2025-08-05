# ShopSphere Project Structure

This document describes the monorepo structure for the ShopSphere eCommerce platform.

## Directory Structure

```
.
├── README.md                           # Project overview and documentation
├── PROJECT_STRUCTURE.md               # This file
├── Makefile                           # Build and development commands
├── docker-compose.yml                # Local development environment
├── go.work                           # Go workspace configuration
├── go.work.sum                       # Go workspace checksums
├── .env.example                      # Environment variables template
│
├── services/                         # Microservices
│   ├── auth-service/                 # JWT authentication and authorization
│   ├── user-service/                 # User registration and profile management
│   ├── product-service/              # Product catalog and inventory
│   ├── cart-service/                 # Shopping cart operations
│   ├── order-service/                # Order processing and management
│   ├── payment-service/              # Payment processing and gateway integration
│   ├── shipping-service/             # Shipping calculations and tracking
│   ├── review-service/               # Product reviews and ratings
│   ├── notification-service/         # Email, SMS, and push notifications
│   ├── admin-service/                # Administrative operations
│   ├── search-service/               # Full-text search and filtering
│   └── recommendation-service/       # ML-based product recommendations
│
├── shared/                           # Common libraries and utilities
│   ├── models/                       # Domain models (User, Product, Order)
│   ├── utils/                        # Utilities (logger, errors, database)
│   └── middleware/                   # Common middleware (auth, CORS, etc.)
│
├── deployments/                      # Kubernetes manifests and configurations
│   ├── kong/                         # API Gateway configuration
│   └── monitoring/                   # Observability stack configuration
│       ├── prometheus.yml            # Prometheus configuration
│       └── grafana/                  # Grafana dashboards and datasources
│
└── scripts/                          # Development and deployment scripts
    └── init-db.sql                   # Database initialization script
```

## Service Architecture

Each microservice follows a consistent structure:

```
services/{service-name}/
├── main.go                           # Service entry point
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksums
├── Dockerfile                       # Container build configuration
└── bin/                             # Compiled binaries (generated)
    └── {service-name}
```

## Shared Libraries

The `shared/` directory contains common code used across all services:

- **models/**: Domain models with consistent data structures
- **utils/**: Utility functions for logging, error handling, and database connections
- **middleware/**: HTTP middleware for authentication, CORS, and request handling

## Development Environment

The project includes a complete development environment with:

- **PostgreSQL**: Primary database for transactional data
- **Redis**: Caching and session storage
- **MongoDB**: Document storage for product catalogs
- **Elasticsearch**: Full-text search and analytics
- **Kafka**: Event streaming and message queues
- **Kong**: API Gateway with rate limiting and authentication
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization
- **Jaeger**: Distributed tracing

## Getting Started

1. **Prerequisites**:
   - Go 1.21+
   - Docker and Docker Compose
   - Make

2. **Start Development Environment**:
   ```bash
   make dev-up
   ```

3. **Build All Services**:
   ```bash
   make build
   ```

4. **Run Tests**:
   ```bash
   make test
   ```

5. **Stop Development Environment**:
   ```bash
   make dev-down
   ```

## Service Ports

- Auth Service: 8001
- User Service: 8002
- Product Service: 8003
- Cart Service: 8004
- Order Service: 8005
- Payment Service: 8006
- Shipping Service: 8007
- Review Service: 8008
- Notification Service: 8009
- Admin Service: 8010
- Search Service: 8011
- Recommendation Service: 8012

## Infrastructure Ports

- PostgreSQL: 5432
- Redis: 6379
- MongoDB: 27017
- Elasticsearch: 9200
- Kafka: 9092
- Kong API Gateway: 8000
- Prometheus: 9090
- Grafana: 3000
- Jaeger: 16686

## Next Steps

This foundation provides:

✅ Complete monorepo structure with 12 microservices
✅ Shared libraries for common functionality
✅ Go workspace configuration for dependency management
✅ Docker Compose development environment
✅ Build system with Makefile
✅ Basic service health checks
✅ API Gateway configuration
✅ Observability stack setup
✅ Database initialization scripts

The project is now ready for implementing individual service functionality according to the requirements and design specifications.