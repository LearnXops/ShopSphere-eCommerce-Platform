package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// DatabaseConnection wraps sql.DB with additional functionality
type DatabaseConnection struct {
	DB     *sql.DB
	Config DatabaseConfig
}

// NewPostgresConnection creates a new PostgreSQL connection with connection pooling
func NewPostgresConnection(config DatabaseConfig) (*DatabaseConnection, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, NewInternalError("failed to open database connection", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, NewInternalError("failed to ping database", err)
	}

	return &DatabaseConnection{
		DB:     db,
		Config: config,
	}, nil
}

// Close closes the database connection
func (dc *DatabaseConnection) Close() error {
	if dc.DB != nil {
		return dc.DB.Close()
	}
	return nil
}

// Health checks the database connection health
func (dc *DatabaseConnection) Health(ctx context.Context) error {
	if dc.DB == nil {
		return NewInternalError("database connection is nil", nil)
	}

	if err := dc.DB.PingContext(ctx); err != nil {
		return NewInternalError("database ping failed", err)
	}

	return nil
}

// Stats returns database connection statistics
func (dc *DatabaseConnection) Stats() sql.DBStats {
	if dc.DB == nil {
		return sql.DBStats{}
	}
	return dc.DB.Stats()
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Database:     0,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}