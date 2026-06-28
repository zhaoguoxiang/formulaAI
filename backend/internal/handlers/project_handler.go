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
)

// ProjectHandler exposes HTTP handlers for project (workspace) CRUD.
type ProjectHandler struct {
	repo *repository.ProjectRepo
	db   *sql.DB
}

func NewProjectHandler(db *sql.DB) *ProjectHandler {
	return &ProjectHandler{
		repo: repository.NewProjectRepo(),
		db:   db,
	}
}

// ── POST /api/projects ──

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body", "details": err.Error()})
		return
	}
	if p.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := h.repo.Create(c.Request.Context(), h.db, &p); err != nil {
		serverError(c, "failed to create project", err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ── GET /api/projects ──

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects, err := h.repo.List(c.Request.Context(), h.db)
	if err != nil {
		serverError(c, "failed to list projects", err)
		return
	}
	c.JSON(http.StatusOK, projects)
}

// ── GET /api/projects/:id ──

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	project, err := h.repo.GetByID(c.Request.Context(), h.db, id)
	if err != nil {
		notFoundOrError(c, err, "project")
		return
	}
	c.JSON(http.StatusOK, project)
}

// ── PUT /api/projects/:id ──

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body", "details": err.Error()})
		return
	}
	if p.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	p.ID = id
	if err := h.repo.Update(c.Request.Context(), h.db, &p); err != nil {
		notFoundOrError(c, err, "project")
		return
	}
	c.JSON(http.StatusOK, p)
}

// ── DELETE /api/projects/:id ──

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), h.db, id); err != nil {
		notFoundOrError(c, err, "project")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func notFoundOrError(c *gin.Context, err error, resource string) {
	if isNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": resource + " not found"})
	} else {
		serverError(c, "failed to process "+resource, err)
	}
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") || strings.Contains(msg, "no rows in result set")
}
