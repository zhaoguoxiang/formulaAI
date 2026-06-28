package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
)

// ProjectRepo provides CRUD operations for projects (workspaces).
// Deleting a project cascades to all formulas and test outlines within it.
type ProjectRepo struct{}

func NewProjectRepo() *ProjectRepo {
	return &ProjectRepo{}
}

func (r *ProjectRepo) Create(ctx context.Context, db *sql.DB, p *models.Project) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := db.ExecContext(ctx,
		`INSERT INTO projects (id, name, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		p.ID, p.Name, p.Description, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (r *ProjectRepo) GetByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	err := db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project %s: %w", id, sql.ErrNoRows)
	}
	if err != nil {
		return nil, fmt.Errorf("query project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepo) List(ctx context.Context, db *sql.DB) ([]*models.Project, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	if projects == nil {
		projects = []*models.Project{}
	}
	return projects, nil
}

func (r *ProjectRepo) Update(ctx context.Context, db *sql.DB, p *models.Project) error {
	p.UpdatedAt = time.Now()
	result, err := db.ExecContext(ctx,
		`UPDATE projects SET name = $1, description = $2, updated_at = $3 WHERE id = $4`,
		p.Name, p.Description, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("project %s not found", p.ID)
	}
	return nil
}

func (r *ProjectRepo) Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`DELETE FROM projects WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("project %s not found", id)
	}
	return nil
}
