package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes configures all API routes, CORS middleware, and error recovery
// on the given Gin engine using the provided database connection.
func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	// ── CORS middleware (configurable via CORS_ORIGINS env var) ──
	origins := strings.FieldsFunc(os.Getenv("CORS_ORIGINS"), func(r rune) bool { return r == ',' })
	if len(origins) == 0 {
		origins = []string{"http://localhost:4200"}
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── Formula handlers ──
	formulaHandler := NewFormulaHandler(db)
	listHandler := NewFormulaListHandler(db)
	testOutlineHandler := NewTestOutlineHandler(db)
	analysisHandler := NewAnalysisHandler(db)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.POST("/upload", HandleUpload)

		// ── Project routes (no project_id header required) ──
		projectHandler := NewProjectHandler(db)
		projects := api.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.ListProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
		}

		// ── Scoped routes (X-Project-Id header required) ──
		scoped := api.Group("")
		scoped.Use(ProjectMiddleware())
		{
			formulas := scoped.Group("/formulas")
			{
				formulas.POST("", formulaHandler.CreateFormula)
				formulas.GET("", formulaHandler.ListFormulas)
				formulas.GET("/list", listHandler.HandleList)
				formulas.GET("/:id", formulaHandler.GetFormula)
				formulas.PUT("/:id", formulaHandler.UpdateFormula)
				formulas.DELETE("/:id", formulaHandler.DeleteFormula)
			}

			testOutlines := scoped.Group("/test-outlines")
			{
				testOutlines.POST("", testOutlineHandler.CreateTestOutline)
				testOutlines.GET("", testOutlineHandler.ListTestOutlines)
				testOutlines.GET("/:id", testOutlineHandler.GetTestOutline)
				testOutlines.PUT("/:id", testOutlineHandler.SaveVersion)
				testOutlines.PUT("/:id/archive", testOutlineHandler.ArchiveTestOutline)
				testOutlines.GET("/:id/versions", testOutlineHandler.ListVersions)
				testOutlines.PUT("/:id/activate", testOutlineHandler.ActivateVersion)
			}

			analysis := scoped.Group("/analysis")
			{
				analysis.GET("/ingredient-distribution", analysisHandler.IngredientDistribution)
				analysis.GET("/component-mode-ratio", analysisHandler.ComponentModeRatio)
				analysis.GET("/step-count-distribution", analysisHandler.StepCountDistribution)
				analysis.GET("/dosing-method-stats", analysisHandler.DosingMethodStats)
			}
		}
	}

	// ── Static file serving for uploaded images ──
	router.Static("/uploads", "./uploads")
}

// serverError logs the error and returns a generic 500 response to avoid leaking internals.
func serverError(c *gin.Context, msg string, err error) {
	slog.Error(msg, "path", c.Request.URL.Path, "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
