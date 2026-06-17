package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all relevant env vars to test defaults — but DB_PASSWORD is required
	envVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "SERVER_PORT"}
	for _, k := range envVars {
		os.Unsetenv(k)
	}
	// DB_PASSWORD is required, so LoadConfig must fail without it
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() should fail when DB_PASSWORD is unset")
	}
}

func TestLoadConfig_Defaults_WithPassword(t *testing.T) {
	envVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "SERVER_PORT"}
	for _, k := range envVars {
		os.Unsetenv(k)
	}
	os.Setenv("DB_PASSWORD", "testpass")
	defer os.Unsetenv("DB_PASSWORD")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost=localhost, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("expected DBPort=5432, got %s", cfg.DBPort)
	}
	if cfg.DBUser != "formula" {
		t.Errorf("expected DBUser=formula, got %s", cfg.DBUser)
	}
	if cfg.DBPassword != "testpass" {
		t.Errorf("expected DBPassword=testpass, got %s", cfg.DBPassword)
	}
	if cfg.DBName != "formula_ai" {
		t.Errorf("expected DBName=formula_ai, got %s", cfg.DBName)
	}
	if cfg.ServerPort != "8080" {
		t.Errorf("expected ServerPort=8080, got %s", cfg.ServerPort)
	}
	if cfg.DBSSLMode != "disable" {
		t.Errorf("expected DBSSLMode=disable, got %s", cfg.DBSSLMode)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_SSLMODE", "require")
	defer func() {
		for _, k := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "SERVER_PORT", "DB_SSLMODE"} {
			os.Unsetenv(k)
		}
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if cfg.DBHost != "db.example.com" {
		t.Errorf("expected DBHost=db.example.com, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "5433" {
		t.Errorf("expected DBPort=5433, got %s", cfg.DBPort)
	}
	if cfg.DBUser != "admin" {
		t.Errorf("expected DBUser=admin, got %s", cfg.DBUser)
	}
	if cfg.DBPassword != "secret" {
		t.Errorf("expected DBPassword=secret, got %s", cfg.DBPassword)
	}
	if cfg.DBName != "test_db" {
		t.Errorf("expected DBName=test_db, got %s", cfg.DBName)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort=9090, got %s", cfg.ServerPort)
	}
	if cfg.DBSSLMode != "require" {
		t.Errorf("expected DBSSLMode=require, got %s", cfg.DBSSLMode)
	}
}

func TestValidate_MissingFields(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "missing DB_HOST",
			config: Config{DBPort: "5432", DBUser: "u", DBPassword: "p", DBName: "d", ServerPort: "8080"},
		},
		{
			name:   "missing DB_USER",
			config: Config{DBHost: "h", DBPort: "5432", DBPassword: "p", DBName: "d", ServerPort: "8080"},
		},
		{
			name:   "missing DB_PASSWORD",
			config: Config{DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "d", ServerPort: "8080"},
		},
		{
			name:   "missing DB_NAME",
			config: Config{DBHost: "h", DBPort: "5432", DBUser: "u", DBPassword: "p", ServerPort: "8080"},
		},
		{
			name:   "missing DB_PORT",
			config: Config{DBHost: "h", DBUser: "u", DBPassword: "p", DBName: "d", ServerPort: "8080"},
		},
		{
			name:   "missing SERVER_PORT",
			config: Config{DBHost: "h", DBPort: "5432", DBUser: "u", DBPassword: "p", DBName: "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "formula",
		DBPassword: "changeme",
		DBName:     "formula_ai",
		DBSSLMode:  "disable",
		ServerPort: "8080",
	}

	dsn := cfg.DSN()
	expected := "postgres://formula:changeme@localhost:5432/formula_ai?sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN=%s, got %s", expected, dsn)
	}
}

func TestDSN_PasswordWithSpecialChars(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "formula",
		DBPassword: "p@ss:word!",
		DBName:     "formula_ai",
		DBSSLMode:  "disable",
	}

	dsn := cfg.DSN()
	// Password should be URL-encoded
	expected := "postgres://formula:p%40ss%3Aword%21@localhost:5432/formula_ai?sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN=%s, got %s", expected, dsn)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{"empty", ""},
		{"not a number", "abc"},
		{"out of range low", "0"},
		{"out of range high", "65536"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if err == nil {
				t.Errorf("expected error for port %q", tt.port)
			}
		})
	}
}
