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
	@echo "Linting complete!"