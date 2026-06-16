package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
)

// DosingRepo provides CRUD operations for formula dosing actions.
type DosingRepo struct{}

// NewDosingRepo creates a new DosingRepo instance.
func NewDosingRepo() *DosingRepo {
	return &DosingRepo{}
}

// Create inserts a new dosing action.
func (r *DosingRepo) Create(ctx context.Context, db *sql.DB, da *models.FormulaDosingAction) error {
	if da.ID == uuid.Nil {
		da.ID = uuid.New()
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO formula_dosing_actions (id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		da.ID, da.StepID, da.IngredientID, da.DosingOrder, da.UseRatio, da.DosingMethod,
	)
	if err != nil {
		return fmt.Errorf("insert dosing action: %w", err)
	}

	return nil
}

// GetByID retrieves a single dosing action by its ID.
func (r *DosingRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.FormulaDosingAction, error) {
	var da models.FormulaDosingAction
	var dosingMethod sql.NullString

	err := db.QueryRowContext(ctx,
		`SELECT id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method
		 FROM formula_dosing_actions WHERE id = $1`, id,
	).Scan(&da.ID, &da.StepID, &da.IngredientID, &da.DosingOrder, &da.UseRatio, &dosingMethod)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dosing action %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query dosing action: %w", err)
	}
	if dosingMethod.Valid {
		da.DosingMethod = dosingMethod.String
	}

	return &da, nil
}

// ListByStepID retrieves all dosing actions for a given step, ordered by dosing_order.
func (r *DosingRepo) ListByStepID(ctx context.Context, db *sql.DB, stepID uuid.UUID) ([]models.FormulaDosingAction, error) {
	return r.queryDosingActions(ctx, db,
		`SELECT id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method
		 FROM formula_dosing_actions WHERE step_id = $1 ORDER BY dosing_order`, stepID)
}

// ListByIngredientID retrieves all dosing actions for a given ingredient, ordered by dosing_order.
func (r *DosingRepo) ListByIngredientID(ctx context.Context, db *sql.DB, ingredientID uuid.UUID) ([]models.FormulaDosingAction, error) {
	return r.queryDosingActions(ctx, db,
		`SELECT id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method
		 FROM formula_dosing_actions WHERE ingredient_id = $1 ORDER BY dosing_order`, ingredientID)
}

// queryDosingActions is a shared helper for loading dosing action rows.
func (r *DosingRepo) queryDosingActions(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, query string, arg interface{}) ([]models.FormulaDosingAction, error) {
	rows, err := db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("query dosing actions: %w", err)
	}
	defer rows.Close()

	var actions []models.FormulaDosingAction
	for rows.Next() {
		var da models.FormulaDosingAction
		var dosingMethod sql.NullString
		if err := rows.Scan(&da.ID, &da.StepID, &da.IngredientID, &da.DosingOrder, &da.UseRatio, &dosingMethod); err != nil {
			return nil, fmt.Errorf("scan dosing action: %w", err)
		}
		if dosingMethod.Valid {
			da.DosingMethod = dosingMethod.String
		}
		actions = append(actions, da)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if actions == nil {
		actions = []models.FormulaDosingAction{}
	}

	return actions, nil
}

// Update modifies an existing dosing action.
func (r *DosingRepo) Update(ctx context.Context, db *sql.DB, da *models.FormulaDosingAction) error {
	result, err := db.ExecContext(ctx,
		`UPDATE formula_dosing_actions SET step_id=$1, ingredient_id=$2, dosing_order=$3, use_ratio=$4, dosing_method=$5
		 WHERE id=$6`,
		da.StepID, da.IngredientID, da.DosingOrder, da.UseRatio, da.DosingMethod, da.ID,
	)
	if err != nil {
		return fmt.Errorf("update dosing action: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("dosing action %s not found", da.ID)
	}

	return nil
}

// Delete removes a dosing action.
func (r *DosingRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM formula_dosing_actions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete dosing action: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("dosing action %s not found", id)
	}

	return nil
}
