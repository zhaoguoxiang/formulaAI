package handlers

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	"formula-ai-system/backend/internal/config"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	cfg := &config.Config{
		DBHost:     config.GetEnv("DB_HOST", "localhost"),
		DBPort:     config.GetEnv("DB_PORT", "5432"),
		DBUser:     config.GetEnv("DB_USER", "formula"),
		DBPassword: config.GetEnv("DB_PASSWORD", "changeme"),
		DBName:     config.GetEnv("DB_NAME", "formula_ai"),
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
		"formula_step_materials",
		"formula_step_material_categories",
		"formula_step_parameters",
		"formula_steps",
		"formula_parts",
		"test_indicators",
		"test_items",
		"test_outlines",
		"formulas",
	}
	for _, tbl := range tables {
		_, _ = db.Exec("DELETE FROM " + tbl)
	}
}
