package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"formula-ai-system/backend/internal/config"
	"formula-ai-system/backend/internal/models"
)

// --- Helpers ---

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	cfg := &config.Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "formula"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "changeme"),
		DBName:     getEnvOrDefault("DB_NAME", "formula_ai"),
	}
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("db.Ping: %v (is PostgreSQL running?)", err)
	}
	// Truncate all tables for test isolation
	tables := []string{"formula_dosing_actions", "formula_ingredients", "formula_steps", "formula_parts", "formulas"}
	for _, tbl := range tables {
		db.Exec("DELETE FROM " + tbl)
	}
	return db
}

func makeSingleFormula() *models.Formula {
	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Test Single Formula",
		Code:          "TSF-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts: []models.FormulaPart{
			{
				ID:        uuid.New(),
				Name:      models.PartMain,
				MixRatio:  100,
				SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{ID: uuid.New(), Material: "Resin", Percentage: 60, SortOrder: 1},
					{ID: uuid.New(), Material: "Hardener", Percentage: 40, SortOrder: 2},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: uuid.New(), StepNo: 1, Name: "Mixing", Temperature: "25C", Duration: "10min"},
			{ID: uuid.New(), StepNo: 2, Name: "Curing", Temperature: "60C", Duration: "2h"},
		},
	}
}

func makeDoubleFormula() *models.Formula {
	partAID := uuid.New()
	partBID := uuid.New()
	ingA1 := uuid.New()
	ingA2 := uuid.New()
	ingB1 := uuid.New()
	ingB2 := uuid.New()
	step1 := uuid.New()
	step2 := uuid.New()
	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Test Double Formula",
		Code:          "TDF-001",
		ComponentMode: models.ComponentModeDouble,
		Status:        models.StatusActive,
		Parts: []models.FormulaPart{
			{
				ID: partAID, Name: models.PartA, MixRatio: 50, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{ID: ingA1, Material: "Resin-A", Percentage: 70, SortOrder: 1},
					{ID: ingA2, Material: "Additive-A", Percentage: 30, SortOrder: 2},
				},
			},
			{
				ID: partBID, Name: models.PartB, MixRatio: 50, SortOrder: 2,
				Ingredients: []models.FormulaIngredient{
					{ID: ingB1, Material: "Hardener-B", Percentage: 55, SortOrder: 1},
					{ID: ingB2, Material: "Catalyst-B", Percentage: 45, SortOrder: 2},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: step1, StepNo: 1, Name: "Premix A", Temperature: "25C", Duration: "5min"},
			{ID: step2, StepNo: 2, Name: "Combine A+B", Temperature: "30C", Duration: "20min"},
		},
	}
}

func makeFormulaWithDosing() *models.Formula {
	ing1ID := uuid.New()
	ing2ID := uuid.New()
	step1ID := uuid.New()
	step2ID := uuid.New()

	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Formula With Dosing",
		Code:          "FWD-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts: []models.FormulaPart{
			{
				ID: uuid.New(), Name: models.PartMain, MixRatio: 100, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{
						ID: ing1ID, Material: "Resin", Percentage: 60, SortOrder: 1,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step1ID, DosingOrder: 1, UseRatio: 60, DosingMethod: "pump"},
							{ID: uuid.New(), StepID: step2ID, DosingOrder: 2, UseRatio: 40, DosingMethod: "manual"},
						},
					},
					{
						ID: ing2ID, Material: "Hardener", Percentage: 40, SortOrder: 2,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step1ID, DosingOrder: 1, UseRatio: 100, DosingMethod: "pump"},
						},
					},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: step1ID, StepNo: 1, Name: "Dosing", Temperature: "25C", Duration: "5min"},
			{ID: step2ID, StepNo: 2, Name: "Mixing", Temperature: "30C", Duration: "10min"},
		},
	}
}

// --- Test FormulaRepo.Create ---

func TestFormulaRepo_Create_SingleMode(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeSingleFormula()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Verify it was persisted
	retrieved, err := repo.GetByID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if retrieved.Name != f.Name {
		t.Errorf("Name = %q, want %q", retrieved.Name, f.Name)
	}
	if retrieved.Code != f.Code {
		t.Errorf("Code = %q, want %q", retrieved.Code, f.Code)
	}
	if retrieved.ComponentMode != f.ComponentMode {
		t.Errorf("ComponentMode = %q, want %q", retrieved.ComponentMode, f.ComponentMode)
	}
	if retrieved.Status != f.Status {
		t.Errorf("Status = %q, want %q", retrieved.Status, f.Status)
	}
	if len(retrieved.Parts) != 1 {
		t.Fatalf("len(Parts) = %d, want 1", len(retrieved.Parts))
	}
	if len(retrieved.Parts[0].Ingredients) != 2 {
		t.Fatalf("len(Ingredients) = %d, want 2", len(retrieved.Parts[0].Ingredients))
	}
	if len(retrieved.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want 2", len(retrieved.Steps))
	}
}

func TestFormulaRepo_Create_DoubleMode(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeDoubleFormula()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if len(retrieved.Parts) != 2 {
		t.Fatalf("len(Parts) = %d, want 2", len(retrieved.Parts))
	}
	// PartA
	if retrieved.Parts[0].Name != models.PartA {
		t.Errorf("Parts[0].Name = %q, want PartA", retrieved.Parts[0].Name)
	}
	if len(retrieved.Parts[0].Ingredients) != 2 {
		t.Errorf("len(Parts[0].Ingredients) = %d, want 2", len(retrieved.Parts[0].Ingredients))
	}
	// PartB
	if retrieved.Parts[1].Name != models.PartB {
		t.Errorf("Parts[1].Name = %q, want PartB", retrieved.Parts[1].Name)
	}
	if len(retrieved.Parts[1].Ingredients) != 2 {
		t.Errorf("len(Parts[1].Ingredients) = %d, want 2", len(retrieved.Parts[1].Ingredients))
	}
	if len(retrieved.Steps) != 2 {
		t.Errorf("len(Steps) = %d, want 2", len(retrieved.Steps))
	}
}

func TestFormulaRepo_Create_WithDosingActions(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeFormulaWithDosing()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if len(retrieved.Parts) != 1 {
		t.Fatalf("len(Parts) = %d, want 1", len(retrieved.Parts))
	}
	ings := retrieved.Parts[0].Ingredients
	if len(ings) != 2 {
		t.Fatalf("len(Ingredients) = %d, want 2", len(ings))
	}
	if len(ings[0].DosingActions) != 2 {
		t.Errorf("len(DosingActions[0]) = %d, want 2", len(ings[0].DosingActions))
	}
	if len(ings[1].DosingActions) != 1 {
		t.Errorf("len(DosingActions[1]) = %d, want 1", len(ings[1].DosingActions))
	}
}

// --- Test FormulaRepo.GetByID ---

func TestFormulaRepo_GetByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	_, err := repo.GetByID(ctx, db, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent formula")
	}
}

// --- Test FormulaRepo.List ---

func TestFormulaRepo_List_Empty(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	formulas, err := repo.List(ctx, db)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(formulas) != 0 {
		t.Errorf("len(formulas) = %d, want 0", len(formulas))
	}
}

func TestFormulaRepo_List_Multiple(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f1 := makeSingleFormula()
	f2 := makeDoubleFormula()
	f2.Code = "TDF-002" // unique code

	if err := repo.Create(ctx, db, f1); err != nil {
		t.Fatalf("Create f1: %v", err)
	}
	if err := repo.Create(ctx, db, f2); err != nil {
		t.Fatalf("Create f2: %v", err)
	}

	formulas, err := repo.List(ctx, db)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(formulas) != 2 {
		t.Fatalf("len(formulas) = %d, want 2", len(formulas))
	}

	// Verify nested data is populated
	for _, f := range formulas {
		if len(f.Parts) == 0 {
			t.Errorf("formula %s has no parts", f.ID)
		}
		if len(f.Steps) == 0 {
			t.Errorf("formula %s has no steps", f.ID)
		}
	}
}

// --- Test FormulaRepo.Update ---

func TestFormulaRepo_Update_ChangeData(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeSingleFormula()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	// Modify the formula
	f.Name = "Updated Formula"
	f.Code = "UPD-001"
	f.Parts[0].Ingredients = append(f.Parts[0].Ingredients, models.FormulaIngredient{
		ID: uuid.New(), Material: "Catalyst", Percentage: 0, SortOrder: 3,
	})
	// Adjust percentages to sum to 100
	f.Parts[0].Ingredients[0].Percentage = 50
	f.Parts[0].Ingredients[1].Percentage = 30
	f.Parts[0].Ingredients[2].Percentage = 20

	f.Steps = []models.FormulaStep{
		{ID: uuid.New(), StepNo: 1, Name: "New Step", Temperature: "40C", Duration: "15min"},
	}

	if err := repo.Update(ctx, db, f); err != nil {
		t.Fatalf("Update(): %v", err)
	}

	retrieved, err := repo.GetByID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("GetByID(): %v", err)
	}

	if retrieved.Name != "Updated Formula" {
		t.Errorf("Name = %q, want %q", retrieved.Name, "Updated Formula")
	}
	if len(retrieved.Parts[0].Ingredients) != 3 {
		t.Errorf("len(Ingredients) = %d, want 3", len(retrieved.Parts[0].Ingredients))
	}
	if len(retrieved.Steps) != 1 {
		t.Errorf("len(Steps) = %d, want 1", len(retrieved.Steps))
	}
	if retrieved.Steps[0].Name != "New Step" {
		t.Errorf("Step name = %q, want %q", retrieved.Steps[0].Name, "New Step")
	}
}

func TestFormulaRepo_Update_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeSingleFormula()
	err := repo.Update(ctx, db, f)
	if err == nil {
		t.Fatal("expected error for updating non-existent formula")
	}
}

// --- Test FormulaRepo.Delete ---

func TestFormulaRepo_Delete_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeSingleFormula()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	if err := repo.Delete(ctx, db, f.ID); err != nil {
		t.Fatalf("Delete(): %v", err)
	}

	// Verify formula is gone
	_, err := repo.GetByID(ctx, db, f.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestFormulaRepo_Delete_CascadeRemovesNested(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeFormulaWithDosing()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	// Get part and ingredient IDs for verification
	partID := f.Parts[0].ID
	ingID := f.Parts[0].Ingredients[0].ID

	if err := repo.Delete(ctx, db, f.ID); err != nil {
		t.Fatalf("Delete(): %v", err)
	}

	// Part should be cascade-deleted
	var partCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM formula_parts WHERE id = $1", partID).Scan(&partCount)
	if partCount != 0 {
		t.Errorf("part still exists after cascade delete: count=%d", partCount)
	}

	// Ingredient should be cascade-deleted
	var ingCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM formula_ingredients WHERE id = $1", ingID).Scan(&ingCount)
	if ingCount != 0 {
		t.Errorf("ingredient still exists after cascade delete: count=%d", ingCount)
	}

	// Steps should be cascade-deleted
	var stepCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM formula_steps WHERE formula_id = $1", f.ID).Scan(&stepCount)
	if stepCount != 0 {
		t.Errorf("steps still exist after cascade delete: count=%d", stepCount)
	}
}

func TestFormulaRepo_Delete_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	err := repo.Delete(ctx, db, uuid.New())
	if err == nil {
		t.Fatal("expected error for deleting non-existent formula")
	}
}

// --- Test FormulaRepo.Create with auto-generated UUIDs ---

func TestFormulaRepo_Create_AutoGenerateUUIDs(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	// All zero UUIDs - repo should generate them
	f := &models.Formula{
		Name:          "Auto UUID Test",
		Code:          "AUTO-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusDraft,
		Parts: []models.FormulaPart{
			{
				Name: models.PartMain, MixRatio: 100, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{Material: "A", Percentage: 100},
				},
			},
		},
		Steps: []models.FormulaStep{
			{StepNo: 1, Name: "Step 1"},
		},
	}

	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	if f.ID == uuid.Nil {
		t.Error("Formula ID was not generated")
	}
	if f.Parts[0].ID == uuid.Nil {
		t.Error("Part ID was not generated")
	}
	if f.Parts[0].Ingredients[0].ID == uuid.Nil {
		t.Error("Ingredient ID was not generated")
	}
	if f.Steps[0].ID == uuid.Nil {
		t.Error("Step ID was not generated")
	}
}

// --- Test Formula Update with dosing actions ---

func TestFormulaRepo_Update_WithDosingActions(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewFormulaRepo()

	f := makeFormulaWithDosing()
	if err := repo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	// Change dosing actions
	f.Parts[0].Ingredients[0].DosingActions = []models.FormulaDosingAction{
		{ID: uuid.New(), StepID: f.Steps[0].ID, DosingOrder: 1, UseRatio: 100, DosingMethod: "updated_method"},
	}

	if err := repo.Update(ctx, db, f); err != nil {
		t.Fatalf("Update(): %v", err)
	}

	retrieved, err := repo.GetByID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("GetByID(): %v", err)
	}

	da := retrieved.Parts[0].Ingredients[0].DosingActions
	if len(da) != 1 {
		t.Fatalf("len(DosingActions) = %d, want 1", len(da))
	}
	if da[0].DosingMethod != "updated_method" {
		t.Errorf("DosingMethod = %q, want %q", da[0].DosingMethod, "updated_method")
	}
}
