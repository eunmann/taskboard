// Package app provides the application dependency container and configuration.
package app

import (
	"github.com/OWNER/PROJECT_NAME/internal/platform/env"
)

// Config holds all application configuration.
type Config struct {
	Backend  BackendConfig
	Database DatabaseConfig
}

// BackendConfig holds HTTP server configuration.
type BackendConfig struct {
	Port             string
	SiteURL          string
	LocalAuthEnabled bool
	CookieSecure     bool
}

// DatabaseConfig holds database connection parameters.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// ConfigFromEnv loads configuration from environment variables.
func ConfigFromEnv() Config {
	return Config{
		Backend: BackendConfig{
			Port:             env.GetOrDefault(env.Port, "8080"),
			SiteURL:          env.GetOrDefault(env.SiteURL, "http://localhost:8080"),
			LocalAuthEnabled: env.GetOrDefault(env.LocalAuthEnabled, "true") == "true",
			CookieSecure:     env.GetOrDefault(env.CookieSecure, "false") == "true",
		},
		Database: DatabaseConfig{
			Host:     env.GetOrDefault(env.PostgresHost, "localhost"),
			Port:     env.GetOrDefault(env.PostgresPort, "5432"),
			User:     env.GetOrDefault(env.PostgresUser, "platform"),
			Password: env.GetOrDefault(env.PostgresPassword, "platform_dev"),
			Database: env.GetOrDefault(env.PostgresDB, "platform"),
			SSLMode:  env.GetOrDefault(env.PostgresSSLMode, "disable"),
		},
	}
}
