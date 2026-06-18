package models

import "github.com/google/uuid"

// FormulaStep represents a process step in a formula.
// PartID is nullable (nil means step applies to the whole formula).
type FormulaStep struct {
	ID             uuid.UUID                        `json:"id"              db:"id"`
	FormulaID      uuid.UUID                        `json:"formula_id"      db:"formula_id"`
	PartID         *uuid.UUID                       `json:"part_id"         db:"part_id"`
	StepNo         int                              `json:"step_no"         db:"step_no"`
	Name           string                           `json:"name"            db:"name"`
	Description    string                           `json:"description"     db:"description"`
	InstrumentName string                           `json:"instrument_name" db:"instrument_name"`
	Temperature    string                           `json:"temperature"     db:"temperature"`
	Duration       string                           `json:"duration"        db:"duration"`
	Categories     []FormulaStepMaterialCategory    `json:"categories"`
	Parameters     []FormulaStepParameter           `json:"parameters"`
}
