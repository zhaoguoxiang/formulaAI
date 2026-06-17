package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"formula-ai-system/backend/internal/models"
	"formula-ai-system/backend/internal/repository"
)

// TestOutlineHandler exposes HTTP handlers for test outline CRUD operations.
type TestOutlineHandler struct {
	repo *repository.TestOutlineRepo
	db   *sql.DB
}

func NewTestOutlineHandler(db *sql.DB) *TestOutlineHandler {
	return &TestOutlineHandler{
		repo: repository.NewTestOutlineRepo(),
		db:   db,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/test-outlines
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) CreateTestOutline(c *gin.Context) {
	var o models.TestOutline

	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid JSON body",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.Create(c.Request.Context(), h.db, &o); err != nil {
		serverError(c, "failed to create test outline", err)
		return
	}

	c.JSON(http.StatusCreated, o)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/test-outlines
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) ListTestOutlines(c *gin.Context) {
	outlines, err := h.repo.List(c.Request.Context(), h.db)
	if err != nil {
		serverError(c, "failed to list test outlines", err)
		return
	}

	c.JSON(http.StatusOK, outlines)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/test-outlines/:id
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) GetTestOutline(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid test outline ID",
			"details": err.Error(),
		})
		return
	}

	o, err := h.repo.GetByID(c.Request.Context(), h.db, id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "test outline not found"})
			return
		}
		serverError(c, "failed to get test outline", err)
		return
	}

	c.JSON(http.StatusOK, o)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/test-outlines/:id — creates a new version of the outline
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) SaveVersion(c *gin.Context) {
	var o models.TestOutline
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid JSON body",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.SaveVersion(c.Request.Context(), h.db, &o); err != nil {
		serverError(c, "failed to save new version", err)
		return
	}

	c.JSON(http.StatusOK, o)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/test-outlines/:id/archive — soft-delete (status = 'archived')
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) ArchiveTestOutline(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid test outline ID",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.Archive(c.Request.Context(), h.db, id); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "test outline not found"})
			return
		}
		serverError(c, "failed to archive test outline", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "archived"})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/test-outlines/:id/versions — list all versions by outline name
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) ListVersions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid test outline ID",
			"details": err.Error(),
		})
		return
	}

	o, err := h.repo.GetByID(c.Request.Context(), h.db, id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "test outline not found"})
			return
		}
		serverError(c, "failed to get test outline", err)
		return
	}

	versions, err := h.repo.ListVersions(c.Request.Context(), h.db, o.Name)
	if err != nil {
		serverError(c, "failed to list versions", err)
		return
	}

	c.JSON(http.StatusOK, versions)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/test-outlines/:id/activate — set a specific version as active
// ─────────────────────────────────────────────────────────────────────────────

func (h *TestOutlineHandler) ActivateVersion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid test outline ID",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.ActivateVersion(c.Request.Context(), h.db, id); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "test outline not found"})
			return
		}
		serverError(c, "failed to activate version", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "activated"})
}
