package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisConfig creates a new Redis configuration with defaults
func NewRedisConfig() *RedisConfig {
	port := 6379
	db := 0

	if portStr := os.Getenv("REDIS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if d, err := strconv.Atoi(dbStr); err == nil {
			db = d
		}
	}

	return &RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     port,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
}

// Connect establishes a connection to Redis
func (c *RedisConfig) Connect() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password:     c.Password,
		DB:           c.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// CheckRedisConnection verifies Redis connectivity
func CheckRedisConnection() error {
	config := NewRedisConfig()
	client, err := config.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	return nil
}

// RedisHealthCheck performs a health check on Redis
func RedisHealthCheck() (bool, error) {
	config := NewRedisConfig()
	client, err := config.Connect()
	if err != nil {
		return false, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple ping to check if Redis is responsive
	err = client.Ping(ctx).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
