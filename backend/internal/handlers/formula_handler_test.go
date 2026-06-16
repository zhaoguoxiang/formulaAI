package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"formula-ai-system/backend/internal/config"
	"formula-ai-system/backend/internal/models"
)

// ── Test setup ──────────────────────────────────────────────────────────────

func setupTestRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	t.Helper()

	cfg := &config.Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "formula"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "changeme"),
		DBName:     getEnvOrDefault("DB_NAME", "formula_ai"),
	}

	db, err := sql.Open("postgres", cfg.DSN())
	require.NoError(t, err, "failed to open database connection")
	require.NoError(t, db.Ping(), "failed to ping database - is PostgreSQL running?")

	truncateAllTables(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	formulaHandler := NewFormulaHandler(db)
	formulas := router.Group("/api/formulas")
	{
		formulas.POST("", formulaHandler.CreateFormula)
		formulas.GET("", formulaHandler.ListFormulas)
		formulas.GET("/:id", formulaHandler.GetFormula)
		formulas.PUT("/:id", formulaHandler.UpdateFormula)
		formulas.DELETE("/:id", formulaHandler.DeleteFormula)
	}

	return router, db
}

// ── Test fixtures ───────────────────────────────────────────────────────────

func validCreateRequest() map[string]interface{} {
	stepID := uuid.New().String()
	return map[string]interface{}{
		"name":           "Test Formula",
		"code":           "TF-001",
		"component_mode": "single",
		"status":         "draft",
		"parts": []map[string]interface{}{
			{
				"name":      "PartMain",
				"mix_ratio": 100,
				"sort_order": 1,
				"ingredients": []map[string]interface{}{
					{
						"material":   "Resin",
						"percentage": 60,
						"sort_order": 1,
						"dosing_actions": []map[string]interface{}{
							{
								"step_id":       stepID,
								"dosing_order":  1,
								"use_ratio":     100,
								"dosing_method": "pump",
							},
						},
					},
					{
						"material":   "Hardener",
						"percentage": 40,
						"sort_order": 2,
						"dosing_actions": []map[string]interface{}{
							{
								"step_id":       stepID,
								"dosing_order":  1,
								"use_ratio":     100,
								"dosing_method": "pour",
							},
						},
					},
				},
			},
		},
		"steps": []map[string]interface{}{
			{
				"id":          stepID,
				"step_no":     1,
				"name":        "Mixing",
				"temperature": "25C",
				"duration":    "10min",
			},
		},
	}
}

func createFormula(t *testing.T, router *gin.Engine) models.Formula {
	t.Helper()
	body, _ := json.Marshal(validCreateRequest())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "create failed: %s", w.Body.String())

	var f models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &f))
	return f
}

// ── POST /api/formulas ──────────────────────────────────────────────────────

func TestCreateFormula_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	body, _ := json.Marshal(validCreateRequest())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var f models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &f))
	assert.NotEqual(t, uuid.Nil, f.ID)
	assert.Equal(t, "Test Formula", f.Name)
	assert.Equal(t, models.ComponentModeSingle, f.ComponentMode)
	assert.Len(t, f.Parts, 1)
	assert.Len(t, f.Steps, 1)
}

func TestCreateFormula_InvalidJSON(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid JSON body", resp["error"])
}

func TestCreateFormula_EmptyBody(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "validation failed", resp["error"])
}

func TestCreateFormula_ValidationErrors(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// Formula with wrong parts for single mode
	payload := map[string]interface{}{
		"name":           "Bad Formula",
		"code":           "BF-001",
		"component_mode": "single",
		"status":         "draft",
		"parts": []map[string]interface{}{
			{"name": "PartA", "mix_ratio": 50, "sort_order": 1, "ingredients": []map[string]interface{}{}},
		},
		"steps": []map[string]interface{}{},
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	errors, ok := resp["errors"].([]interface{})
	assert.True(t, ok, "expected errors array")
	assert.NotEmpty(t, errors)
}

// ── GET /api/formulas ───────────────────────────────────────────────────────

func TestListFormulas_Empty(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var formulas []models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &formulas))
	assert.Empty(t, formulas)
}

func TestListFormulas_WithData(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	createFormula(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var formulas []models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &formulas))
	assert.Len(t, formulas, 1)
	assert.Equal(t, "Test Formula", formulas[0].Name)
}

// ── GET /api/formulas/:id ───────────────────────────────────────────────────

func TestGetFormula_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	f := createFormula(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/"+f.ID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var retrieved models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &retrieved))
	assert.Equal(t, f.ID, retrieved.ID)
	assert.Equal(t, f.Name, retrieved.Name)
}

func TestGetFormula_NotFound(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "formula not found", resp["error"])
}

func TestGetFormula_InvalidUUID(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid formula ID", resp["error"])
}

// ── PUT /api/formulas/:id ───────────────────────────────────────────────────

func TestUpdateFormula_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	f := createFormula(t, router)
	stepID := uuid.New().String()

	updatePayload := map[string]interface{}{
		"name":           "Updated Formula",
		"code":           "UF-001",
		"component_mode": "single",
		"status":         "draft",
		"parts": []map[string]interface{}{
			{
				"name":      "PartMain",
				"mix_ratio": 100,
				"sort_order": 1,
				"ingredients": []map[string]interface{}{
					{
						"material":   "Updated Resin",
						"percentage": 100,
						"sort_order": 1,
						"dosing_actions": []map[string]interface{}{
							{"step_id": stepID, "dosing_order": 1, "use_ratio": 100, "dosing_method": "pump"},
						},
					},
				},
			},
		},
		"steps": []map[string]interface{}{
			{"id": stepID, "step_no": 1, "name": "New Step", "temperature": "30C", "duration": "5min"},
		},
	}
	body, _ := json.Marshal(updatePayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/formulas/"+f.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, f.ID, updated.ID)
	assert.Equal(t, "Updated Formula", updated.Name)
	assert.Len(t, updated.Steps, 1)
	assert.Equal(t, "New Step", updated.Steps[0].Name)
}

func TestUpdateFormula_NotFound(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	body, _ := json.Marshal(validCreateRequest())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/formulas/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateFormula_InvalidUUID(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	body, _ := json.Marshal(validCreateRequest())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/formulas/not-a-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid formula ID", resp["error"])
}

func TestUpdateFormula_InvalidJSON(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	f := createFormula(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/formulas/"+f.ID.String(), strings.NewReader("bad json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid JSON body", resp["error"])
}

// ── DELETE /api/formulas/:id ────────────────────────────────────────────────

func TestDeleteFormula_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	f := createFormula(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/formulas/"+f.ID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify it's gone
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/formulas/"+f.ID.String(), nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestDeleteFormula_NotFound(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/formulas/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteFormula_InvalidUUID(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/formulas/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid formula ID", resp["error"])
}

// ── Integration: workflow with validation warnings only ─────────────────────

func TestCreateFormula_WarningsOnly_Proceeds(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	stepID := uuid.New().String()
	payload := map[string]interface{}{
		"name":           "Warning Only",
		"code":           "WO-001",
		"component_mode": "single",
		"status":         "draft",
		"parts": []map[string]interface{}{
			{
				"name":      "PartMain",
				"mix_ratio": 100,
				"sort_order": 1,
				"ingredients": []map[string]interface{}{
					{
						"material":   "Material A",
						"percentage": 100,
						"sort_order": 1,
						"dosing_actions": []map[string]interface{}{
							{"step_id": stepID, "dosing_order": 1, "use_ratio": 100, "dosing_method": "pour"},
						},
					},
				},
			},
		},
		"steps": []map[string]interface{}{
			{"id": stepID, "step_no": 1, "name": "Step 1", "temperature": "25C", "duration": "5min"},
		},
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateFormula_DoubleMode_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	step1ID := uuid.New().String()
	step2ID := uuid.New().String()

	payload := map[string]interface{}{
		"name":           "Double Mode Formula",
		"code":           "DMF-001",
		"component_mode": "double",
		"status":         "draft",
		"parts": []map[string]interface{}{
			{
				"name": "PartA", "mix_ratio": 50, "sort_order": 1,
				"ingredients": []map[string]interface{}{
					{
						"material": "Resin A", "percentage": 100, "sort_order": 1,
						"dosing_actions": []map[string]interface{}{
							{"step_id": step1ID, "dosing_order": 1, "use_ratio": 100, "dosing_method": "pump"},
						},
					},
				},
			},
			{
				"name": "PartB", "mix_ratio": 50, "sort_order": 2,
				"ingredients": []map[string]interface{}{
					{
						"material": "Hardener B", "percentage": 100, "sort_order": 1,
						"dosing_actions": []map[string]interface{}{
							{"step_id": step2ID, "dosing_order": 1, "use_ratio": 100, "dosing_method": "pour"},
						},
					},
				},
			},
		},
		"steps": []map[string]interface{}{
			{"id": step1ID, "step_no": 1, "name": "Premix A", "temperature": "25C", "duration": "5min"},
			{"id": step2ID, "step_no": 2, "name": "Combine", "temperature": "30C", "duration": "10min"},
		},
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var f models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &f))
	assert.Equal(t, models.ComponentModeDouble, f.ComponentMode)
	assert.Len(t, f.Parts, 2)
}

func TestFormula_RepositoryDoesNotExist(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// GET a formula that doesn't exist - should get 404
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFormula_CreateWithDosingActions(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	reqBody := validCreateRequest()
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var f models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &f))
	assert.NotEmpty(t, f.Steps)
	assert.NotEmpty(t, f.Parts[0].Ingredients[0].DosingActions)
	assert.Equal(t, "pump", f.Parts[0].Ingredients[0].DosingActions[0].DosingMethod)
}

func TestFormula_ListMultiple(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// Create two formulas
	req1 := validCreateRequest()
	body1, _ := json.Marshal(req1)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body1))
	r1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, r1)
	require.Equal(t, http.StatusCreated, w1.Code)

	req2 := validCreateRequest()
	req2["name"] = "Second Formula"
	req2["code"] = "SF-002"
	body2, _ := json.Marshal(req2)
	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest(http.MethodPost, "/api/formulas", bytes.NewReader(body2))
	r2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, r2)
	require.Equal(t, http.StatusCreated, w2.Code)

	// List all
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/formulas", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var formulas []models.Formula
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &formulas))
	assert.Len(t, formulas, 2)
}

// ── Handler struct initialization ───────────────────────────────────────────

func TestNewFormulaHandler_CreatesRepo(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://nonexistent:nonexistent@localhost:5432/test?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	h := NewFormulaHandler(db)
	assert.NotNil(t, h)
	assert.NotNil(t, h.repo)
	assert.NotNil(t, h.db)
}
