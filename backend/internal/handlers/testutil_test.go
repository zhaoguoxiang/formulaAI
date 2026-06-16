package handlers

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"formula-ai-system/backend/internal/config"
)

func openTestDB(t *testing.T) *sql.DB {
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
	truncateAllTables(t, db)
	return db
}

func truncateAllTables(t *testing.T, db *sql.DB) {
	t.Helper()
	tables := []string{
		"formula_dosing_actions",
		"formula_ingredients",
		"formula_steps",
		"formula_parts",
		"formulas",
	}
	for _, tbl := range tables {
		_, _ = db.Exec("DELETE FROM " + tbl)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
