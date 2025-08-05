# Database Setup Guide

This guide covers the complete database setup for the ShopSphere eCommerce platform, including PostgreSQL, MongoDB, and Redis configurations.

## Overview

The ShopSphere platform uses multiple database technologies:

- **PostgreSQL**: Primary relational database for transactional data
- **MongoDB**: Document store for product catalog and analytics
- **Redis**: In-memory cache for sessions and temporary data

## Quick Start

### Prerequisites

1. Docker and Docker Compose installed
2. Go 1.19+ installed (for migration tool)
3. Make utility installed

### Automated Setup

1. Start the development environment:
```bash
make dev-up
```

2. Run the complete database setup:
```bash
make db-setup
```

This will:
- Install golang-migrate tool
- Run all PostgreSQL migrations
- Initialize MongoDB collections
- Configure Redis
- Load development seed data

### Manual Setup

If you prefer to set up databases manually:

```bash
# Install migration tool
make migrate-install

# Run PostgreSQL migrations
make migrate-up

# Initialize MongoDB
make mongo-init

# Configure Redis
make redis-config

# Load seed data
make seed-dev
```

## Database Schemas

### PostgreSQL Databases

The platform uses separate PostgreSQL databases for each service:

#### User Service Database (`user_service`)
- `users` - User accounts and authentication
- `addresses` - User shipping/billing addresses
- `user_sessions` - Active user sessions
- `password_reset_tokens` - Password reset tokens
- `email_verification_tokens` - Email verification tokens

#### Product Service Database (`product_service`)
- `categories` - Product categories (hierarchical)
- `products` - Product catalog
- `product_variants` - Product variations (size, color, etc.)
- `product_images` - Product image management
- `product_tags` - Flexible tagging system
- `product_tag_relations` - Product-tag relationships
- `inventory_movements` - Stock movement tracking

#### Order Service Database (`order_service`)
- `orders` - Customer orders
- `order_items` - Order line items
- `order_status_history` - Order status tracking
- `order_discounts` - Applied discounts
- `shopping_carts` - Shopping cart storage
- `shopping_cart_items` - Cart items
- `order_fulfillments` - Fulfillment tracking
- `order_fulfillment_items` - Fulfillment line items

#### Review Service Database (`review_service`)
- `reviews` - Product reviews and ratings
- `review_votes` - Review helpfulness votes
- `review_reports` - Reported inappropriate reviews
- `product_review_summaries` - Aggregated review data
- `review_responses` - Merchant/admin responses
- `review_response_templates` - Response templates

### MongoDB Collections

The MongoDB database (`shopsphere`) contains:

#### Product Catalog
- `product_catalog` - Rich product descriptions and media
- `category_hierarchy` - Complex category structures

#### Analytics
- `user_analytics` - User behavior tracking
- `product_analytics` - Product performance metrics
- `search_analytics` - Search query analysis
- `inventory_snapshots` - Historical inventory data

#### Recommendations
- `user_preferences` - User preference profiles
- `product_recommendations` - Cached recommendations

### Redis Configuration

Redis is configured for:
- Session storage (database 0)
- Product caching (database 1)
- Cart data (database 2)
- Rate limiting (database 3)

## Migration Management

### Creating New Migrations

```bash
make migrate-create NAME=add_user_preferences
```

This creates two files:
- `migrations/postgresql/YYYYMMDDHHMMSS_add_user_preferences.up.sql`
- `migrations/postgresql/YYYYMMDDHHMMSS_add_user_preferences.down.sql`

### Running Migrations

```bash
# Run all pending migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-version
```

### Migration Best Practices

1. **Always create both up and down migrations**
2. **Test migrations on a copy of production data**
3. **Use transactions for complex migrations**
4. **Add indexes after data insertion for better performance**
5. **Document breaking changes in migration comments**

## Seed Data

### Development Data

The development seed data includes:
- Sample users (admin, customers, moderator)
- Product categories and products
- Sample orders and reviews
- Test shopping carts

Load with:
```bash
make seed-dev
```

### Test Data

Minimal test data for automated testing:
- Basic test users and products
- Simple test scenarios

Load with:
```bash
make seed-test
```

## Database Connections

### Connection Strings

#### PostgreSQL
```
postgres://shopsphere:shopsphere123@localhost:5432/{service_name}?sslmode=disable
```

#### MongoDB
```
mongodb://shopsphere:shopsphere123@localhost:27017/shopsphere
```

#### Redis
```
redis://localhost:6379/0
```

### Connection Pooling

PostgreSQL connections use the following pool settings:
- Max Open Connections: 25
- Max Idle Connections: 5
- Connection Max Lifetime: 5 minutes

## Monitoring and Maintenance

### Health Checks

Each service should implement database health checks:

```go
func (s *Service) HealthCheck() error {
    return utils.DatabaseHealthCheck(s.serviceName)
}
```

### Backup Strategy

1. **PostgreSQL**: Use `pg_dump` for logical backups
2. **MongoDB**: Use `mongodump` for document backups
3. **Redis**: Use RDB snapshots for persistence

### Performance Monitoring

Monitor these key metrics:
- Connection pool utilization
- Query execution times
- Index usage statistics
- Cache hit rates (Redis)

## Troubleshooting

### Common Issues

#### Migration Failures
```bash
# Check current version
make migrate-version

# Force to specific version (use with caution)
make migrate-force VERSION=1 SERVICE=user_service
```

#### Connection Issues
```bash
# Test database connectivity
docker exec -it shopsphere-postgres psql -U shopsphere -d user_service -c "SELECT 1;"
```

#### Reset Everything
```bash
# Complete database reset (destructive!)
make db-reset
```

### Logs and Debugging

Check container logs:
```bash
docker logs shopsphere-postgres
docker logs shopsphere-mongodb
docker logs shopsphere-redis
```

## Security Considerations

1. **Use environment variables for credentials**
2. **Enable SSL/TLS in production**
3. **Implement proper access controls**
4. **Regular security updates**
5. **Audit database access logs**

## Production Deployment

### Environment Variables

Set these environment variables for production:

```bash
# PostgreSQL
POSTGRES_HOST=your-postgres-host
POSTGRES_PORT=5432
POSTGRES_USER=your-username
POSTGRES_PASSWORD=your-secure-password

# MongoDB
MONGODB_HOST=your-mongodb-host
MONGODB_PORT=27017
MONGODB_USER=your-username
MONGODB_PASSWORD=your-secure-password

# Redis
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=your-secure-password
```

### High Availability

Consider these for production:
- PostgreSQL streaming replication
- MongoDB replica sets
- Redis Sentinel for failover
- Connection pooling with PgBouncer
- Database monitoring with Prometheus

## API Documentation

Database utility functions are available in `shared/utils/database.go`:

- `NewDatabaseConfig()` - Create database configuration
- `Connect()` - Establish database connection
- `CheckDatabaseConnection()` - Verify connectivity
- `DatabaseHealthCheck()` - Health check endpoint
- `ExecuteInTransaction()` - Transaction wrapper

For more details, see the source code documentation.