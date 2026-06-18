package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"formula-ai-system/backend/internal/models"
	"formula-ai-system/backend/internal/repository"
)

// MatrixHandler serves the formula matrix query API.
type MatrixHandler struct {
	repo *repository.FormulaRepo
	db   *sql.DB
}

// NewMatrixHandler creates a new MatrixHandler wired to the given database.
func NewMatrixHandler(db *sql.DB) *MatrixHandler {
	return &MatrixHandler{
		repo: repository.NewFormulaRepo(),
		db:   db,
	}
}

// --- Response types (flat / restructured for frontend matrix) ---

type matrixFormulaResponse struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Code          string                  `json:"code"`
	ComponentMode string                  `json:"component_mode"`
	Status        string                  `json:"status"`
	Parts         []matrixPartResponse    `json:"parts"`
	Steps         []matrixStepResponse    `json:"steps"`
}

type matrixPartResponse struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type matrixIngredientResponse struct {
	Material   string  `json:"material"`
	Percentage float64 `json:"percentage"`
	Weight     float64 `json:"weight"`
}

type matrixMaterialResponse struct {
	Category   string                    `json:"category"`
	Materials  []matrixIngredientResponse `json:"materials"`
}

type matrixStepResponse struct {
	StepNo         int                        `json:"step_no"`
	Name           string                     `json:"name"`
	Description    string                     `json:"description"`
	InstrumentName string                     `json:"instrument_name"`
	Temperature    string                     `json:"temperature"`
	Duration       string                     `json:"duration"`
	Materials      []matrixMaterialResponse   `json:"materials"`
	Parameters     []matrixStepParamResponse  `json:"parameters"`
}

type matrixStepParamResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// HandleMatrix returns all formulas with a flat restructured view for the
// frontend matrix table. Supports component_mode and formula_type query params.
//
//	GET /api/formulas/matrix?component_mode=single|double&formula_type=formula|material
func (h *MatrixHandler) HandleMatrix(c *gin.Context) {
	ctx := c.Request.Context()

	modeFilter := c.Query("component_mode")
	typeFilter := c.Query("formula_type")

	formulas, err := h.repo.List(ctx, h.db, repository.ListOptions{
		ComponentMode: modeFilter,
		FormulaType:   typeFilter,
	})
	if err != nil {
		serverError(c, "failed to load matrix", err)
		return
	}

	result := make([]matrixFormulaResponse, 0, len(formulas))
	for _, f := range formulas {
		result = append(result, restructureFormula(f))
	}

	c.JSON(http.StatusOK, gin.H{"formulas": result})
}

// restructureFormula converts a domain Formula into the matrix response shape:
// dosing_actions live under steps (resolved to ingredient_material) instead of
// under ingredients.
func restructureFormula(f *models.Formula) matrixFormulaResponse {
	// Build parts (simplified)
	parts := make([]matrixPartResponse, 0, len(f.Parts))
	for _, p := range f.Parts {
		parts = append(parts, matrixPartResponse{
			Name:      string(p.Name),
			SortOrder: p.SortOrder,
		})
	}

	// Build steps with materials and parameters
	steps := make([]matrixStepResponse, 0, len(f.Steps))
	for _, s := range f.Steps {
		mats := make([]matrixMaterialResponse, 0, len(s.Categories))
		for _, cat := range s.Categories {
			ings := make([]matrixIngredientResponse, 0, len(cat.Materials))
			for _, m := range cat.Materials {
				ings = append(ings, matrixIngredientResponse{
					Material:   m.Material,
					Percentage: m.Percentage,
					Weight:     m.Weight,
				})
			}
			mats = append(mats, matrixMaterialResponse{
				Category:  cat.Name,
				Materials: ings,
			})
		}

		params := make([]matrixStepParamResponse, 0, len(s.Parameters))
		for _, p := range s.Parameters {
			params = append(params, matrixStepParamResponse{
				Name:  p.Name,
				Value: p.Value,
				Unit:  p.Unit,
			})
		}

		steps = append(steps, matrixStepResponse{
			StepNo:         s.StepNo,
			Name:           s.Name,
			Description:    s.Description,
			InstrumentName: s.InstrumentName,
			Temperature:    s.Temperature,
			Duration:       s.Duration,
			Materials:      mats,
			Parameters:     params,
		})
	}

	return matrixFormulaResponse{
		ID:            f.ID.String(),
		Name:          f.Name,
		Code:          f.Code,
		ComponentMode: string(f.ComponentMode),
		Status:        string(f.Status),
		Parts:         parts,
		Steps:         steps,
	}
}
