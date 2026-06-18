package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// Config holds all application configuration read from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	ServerPort string
}

// LoadConfig reads configuration from environment variables.
// Returns an error if any required variable is missing or invalid.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBPort:     GetEnv("DB_PORT", "5432"),
		DBUser:     GetEnv("DB_USER", "formula"),
		DBPassword: GetEnv("DB_PASSWORD", ""),
		DBName:     GetEnv("DB_NAME", "formula_ai"),
		DBSSLMode:  GetEnv("DB_SSLMODE", "require"),
		ServerPort: GetEnv("SERVER_PORT", "8080"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN returns a PostgreSQL connection string in URL format.
func (c *Config) DSN() string {
	password := url.QueryEscape(c.DBPassword)
	sslMode := c.DBSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, password, c.DBHost, c.DBPort, c.DBName, sslMode,
	)
}

// validate checks that all required configuration fields are valid.
func (c *Config) validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if err := validatePort(c.DBPort); err != nil {
		return fmt.Errorf("DB_PORT: %w", err)
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
	if err := validatePort(c.ServerPort); err != nil {
		return fmt.Errorf("SERVER_PORT: %w", err)
	}
	return nil
}

func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port %q: not a number", port)
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("invalid port %q: out of range (1-65535)", port)
	}
	return nil
}

// GetEnv returns the value of an environment variable, or a fallback if not set.
func GetEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
