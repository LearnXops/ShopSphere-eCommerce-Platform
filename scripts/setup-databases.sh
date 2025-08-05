#!/bin/bash

# Database setup script for ShopSphere eCommerce platform
# This script sets up all databases, runs migrations, and loads seed data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_USER=${POSTGRES_USER:-shopsphere}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-shopsphere123}
POSTGRES_DB=${POSTGRES_DB:-shopsphere}

MONGODB_HOST=${MONGODB_HOST:-localhost}
MONGODB_PORT=${MONGODB_PORT:-27017}
MONGODB_USER=${MONGODB_USER:-shopsphere}
MONGODB_PASSWORD=${MONGODB_PASSWORD:-shopsphere123}
MONGODB_DB=${MONGODB_DB:-shopsphere}

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-shopsphere123}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for service to be ready
wait_for_service() {
    local service=$1
    local host=$2
    local port=$3
    local max_attempts=30
    local attempt=1

    print_status "Waiting for $service to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if nc -z "$host" "$port" 2>/dev/null; then
            print_success "$service is ready!"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts: $service not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "$service failed to start within expected time"
    return 1
}

# Function to check PostgreSQL connection
check_postgres() {
    print_status "Checking PostgreSQL connection..."
    
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d postgres -c "SELECT 1;" >/dev/null 2>&1; then
        print_success "PostgreSQL connection successful"
        return 0
    else
        print_error "Failed to connect to PostgreSQL"
        return 1
    fi
}

# Function to create databases
create_databases() {
    print_status "Creating databases..."
    
    local databases=("user_service" "product_service" "order_service" "review_service" "auth_service" "payment_service" "shipping_service" "admin_service")
    
    for db in "${databases[@]}"; do
        print_status "Creating database: $db"
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d postgres -c "CREATE DATABASE $db;" 2>/dev/null || print_warning "Database $db may already exist"
    done
    
    print_success "Database creation completed"
}

# Function to install golang-migrate
install_migrate() {
    if command_exists migrate; then
        print_success "golang-migrate is already installed"
        return 0
    fi
    
    print_status "Installing golang-migrate..."
    
    if command_exists go; then
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        print_success "golang-migrate installed successfully"
    else
        print_error "Go is not installed. Please install Go first."
        return 1
    fi
}

# Function to run migrations
run_migrations() {
    print_status "Running database migrations..."
    
    local services=("user_service" "product_service" "order_service" "review_service")
    
    for service in "${services[@]}"; do
        print_status "Running migrations for $service..."
        local db_url="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$service?sslmode=disable"
        
        if migrate -path migrations/postgresql -database "$db_url" up; then
            print_success "Migrations completed for $service"
        else
            print_error "Failed to run migrations for $service"
            return 1
        fi
    done
    
    print_success "All migrations completed successfully"
}

# Function to setup MongoDB
setup_mongodb() {
    print_status "Setting up MongoDB..."
    
    if command_exists mongosh; then
        print_status "Initializing MongoDB collections..."
        mongosh --host $MONGODB_HOST:$MONGODB_PORT --username $MONGODB_USER --password $MONGODB_PASSWORD --authenticationDatabase admin $MONGODB_DB < migrations/mongodb/init-collections.js
        print_success "MongoDB setup completed"
    else
        print_warning "mongosh not found. Please install MongoDB shell to initialize collections."
    fi
}

# Function to setup Redis
setup_redis() {
    print_status "Setting up Redis..."
    
    if command_exists redis-cli; then
        print_status "Testing Redis connection..."
        if redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping | grep -q "PONG"; then
            print_success "Redis connection successful"
        else
            print_warning "Redis connection failed or authentication required"
        fi
    else
        print_warning "redis-cli not found. Please install Redis CLI to test connection."
    fi
}

# Function to load seed data
load_seed_data() {
    local env=${1:-dev}
    print_status "Loading $env seed data..."
    
    local seed_file="migrations/seed/${env}_seed_data.sql"
    
    if [ -f "$seed_file" ]; then
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d user_service -f "$seed_file"
        print_success "$env seed data loaded successfully"
    else
        print_error "Seed file $seed_file not found"
        return 1
    fi
}

# Function to verify setup
verify_setup() {
    print_status "Verifying database setup..."
    
    local services=("user_service" "product_service" "order_service" "review_service")
    
    for service in "${services[@]}"; do
        print_status "Checking $service database..."
        if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $service -c "SELECT COUNT(*) FROM schema_migrations;" >/dev/null 2>&1; then
            local migration_count=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $service -t -c "SELECT COUNT(*) FROM schema_migrations;" | xargs)
            print_success "$service: $migration_count migrations applied"
        else
            print_error "$service: Failed to verify migrations"
        fi
    done
}

# Main setup function
main() {
    print_status "Starting ShopSphere database setup..."
    
    # Check if running in Docker environment
    if [ -n "$DOCKER_ENV" ]; then
        print_status "Running in Docker environment"
        wait_for_service "PostgreSQL" $POSTGRES_HOST $POSTGRES_PORT
        wait_for_service "MongoDB" $MONGODB_HOST $MONGODB_PORT
        wait_for_service "Redis" $REDIS_HOST $REDIS_PORT
    fi
    
    # Check PostgreSQL connection
    if ! check_postgres; then
        print_error "Cannot proceed without PostgreSQL connection"
        exit 1
    fi
    
    # Create databases
    create_databases
    
    # Install migration tool
    install_migrate
    
    # Run migrations
    run_migrations
    
    # Setup MongoDB
    setup_mongodb
    
    # Setup Redis
    setup_redis
    
    # Load seed data (default to dev)
    local environment=${1:-dev}
    load_seed_data $environment
    
    # Verify setup
    verify_setup
    
    print_success "Database setup completed successfully!"
    print_status "You can now start the ShopSphere services."
}

# Handle command line arguments
case "${1:-setup}" in
    "setup")
        main "${2:-dev}"
        ;;
    "migrate")
        install_migrate
        run_migrations
        ;;
    "seed")
        load_seed_data "${2:-dev}"
        ;;
    "verify")
        verify_setup
        ;;
    "help")
        echo "Usage: $0 [command] [environment]"
        echo ""
        echo "Commands:"
        echo "  setup [env]   - Complete database setup (default: dev)"
        echo "  migrate       - Run migrations only"
        echo "  seed [env]    - Load seed data only (dev|test)"
        echo "  verify        - Verify database setup"
        echo "  help          - Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD"
        echo "  MONGODB_HOST, MONGODB_PORT, MONGODB_USER, MONGODB_PASSWORD"
        echo "  REDIS_HOST, REDIS_PORT, REDIS_PASSWORD"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac