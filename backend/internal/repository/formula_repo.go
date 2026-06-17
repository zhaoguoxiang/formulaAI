package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"formula-ai-system/backend/internal/models"
)

// ListOptions controls filtering and pagination for List queries.
type ListOptions struct {
	ComponentMode string // optional filter: "single", "double", "" for all
	Limit         int    // max results, 0 means no limit
	Offset        int    // pagination offset
}

// FormulaRepo provides CRUD operations for formulas with nested data.
// All nested operations (parts, ingredients, steps, dosing actions) are
// executed within PostgreSQL transactions for atomicity.
type FormulaRepo struct{}

// NewFormulaRepo creates a new FormulaRepo instance.
func NewFormulaRepo() *FormulaRepo {
	return &FormulaRepo{}
}

// Create inserts a formula and all its nested data in a single transaction.
// UUIDs are generated for any entity that doesn't already have one.
func (r *FormulaRepo) Create(ctx context.Context, db *sql.DB, f *models.Formula) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Ensure formula has an ID
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	now := time.Now()
	f.CreatedAt = now
	f.UpdatedAt = now

	// 1. Insert formula
	_, err = tx.ExecContext(ctx,
		`INSERT INTO formulas (id, name, code, component_mode, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		f.ID, f.Name, f.Code, string(f.ComponentMode), string(f.Status), f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert formula: %w", err)
	}

	// 2. Insert steps (need step IDs for dosing actions)
	stepIDs := make([]uuid.UUID, len(f.Steps))
	for i := range f.Steps {
		step := &f.Steps[i]
		if step.ID == uuid.Nil {
			step.ID = uuid.New()
		}
		step.FormulaID = f.ID

		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_steps (id, formula_id, part_id, step_no, name, temperature, duration)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			step.ID, step.FormulaID, step.PartID, step.StepNo, step.Name, step.Temperature, step.Duration,
		)
		if err != nil {
			return fmt.Errorf("insert step %d: %w", i, err)
		}
		stepIDs[i] = step.ID
	}

	// 3. Insert parts and ingredients (with dosing actions)
	for pi := range f.Parts {
		part := &f.Parts[pi]
		if part.ID == uuid.Nil {
			part.ID = uuid.New()
		}
		part.FormulaID = f.ID

		dbPartName := partNameToDB(part.Name)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_parts (id, formula_id, name, mix_ratio, sort_order)
			 VALUES ($1, $2, $3, $4, $5)`,
			part.ID, part.FormulaID, dbPartName, part.MixRatio, part.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert part %d: %w", pi, err)
		}

		for ii := range part.Ingredients {
			ing := &part.Ingredients[ii]
			if ing.ID == uuid.Nil {
				ing.ID = uuid.New()
			}
			ing.PartID = part.ID

			var weight interface{}
			if ing.Weight != 0 {
				weight = ing.Weight
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO formula_ingredients (id, part_id, sort_order, material, percentage, weight)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				ing.ID, ing.PartID, ing.SortOrder, ing.Material, ing.Percentage, weight,
			)
			if err != nil {
				return fmt.Errorf("insert ingredient %d in part %d: %w", ii, pi, err)
			}

			// Insert dosing actions for this ingredient
			for di := range ing.DosingActions {
				da := &ing.DosingActions[di]
				if da.ID == uuid.Nil {
					da.ID = uuid.New()
				}
				// ingredient_id is always the parent ingredient
				da.IngredientID = ing.ID

				_, err = tx.ExecContext(ctx,
					`INSERT INTO formula_dosing_actions (id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method)
					 VALUES ($1, $2, $3, $4, $5, $6)`,
					da.ID, da.StepID, da.IngredientID, da.DosingOrder, da.UseRatio, da.DosingMethod,
				)
				if err != nil {
					return fmt.Errorf("insert dosing action %d for ingredient %d: %w", di, ii, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// GetByID retrieves a formula with all nested data (parts, ingredients, steps, dosing actions).
func (r *FormulaRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.Formula, error) {
	// 1. Query formula
	f := &models.Formula{}
	var componentMode, status string
	err := db.QueryRowContext(ctx,
		`SELECT id, name, code, component_mode, status, created_at, updated_at
		 FROM formulas WHERE id = $1`, id,
	).Scan(&f.ID, &f.Name, &f.Code, &componentMode, &status, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("formula %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query formula: %w", err)
	}
	f.ComponentMode = models.ComponentMode(componentMode)
	f.Status = models.Status(status)

	// 2. Query parts
	parts, err := r.queryParts(ctx, db, f.ID)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	f.Parts = parts

	// 3. Query steps
	steps, err := r.querySteps(ctx, db, f.ID)
	if err != nil {
		return nil, fmt.Errorf("query steps: %w", err)
	}
	f.Steps = steps

	return f, nil
}

// List retrieves formulas with their nested data, supporting optional mode filter and pagination.
func (r *FormulaRepo) List(ctx context.Context, db *sql.DB, opts ListOptions) ([]*models.Formula, error) {
	args := []interface{}{}
	where := ""

	if opts.ComponentMode != "" {
		where = " WHERE component_mode = $1"
		args = append(args, opts.ComponentMode)
	}

	query := `SELECT id, name, code, component_mode, status, created_at, updated_at
		FROM formulas` + where + ` ORDER BY created_at DESC`

	if opts.Limit > 0 {
		argIdx := len(args) + 1
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		argIdx := len(args) + 1
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, opts.Offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query formulas: %w", err)
	}
	defer rows.Close()

	var formulas []*models.Formula
	var formulaIDs []uuid.UUID
	for rows.Next() {
		f := &models.Formula{}
		var componentMode, status string
		if err := rows.Scan(&f.ID, &f.Name, &f.Code, &componentMode, &status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan formula: %w", err)
		}
		f.ComponentMode = models.ComponentMode(componentMode)
		f.Status = models.Status(status)
		formulas = append(formulas, f)
		formulaIDs = append(formulaIDs, f.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if len(formulas) == 0 {
		return []*models.Formula{}, nil
	}

	// Batch query all nested data
	partsByFormula, err := r.queryPartsBulk(ctx, db, formulaIDs)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	stepsByFormula, err := r.queryStepsBulk(ctx, db, formulaIDs)
	if err != nil {
		return nil, fmt.Errorf("query steps: %w", err)
	}

	// Collect part IDs for batch ingredient query
	var partIDs []uuid.UUID
	for _, parts := range partsByFormula {
		for _, p := range parts {
			partIDs = append(partIDs, p.ID)
		}
	}

	var ingredientsByPart map[uuid.UUID][]models.FormulaIngredient
	var dosingByIngredient map[uuid.UUID][]models.FormulaDosingAction
	if len(partIDs) > 0 {
		ingredientsByPart, err = r.queryIngredientsBulk(ctx, db, partIDs)
		if err != nil {
			return nil, fmt.Errorf("query ingredients: %w", err)
		}

		var ingredientIDs []uuid.UUID
		for _, ingredients := range ingredientsByPart {
			for _, ing := range ingredients {
				ingredientIDs = append(ingredientIDs, ing.ID)
			}
		}
		if len(ingredientIDs) > 0 {
			dosingByIngredient, err = r.queryDosingActionsBulk(ctx, db, ingredientIDs)
			if err != nil {
				return nil, fmt.Errorf("query dosing actions: %w", err)
			}
		}
	}

	if ingredientsByPart == nil {
		ingredientsByPart = make(map[uuid.UUID][]models.FormulaIngredient)
	}
	if dosingByIngredient == nil {
		dosingByIngredient = make(map[uuid.UUID][]models.FormulaDosingAction)
	}

	// Assemble formulas
	for _, f := range formulas {
		parts := partsByFormula[f.ID]
		if parts == nil {
			parts = []models.FormulaPart{}
		}
		for i := range parts {
			ingredients := ingredientsByPart[parts[i].ID]
			if ingredients == nil {
				ingredients = []models.FormulaIngredient{}
			}
			for j := range ingredients {
				ingredients[j].DosingActions = dosingByIngredient[ingredients[j].ID]
				if ingredients[j].DosingActions == nil {
					ingredients[j].DosingActions = []models.FormulaDosingAction{}
				}
			}
			parts[i].Ingredients = ingredients
		}
		f.Parts = parts

		steps := stepsByFormula[f.ID]
		if steps == nil {
			steps = []models.FormulaStep{}
		}
		f.Steps = steps
	}

	return formulas, nil
}

// Update replaces a formula's nested data atomically.
// Strategy: UPDATE formula row, DELETE all nested rows, re-INSERT new nested data.
func (r *FormulaRepo) Update(ctx context.Context, db *sql.DB, f *models.Formula) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	f.UpdatedAt = time.Now()

	// 1. Update formula row
	result, err := tx.ExecContext(ctx,
		`UPDATE formulas SET name=$1, code=$2, component_mode=$3, status=$4, updated_at=$5
		 WHERE id=$6`,
		f.Name, f.Code, string(f.ComponentMode), string(f.Status), f.UpdatedAt, f.ID,
	)
	if err != nil {
		return fmt.Errorf("update formula: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("formula %s not found", f.ID)
	}

	// 2. Delete all nested rows (steps first, then parts - CASCADE handles deeper nesting)
	_, err = tx.ExecContext(ctx, `DELETE FROM formula_steps WHERE formula_id = $1`, f.ID)
	if err != nil {
		return fmt.Errorf("delete nested steps: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM formula_parts WHERE formula_id = $1`, f.ID)
	if err != nil {
		return fmt.Errorf("delete nested parts: %w", err)
	}

	// 3. Re-insert all nested data (same logic as Create minus formula INSERT)
	stepIDs := make([]uuid.UUID, len(f.Steps))
	for i := range f.Steps {
		step := &f.Steps[i]
		if step.ID == uuid.Nil {
			step.ID = uuid.New()
		}
		step.FormulaID = f.ID

		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_steps (id, formula_id, part_id, step_no, name, temperature, duration)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			step.ID, step.FormulaID, step.PartID, step.StepNo, step.Name, step.Temperature, step.Duration,
		)
		if err != nil {
			return fmt.Errorf("re-insert step %d: %w", i, err)
		}
		stepIDs[i] = step.ID
	}
	_ = stepIDs // available for validation if needed

	for pi := range f.Parts {
		part := &f.Parts[pi]
		if part.ID == uuid.Nil {
			part.ID = uuid.New()
		}
		part.FormulaID = f.ID

		dbPartName := partNameToDB(part.Name)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_parts (id, formula_id, name, mix_ratio, sort_order)
			 VALUES ($1, $2, $3, $4, $5)`,
			part.ID, part.FormulaID, dbPartName, part.MixRatio, part.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("re-insert part %d: %w", pi, err)
		}

		for ii := range part.Ingredients {
			ing := &part.Ingredients[ii]
			if ing.ID == uuid.Nil {
				ing.ID = uuid.New()
			}
			ing.PartID = part.ID

			var weight interface{}
			if ing.Weight != 0 {
				weight = ing.Weight
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO formula_ingredients (id, part_id, sort_order, material, percentage, weight)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				ing.ID, ing.PartID, ing.SortOrder, ing.Material, ing.Percentage, weight,
			)
			if err != nil {
				return fmt.Errorf("re-insert ingredient %d in part %d: %w", ii, pi, err)
			}

			for di := range ing.DosingActions {
				da := &ing.DosingActions[di]
				if da.ID == uuid.Nil {
					da.ID = uuid.New()
				}
				da.IngredientID = ing.ID

				_, err = tx.ExecContext(ctx,
					`INSERT INTO formula_dosing_actions (id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method)
					 VALUES ($1, $2, $3, $4, $5, $6)`,
					da.ID, da.StepID, da.IngredientID, da.DosingOrder, da.UseRatio, da.DosingMethod,
				)
				if err != nil {
					return fmt.Errorf("re-insert dosing action %d for ingredient %d: %w", di, ii, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// Delete removes a formula. CASCADE constraints handle deletion of all nested data.
func (r *FormulaRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM formulas WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete formula: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("formula %s not found", id)
	}

	return nil
}

// queryParts loads all parts for a formula with their nested ingredients and dosing actions.
func (r *FormulaRepo) queryParts(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, formulaID uuid.UUID) ([]models.FormulaPart, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, name, mix_ratio, sort_order
		 FROM formula_parts WHERE formula_id = $1 ORDER BY sort_order`, formulaID,
	)
	if err != nil {
		return nil, err
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

		ingredients, err := r.queryIngredients(ctx, db, p.ID)
		if err != nil {
			return nil, fmt.Errorf("query ingredients for part %s: %w", p.ID, err)
		}
		p.Ingredients = ingredients

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

// queryIngredients loads all ingredients for a part with their nested dosing actions.
func (r *FormulaRepo) queryIngredients(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, partID uuid.UUID) ([]models.FormulaIngredient, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, part_id, sort_order, material, percentage, weight
		 FROM formula_ingredients WHERE part_id = $1 ORDER BY sort_order`, partID,
	)
	if err != nil {
		return nil, err
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

		dosingActions, err := r.queryDosingActions(ctx, db, ing.ID)
		if err != nil {
			return nil, fmt.Errorf("query dosing actions for ingredient %s: %w", ing.ID, err)
		}
		ing.DosingActions = dosingActions

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

// queryDosingActions loads all dosing actions for an ingredient.
func (r *FormulaRepo) queryDosingActions(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, ingredientID uuid.UUID) ([]models.FormulaDosingAction, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method
		 FROM formula_dosing_actions WHERE ingredient_id = $1 ORDER BY dosing_order`, ingredientID,
	)
	if err != nil {
		return nil, err
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

// querySteps loads all steps for a formula.
func (r *FormulaRepo) querySteps(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, formulaID uuid.UUID) ([]models.FormulaStep, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, part_id::text, step_no, name, temperature, duration
		 FROM formula_steps WHERE formula_id = $1 ORDER BY step_no`, formulaID,
	)
	if err != nil {
		return nil, err
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

// queryPartsBulk loads all parts for multiple formula IDs in a single query.
func (r *FormulaRepo) queryPartsBulk(ctx context.Context, db *sql.DB, formulaIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaPart, error) {
	result := make(map[uuid.UUID][]models.FormulaPart)
	if len(formulaIDs) == 0 {
		return result, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, name, mix_ratio, sort_order
		 FROM formula_parts WHERE formula_id = ANY($1) ORDER BY sort_order`,
		pq.Array(formulaIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.FormulaPart
		var dbName string
		if err := rows.Scan(&p.ID, &p.FormulaID, &dbName, &p.MixRatio, &p.SortOrder); err != nil {
			return nil, fmt.Errorf("scan part: %w", err)
		}
		p.Name = partNameFromDB(dbName)
		result[p.FormulaID] = append(result[p.FormulaID], p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// queryIngredientsBulk loads all ingredients for multiple part IDs in a single query.
func (r *FormulaRepo) queryIngredientsBulk(ctx context.Context, db *sql.DB, partIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaIngredient, error) {
	result := make(map[uuid.UUID][]models.FormulaIngredient)
	if len(partIDs) == 0 {
		return result, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, part_id, sort_order, material, percentage, weight
		 FROM formula_ingredients WHERE part_id = ANY($1) ORDER BY sort_order`,
		pq.Array(partIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ing models.FormulaIngredient
		var weight sql.NullFloat64
		if err := rows.Scan(&ing.ID, &ing.PartID, &ing.SortOrder, &ing.Material, &ing.Percentage, &weight); err != nil {
			return nil, fmt.Errorf("scan ingredient: %w", err)
		}
		if weight.Valid {
			ing.Weight = weight.Float64
		}
		result[ing.PartID] = append(result[ing.PartID], ing)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// queryDosingActionsBulk loads all dosing actions for multiple ingredient IDs in a single query.
func (r *FormulaRepo) queryDosingActionsBulk(ctx context.Context, db *sql.DB, ingredientIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaDosingAction, error) {
	result := make(map[uuid.UUID][]models.FormulaDosingAction)
	if len(ingredientIDs) == 0 {
		return result, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, step_id, ingredient_id, dosing_order, use_ratio, dosing_method
		 FROM formula_dosing_actions WHERE ingredient_id = ANY($1) ORDER BY dosing_order`,
		pq.Array(ingredientIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var da models.FormulaDosingAction
		var dosingMethod sql.NullString
		if err := rows.Scan(&da.ID, &da.StepID, &da.IngredientID, &da.DosingOrder, &da.UseRatio, &dosingMethod); err != nil {
			return nil, fmt.Errorf("scan dosing action: %w", err)
		}
		if dosingMethod.Valid {
			da.DosingMethod = dosingMethod.String
		}
		result[da.IngredientID] = append(result[da.IngredientID], da)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// queryStepsBulk loads all steps for multiple formula IDs in a single query.
func (r *FormulaRepo) queryStepsBulk(ctx context.Context, db *sql.DB, formulaIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaStep, error) {
	result := make(map[uuid.UUID][]models.FormulaStep)
	if len(formulaIDs) == 0 {
		return result, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, part_id::text, step_no, name, temperature, duration
		 FROM formula_steps WHERE formula_id = ANY($1) ORDER BY step_no`,
		pq.Array(formulaIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s models.FormulaStep
		var partIDStr sql.NullString
		if err := rows.Scan(&s.ID, &s.FormulaID, &partIDStr, &s.StepNo, &s.Name, &s.Temperature, &s.Duration); err != nil {
			return nil, fmt.Errorf("scan step: %w", err)
		}
		if partIDStr.Valid {
			if uid, err := uuid.Parse(partIDStr.String); err == nil {
				s.PartID = &uid
			}
		}
		result[s.FormulaID] = append(result[s.FormulaID], s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// partNameToDB converts a model PartName to the database representation.
func partNameToDB(name models.PartName) string {
	switch name {
	case models.PartA:
		return "A"
	case models.PartB:
		return "B"
	case models.PartMain:
		return "MAIN"
	default:
		return string(name)
	}
}

// partNameFromDB converts a database part name to the model PartName.
func partNameFromDB(s string) models.PartName {
	switch s {
	case "A":
		return models.PartA
	case "B":
		return models.PartB
	case "MAIN":
		return models.PartMain
	default:
		return models.PartName(s)
	}
}
