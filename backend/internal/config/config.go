package config

import (
	"fmt"
	"net/url"
	"os"
)

// Config holds all application configuration read from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
// Returns an error if any required variable is missing.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "formula"),
		DBPassword: getEnv("DB_PASSWORD", "changeme"),
		DBName:     getEnv("DB_NAME", "formula_ai"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN returns a PostgreSQL connection string in URL format:
//
//	postgres://user:password@host:port/dbname?sslmode=disable
func (c *Config) DSN() string {
	password := url.QueryEscape(c.DBPassword)
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, password, c.DBHost, c.DBPort, c.DBName,
	)
}

// validate checks that all required configuration fields are non-empty.
func (c *Config) validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.ServerPort == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	return nil
}

// getEnv returns the value of an environment variable, or a fallback if not set.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
