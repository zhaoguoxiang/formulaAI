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
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"formula-ai-system/backend/internal/models"
	"formula-ai-system/backend/internal/repository"
)



// seedTestData inserts known formulas with ingredients, steps, and dosing actions
// so aggregation queries return predictable results.
func seedTestData(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	formulaRepo := repository.NewFormulaRepo()

	// Formula 1: single mode, 3 steps, ingredients: Water (30%), Resin (70%)
	f1 := makeTestFormula("F-001", "single", []testIngredient{
		{Material: "Water", Percentage: 30},
		{Material: "Resin", Percentage: 70},
	}, 3, []testDosing{
		{Method: "spray"},
		{Method: "spray"},
	})

	// Formula 2: double mode, 4 steps, ingredients: Water (40%), Resin (60%)
	f2 := makeTestFormula("F-002", "double", []testIngredient{
		{Material: "Water", Percentage: 40},
		{Material: "Resin", Percentage: 60},
	}, 4, []testDosing{
		{Method: "pour"},
		{Method: "spray"},
	})

	// Formula 3: single mode, 1 step, ingredient: Pigment (100%)
	f3 := makeTestFormula("F-003", "single", []testIngredient{
		{Material: "Pigment", Percentage: 100},
	}, 1, []testDosing{
		{Method: "spray"},
	})

	for _, f := range []*models.Formula{f1, f2, f3} {
		if err := formulaRepo.Create(ctx, db, f); err != nil {
			t.Fatalf("seed formula %s: %v", f.Code, err)
		}
	}
}

type testIngredient struct {
	Material   string
	Percentage float64
}

type testDosing struct {
	Method string
}

func makeTestFormula(code, componentMode string, ingredients []testIngredient, stepCount int, dosings []testDosing) *models.Formula {
	partID := uuid.New()
	part := models.FormulaPart{
		ID:        partID,
		Name:      models.PartMain,
		MixRatio:  100,
		SortOrder: 1,
	}

	for i, ing := range ingredients {
		ingID := uuid.New()
		ingredient := models.FormulaIngredient{
			ID:         ingID,
			PartID:     partID,
			Material:   ing.Material,
			Percentage: ing.Percentage,
			SortOrder:  i + 1,
		}
		// Attach dosing actions to first ingredient for simplicity
		if i == 0 {
			for j, d := range dosings {
				da := models.FormulaDosingAction{
					ID:           uuid.New(),
					StepID:       uuid.New(), // will be overwritten after step creation
					DosingOrder:  j + 1,
					UseRatio:     100,
					DosingMethod: d.Method,
				}
				ingredient.DosingActions = append(ingredient.DosingActions, da)
			}
		}
		part.Ingredients = append(part.Ingredients, ingredient)
	}

	steps := make([]models.FormulaStep, stepCount)
	for i := 0; i < stepCount; i++ {
		stepID := uuid.New()
		steps[i] = models.FormulaStep{
			ID:     stepID,
			StepNo: i + 1,
			Name:   "Step " + string(rune('A'+i)),
		}
	}

	// Link dosing action step IDs to actual step UUIDs
	if len(dosings) > 0 && len(part.Ingredients) > 0 && len(part.Ingredients[0].DosingActions) > 0 {
		for k := range part.Ingredients[0].DosingActions {
			if k < len(steps) {
				part.Ingredients[0].DosingActions[k].StepID = steps[k].ID
			}
		}
	}

	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Test " + code,
		Code:          code,
		ComponentMode: models.ComponentMode(componentMode),
		Status:        models.StatusActive,
		Parts:         []models.FormulaPart{part},
		Steps:         steps,
	}
}

func performRequest(t *testing.T, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	handler(c)
	return w
}

func TestIngredientDistribution(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)
	seedTestData(t, db)

	h := NewAnalysisHandler(db)
	w := performRequest(t, h.IngredientDistribution)

	assert.Equal(t, http.StatusOK, w.Code)

	var results []IngredientDistributionItem
	err := json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 3) // Water, Resin, Pigment

	// Build a map for easy assertion
	m := make(map[string]IngredientDistributionItem)
	for _, r := range results {
		m[r.Material] = r
	}

	// Water: 2 occurrences, avg (30+40)/2 = 35
	assert.Equal(t, 2, m["Water"].Count)
	assert.InDelta(t, 35.0, m["Water"].AvgPercentage, 0.01)

	// Resin: 2 occurrences, avg (70+60)/2 = 65
	assert.Equal(t, 2, m["Resin"].Count)
	assert.InDelta(t, 65.0, m["Resin"].AvgPercentage, 0.01)

	// Pigment: 1 occurrence, avg 100
	assert.Equal(t, 1, m["Pigment"].Count)
	assert.InDelta(t, 100.0, m["Pigment"].AvgPercentage, 0.01)
}

func TestComponentModeRatio(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)
	seedTestData(t, db)

	h := NewAnalysisHandler(db)
	w := performRequest(t, h.ComponentModeRatio)

	assert.Equal(t, http.StatusOK, w.Code)

	var results []ComponentModeRatioItem
	err := json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	m := make(map[string]ComponentModeRatioItem)
	for _, r := range results {
		m[r.Mode] = r
	}

	// single: 2 out of 3 = 66.67%
	assert.Equal(t, 2, m["single"].Count)
	assert.InDelta(t, 66.67, m["single"].Percentage, 0.1)

	// double: 1 out of 3 = 33.33%
	assert.Equal(t, 1, m["double"].Count)
	assert.InDelta(t, 33.33, m["double"].Percentage, 0.1)
}

func TestStepCountDistribution(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)
	seedTestData(t, db)

	h := NewAnalysisHandler(db)
	w := performRequest(t, h.StepCountDistribution)

	assert.Equal(t, http.StatusOK, w.Code)

	var results []StepCountDistributionItem
	err := json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)
	// 1-2: F-003(1 step), 3-5: F-001(3 steps) + F-002(4 steps)
	// 5+: none, so only 2 buckets appear in results
	assert.Len(t, results, 2)

	m := make(map[string]int)
	for _, r := range results {
		m[r.Range] = r.Count
	}

	assert.Equal(t, 1, m["1-2"])
	assert.Equal(t, 2, m["3-5"])
}

func TestDosingMethodStats(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	truncateAllTables(t, db)
	seedTestData(t, db)

	h := NewAnalysisHandler(db)
	w := performRequest(t, h.DosingMethodStats)

	assert.Equal(t, http.StatusOK, w.Code)

	var results []DosingMethodStatsItem
	err := json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	m := make(map[string]int)
	for _, r := range results {
		m[r.Method] = r.Count
	}

	// spray: F-001(2) + F-002(1) + F-003(1) = 4
	assert.Equal(t, 4, m["spray"])
	// pour: F-002(1) = 1
	assert.Equal(t, 1, m["pour"])
}
