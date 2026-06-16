package models

import "github.com/google/uuid"

// FormulaIngredient represents a material and its percentage in a formula part.
type FormulaIngredient struct {
	ID            uuid.UUID             `json:"id"             db:"id"`
	PartID        uuid.UUID             `json:"part_id"        db:"part_id"`
	SortOrder     int                   `json:"sort_order"     db:"sort_order"`
	Material      string                `json:"material"       db:"material"`
	Percentage    float64               `json:"percentage"     db:"percentage"`
	Weight        float64               `json:"weight"         db:"weight"`
	DosingActions []FormulaDosingAction `json:"dosing_actions"`
}
