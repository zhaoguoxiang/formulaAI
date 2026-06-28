package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AnalysisHandler provides HTTP handlers for statistical analysis endpoints.
// All handlers use SQL GROUP BY aggregation queries against the formula database.
type AnalysisHandler struct {
	db *sql.DB
}

// NewAnalysisHandler creates a new AnalysisHandler with the given database connection.
func NewAnalysisHandler(db *sql.DB) *AnalysisHandler {
	return &AnalysisHandler{db: db}
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
//	FROM formula_step_materials GROUP BY material ORDER BY count DESC
func (h *AnalysisHandler) IngredientDistribution(c *gin.Context) {
	projectID := GetProjectID(c)
	rows, err := h.db.Query(
		`SELECT m.material, COUNT(*) as count, AVG(m.percentage) as avg_percentage
		 FROM formula_step_materials m
		 JOIN formula_step_material_categories c ON c.id = m.category_id
		 JOIN formula_steps s ON s.id = c.step_id
		 JOIN formulas f ON f.id = s.formula_id
		 WHERE f.project_id = $1
		 GROUP BY m.material
		 ORDER BY count DESC`, projectID,
	)
	if err != nil {
		serverError(c, "analysis query failed", err)
		return
	}
	defer rows.Close()

	var results []IngredientDistributionItem
	for rows.Next() {
		var item IngredientDistributionItem
		if err := rows.Scan(&item.Material, &item.Count, &item.AvgPercentage); err != nil {
			serverError(c, "analysis query failed", err)
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		serverError(c, "analysis query failed", err)
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
	projectID := GetProjectID(c)
	rows, err := h.db.Query(
		`SELECT component_mode,
		        COUNT(*) as count,
		        ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
		 FROM formulas
		 WHERE project_id = $1
		 GROUP BY component_mode
		 ORDER BY count DESC`, projectID,
	)
	if err != nil {
		serverError(c, "analysis query failed", err)
		return
	}
	defer rows.Close()

	var results []ComponentModeRatioItem
	for rows.Next() {
		var item ComponentModeRatioItem
		if err := rows.Scan(&item.Mode, &item.Count, &item.Percentage); err != nil {
			serverError(c, "analysis query failed", err)
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		serverError(c, "analysis query failed", err)
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
	projectID := GetProjectID(c)
	rows, err := h.db.Query(
		`WITH step_counts AS (
		     SELECT s.formula_id, COUNT(*) as cnt
		     FROM formula_steps s
		     JOIN formulas f ON f.id = s.formula_id
		     WHERE f.project_id = $1
		     GROUP BY s.formula_id
		 )
		 SELECT CASE
		          WHEN cnt BETWEEN 1 AND 2 THEN '1-2'
		          WHEN cnt BETWEEN 3 AND 5 THEN '3-5'
		          ELSE '5+'
		        END as range,
		        COUNT(*) as count
		 FROM step_counts
		 GROUP BY range
		 ORDER BY range`, projectID,
	)
	if err != nil {
		serverError(c, "analysis query failed", err)
		return
	}
	defer rows.Close()

	var results []StepCountDistributionItem
	for rows.Next() {
		var item StepCountDistributionItem
		if err := rows.Scan(&item.Range, &item.Count); err != nil {
			serverError(c, "analysis query failed", err)
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		serverError(c, "analysis query failed", err)
		return
	}

	if results == nil {
		results = []StepCountDistributionItem{}
	}

	c.JSON(http.StatusOK, results)
}

// CategoryMaterialCountItem represents one row in the category material count result.
type CategoryMaterialCountItem struct {
	CategoryName string `json:"category_name"`
	MaterialCount int   `json:"material_count"`
}

// CategoryMaterialCount returns the distribution of materials per category across steps.
//
// GET /api/analysis/dosing-method-stats
//
// SQL: SELECT c.name, COUNT(m.id) FROM formula_step_material_categories c
//
//	JOIN formula_step_materials m ON m.category_id = c.id
//	GROUP BY c.name ORDER BY material_count DESC
func (h *AnalysisHandler) DosingMethodStats(c *gin.Context) {
	projectID := GetProjectID(c)
	rows, err := h.db.Query(
		`SELECT c.name, COUNT(m.id) as material_count
		 FROM formula_step_material_categories c
		 JOIN formula_step_materials m ON m.category_id = c.id
		 JOIN formula_steps s ON s.id = c.step_id
		 JOIN formulas f ON f.id = s.formula_id
		 WHERE f.project_id = $1
		 GROUP BY c.name
		 ORDER BY material_count DESC`, projectID,
	)
	if err != nil {
		serverError(c, "analysis query failed", err)
		return
	}
	defer rows.Close()

	var results []CategoryMaterialCountItem
	for rows.Next() {
		var item CategoryMaterialCountItem
		if err := rows.Scan(&item.CategoryName, &item.MaterialCount); err != nil {
			serverError(c, "analysis query failed", err)
			return
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		serverError(c, "analysis query failed", err)
		return
	}

	if results == nil {
		results = []CategoryMaterialCountItem{}
	}

	c.JSON(http.StatusOK, results)
}
