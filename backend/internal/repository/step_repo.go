package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
)

// StepRepo provides CRUD operations for formula steps.
type StepRepo struct{}

// NewStepRepo creates a new StepRepo instance.
func NewStepRepo() *StepRepo {
	return &StepRepo{}
}

// Create inserts a new formula step.
func (r *StepRepo) Create(ctx context.Context, db *sql.DB, step *models.FormulaStep) error {
	if step.ID == uuid.Nil {
		step.ID = uuid.New()
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO formula_steps (id, formula_id, part_id, step_no, name, temperature, duration)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		step.ID, step.FormulaID, step.PartID, step.StepNo, step.Name, step.Temperature, step.Duration,
	)
	if err != nil {
		return fmt.Errorf("insert step: %w", err)
	}

	return nil
}

// GetByID retrieves a single step by its ID.
func (r *StepRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.FormulaStep, error) {
	var s models.FormulaStep
	var partIDStr sql.NullString
	var temperature, duration sql.NullString

	err := db.QueryRowContext(ctx,
		`SELECT id, formula_id, part_id::text, step_no, name, temperature, duration
		 FROM formula_steps WHERE id = $1`, id,
	).Scan(&s.ID, &s.FormulaID, &partIDStr, &s.StepNo, &s.Name, &temperature, &duration)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("step %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query step: %w", err)
	}

	if partIDStr.Valid {
		if uid, err := uuid.Parse(partIDStr.String); err == nil {
			s.PartID = &uid
		}
	}
	if temperature.Valid {
		s.Temperature = temperature.String
	}
	if duration.Valid {
		s.Duration = duration.String
	}

	return &s, nil
}

// ListByFormulaID retrieves all steps for a given formula.
func (r *StepRepo) ListByFormulaID(ctx context.Context, db *sql.DB, formulaID uuid.UUID) ([]models.FormulaStep, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, part_id::text, step_no, name, temperature, duration
		 FROM formula_steps WHERE formula_id = $1 ORDER BY step_no`, formulaID,
	)
	if err != nil {
		return nil, fmt.Errorf("query steps: %w", err)
	}
	defer rows.Close()

	var steps []models.FormulaStep
	for rows.Next() {
		var s models.FormulaStep
		var partIDStr sql.NullString
		var temperature, duration sql.NullString
		if err := rows.Scan(&s.ID, &s.FormulaID, &partIDStr, &s.StepNo, &s.Name, &temperature, &duration); err != nil {
			return nil, fmt.Errorf("scan step: %w", err)
		}
		if partIDStr.Valid {
			if uid, err := uuid.Parse(partIDStr.String); err == nil {
				s.PartID = &uid
			}
		}
		if temperature.Valid {
			s.Temperature = temperature.String
		}
		if duration.Valid {
			s.Duration = duration.String
		}
		steps = append(steps, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if steps == nil {
		steps = []models.FormulaStep{}
	}

	return steps, nil
}

// Update modifies an existing step.
func (r *StepRepo) Update(ctx context.Context, db *sql.DB, step *models.FormulaStep) error {
	result, err := db.ExecContext(ctx,
		`UPDATE formula_steps SET part_id=$1, step_no=$2, name=$3, temperature=$4, duration=$5
		 WHERE id=$6`,
		step.PartID, step.StepNo, step.Name, step.Temperature, step.Duration, step.ID,
	)
	if err != nil {
		return fmt.Errorf("update step: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("step %s not found", step.ID)
	}

	return nil
}

// Delete removes a step. CASCADE handles nested dosing actions.
func (r *StepRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM formula_steps WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete step: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("step %s not found", id)
	}

	return nil
}
