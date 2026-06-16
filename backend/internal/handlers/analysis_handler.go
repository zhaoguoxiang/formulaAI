package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AnalysisHandler provides HTTP handlers for statistical analysis endpoints.
// All handlers use SQL GROUP BY aggregation queries against the formula database.
type AnalysisHandler struct {
	DB *sql.DB
}

// NewAnalysisHandler creates a new AnalysisHandler with the given database connection.
func NewAnalysisHandler(db *sql.DB) *AnalysisHandler {
	return &AnalysisHandler{DB: db}
}

// IngredientDistributionItem represents one row in the ingredient distribution result.
type IngredientDistributionItem struct {
	Material      string  `json:"material"`
	Count         int     `json:"count"`
	AvgPercentage float64 `json:"avg_percentage"`
}

// IngredientDistribution returns material usage statistics across all formulas.
//
// GET /api/analysis/ingredient-distribution
//
// SQL: SELECT material, COUNT(*) as count, AVG(percentage) as avg_percentage
//
//	FROM formula_ingredients GROUP BY material ORDER BY count DESC
func (h *AnalysisHandler) IngredientDistribution(c *gin.Context) {
	rows, err := h.DB.Query(
		`SELECT material, COUNT(*) as count, AVG(percentage) as avg_percentage
		 FROM formula_ingredients
		 GROUP BY material
		 ORDER BY count DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []IngredientDistributionItem
	for rows.Next() {
		var item IngredientDistributionItem
		if err := rows.Scan(&item.Material, &item.Count, &item.AvgPercentage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []IngredientDistributionItem{}
	}

	c.JSON(http.StatusOK, results)
}

// ComponentModeRatioItem represents one row in the component mode ratio result.
type ComponentModeRatioItem struct {
	Mode       string  `json:"mode"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ComponentModeRatio returns the distribution of component_mode (single vs double) across formulas.
//
// GET /api/analysis/component-mode-ratio
//
// SQL: SELECT component_mode, COUNT(*), ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2)
//
//	FROM formulas GROUP BY component_mode
func (h *AnalysisHandler) ComponentModeRatio(c *gin.Context) {
	rows, err := h.DB.Query(
		`SELECT component_mode,
		        COUNT(*) as count,
		        ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
		 FROM formulas
		 GROUP BY component_mode
		 ORDER BY count DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []ComponentModeRatioItem
	for rows.Next() {
		var item ComponentModeRatioItem
		if err := rows.Scan(&item.Mode, &item.Count, &item.Percentage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []ComponentModeRatioItem{}
	}

	c.JSON(http.StatusOK, results)
}

// StepCountDistributionItem represents one row in the step count distribution result.
type StepCountDistributionItem struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// StepCountDistribution returns the distribution of step counts per formula,
// bucketed into ranges: "1-2", "3-5", "5+".
//
// GET /api/analysis/step-count-distribution
//
// SQL: Uses a CTE to count steps per formula, then buckets the counts.
func (h *AnalysisHandler) StepCountDistribution(c *gin.Context) {
	rows, err := h.DB.Query(
		`WITH step_counts AS (
		     SELECT formula_id, COUNT(*) as cnt
		     FROM formula_steps
		     GROUP BY formula_id
		 )
		 SELECT CASE
		          WHEN cnt BETWEEN 1 AND 2 THEN '1-2'
		          WHEN cnt BETWEEN 3 AND 5 THEN '3-5'
		          ELSE '5+'
		        END as range,
		        COUNT(*) as count
		 FROM step_counts
		 GROUP BY range
		 ORDER BY range`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []StepCountDistributionItem
	for rows.Next() {
		var item StepCountDistributionItem
		if err := rows.Scan(&item.Range, &item.Count); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []StepCountDistributionItem{}
	}

	c.JSON(http.StatusOK, results)
}

// DosingMethodStatsItem represents one row in the dosing method stats result.
type DosingMethodStatsItem struct {
	Method string `json:"method"`
	Count  int    `json:"count"`
}

// DosingMethodStats returns the distribution of dosing methods used across all formulas.
//
// GET /api/analysis/dosing-method-stats
//
// SQL: SELECT dosing_method, COUNT(*) FROM formula_dosing_actions
//
//	WHERE dosing_method IS NOT NULL GROUP BY dosing_method ORDER BY count DESC
func (h *AnalysisHandler) DosingMethodStats(c *gin.Context) {
	rows, err := h.DB.Query(
		`SELECT dosing_method, COUNT(*) as count
		 FROM formula_dosing_actions
		 WHERE dosing_method IS NOT NULL AND dosing_method != ''
		 GROUP BY dosing_method
		 ORDER BY count DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []DosingMethodStatsItem
	for rows.Next() {
		var item DosingMethodStatsItem
		if err := rows.Scan(&item.Method, &item.Count); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []DosingMethodStatsItem{}
	}

	c.JSON(http.StatusOK, results)
}

// RegisterRoutes registers all analysis endpoints on the given Gin engine.
func (h *AnalysisHandler) RegisterRoutes(r *gin.Engine) {
	analysis := r.Group("/api/analysis")
	{
		analysis.GET("/ingredient-distribution", h.IngredientDistribution)
		analysis.GET("/component-mode-ratio", h.ComponentModeRatio)
		analysis.GET("/step-count-distribution", h.StepCountDistribution)
		analysis.GET("/dosing-method-stats", h.DosingMethodStats)
	}
}
