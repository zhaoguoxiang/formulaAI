package models

import "github.com/google/uuid"

// FormulaStepParameter represents a named parameter (e.g. temperature, pressure, time) for a process step.
type FormulaStepParameter struct {
	ID        uuid.UUID `json:"id"         db:"id"`
	StepID    uuid.UUID `json:"step_id"    db:"step_id"`
	Name      string    `json:"name"       db:"name"`
	Value     string    `json:"value"      db:"value"`
	Unit      string    `json:"unit"       db:"unit"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
}
