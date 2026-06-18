package models

import "github.com/google/uuid"

// FormulaIngredientCategory groups ingredients under a part.
type FormulaIngredientCategory struct {
	ID          uuid.UUID           `json:"id"          db:"id"`
	PartID      uuid.UUID           `json:"part_id"     db:"part_id"`
	Name        string              `json:"name"        db:"name"`
	SortOrder   int                 `json:"sort_order"  db:"sort_order"`
	Ingredients []FormulaIngredient `json:"ingredients"`
}

// FormulaIngredient represents a material in a part's category.
type FormulaIngredient struct {
	ID         uuid.UUID `json:"id"          db:"id"`
	CategoryID uuid.UUID `json:"category_id" db:"category_id"`
	Material   string    `json:"material"    db:"material"`
	Percentage float64   `json:"percentage"  db:"percentage"`
	Weight     float64   `json:"weight"      db:"weight"`
	BatchNo    string    `json:"batch_no"    db:"batch_no"`
	Unit       string    `json:"unit"        db:"unit"`
	SortOrder  int       `json:"sort_order"  db:"sort_order"`
}
