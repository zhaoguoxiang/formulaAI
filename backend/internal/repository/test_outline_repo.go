package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
)

// TestOutlineRepo provides CRUD operations for test outlines with nested test items
// and indicators. All writes use transactions for atomicity.
type TestOutlineRepo struct{}

// NewTestOutlineRepo creates a new TestOutlineRepo instance.
func NewTestOutlineRepo() *TestOutlineRepo {
	return &TestOutlineRepo{}
}

// ─────────────────────────────────────────────────────────────────────────────
// Create
// ─────────────────────────────────────────────────────────────────────────────

// Create inserts a new test outline with version=1 and status='active'.
func (r *TestOutlineRepo) Create(ctx context.Context, db *sql.DB, o *models.TestOutline) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	r.prepareForInsert(o)
	o.Version = 1
	o.Status = models.OutlineStatusActive

	_, err = tx.ExecContext(ctx,
		`INSERT INTO test_outlines (id, name, version, version_note, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		o.ID, o.Name, o.Version, o.VersionNote, o.Status, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert test outline: %w", err)
	}

	if err := r.insertItems(ctx, tx, o); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SaveVersion — archives current active version, creates new with version+1
// ─────────────────────────────────────────────────────────────────────────────

// SaveVersion creates a new version of an existing outline. The current active
// version is set to 'archived', and a new row is inserted with version+1.
func (r *TestOutlineRepo) SaveVersion(ctx context.Context, db *sql.DB, o *models.TestOutline) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Find max version for this outline name (with lock to prevent concurrent version collisions)
	var maxVersion int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM test_outlines WHERE name = $1 FOR UPDATE`,
		o.Name,
	).Scan(&maxVersion)
	if err != nil {
		return fmt.Errorf("query max version: %w", err)
	}

	// Archive all existing versions with this name
	_, err = tx.ExecContext(ctx,
		`UPDATE test_outlines SET status = $1, updated_at = $2 WHERE name = $3`,
		models.OutlineStatusArchived, time.Now(), o.Name,
	)
	if err != nil {
		return fmt.Errorf("archive old versions: %w", err)
	}

	// Insert new version
	r.prepareForInsert(o)
	o.ID = uuid.New()
	o.Version = maxVersion + 1
	o.Status = models.OutlineStatusActive

	_, err = tx.ExecContext(ctx,
		`INSERT INTO test_outlines (id, name, version, version_note, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		o.ID, o.Name, o.Version, o.VersionNote, o.Status, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert new version: %w", err)
	}

	// Re-link items to new outline ID
	for i := range o.Items {
		o.Items[i].ID = uuid.New()
		o.Items[i].OutlineID = o.ID
		for j := range o.Items[i].Indicators {
			o.Items[i].Indicators[j].ID = uuid.New()
			o.Items[i].Indicators[j].ItemID = o.Items[i].ID
			o.Items[i].Indicators[j].CreatedAt = o.CreatedAt
		}
	}

	if err := r.insertItems(ctx, tx, o); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Archive — soft delete by setting status = 'archived'
// ─────────────────────────────────────────────────────────────────────────────

// Archive sets all versions of an outline name to 'archived'.
func (r *TestOutlineRepo) Archive(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	// Find the outline name first
	var name string
	err := db.QueryRowContext(ctx,
		`SELECT name FROM test_outlines WHERE id = $1`, id,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("test outline %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return fmt.Errorf("query outline name: %w", err)
	}

	result, err := db.ExecContext(ctx,
		`UPDATE test_outlines SET status = $1, updated_at = $2 WHERE name = $3`,
		models.OutlineStatusArchived, time.Now(), name,
	)
	if err != nil {
		return fmt.Errorf("archive test outlines: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("test outline %s not found", id)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetByID
// ─────────────────────────────────────────────────────────────────────────────

func (r *TestOutlineRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.TestOutline, error) {
	o := &models.TestOutline{}
	err := db.QueryRowContext(ctx,
		`SELECT id, name, version, version_note, status, created_at, updated_at
		 FROM test_outlines WHERE id = $1`, id,
	).Scan(&o.ID, &o.Name, &o.Version, &o.VersionNote, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test outline %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query test outline: %w", err)
	}

	items, err := r.queryItems(ctx, db, o.ID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	o.Items = items

	return o, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// List — only returns the latest active version per outline name
// ─────────────────────────────────────────────────────────────────────────────

func (r *TestOutlineRepo) List(ctx context.Context, db *sql.DB) ([]*models.TestOutline, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, version, version_note, status, created_at, updated_at
		 FROM test_outlines WHERE status != $1 ORDER BY created_at DESC`,
		models.OutlineStatusArchived,
	)
	if err != nil {
		return nil, fmt.Errorf("query test outlines: %w", err)
	}
	defer rows.Close()

	var outlines []*models.TestOutline
	for rows.Next() {
		o := &models.TestOutline{}
		if err := rows.Scan(&o.ID, &o.Name, &o.Version, &o.VersionNote, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan test outline: %w", err)
		}

		items, err := r.queryItems(ctx, db, o.ID)
		if err != nil {
			return nil, fmt.Errorf("query items for outline %s: %w", o.ID, err)
		}
		o.Items = items

		outlines = append(outlines, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if outlines == nil {
		outlines = []*models.TestOutline{}
	}

	return outlines, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ListVersions — returns all versions of outlines with a given name
// ─────────────────────────────────────────────────────────────────────────────

func (r *TestOutlineRepo) ListVersions(ctx context.Context, db *sql.DB, name string) ([]*models.TestOutline, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, version, version_note, status, created_at, updated_at
		 FROM test_outlines WHERE name = $1 ORDER BY version DESC`, name,
	)
	if err != nil {
		return nil, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var outlines []*models.TestOutline
	for rows.Next() {
		o := &models.TestOutline{}
		if err := rows.Scan(&o.ID, &o.Name, &o.Version, &o.VersionNote, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		outlines = append(outlines, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if outlines == nil {
		outlines = []*models.TestOutline{}
	}

	return outlines, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ActivateVersion — sets a specific version to 'active', archives others
// ─────────────────────────────────────────────────────────────────────────────

func (r *TestOutlineRepo) ActivateVersion(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Get the outline name inside the transaction to avoid TOCTOU
	var name string
	err = tx.QueryRowContext(ctx,
		`SELECT name FROM test_outlines WHERE id = $1 FOR UPDATE`, id,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("test outline %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return fmt.Errorf("query outline name: %w", err)
	}

	// Archive all versions with this name
	_, err = tx.ExecContext(ctx,
		`UPDATE test_outlines SET status = $1, updated_at = $2 WHERE name = $3`,
		models.OutlineStatusArchived, time.Now(), name,
	)
	if err != nil {
		return fmt.Errorf("archive versions: %w", err)
	}

	// Activate the selected version
	_, err = tx.ExecContext(ctx,
		`UPDATE test_outlines SET status = $1, updated_at = $2 WHERE id = $3`,
		models.OutlineStatusActive, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("activate version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Delete — physical delete (kept for admin if needed)
// ─────────────────────────────────────────────────────────────────────────────

// Delete removes a test outline and all nested data via CASCADE constraints.
func (r *TestOutlineRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM test_outlines WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete test outline: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("test outline %s not found", id)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

type querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func (r *TestOutlineRepo) prepareForInsert(o *models.TestOutline) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = now

	for i := range o.Items {
		item := &o.Items[i]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.OutlineID = o.ID
		item.CreatedAt = now

		for j := range item.Indicators {
			ind := &item.Indicators[j]
			if ind.ID == uuid.Nil {
				ind.ID = uuid.New()
			}
			ind.ItemID = item.ID
			ind.CreatedAt = now
		}
	}
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (r *TestOutlineRepo) insertItems(ctx context.Context, tx execer, o *models.TestOutline) error {
	for ii := range o.Items {
		item := &o.Items[ii]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.OutlineID = o.ID
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now()
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO test_items (id, outline_id, name, sort_order, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			item.ID, item.OutlineID, item.Name, item.SortOrder, item.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert item %d: %w", ii, err)
		}

		for ji := range item.Indicators {
			ind := &item.Indicators[ji]
			if ind.ID == uuid.Nil {
				ind.ID = uuid.New()
			}
			ind.ItemID = item.ID
			if ind.CreatedAt.IsZero() {
				ind.CreatedAt = time.Now()
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO test_indicators (id, item_id, name, unit, min_value, max_value,
				     sample_prep_method, test_method, test_condition, version, sort_order, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
				ind.ID, ind.ItemID, ind.Name,
				nullStr(ind.Unit),
				nullFloat(ind.MinValue),
				nullFloat(ind.MaxValue),
				nullStr(ind.SamplePrepMethod),
				nullStr(ind.TestMethod),
				nullStr(ind.TestCondition),
				nullStr(ind.Version),
				ind.SortOrder, ind.CreatedAt,
			)
			if err != nil {
				return fmt.Errorf("insert indicator %d in item %d: %w", ji, ii, err)
			}
		}
	}
	return nil
}

func (r *TestOutlineRepo) queryItems(ctx context.Context, q querier, outlineID uuid.UUID) ([]models.TestItem, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, outline_id, name, sort_order, created_at
		 FROM test_items WHERE outline_id = $1 ORDER BY sort_order, created_at`, outlineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.TestItem
	for rows.Next() {
		var item models.TestItem
		if err := rows.Scan(&item.ID, &item.OutlineID, &item.Name, &item.SortOrder, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		indicators, err := r.queryIndicators(ctx, q, item.ID)
		if err != nil {
			return nil, fmt.Errorf("query indicators for item %s: %w", item.ID, err)
		}
		item.Indicators = indicators

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []models.TestItem{}
	}

	return items, nil
}

func (r *TestOutlineRepo) queryIndicators(ctx context.Context, q querier, itemID uuid.UUID) ([]models.TestIndicator, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, item_id, name, unit, min_value, max_value,
		        sample_prep_method, test_method, test_condition, version, sort_order, created_at
		 FROM test_indicators WHERE item_id = $1 ORDER BY sort_order, created_at`, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []models.TestIndicator
	for rows.Next() {
		var ind models.TestIndicator
		var unit, samplePrepMethod, testMethod, testCondition, version sql.NullString
		var minValue, maxValue sql.NullFloat64

		if err := rows.Scan(
			&ind.ID, &ind.ItemID, &ind.Name,
			&unit, &minValue, &maxValue,
			&samplePrepMethod, &testMethod, &testCondition, &version,
			&ind.SortOrder, &ind.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan indicator: %w", err)
		}

		if unit.Valid {
			ind.Unit = unit.String
		}
		if minValue.Valid {
			v := minValue.Float64
			ind.MinValue = &v
		}
		if maxValue.Valid {
			v := maxValue.Float64
			ind.MaxValue = &v
		}
		if samplePrepMethod.Valid {
			ind.SamplePrepMethod = samplePrepMethod.String
		}
		if testMethod.Valid {
			ind.TestMethod = testMethod.String
		}
		if testCondition.Valid {
			ind.TestCondition = testCondition.String
		}
		if version.Valid {
			ind.Version = version.String
		}

		indicators = append(indicators, ind)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if indicators == nil {
		indicators = []models.TestIndicator{}
	}

	return indicators, nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullFloat(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
