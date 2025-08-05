package utils

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabaseConfig creates a new database configuration with defaults
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:    "localhost",
		Port:    5432,
		User:    "shopsphere",
		Password: "shopsphere123",
		SSLMode: "disable",
	}
}

// ConnectionString returns the PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Connect establishes a connection to the PostgreSQL database
func (c *DatabaseConfig) Connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// MigrateDatabase runs database migrations for a specific service
func MigrateDatabase(serviceName string) error {
	config := NewDatabaseConfig()
	config.DBName = serviceName

	db, err := config.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to %s database: %w", serviceName, err)
	}
	defer db.Close()

	// Create schema_migrations table if it doesn't exist
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL PRIMARY KEY,
			dirty boolean NOT NULL
		);
	`
	
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// CheckDatabaseConnection verifies database connectivity
func CheckDatabaseConnection(serviceName string) error {
	config := NewDatabaseConfig()
	config.DBName = serviceName

	db, err := config.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// DatabaseHealthCheck performs a health check on the database
func DatabaseHealthCheck(serviceName string) (bool, error) {
	config := NewDatabaseConfig()
	config.DBName = serviceName

	db, err := config.Connect()
	if err != nil {
		return false, err
	}
	defer db.Close()

	// Simple query to check if database is responsive
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// GetDatabaseVersion returns the current migration version
func GetDatabaseVersion(serviceName string) (int64, error) {
	config := NewDatabaseConfig()
	config.DBName = serviceName

	db, err := config.Connect()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var version int64
	err = db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No migrations applied yet
		}
		return 0, err
	}

	return version, nil
}

// ExecuteInTransaction executes a function within a database transaction
func ExecuteInTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}