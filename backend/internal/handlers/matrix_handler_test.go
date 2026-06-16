package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"formula-ai-system/backend/internal/models"
	"formula-ai-system/backend/internal/repository"
)

// makeFormulaWithDosing creates a formula with dosing actions for testing.
func makeFormulaWithDosing() *models.Formula {
	ing1ID := uuid.New()
	ing2ID := uuid.New()
	step1ID := uuid.New()
	step2ID := uuid.New()

	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Test Formula",
		Code:          "TF-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts: []models.FormulaPart{
			{
				ID: uuid.New(), Name: models.PartMain, MixRatio: 100, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{
						ID: ing1ID, Material: "Resin", Percentage: 60, SortOrder: 1,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step1ID, IngredientID: ing1ID, DosingOrder: 1, UseRatio: 60, DosingMethod: "pump"},
							{ID: uuid.New(), StepID: step2ID, IngredientID: ing1ID, DosingOrder: 2, UseRatio: 40, DosingMethod: "manual"},
						},
					},
					{
						ID: ing2ID, Material: "Hardener", Percentage: 40, SortOrder: 2,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step1ID, IngredientID: ing2ID, DosingOrder: 1, UseRatio: 40, DosingMethod: "pour"},
						},
					},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: step1ID, StepNo: 1, Name: "Mixing", Temperature: "25C", Duration: "10min"},
			{ID: step2ID, StepNo: 2, Name: "Curing", Temperature: "60C", Duration: "2h"},
		},
	}
}

// --- Unit test: restructureFormula ---

func TestRestructureFormula_MovesDosingToSteps(t *testing.T) {
	f := makeFormulaWithDosing()
	result := restructureFormula(f)

	assert.Equal(t, f.ID.String(), result.ID)
	assert.Equal(t, f.Name, result.Name)
	assert.Equal(t, f.Code, result.Code)
	assert.Equal(t, "single", result.ComponentMode)
	assert.Equal(t, "draft", result.Status)

	// Parts: 1 part with 2 ingredients, no dosing_actions on ingredients
	require.Len(t, result.Parts, 1)
	assert.Equal(t, "PartMain", result.Parts[0].Name)
	assert.Equal(t, float64(100), result.Parts[0].MixRatio)
	require.Len(t, result.Parts[0].Ingredients, 2)
	assert.Equal(t, "Resin", result.Parts[0].Ingredients[0].Material)
	assert.Equal(t, float64(60), result.Parts[0].Ingredients[0].Percentage)

	// Steps: 2 steps, dosing_actions moved here
	require.Len(t, result.Steps, 2)
	assert.Equal(t, 1, result.Steps[0].StepNo)
	assert.Equal(t, "Mixing", result.Steps[0].Name)
	assert.Equal(t, "25C", result.Steps[0].Temperature)
	assert.Equal(t, "10min", result.Steps[0].Duration)

	// Step 1 has 2 dosing actions (Resin + Hardener)
	require.Len(t, result.Steps[0].DosingActions, 2)
	assert.Equal(t, "Resin", result.Steps[0].DosingActions[0].IngredientMaterial)
	assert.Equal(t, 1, result.Steps[0].DosingActions[0].DosingOrder)
	assert.Equal(t, float64(60), result.Steps[0].DosingActions[0].UseRatio)
	assert.Equal(t, "pump", result.Steps[0].DosingActions[0].DosingMethod)
	assert.Equal(t, "Hardener", result.Steps[0].DosingActions[1].IngredientMaterial)

	// Step 2 has 1 dosing action (Resin only)
	require.Len(t, result.Steps[1].DosingActions, 1)
	assert.Equal(t, "Resin", result.Steps[1].DosingActions[0].IngredientMaterial)
	assert.Equal(t, 2, result.Steps[1].DosingActions[0].DosingOrder)
	assert.Equal(t, float64(40), result.Steps[1].DosingActions[0].UseRatio)
	assert.Equal(t, "manual", result.Steps[1].DosingActions[0].DosingMethod)
}

func TestRestructureFormula_EmptyCollections(t *testing.T) {
	f := &models.Formula{
		ID:            uuid.New(),
		Name:          "Empty",
		Code:          "E-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts:         []models.FormulaPart{},
		Steps:         []models.FormulaStep{},
	}

	result := restructureFormula(f)
	assert.Len(t, result.Parts, 0)
	assert.Len(t, result.Steps, 0)
}

func TestRestructureFormula_StepWithNoDosing(t *testing.T) {
	stepID := uuid.New()
	f := &models.Formula{
		ID:            uuid.New(),
		Name:          "No Dosing",
		Code:          "ND-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts: []models.FormulaPart{
			{
				ID: uuid.New(), Name: models.PartMain, MixRatio: 100, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{ID: uuid.New(), Material: "Resin", Percentage: 100, SortOrder: 1},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: stepID, StepNo: 1, Name: "Mix", Temperature: "25C", Duration: "5min"},
		},
	}

	result := restructureFormula(f)
	require.Len(t, result.Steps, 1)
	assert.Len(t, result.Steps[0].DosingActions, 0)
}

// --- Integration test: HandleMatrix ---

func setupMatrixTestRouter(t *testing.T, db *sql.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewMatrixHandler(db)
	router.GET("/api/formulas/matrix", handler.HandleMatrix)
	return router
}

func TestHandleMatrix_ReturnsFormulas(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)

	repo := repository.NewFormulaRepo()

	// Seed a formula with dosing actions
	f := makeFormulaWithDosing()
	err := repo.Create(context.Background(), db, f)
	require.NoError(t, err)

	router := setupMatrixTestRouter(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/matrix", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Formulas []matrixFormulaResponse `json:"formulas"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Len(t, resp.Formulas, 1)
	assert.Equal(t, f.Name, resp.Formulas[0].Name)
	assert.Equal(t, f.Code, resp.Formulas[0].Code)
	require.Len(t, resp.Formulas[0].Steps, 2)
	require.Len(t, resp.Formulas[0].Steps[0].DosingActions, 2)
}

func TestHandleMatrix_FilterByComponentMode(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)

	repo := repository.NewFormulaRepo()

	// Seed a single-mode formula
	single := &models.Formula{
		ID:            uuid.New(),
		Name:          "Single Mode",
		Code:          "SM-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts: []models.FormulaPart{
			{ID: uuid.New(), Name: models.PartMain, MixRatio: 100, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{ID: uuid.New(), Material: "A", Percentage: 100, SortOrder: 1},
				}},
		},
		Steps: []models.FormulaStep{
			{ID: uuid.New(), StepNo: 1, Name: "Mix"},
		},
	}
	err := repo.Create(context.Background(), db, single)
	require.NoError(t, err)

	// Seed a double-mode formula
	double := &models.Formula{
		ID:            uuid.New(),
		Name:          "Double Mode",
		Code:          "DM-001",
		ComponentMode: models.ComponentModeDouble,
		Status:        models.StatusActive,
		Parts: []models.FormulaPart{
			{ID: uuid.New(), Name: models.PartA, MixRatio: 50, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{ID: uuid.New(), Material: "X", Percentage: 100, SortOrder: 1},
				}},
		},
		Steps: []models.FormulaStep{
			{ID: uuid.New(), StepNo: 1, Name: "Combine"},
		},
	}
	err = repo.Create(context.Background(), db, double)
	require.NoError(t, err)

	router := setupMatrixTestRouter(t, db)

	// Filter single
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/matrix?component_mode=single", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Formulas []matrixFormulaResponse `json:"formulas"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	require.Len(t, resp.Formulas, 1)
	assert.Equal(t, "Single Mode", resp.Formulas[0].Name)

	// Filter double
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/formulas/matrix?component_mode=double", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &resp)
	require.Len(t, resp.Formulas, 1)
	assert.Equal(t, "Double Mode", resp.Formulas[0].Name)

	// No filter: both
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/formulas/matrix", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Formulas, 2)
}

func TestHandleMatrix_EmptyDatabase(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)

	router := setupMatrixTestRouter(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/matrix", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Formulas []matrixFormulaResponse `json:"formulas"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Formulas, 0)
}
