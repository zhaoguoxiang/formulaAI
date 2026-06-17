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
	matrixHandler := NewMatrixHandler(db)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.POST("/upload", HandleUpload)

		formulas := api.Group("/formulas")
		{
			formulas.POST("", formulaHandler.CreateFormula)
			formulas.GET("", formulaHandler.ListFormulas)
			formulas.GET("/matrix", matrixHandler.HandleMatrix)
			formulas.GET("/:id", formulaHandler.GetFormula)
			formulas.PUT("/:id", formulaHandler.UpdateFormula)
			formulas.DELETE("/:id", formulaHandler.DeleteFormula)
		}

		// ── Test Outline handlers ──
		testOutlineHandler := NewTestOutlineHandler(db)

		testOutlines := api.Group("/test-outlines")
		{
			testOutlines.POST("", testOutlineHandler.CreateTestOutline)
			testOutlines.GET("", testOutlineHandler.ListTestOutlines)
			testOutlines.GET("/:id", testOutlineHandler.GetTestOutline)
			testOutlines.PUT("/:id", testOutlineHandler.SaveVersion)
			testOutlines.PUT("/:id/archive", testOutlineHandler.ArchiveTestOutline)
			testOutlines.GET("/:id/versions", testOutlineHandler.ListVersions)
			testOutlines.PUT("/:id/activate", testOutlineHandler.ActivateVersion)
		}

		// ── Analysis handlers ──
		analysisHandler := NewAnalysisHandler(db)
		analysisHandler.RegisterRoutes(router)
	}

	// ── Static file serving for uploaded images ──
	router.Static("/uploads", "./uploads")
}

// serverError logs the error and returns a generic 500 response to avoid leaking internals.
func serverError(c *gin.Context, msg string, err error) {
	slog.Error(msg, "path", c.Request.URL.Path, "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
