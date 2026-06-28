package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// contextKey is the key type used for storing project_id in gin.Context.
const ProjectIDKey = "project_id"

// ProjectMiddleware extracts X-Project-Id from the request header and validates
// it as a UUID. The project_id is stored in the gin context for downstream handlers.
//
// Routes that don't require a project context (e.g. /api/projects, /api/health)
// should register before this middleware.
func ProjectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("X-Project-Id")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "X-Project-Id header is required",
			})
			return
		}

		projectID, err := uuid.Parse(header)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "X-Project-Id must be a valid UUID",
				"details": err.Error(),
			})
			return
		}

		c.Set(ProjectIDKey, projectID)
		c.Next()
	}
}

// GetProjectID extracts the project UUID from the gin context.
// Returns uuid.Nil if not present (should only happen on non-project routes).
func GetProjectID(c *gin.Context) uuid.UUID {
	id, exists := c.Get(ProjectIDKey)
	if !exists {
		return uuid.Nil
	}
	return id.(uuid.UUID)
}
