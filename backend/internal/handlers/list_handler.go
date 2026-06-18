package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"formula-ai-system/backend/internal/models"
	"formula-ai-system/backend/internal/repository"
)

// FormulaListHandler serves the formula list query API.
type FormulaListHandler struct {
	repo *repository.FormulaRepo
	db   *sql.DB
}

// NewFormulaListHandler creates a new FormulaListHandler wired to the given database.
func NewFormulaListHandler(db *sql.DB) *FormulaListHandler {
	return &FormulaListHandler{
		repo: repository.NewFormulaRepo(),
		db:   db,
	}
}

// --- Response types (flat / restructured for frontend list) ---

type listFormulaResponse struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Code          string               `json:"code"`
	ComponentMode string               `json:"component_mode"`
	Status        string               `json:"status"`
	Parts         []listPartResponse   `json:"parts"`
	Steps         []listStepResponse   `json:"steps"`
}

type listPartResponse struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type listIngredientResponse struct {
	Material   string  `json:"material"`
	Percentage float64 `json:"percentage"`
	Weight     float64 `json:"weight"`
}

type listMaterialResponse struct {
	Category  string                    `json:"category"`
	Materials []listIngredientResponse  `json:"materials"`
}

type listStepResponse struct {
	StepNo         int                       `json:"step_no"`
	Name           string                    `json:"name"`
	Description    string                    `json:"description"`
	InstrumentName string                    `json:"instrument_name"`
	Temperature    string                    `json:"temperature"`
	Duration       string                    `json:"duration"`
	Materials      []listMaterialResponse    `json:"materials"`
	Parameters     []listStepParamResponse   `json:"parameters"`
}

type listStepParamResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// HandleList returns all formulas with a flat restructured view for the
// frontend list. Supports component_mode and formula_type query params.
//
//	GET /api/formulas/list?component_mode=single|double&formula_type=formula|material
func (h *FormulaListHandler) HandleList(c *gin.Context) {
	ctx := c.Request.Context()

	modeFilter := c.Query("component_mode")
	typeFilter := c.Query("formula_type")

	formulas, err := h.repo.List(ctx, h.db, repository.ListOptions{
		ComponentMode: modeFilter,
		FormulaType:   typeFilter,
	})
	if err != nil {
		serverError(c, "failed to load formula list", err)
		return
	}

	result := make([]listFormulaResponse, 0, len(formulas))
	for _, f := range formulas {
		result = append(result, restructureFormula(f))
	}

	c.JSON(http.StatusOK, gin.H{"formulas": result})
}

// restructureFormula converts a domain Formula into the list response shape:
// dosing_actions live under steps (resolved to ingredient_material) instead of
// under ingredients.
func restructureFormula(f *models.Formula) listFormulaResponse {
	// Build parts (simplified)
	parts := make([]listPartResponse, 0, len(f.Parts))
	for _, p := range f.Parts {
		parts = append(parts, listPartResponse{
			Name:      string(p.Name),
			SortOrder: p.SortOrder,
		})
	}

	// Build steps with materials and parameters
	steps := make([]listStepResponse, 0, len(f.Steps))
	for _, s := range f.Steps {
		mats := make([]listMaterialResponse, 0, len(s.Categories))
		for _, cat := range s.Categories {
			ings := make([]listIngredientResponse, 0, len(cat.Materials))
			for _, m := range cat.Materials {
				ings = append(ings, listIngredientResponse{
					Material:   m.Material,
					Percentage: m.Percentage,
					Weight:     m.Weight,
				})
			}
			mats = append(mats, listMaterialResponse{
				Category:  cat.Name,
				Materials: ings,
			})
		}

		params := make([]listStepParamResponse, 0, len(s.Parameters))
		for _, p := range s.Parameters {
			params = append(params, listStepParamResponse{
				Name:  p.Name,
				Value: p.Value,
				Unit:  p.Unit,
			})
		}

		steps = append(steps, listStepResponse{
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

	return listFormulaResponse{
		ID:            f.ID.String(),
		Name:          f.Name,
		Code:          f.Code,
		ComponentMode: string(f.ComponentMode),
		Status:        string(f.Status),
		Parts:         parts,
		Steps:         steps,
	}
}
