package models

import "github.com/google/uuid"

// FormulaStepMaterialCategory groups materials under a named category within a "投料" step.
type FormulaStepMaterialCategory struct {
	ID        uuid.UUID              `json:"id"         db:"id"`
	StepID    uuid.UUID              `json:"step_id"    db:"step_id"`
	Name      string                 `json:"name"       db:"name"`
	SortOrder int                    `json:"sort_order" db:"sort_order"`
	Materials []FormulaStepMaterial  `json:"materials"`
}

// FormulaStepMaterial represents a single raw material in a step's category.
type FormulaStepMaterial struct {
	ID         uuid.UUID `json:"id"          db:"id"`
	CategoryID uuid.UUID `json:"category_id" db:"category_id"`
	Material   string    `json:"material"    db:"material"`
	Percentage float64   `json:"percentage"  db:"percentage"`
	Weight     float64   `json:"weight"      db:"weight"`
	BatchNo    string    `json:"batch_no"    db:"batch_no"`
	Unit       string    `json:"unit"        db:"unit"`
	SortOrder  int       `json:"sort_order"  db:"sort_order"`
}
