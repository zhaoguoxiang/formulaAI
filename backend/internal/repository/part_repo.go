package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
)

// PartRepo provides CRUD operations for formula parts.
type PartRepo struct{}

// NewPartRepo creates a new PartRepo instance.
func NewPartRepo() *PartRepo {
	return &PartRepo{}
}

// Create inserts a new formula part.
func (r *PartRepo) Create(ctx context.Context, db *sql.DB, part *models.FormulaPart) error {
	if part.ID == uuid.Nil {
		part.ID = uuid.New()
	}

	dbPartName := partNameToDB(part.Name)
	_, err := db.ExecContext(ctx,
		`INSERT INTO formula_parts (id, formula_id, name, mix_ratio, sort_order)
		 VALUES ($1, $2, $3, $4, $5)`,
		part.ID, part.FormulaID, dbPartName, part.MixRatio, part.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("insert part: %w", err)
	}

	return nil
}

// GetByID retrieves a single part by its ID, including nested ingredients.
func (r *PartRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.FormulaPart, error) {
	var p models.FormulaPart
	var dbName string

	err := db.QueryRowContext(ctx,
		`SELECT id, formula_id, name, mix_ratio, sort_order
		 FROM formula_parts WHERE id = $1`, id,
	).Scan(&p.ID, &p.FormulaID, &dbName, &p.MixRatio, &p.SortOrder)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("part %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query part: %w", err)
	}
	p.Name = partNameFromDB(dbName)

	// Load nested ingredients
	formulaRepo := &FormulaRepo{}
	ingredients, err := formulaRepo.queryIngredients(ctx, db, p.ID)
	if err != nil {
		return nil, fmt.Errorf("query ingredients: %w", err)
	}
	p.Ingredients = ingredients

	return &p, nil
}

// ListByFormulaID retrieves all parts for a given formula.
func (r *PartRepo) ListByFormulaID(ctx context.Context, db *sql.DB, formulaID uuid.UUID) ([]models.FormulaPart, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, name, mix_ratio, sort_order
		 FROM formula_parts WHERE formula_id = $1 ORDER BY sort_order`, formulaID,
	)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	defer rows.Close()

	var parts []models.FormulaPart
	for rows.Next() {
		var p models.FormulaPart
		var dbName string
		if err := rows.Scan(&p.ID, &p.FormulaID, &dbName, &p.MixRatio, &p.SortOrder); err != nil {
			return nil, fmt.Errorf("scan part: %w", err)
		}
		p.Name = partNameFromDB(dbName)
		parts = append(parts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if parts == nil {
		parts = []models.FormulaPart{}
	}

	return parts, nil
}

// Update modifies an existing part.
func (r *PartRepo) Update(ctx context.Context, db *sql.DB, part *models.FormulaPart) error {
	dbPartName := partNameToDB(part.Name)
	result, err := db.ExecContext(ctx,
		`UPDATE formula_parts SET name=$1, mix_ratio=$2, sort_order=$3
		 WHERE id=$4`,
		dbPartName, part.MixRatio, part.SortOrder, part.ID,
	)
	if err != nil {
		return fmt.Errorf("update part: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("part %s not found", part.ID)
	}

	return nil
}

// Delete removes a part. CASCADE handles nested ingredients and dosing actions.
func (r *PartRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM formula_parts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete part: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("part %s not found", id)
	}

	return nil
}
