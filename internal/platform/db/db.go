// Package db provides PostgreSQL connection utilities.
package db

import (
	"context"
	"fmt"
	"net"

	"github.com/OWNER/PROJECT_NAME/internal/platform/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection parameters.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConfigFromEnv loads database config from environment variables.
func ConfigFromEnv() Config {
	return Config{
		Host:     env.GetOrDefault(env.PostgresHost, "localhost"),
		Port:     env.GetOrDefault(env.PostgresPort, "5432"),
		User:     env.GetOrDefault(env.PostgresUser, "platform"),
		Password: env.GetOrDefault(env.PostgresPassword, "platform_dev"),
		DBName:   env.GetOrDefault(env.PostgresDB, "platform"),
		SSLMode:  env.GetOrDefault(env.PostgresSSLMode, "disable"),
	}
}

// DSN returns the PostgreSQL connection string.
func (c Config) DSN() string {
	host := net.JoinHostPort(c.Host, c.Port)

	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		c.User, c.Password, host, c.DBName, c.SSLMode,
	)
}

// Connect creates a new connection pool and verifies connectivity.
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// ConnectFromEnv creates a connection pool using environment configuration.
func ConnectFromEnv(ctx context.Context) (*pgxpool.Pool, error) {
	return Connect(ctx, ConfigFromEnv())
}
