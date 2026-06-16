package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"formula-ai-system/backend/internal/config"
)

// getTestDB connects to the configured PostgreSQL instance and returns a *sql.DB.
// Tables are truncated before each test for isolation.
func getTestDB(t *testing.T) *sql.DB {
	t.Helper()

	cfg := &config.Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "formula"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "changeme"),
		DBName:     getEnvOrDefault("DB_NAME", "formula_ai"),
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("db.Ping: %v (is PostgreSQL running?)", err)
	}

	return db
}

// truncateAll removes all data from formula tables in correct dependency order.
func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"formula_dosing_actions",
		"formula_ingredients",
		"formula_steps",
		"formula_parts",
		"formulas",
	}

	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			t.Logf("truncate %s: %v (continuing)", table, err)
		}
	}
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
