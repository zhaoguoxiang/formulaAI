package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
	"formula-ai-system/backend/internal/repository"
	"formula-ai-system/backend/internal/services"
)

// FormulaHandler exposes HTTP handlers for formula CRUD operations.
type FormulaHandler struct {
	repo *repository.FormulaRepo
	db   *sql.DB
}

// NewFormulaHandler creates a new FormulaHandler wired to the given database.
func NewFormulaHandler(db *sql.DB) *FormulaHandler {
	return &FormulaHandler{
		repo: repository.NewFormulaRepo(),
		db:   db,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/formulas
// ─────────────────────────────────────────────────────────────────────────────

// CreateFormula handles POST /api/formulas. It parses JSON, validates the
// formula, and persists it. Returns 201 on success or 400/422/500 on failure.
func (h *FormulaHandler) CreateFormula(c *gin.Context) {
	var f models.Formula

	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid JSON body",
			"details": err.Error(),
		})
		return
	}

	// Business-level validation
	if err := services.ValidateAndPrepare(&f); err != nil {
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			if ve.IsBlocking() {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":   "validation failed",
					"details": ve.Error(),
					"errors":  ve.Errors,
					"warnings": ve.Warnings,
				})
				return
			}
			// warnings only – log but proceed
		}
	}

	if err := h.repo.Create(c.Request.Context(), h.db, &f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create formula",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, f)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/formulas
// ─────────────────────────────────────────────────────────────────────────────

// ListFormulas handles GET /api/formulas. Returns all formulas with their
// nested data as a JSON array.
func (h *FormulaHandler) ListFormulas(c *gin.Context) {
	formulas, err := h.repo.List(c.Request.Context(), h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list formulas",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, formulas)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/formulas/:id
// ─────────────────────────────────────────────────────────────────────────────

// GetFormula handles GET /api/formulas/:id. Parses the UUID path parameter
// and returns the formula or 404.
func (h *FormulaHandler) GetFormula(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid formula ID",
			"details": err.Error(),
		})
		return
	}

	f, err := h.repo.GetByID(c.Request.Context(), h.db, id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "formula not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get formula",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, f)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/formulas/:id
// ─────────────────────────────────────────────────────────────────────────────

// UpdateFormula handles PUT /api/formulas/:id. Parses UUID and JSON, validates,
// and updates the formula atomically.
func (h *FormulaHandler) UpdateFormula(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid formula ID",
			"details": err.Error(),
		})
		return
	}

	var f models.Formula
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid JSON body",
			"details": err.Error(),
		})
		return
	}

	// Enforce path ID matches body ID when both are present
	if f.ID != uuid.Nil && f.ID != id {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "path ID does not match body ID",
		})
		return
	}
	f.ID = id

	// Business-level validation
	if err := services.ValidateAndPrepare(&f); err != nil {
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			if ve.IsBlocking() {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":   "validation failed",
					"details": ve.Error(),
					"errors":  ve.Errors,
					"warnings": ve.Warnings,
				})
				return
			}
		}
	}

	if err := h.repo.Update(c.Request.Context(), h.db, &f); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "formula not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update formula",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, f)
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/formulas/:id
// ─────────────────────────────────────────────────────────────────────────────

// DeleteFormula handles DELETE /api/formulas/:id. Removes the formula and all
// nested data via cascade. Returns 204 or 404.
func (h *FormulaHandler) DeleteFormula(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid formula ID",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), h.db, id); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "formula not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete formula",
			"details": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// isNotFound checks whether an error indicates that a requested resource was
// not found (i.e. sql.ErrNoRows or a "not found" message from the repo).
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	// sql.ErrNoRows
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	// Repository returns "formula <uuid> not found" or wraps sql.ErrNoRows
	msg := err.Error()
	return strings.Contains(msg, "not found")
}
