package handlers

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"time"

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

	// Auto-generate code if empty: FML-YYMMDD-XXXX
	if f.Code == "" {
		f.Code = generateFormulaCode()
	}

	// Business-level validation
	if err := services.ValidateAndPrepare(&f); err != nil {
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			if ve.IsBlocking() {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":    "validation failed",
					"details":  ve.Error(),
					"errors":   ve.Errors,
					"warnings": ve.Warnings,
				})
				return
			}
			// Log non-blocking validation warnings
			for _, w := range ve.Warnings {
				slog.Warn("formula validation warning", "warning", w)
			}
		}
	}

	if err := h.repo.Create(c.Request.Context(), h.db, &f); err != nil {
		serverError(c, "failed to create formula", err)
		return
	}

	c.JSON(http.StatusCreated, f)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/formulas
// ─────────────────────────────────────────────────────────────────────────────

// ListFormulas handles GET /api/formulas. Returns all formulas with their
// nested data as a JSON array. Supports ?formula_type=formula|material filter.
func (h *FormulaHandler) ListFormulas(c *gin.Context) {
	opts := repository.ListOptions{
		FormulaType: c.Query("formula_type"),
	}
	formulas, err := h.repo.List(c.Request.Context(), h.db, opts)
	if err != nil {
		serverError(c, "failed to list formulas", err)
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
		serverError(c, "failed to get formula", err)
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
					"error":    "validation failed",
					"details":  ve.Error(),
					"errors":   ve.Errors,
					"warnings": ve.Warnings,
				})
				return
			}
			for _, w := range ve.Warnings {
				slog.Warn("formula validation warning", "warning", w)
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
		serverError(c, "failed to update formula", err)
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
		serverError(c, "failed to delete formula", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// generateFormulaCode creates an auto-generated code in format FML-YYMMDD-XXXX.
func generateFormulaCode() string {
	now := time.Now()
	dateStr := fmt.Sprintf("%02d%02d%02d", now.Year()%100, now.Month(), now.Day())
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	suffix := make([]byte, 4)
	for i := range suffix {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			suffix[i] = 'X' // fallback on crypto failure (extremely unlikely)
		} else {
			suffix[i] = chars[n.Int64()]
		}
	}
	return fmt.Sprintf("FML-%s-%s", dateStr, string(suffix))
}

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
