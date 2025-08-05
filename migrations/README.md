# Database Migrations

This directory contains database migration files for the ShopSphere eCommerce platform.

## Structure

- `postgresql/` - PostgreSQL migration files
- `mongodb/` - MongoDB initialization scripts
- `redis/` - Redis configuration files
- `seed/` - Seed data scripts for development and testing

## Usage

### PostgreSQL Migrations

Use golang-migrate to run PostgreSQL migrations:

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations up
make migrate-up

# Run migrations down
make migrate-down

# Create new migration
make migrate-create name=your_migration_name
```

### Seed Data

Load seed data for development:

```bash
make seed-dev
```

Load seed data for testing:

```bash
make seed-test
```

## Migration Files

Migration files follow the naming convention:
- `YYYYMMDDHHMMSS_description.up.sql` - Forward migration
- `YYYYMMDDHHMMSS_description.down.sql` - Rollback migration

Each service has its own database and migration files are organized by service.