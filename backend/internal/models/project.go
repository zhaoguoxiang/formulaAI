package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents an isolated workspace that groups formulas, test outlines,
// and all related data. Each project is completely independent — data is not
// shared across projects.
type Project struct {
	ID          uuid.UUID `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}
