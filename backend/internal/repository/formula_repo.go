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

type ListOptions struct {
	ComponentMode string
	Limit         int
	Offset        int
}

type FormulaRepo struct{}

func NewFormulaRepo() *FormulaRepo {
	return &FormulaRepo{}
}

func (r *FormulaRepo) Create(ctx context.Context, db *sql.DB, f *models.Formula) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

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

	// 2. Insert parts (simplified - no mix_ratio or materials)
	for pi := range f.Parts {
		part := &f.Parts[pi]
		if part.ID == uuid.Nil {
			part.ID = uuid.New()
		}
		part.FormulaID = f.ID

		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_parts (id, formula_id, name, sort_order)
			 VALUES ($1, $2, $3, $4)`,
			part.ID, part.FormulaID, partNameToDB(part.Name), part.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert part %d: %w", pi, err)
		}
	}

	// 3. Insert steps with material categories, materials, and parameters
	for si := range f.Steps {
		step := &f.Steps[si]
		if step.ID == uuid.Nil {
			step.ID = uuid.New()
		}
		step.FormulaID = f.ID

		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_steps (id, formula_id, part_id, step_no, name, description, instrument_name, temperature, duration)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			step.ID, step.FormulaID, step.PartID, step.StepNo, step.Name, step.Description, step.InstrumentName, step.Temperature, step.Duration,
		)
		if err != nil {
			return fmt.Errorf("insert step %d: %w", si, err)
		}

		// Insert material categories and materials (for 投料 steps)
		for ci := range step.Categories {
			cat := &step.Categories[ci]
			if cat.ID == uuid.Nil {
				cat.ID = uuid.New()
			}
			cat.StepID = step.ID

			_, err = tx.ExecContext(ctx,
				`INSERT INTO formula_step_material_categories (id, step_id, name, sort_order)
				 VALUES ($1, $2, $3, $4)`,
				cat.ID, cat.StepID, cat.Name, cat.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("insert material category %d in step %d: %w", ci, si, err)
			}

			for mi := range cat.Materials {
				mat := &cat.Materials[mi]
				if mat.ID == uuid.Nil {
					mat.ID = uuid.New()
				}
				mat.CategoryID = cat.ID

				_, err = tx.ExecContext(ctx,
					`INSERT INTO formula_step_materials (id, category_id, material, percentage, weight, batch_no, unit, sort_order)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
					mat.ID, mat.CategoryID, mat.Material, mat.Percentage, mat.Weight, mat.BatchNo, mat.Unit, mat.SortOrder,
				)
				if err != nil {
					return fmt.Errorf("insert material %d in category %d: %w", mi, ci, err)
				}
			}
		}

		// Insert step parameters
		for pi := range step.Parameters {
			param := &step.Parameters[pi]
			if param.ID == uuid.Nil {
				param.ID = uuid.New()
			}
			param.StepID = step.ID

			_, err = tx.ExecContext(ctx,
				`INSERT INTO formula_step_parameters (id, step_id, name, value, unit, sort_order)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				param.ID, param.StepID, param.Name, param.Value, param.Unit, param.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("insert step parameter %d for step %d: %w", pi, si, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *FormulaRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.Formula, error) {
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

	parts, err := r.queryParts(ctx, db, f.ID)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	f.Parts = parts

	steps, err := r.querySteps(ctx, db, f.ID)
	if err != nil {
		return nil, fmt.Errorf("query steps: %w", err)
	}
	f.Steps = steps

	return f, nil
}

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

	partsByFormula, err := r.queryPartsBulk(ctx, db, formulaIDs)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	stepsByFormula, err := r.queryStepsBulk(ctx, db, formulaIDs)
	if err != nil {
		return nil, fmt.Errorf("query steps: %w", err)
	}

	for _, f := range formulas {
		f.Parts = partsByFormula[f.ID]
		if f.Parts == nil {
			f.Parts = []models.FormulaPart{}
		}
		f.Steps = stepsByFormula[f.ID]
		if f.Steps == nil {
			f.Steps = []models.FormulaStep{}
		}
	}

	return formulas, nil
}

func (r *FormulaRepo) Update(ctx context.Context, db *sql.DB, f *models.Formula) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	f.UpdatedAt = time.Now()

	result, err := tx.ExecContext(ctx,
		`UPDATE formulas SET name=$1, code=$2, component_mode=$3, status=$4, updated_at=$5 WHERE id=$6`,
		f.Name, f.Code, string(f.ComponentMode), string(f.Status), f.UpdatedAt, f.ID,
	)
	if err != nil {
		return fmt.Errorf("update formula: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("formula %s not found", f.ID)
	}

	// Delete all nested data
	_, err = tx.ExecContext(ctx, `DELETE FROM formula_steps WHERE formula_id = $1`, f.ID)
	if err != nil {
		return fmt.Errorf("delete nested steps: %w", err)
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM formula_parts WHERE formula_id = $1`, f.ID)
	if err != nil {
		return fmt.Errorf("delete nested parts: %w", err)
	}

	// Re-insert parts
	for pi := range f.Parts {
		part := &f.Parts[pi]
		if part.ID == uuid.Nil {
			part.ID = uuid.New()
		}
		part.FormulaID = f.ID
		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_parts (id, formula_id, name, sort_order)
			 VALUES ($1, $2, $3, $4)`,
			part.ID, part.FormulaID, partNameToDB(part.Name), part.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("re-insert part %d: %w", pi, err)
		}
	}

	// Re-insert steps (with materials, categories, parameters)
	for si := range f.Steps {
		step := &f.Steps[si]
		if step.ID == uuid.Nil {
			step.ID = uuid.New()
		}
		step.FormulaID = f.ID
		_, err = tx.ExecContext(ctx,
			`INSERT INTO formula_steps (id, formula_id, part_id, step_no, name, description, instrument_name, temperature, duration)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			step.ID, step.FormulaID, step.PartID, step.StepNo, step.Name, step.Description, step.InstrumentName, step.Temperature, step.Duration,
		)
		if err != nil {
			return fmt.Errorf("re-insert step %d: %w", si, err)
		}

		for ci := range step.Categories {
			cat := &step.Categories[ci]
			if cat.ID == uuid.Nil {
				cat.ID = uuid.New()
			}
			cat.StepID = step.ID
			_, err = tx.ExecContext(ctx,
				`INSERT INTO formula_step_material_categories (id, step_id, name, sort_order)
				 VALUES ($1, $2, $3, $4)`,
				cat.ID, cat.StepID, cat.Name, cat.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("re-insert material category %d: %w", ci, err)
			}

			for mi := range cat.Materials {
				mat := &cat.Materials[mi]
				if mat.ID == uuid.Nil {
					mat.ID = uuid.New()
				}
				mat.CategoryID = cat.ID
				_, err = tx.ExecContext(ctx,
					`INSERT INTO formula_step_materials (id, category_id, material, percentage, weight, batch_no, unit, sort_order)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
					mat.ID, mat.CategoryID, mat.Material, mat.Percentage, mat.Weight, mat.BatchNo, mat.Unit, mat.SortOrder,
				)
				if err != nil {
					return fmt.Errorf("re-insert material %d: %w", mi, err)
				}
			}
		}

		for pi := range step.Parameters {
			param := &step.Parameters[pi]
			if param.ID == uuid.Nil {
				param.ID = uuid.New()
			}
			param.StepID = step.ID
			_, err = tx.ExecContext(ctx,
				`INSERT INTO formula_step_parameters (id, step_id, name, value, unit, sort_order)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				param.ID, param.StepID, param.Name, param.Value, param.Unit, param.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("re-insert step parameter %d: %w", pi, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

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

// ─── Query helpers ───

func (r *FormulaRepo) queryParts(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, formulaID uuid.UUID) ([]models.FormulaPart, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, name, sort_order FROM formula_parts WHERE formula_id = $1 ORDER BY sort_order`, formulaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.FormulaPart
	for rows.Next() {
		var p models.FormulaPart
		var dbName string
		if err := rows.Scan(&p.ID, &p.FormulaID, &dbName, &p.SortOrder); err != nil {
			return nil, fmt.Errorf("scan part: %w", err)
		}
		p.Name = partNameFromDB(dbName)
		parts = append(parts, p)
	}
	if parts == nil {
		parts = []models.FormulaPart{}
	}
	return parts, rows.Err()
}

func (r *FormulaRepo) queryPartsBulk(ctx context.Context, db *sql.DB, formulaIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaPart, error) {
	result := make(map[uuid.UUID][]models.FormulaPart)
	if len(formulaIDs) == 0 {
		return result, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, name, sort_order FROM formula_parts WHERE formula_id = ANY($1) ORDER BY sort_order`,
		pq.Array(formulaIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.FormulaPart
		var dbName string
		if err := rows.Scan(&p.ID, &p.FormulaID, &dbName, &p.SortOrder); err != nil {
			return nil, fmt.Errorf("scan part: %w", err)
		}
		p.Name = partNameFromDB(dbName)
		result[p.FormulaID] = append(result[p.FormulaID], p)
	}
	return result, rows.Err()
}

func (r *FormulaRepo) querySteps(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, formulaID uuid.UUID) ([]models.FormulaStep, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, part_id::text, step_no, name, description, instrument_name, temperature, duration
		 FROM formula_steps WHERE formula_id = $1 ORDER BY step_no`, formulaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.FormulaStep
	for rows.Next() {
		s, err := scanStep(rows)
		if err != nil {
			return nil, err
		}
		s.Parameters, err = r.queryStepParameters(ctx, db, s.ID)
		if err != nil {
			return nil, err
		}
		s.Categories, err = r.queryMaterialCategories(ctx, db, s.ID)
		if err != nil {
			return nil, err
		}
		steps = append(steps, *s)
	}
	if steps == nil {
		steps = []models.FormulaStep{}
	}
	return steps, rows.Err()
}

func (r *FormulaRepo) queryStepsBulk(ctx context.Context, db *sql.DB, formulaIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaStep, error) {
	result := make(map[uuid.UUID][]models.FormulaStep)
	if len(formulaIDs) == 0 {
		return result, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, formula_id, part_id::text, step_no, name, description, instrument_name, temperature, duration
		 FROM formula_steps WHERE formula_id = ANY($1) ORDER BY step_no`,
		pq.Array(formulaIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stepIDs []uuid.UUID
	for rows.Next() {
		s, err := scanStep(rows)
		if err != nil {
			return nil, err
		}
		stepIDs = append(stepIDs, s.ID)
		result[s.FormulaID] = append(result[s.FormulaID], *s)
	}

	if len(stepIDs) > 0 {
		paramsByStep, _ := r.queryStepParametersBulk(ctx, db, stepIDs)
		materialsByStep, _ := r.queryMaterialCategoriesBulk(ctx, db, stepIDs)
		for fid, stps := range result {
			for i := range stps {
				stps[i].Parameters = paramsByStep[stps[i].ID]
				if stps[i].Parameters == nil {
					stps[i].Parameters = []models.FormulaStepParameter{}
				}
				stps[i].Categories = materialsByStep[stps[i].ID]
				if stps[i].Categories == nil {
					stps[i].Categories = []models.FormulaStepMaterialCategory{}
				}
			}
			result[fid] = stps
		}
	}
	return result, rows.Err()
}

func (r *FormulaRepo) queryStepParameters(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, stepID uuid.UUID) ([]models.FormulaStepParameter, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, step_id, name, value, unit, sort_order
		 FROM formula_step_parameters WHERE step_id = $1 ORDER BY sort_order`, stepID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	params := scanStepParameters(rows)
	if params == nil {
		params = []models.FormulaStepParameter{}
	}
	return params, nil
}

func (r *FormulaRepo) queryStepParametersBulk(ctx context.Context, db *sql.DB, stepIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaStepParameter, error) {
	result := make(map[uuid.UUID][]models.FormulaStepParameter)
	if len(stepIDs) == 0 {
		return result, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, step_id, name, value, unit, sort_order
		 FROM formula_step_parameters WHERE step_id = ANY($1) ORDER BY sort_order`,
		pq.Array(stepIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for _, p := range scanStepParameters(rows) {
		result[p.StepID] = append(result[p.StepID], p)
	}
	return result, nil
}

func (r *FormulaRepo) queryMaterialCategories(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, stepID uuid.UUID) ([]models.FormulaStepMaterialCategory, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, step_id, name, sort_order
		 FROM formula_step_material_categories WHERE step_id = $1 ORDER BY sort_order`, stepID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.FormulaStepMaterialCategory
	for rows.Next() {
		var c models.FormulaStepMaterialCategory
		if err := rows.Scan(&c.ID, &c.StepID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		mats, err := r.queryMaterials(ctx, db, c.ID)
		if err != nil {
			return nil, err
		}
		c.Materials = mats
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []models.FormulaStepMaterialCategory{}
	}
	return cats, rows.Err()
}

func (r *FormulaRepo) queryMaterialCategoriesBulk(ctx context.Context, db *sql.DB, stepIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaStepMaterialCategory, error) {
	result := make(map[uuid.UUID][]models.FormulaStepMaterialCategory)
	if len(stepIDs) == 0 {
		return result, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, step_id, name, sort_order
		 FROM formula_step_material_categories WHERE step_id = ANY($1) ORDER BY sort_order`,
		pq.Array(stepIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var catIDs []uuid.UUID
	type catEntry struct {
		cat    models.FormulaStepMaterialCategory
		stepID uuid.UUID
	}
	var entries []catEntry
	for rows.Next() {
		var c models.FormulaStepMaterialCategory
		if err := rows.Scan(&c.ID, &c.StepID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		catIDs = append(catIDs, c.ID)
		entries = append(entries, catEntry{cat: c, stepID: c.StepID})
	}

	if len(catIDs) > 0 {
		matsByCat, _ := r.queryMaterialsBulk(ctx, db, catIDs)
		for i, e := range entries {
			entries[i].cat.Materials = matsByCat[e.cat.ID]
			if entries[i].cat.Materials == nil {
				entries[i].cat.Materials = []models.FormulaStepMaterial{}
			}
		}
	}

	for _, e := range entries {
		result[e.stepID] = append(result[e.stepID], e.cat)
	}
	return result, rows.Err()
}

func (r *FormulaRepo) queryMaterials(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, categoryID uuid.UUID) ([]models.FormulaStepMaterial, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, category_id, material, percentage, weight, batch_no, unit, sort_order
		 FROM formula_step_materials WHERE category_id = $1 ORDER BY sort_order`, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mats := scanMaterials(rows)
	if mats == nil {
		mats = []models.FormulaStepMaterial{}
	}
	return mats, nil
}

func (r *FormulaRepo) queryMaterialsBulk(ctx context.Context, db *sql.DB, categoryIDs []uuid.UUID) (map[uuid.UUID][]models.FormulaStepMaterial, error) {
	result := make(map[uuid.UUID][]models.FormulaStepMaterial)
	if len(categoryIDs) == 0 {
		return result, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, category_id, material, percentage, weight, batch_no, unit, sort_order
		 FROM formula_step_materials WHERE category_id = ANY($1) ORDER BY sort_order`,
		pq.Array(categoryIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for _, m := range scanMaterials(rows) {
		result[m.CategoryID] = append(result[m.CategoryID], m)
	}
	return result, nil
}

// ─── Scan helpers ───

func scanStep(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.FormulaStep, error) {
	s := &models.FormulaStep{}
	var partIDStr sql.NullString
	var temperature, duration, instrumentName, description sql.NullString
	if err := scanner.Scan(&s.ID, &s.FormulaID, &partIDStr, &s.StepNo, &s.Name, &description, &instrumentName, &temperature, &duration); err != nil {
		return nil, fmt.Errorf("scan step: %w", err)
	}
	if partIDStr.Valid {
		if uid, err := uuid.Parse(partIDStr.String); err == nil {
			s.PartID = &uid
		}
	}
	if instrumentName.Valid {
		s.InstrumentName = instrumentName.String
	}
	if description.Valid {
		s.Description = description.String
	}
	if temperature.Valid {
		s.Temperature = temperature.String
	}
	if duration.Valid {
		s.Duration = duration.String
	}
	return s, nil
}

func scanStepParameters(rows *sql.Rows) []models.FormulaStepParameter {
	var params []models.FormulaStepParameter
	for rows.Next() {
		var p models.FormulaStepParameter
		if err := rows.Scan(&p.ID, &p.StepID, &p.Name, &p.Value, &p.Unit, &p.SortOrder); err == nil {
			params = append(params, p)
		}
	}
	return params
}

func scanMaterials(rows *sql.Rows) []models.FormulaStepMaterial {
	var mats []models.FormulaStepMaterial
	for rows.Next() {
		var m models.FormulaStepMaterial
		if err := rows.Scan(&m.ID, &m.CategoryID, &m.Material, &m.Percentage, &m.Weight, &m.BatchNo, &m.Unit, &m.SortOrder); err == nil {
			mats = append(mats, m)
		}
	}
	return mats
}

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
