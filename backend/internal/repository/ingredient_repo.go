package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
)

// IngredientRepo provides CRUD operations for formula ingredients.
type IngredientRepo struct{}

// NewIngredientRepo creates a new IngredientRepo instance.
func NewIngredientRepo() *IngredientRepo {
	return &IngredientRepo{}
}

// Create inserts a new formula ingredient.
func (r *IngredientRepo) Create(ctx context.Context, db *sql.DB, ing *models.FormulaIngredient) error {
	if ing.ID == uuid.Nil {
		ing.ID = uuid.New()
	}

	var weight interface{}
	if ing.Weight != 0 {
		weight = ing.Weight
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO formula_ingredients (id, part_id, sort_order, material, percentage, weight)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		ing.ID, ing.PartID, ing.SortOrder, ing.Material, ing.Percentage, weight,
	)
	if err != nil {
		return fmt.Errorf("insert ingredient: %w", err)
	}

	return nil
}

// GetByID retrieves a single ingredient by its ID, including nested dosing actions.
func (r *IngredientRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.FormulaIngredient, error) {
	var ing models.FormulaIngredient
	var weight sql.NullFloat64

	err := db.QueryRowContext(ctx,
		`SELECT id, part_id, sort_order, material, percentage, weight
		 FROM formula_ingredients WHERE id = $1`, id,
	).Scan(&ing.ID, &ing.PartID, &ing.SortOrder, &ing.Material, &ing.Percentage, &weight)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ingredient %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query ingredient: %w", err)
	}
	if weight.Valid {
		ing.Weight = weight.Float64
	}

	// Load nested dosing actions
	formulaRepo := &FormulaRepo{}
	dosingActions, err := formulaRepo.queryDosingActions(ctx, db, ing.ID)
	if err != nil {
		return nil, fmt.Errorf("query dosing actions: %w", err)
	}
	ing.DosingActions = dosingActions

	return &ing, nil
}

// ListByPartID retrieves all ingredients for a given part.
func (r *IngredientRepo) ListByPartID(ctx context.Context, db *sql.DB, partID uuid.UUID) ([]models.FormulaIngredient, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, part_id, sort_order, material, percentage, weight
		 FROM formula_ingredients WHERE part_id = $1 ORDER BY sort_order`, partID,
	)
	if err != nil {
		return nil, fmt.Errorf("query ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []models.FormulaIngredient
	for rows.Next() {
		var ing models.FormulaIngredient
		var weight sql.NullFloat64
		if err := rows.Scan(&ing.ID, &ing.PartID, &ing.SortOrder, &ing.Material, &ing.Percentage, &weight); err != nil {
			return nil, fmt.Errorf("scan ingredient: %w", err)
		}
		if weight.Valid {
			ing.Weight = weight.Float64
		}
		ingredients = append(ingredients, ing)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if ingredients == nil {
		ingredients = []models.FormulaIngredient{}
	}

	return ingredients, nil
}

// Update modifies an existing ingredient.
func (r *IngredientRepo) Update(ctx context.Context, db *sql.DB, ing *models.FormulaIngredient) error {
	var weight interface{}
	if ing.Weight != 0 {
		weight = ing.Weight
	}

	result, err := db.ExecContext(ctx,
		`UPDATE formula_ingredients SET sort_order=$1, material=$2, percentage=$3, weight=$4
		 WHERE id=$5`,
		ing.SortOrder, ing.Material, ing.Percentage, weight, ing.ID,
	)
	if err != nil {
		return fmt.Errorf("update ingredient: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ingredient %s not found", ing.ID)
	}

	return nil
}

// Delete removes an ingredient. CASCADE handles nested dosing actions.
func (r *IngredientRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM formula_ingredients WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete ingredient: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ingredient %s not found", id)
	}

	return nil
}
