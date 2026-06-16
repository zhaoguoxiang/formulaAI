package models

import "github.com/google/uuid"

// PartName identifies a standard part role in a formula.
type PartName string

const (
	PartA    PartName = "PartA"
	PartB    PartName = "PartB"
	PartMain PartName = "PartMain"
)

// FormulaPart represents a component part (A, B, or Main) of a formula.
type FormulaPart struct {
	ID          uuid.UUID           `json:"id"          db:"id"`
	FormulaID   uuid.UUID           `json:"formula_id"  db:"formula_id"`
	Name        PartName            `json:"name"        db:"name"`
	MixRatio    float64             `json:"mix_ratio"   db:"mix_ratio"`
	SortOrder   int                 `json:"sort_order"  db:"sort_order"`
	Ingredients []FormulaIngredient `json:"ingredients"`
}
