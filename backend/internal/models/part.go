package models

import "github.com/google/uuid"

type PartName string

const (
	PartA    PartName = "PartA"
	PartB    PartName = "PartB"
	PartMain PartName = "PartMain"
)

type FormulaPart struct {
	ID         uuid.UUID                   `json:"id"          db:"id"`
	FormulaID  uuid.UUID                   `json:"formula_id"  db:"formula_id"`
	Name       PartName                    `json:"name"        db:"name"`
	SortOrder  int                         `json:"sort_order"  db:"sort_order"`
	Categories []FormulaIngredientCategory `json:"categories"`
}
