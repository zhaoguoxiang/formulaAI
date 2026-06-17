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
	Name        string                    `json:"name"`
	MixRatio    float64                   `json:"mix_ratio"`
	Ingredients []matrixIngredientResponse `json:"ingredients"`
}

type matrixIngredientResponse struct {
	Material   string  `json:"material"`
	Percentage float64 `json:"percentage"`
	Weight     float64 `json:"weight"`
}

type matrixStepResponse struct {
	StepNo        int                          `json:"step_no"`
	Name          string                       `json:"name"`
	Temperature   string                       `json:"temperature"`
	Duration      string                       `json:"duration"`
	DosingActions []matrixDosingActionResponse `json:"dosing_actions"`
}

type matrixDosingActionResponse struct {
	IngredientMaterial string  `json:"ingredient_material"`
	DosingOrder        int     `json:"dosing_order"`
	UseRatio           float64 `json:"use_ratio"`
	DosingMethod       string  `json:"dosing_method"`
}

// HandleMatrix returns all formulas with a flat restructured view for the
// frontend matrix table. Dosing actions are moved from ingredients to steps,
// and ingredient IDs are resolved to material names.
//
//	GET /api/formulas/matrix?component_mode=single|double
func (h *MatrixHandler) HandleMatrix(c *gin.Context) {
	ctx := c.Request.Context()

	// Optional component_mode filter — pushed to SQL layer
	modeFilter := c.Query("component_mode")

	formulas, err := h.repo.List(ctx, h.db, repository.ListOptions{
		ComponentMode: modeFilter,
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
	// Build ingredient ID → material lookup
	ingMaterial := make(map[string]string)
	for _, p := range f.Parts {
		for _, ing := range p.Ingredients {
			ingMaterial[ing.ID.String()] = ing.Material
		}
	}

	// Build parts (without dosing_actions on ingredients)
	parts := make([]matrixPartResponse, 0, len(f.Parts))
	for _, p := range f.Parts {
		ings := make([]matrixIngredientResponse, 0, len(p.Ingredients))
		for _, ing := range p.Ingredients {
			ings = append(ings, matrixIngredientResponse{
				Material:   ing.Material,
				Percentage: ing.Percentage,
				Weight:     ing.Weight,
			})
		}
		parts = append(parts, matrixPartResponse{
			Name:        string(p.Name),
			MixRatio:    p.MixRatio,
			Ingredients: ings,
		})
	}

	// Build steps with dosing_actions resolved from all ingredients
	steps := make([]matrixStepResponse, 0, len(f.Steps))
	for _, s := range f.Steps {
		var das []matrixDosingActionResponse
		for _, p := range f.Parts {
			for _, ing := range p.Ingredients {
				for _, da := range ing.DosingActions {
					if da.StepID.String() == s.ID.String() {
						das = append(das, matrixDosingActionResponse{
							IngredientMaterial: ingMaterial[da.IngredientID.String()],
							DosingOrder:        da.DosingOrder,
							UseRatio:           da.UseRatio,
							DosingMethod:       da.DosingMethod,
						})
					}
				}
			}
		}
		// Ensure empty slice instead of null
		if das == nil {
			das = []matrixDosingActionResponse{}
		}
		steps = append(steps, matrixStepResponse{
			StepNo:        s.StepNo,
			Name:          s.Name,
			Temperature:   s.Temperature,
			Duration:      s.Duration,
			DosingActions: das,
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
