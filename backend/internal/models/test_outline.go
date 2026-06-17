package models

import (
	"time"

	"github.com/google/uuid"
)

// TestOutline is the root entity representing a test outline with nested items and indicators.
// Version numbers auto-increment on each publish. Status tracks the lifecycle.
type TestOutline struct {
	ID        uuid.UUID     `json:"id"         db:"id"`
	Name      string        `json:"name"       db:"name"`
	Version   int           `json:"version"    db:"version"`
	Status    string        `json:"status"     db:"status"`
	Items     []TestItem    `json:"items"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at"`
}

const (
	OutlineStatusDraft    = "draft"
	OutlineStatusActive   = "active"
	OutlineStatusArchived = "archived"
)

// TestItem represents a grouped set of test indicators under a test outline.
type TestItem struct {
	ID         uuid.UUID       `json:"id"          db:"id"`
	OutlineID  uuid.UUID       `json:"outline_id"  db:"outline_id"`
	Name       string          `json:"name"        db:"name"`
	SortOrder  int             `json:"sort_order"  db:"sort_order"`
	Indicators []TestIndicator `json:"indicators"`
	CreatedAt  time.Time       `json:"created_at"  db:"created_at"`
}

// TestIndicator represents a single testable metric within a test item.
type TestIndicator struct {
	ID               uuid.UUID `json:"id"                 db:"id"`
	ItemID           uuid.UUID `json:"item_id"            db:"item_id"`
	Name             string    `json:"name"               db:"name"`
	Unit             string    `json:"unit,omitempty"              db:"unit"`
	MinValue         *float64  `json:"min_value,omitempty"         db:"min_value"`
	MaxValue         *float64  `json:"max_value,omitempty"         db:"max_value"`
	SamplePrepMethod string    `json:"sample_prep_method,omitempty" db:"sample_prep_method"`
	TestMethod       string    `json:"test_method,omitempty"        db:"test_method"`
	TestCondition    string    `json:"test_condition,omitempty"     db:"test_condition"`
	Version          string    `json:"version,omitempty"            db:"version"`
	SortOrder        int       `json:"sort_order"          db:"sort_order"`
	CreatedAt        time.Time `json:"created_at"          db:"created_at"`
}
