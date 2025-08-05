# ShopSphere Makefile

.PHONY: help build test clean dev-up dev-down docker-build docker-push

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build all services"
	@echo "  test         - Run tests for all services"
	@echo "  clean        - Clean build artifacts"
	@echo "  dev-up       - Start development environment"
	@echo "  dev-down     - Stop development environment"
	@echo "  docker-build - Build Docker images for all services"
	@echo "  docker-push  - Push Docker images to registry"
	@echo ""
	@echo "Database commands:"
	@echo "  migrate-install - Install golang-migrate tool"
	@echo "  migrate-up      - Run database migrations up"
	@echo "  migrate-down    - Roll back database migrations"
	@echo "  migrate-version - Check migration versions"
	@echo "  migrate-create  - Create new migration (NAME=migration_name)"
	@echo "  seed-dev        - Load development seed data"
	@echo "  seed-test       - Load test seed data"
	@echo "  mongo-init      - Initialize MongoDB collections"
	@echo "  redis-config    - Apply Redis configuration"
	@echo "  db-setup        - Complete database setup"
	@echo "  db-reset        - Reset all databases (destructive)"

# Build all services
build:
	@echo "Building all services..."
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		echo "Building $$service-service..."; \
		cd services/$$service-service && go build -o bin/$$service-service ./... && cd ../..; \
	done
	@echo "Build complete!"

# Run tests for all services
test:
	@echo "Running tests for all services..."
	@cd shared && go test ./... -v
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		echo "Testing $$service-service..."; \
		cd services/$$service-service && go test ./... -v && cd ../..; \
	done
	@echo "Tests complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		rm -rf services/$$service-service/bin; \
	done
	@echo "Clean complete!"

# Start development environment
dev-up:
	@echo "Starting development environment..."
	@docker-compose up -d
	@echo "Development environment started!"
	@echo "Services available at:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - MongoDB: localhost:27017"
	@echo "  - Elasticsearch: localhost:9200"
	@echo "  - Kafka: localhost:9092"
	@echo "  - Kong API Gateway: localhost:8000"
	@echo "  - Prometheus: localhost:9090"
	@echo "  - Grafana: localhost:3000 (admin/admin123)"
	@echo "  - Jaeger: localhost:16686"

# Stop development environment
dev-down:
	@echo "Stopping development environment..."
	@docker-compose down
	@echo "Development environment stopped!"

# Build Docker images for all services
docker-build:
	@echo "Building Docker images..."
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		echo "Building Docker image for $$service-service..."; \
		docker build -t shopsphere/$$service-service:latest -f services/$$service-service/Dockerfile .; \
	done
	@echo "Docker images built!"

# Push Docker images to registry
docker-push:
	@echo "Pushing Docker images to registry..."
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		echo "Pushing $$service-service..."; \
		docker push shopsphere/$$service-service:latest; \
	done
	@echo "Docker images pushed!"

# Initialize Go modules
init-modules:
	@echo "Initializing Go modules..."
	@cd shared && go mod tidy
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		echo "Initializing $$service-service module..."; \
		cd services/$$service-service && go mod tidy && cd ../..; \
	done
	@echo "Go modules initialized!"

# Format code
fmt:
	@echo "Formatting code..."
	@cd shared && go fmt ./...
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		cd services/$$service-service && go fmt ./... && cd ../..; \
	done
	@echo "Code formatted!"

# Lint code
lint:
	@echo "Linting code..."
	@cd shared && golangci-lint run
	@for service in auth user product cart order payment shipping review notification admin search recommendation; do \
		cd services/$$service-service && golangci-lint run && cd ../..; \
	done
	@echo "Linting complete!"# Dat
abase migration commands
migrate-install:
	@echo "Installing golang-migrate..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "golang-migrate installed!"

migrate-up:
	@echo "Running database migrations up..."
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/user_service?sslmode=disable" up
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/product_service?sslmode=disable" up
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/order_service?sslmode=disable" up
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/review_service?sslmode=disable" up
	@echo "Database migrations completed!"

migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/review_service?sslmode=disable" down
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/order_service?sslmode=disable" down
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/product_service?sslmode=disable" down
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/user_service?sslmode=disable" down
	@echo "Database migrations rolled back!"

migrate-force:
	@echo "Forcing migration version (use with caution)..."
	@echo "Usage: make migrate-force VERSION=<version> SERVICE=<service>"
	@if [ -z "$(VERSION)" ] || [ -z "$(SERVICE)" ]; then \
		echo "Error: VERSION and SERVICE parameters are required"; \
		echo "Example: make migrate-force VERSION=1 SERVICE=user_service"; \
		exit 1; \
	fi
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/$(SERVICE)?sslmode=disable" force $(VERSION)

migrate-version:
	@echo "Checking migration versions..."
	@echo "User Service:"
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/user_service?sslmode=disable" version || echo "No migrations applied"
	@echo "Product Service:"
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/product_service?sslmode=disable" version || echo "No migrations applied"
	@echo "Order Service:"
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/order_service?sslmode=disable" version || echo "No migrations applied"
	@echo "Review Service:"
	@migrate -path migrations/postgresql -database "postgres://shopsphere:shopsphere123@localhost:5432/review_service?sslmode=disable" version || echo "No migrations applied"

migrate-create:
	@echo "Creating new migration..."
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME parameter is required"; \
		echo "Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir migrations/postgresql -seq $(NAME)
	@echo "Migration files created for: $(NAME)"

# Database seeding commands
seed-dev:
	@echo "Loading development seed data..."
	@docker exec -i shopsphere-postgres psql -U shopsphere -d user_service < migrations/seed/dev_seed_data.sql
	@echo "Development seed data loaded!"

seed-test:
	@echo "Loading test seed data..."
	@docker exec -i shopsphere-postgres psql -U shopsphere -d user_service < migrations/seed/test_seed_data.sql
	@echo "Test seed data loaded!"

# MongoDB setup commands
mongo-init:
	@echo "Initializing MongoDB collections..."
	@docker exec -i shopsphere-mongodb mongosh --username shopsphere --password shopsphere123 --authenticationDatabase admin < migrations/mongodb/init-collections.js
	@echo "MongoDB collections initialized!"

# Redis setup commands
redis-config:
	@echo "Applying Redis configuration..."
	@docker cp migrations/redis/redis.conf shopsphere-redis:/usr/local/etc/redis/redis.conf
	@docker restart shopsphere-redis
	@echo "Redis configuration applied!"

# Complete database setup
db-setup: migrate-install migrate-up mongo-init redis-config seed-dev
	@echo "Complete database setup finished!"
	@echo "All databases, collections, and seed data have been configured."

# Reset all databases (use with caution)
db-reset:
	@echo "Resetting all databases..."
	@docker-compose down -v
	@docker-compose up -d postgres mongodb redis
	@sleep 10
	@make db-setup
	@echo "Database reset complete!"