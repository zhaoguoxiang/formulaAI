package models

import "github.com/google/uuid"

// FormulaDosingAction represents a dosing action that ties a step to an ingredient.
type FormulaDosingAction struct {
	ID           uuid.UUID `json:"id"            db:"id"`
	StepID       uuid.UUID `json:"step_id"       db:"step_id"`
	IngredientID uuid.UUID `json:"ingredient_id" db:"ingredient_id"`
	DosingOrder  int       `json:"dosing_order"  db:"dosing_order"`
	UseRatio     float64   `json:"use_ratio"     db:"use_ratio"`
	DosingMethod string    `json:"dosing_method" db:"dosing_method"`
}
