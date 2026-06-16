package handlers

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes configures all API routes, CORS middleware, and error recovery
// on the given Gin engine using the provided database connection.
func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	// ── CORS middleware ──
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── Panic recovery middleware ──
	router.Use(gin.Recovery())

	// ── Formula handlers ──
	formulaHandler := NewFormulaHandler(db)
	matrixHandler := NewMatrixHandler(db)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		formulas := api.Group("/formulas")
		{
			formulas.POST("", formulaHandler.CreateFormula)
			formulas.GET("", formulaHandler.ListFormulas)
			formulas.GET("/matrix", matrixHandler.HandleMatrix)
			formulas.GET("/:id", formulaHandler.GetFormula)
			formulas.PUT("/:id", formulaHandler.UpdateFormula)
			formulas.DELETE("/:id", formulaHandler.DeleteFormula)
		}

		// ── Analysis handlers ──
		analysisHandler := NewAnalysisHandler(db)
		analysisHandler.RegisterRoutes(router)
	}
}
