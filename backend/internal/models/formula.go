package models

import (
	"time"

	"github.com/google/uuid"
)

// ComponentMode defines how components are organized in a formula.
type ComponentMode string

const (
	ComponentModeSingle ComponentMode = "single"
	ComponentModeDouble ComponentMode = "double"
)

// Status represents the lifecycle state of a formula.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// Formula is the root domain entity representing a complete formula.
type Formula struct {
	ID            uuid.UUID      `json:"id"            db:"id"`
	Name          string         `json:"name"          db:"name"`
	Code          string         `json:"code"          db:"code"`
	ComponentMode ComponentMode  `json:"component_mode" db:"component_mode"`
	Status        Status         `json:"status"        db:"status"`
	Parts         []FormulaPart  `json:"parts"`
	Steps         []FormulaStep  `json:"steps"`
	CreatedAt     time.Time      `json:"created_at"    db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"    db:"updated_at"`
}
