package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"formula-ai-system/backend/internal/models"
)

// --- PartRepo Tests ---

func TestPartRepo_CreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a parent formula first
	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	if err := formulaRepo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create formula: %v", err)
	}

	partRepo := NewPartRepo()
	part := &models.FormulaPart{
		ID:        uuid.New(),
		FormulaID: f.ID,
		Name:      models.PartMain,
		MixRatio:  100,
		SortOrder: 1,
	}
	if err := partRepo.Create(ctx, db, part); err != nil {
		t.Fatalf("Create part: %v", err)
	}

	retrieved, err := partRepo.GetByID(ctx, db, part.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if retrieved.Name != models.PartMain {
		t.Errorf("Name = %q, want PartMain", retrieved.Name)
	}
	if retrieved.FormulaID != f.ID {
		t.Errorf("FormulaID mismatch")
	}
}

func TestPartRepo_ListByFormulaID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	if err := formulaRepo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create formula: %v", err)
	}

	partRepo := NewPartRepo()
	partA := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartA, MixRatio: 50, SortOrder: 1}
	partB := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartB, MixRatio: 50, SortOrder: 2}
	partRepo.Create(ctx, db, partA)
	partRepo.Create(ctx, db, partB)

	parts, err := partRepo.ListByFormulaID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("ListByFormulaID: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("len(parts) = %d, want 2", len(parts))
	}
}

func TestPartRepo_Update(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	if err := formulaRepo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create formula: %v", err)
	}

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	part.MixRatio = 75
	part.SortOrder = 5
	if err := partRepo.Update(ctx, db, part); err != nil {
		t.Fatalf("Update: %v", err)
	}

	retrieved, _ := partRepo.GetByID(ctx, db, part.ID)
	if retrieved.MixRatio != 75 {
		t.Errorf("MixRatio = %f, want 75", retrieved.MixRatio)
	}
	if retrieved.SortOrder != 5 {
		t.Errorf("SortOrder = %d, want 5", retrieved.SortOrder)
	}
}

func TestPartRepo_Delete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	if err := formulaRepo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create formula: %v", err)
	}

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	if err := partRepo.Delete(ctx, db, part.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := partRepo.GetByID(ctx, db, part.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

// --- IngredientRepo Tests ---

func TestIngredientRepo_CreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create parent formula and part
	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	if err := formulaRepo.Create(ctx, db, f); err != nil {
		t.Fatalf("Create formula: %v", err)
	}

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	ingRepo := NewIngredientRepo()
	ing := &models.FormulaIngredient{
		ID:         uuid.New(),
		PartID:     part.ID,
		Material:   "Test Material",
		Percentage: 50,
		SortOrder:  1,
	}
	if err := ingRepo.Create(ctx, db, ing); err != nil {
		t.Fatalf("Create: %v", err)
	}

	retrieved, err := ingRepo.GetByID(ctx, db, ing.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if retrieved.Material != "Test Material" {
		t.Errorf("Material = %q", retrieved.Material)
	}
	if retrieved.Percentage != 50 {
		t.Errorf("Percentage = %f", retrieved.Percentage)
	}
}

func TestIngredientRepo_ListByPartID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	formulaRepo.Create(ctx, db, f)

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	ingRepo := NewIngredientRepo()
	ingRepo.Create(ctx, db, &models.FormulaIngredient{ID: uuid.New(), PartID: part.ID, Material: "A", Percentage: 30, SortOrder: 1})
	ingRepo.Create(ctx, db, &models.FormulaIngredient{ID: uuid.New(), PartID: part.ID, Material: "B", Percentage: 70, SortOrder: 2})

	ings, err := ingRepo.ListByPartID(ctx, db, part.ID)
	if err != nil {
		t.Fatalf("ListByPartID: %v", err)
	}
	if len(ings) != 2 {
		t.Fatalf("len = %d, want 2", len(ings))
	}
}

// --- StepRepo Tests ---

func TestStepRepo_CreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	formulaRepo.Create(ctx, db, f)

	stepRepo := NewStepRepo()
	step := &models.FormulaStep{
		ID:          uuid.New(),
		FormulaID:   f.ID,
		StepNo:      1,
		Name:        "Test Step",
		Temperature: "30C",
		Duration:    "5min",
	}
	if err := stepRepo.Create(ctx, db, step); err != nil {
		t.Fatalf("Create: %v", err)
	}

	retrieved, err := stepRepo.GetByID(ctx, db, step.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if retrieved.Name != "Test Step" {
		t.Errorf("Name = %q", retrieved.Name)
	}
	if retrieved.Temperature != "30C" {
		t.Errorf("Temperature = %q", retrieved.Temperature)
	}
}

func TestStepRepo_ListByFormulaID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	formulaRepo.Create(ctx, db, f)

	stepRepo := NewStepRepo()
	stepRepo.Create(ctx, db, &models.FormulaStep{ID: uuid.New(), FormulaID: f.ID, StepNo: 1, Name: "S1"})
	stepRepo.Create(ctx, db, &models.FormulaStep{ID: uuid.New(), FormulaID: f.ID, StepNo: 2, Name: "S2"})

	steps, err := stepRepo.ListByFormulaID(ctx, db, f.ID)
	if err != nil {
		t.Fatalf("ListByFormulaID: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("len = %d, want 2", len(steps))
	}
}

// --- DosingRepo Tests ---

func TestDosingRepo_CreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create parent formula, part, ingredient, step
	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	formulaRepo.Create(ctx, db, f)

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	ingRepo := NewIngredientRepo()
	ing := &models.FormulaIngredient{ID: uuid.New(), PartID: part.ID, Material: "M", Percentage: 100, SortOrder: 1}
	ingRepo.Create(ctx, db, ing)

	stepRepo := NewStepRepo()
	step := &models.FormulaStep{ID: uuid.New(), FormulaID: f.ID, StepNo: 1, Name: "S"}
	stepRepo.Create(ctx, db, step)

	dosingRepo := NewDosingRepo()
	da := &models.FormulaDosingAction{
		ID:           uuid.New(),
		StepID:       step.ID,
		IngredientID: ing.ID,
		DosingOrder:  1,
		UseRatio:     100,
		DosingMethod: "pump",
	}
	if err := dosingRepo.Create(ctx, db, da); err != nil {
		t.Fatalf("Create: %v", err)
	}

	retrieved, err := dosingRepo.GetByID(ctx, db, da.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if retrieved.DosingMethod != "pump" {
		t.Errorf("DosingMethod = %q", retrieved.DosingMethod)
	}
	if retrieved.UseRatio != 100 {
		t.Errorf("UseRatio = %f", retrieved.UseRatio)
	}
}

func TestDosingRepo_ListByStepID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	formulaRepo.Create(ctx, db, f)

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	ingRepo := NewIngredientRepo()
	ing1 := &models.FormulaIngredient{ID: uuid.New(), PartID: part.ID, Material: "A", Percentage: 60, SortOrder: 1}
	ing2 := &models.FormulaIngredient{ID: uuid.New(), PartID: part.ID, Material: "B", Percentage: 40, SortOrder: 2}
	ingRepo.Create(ctx, db, ing1)
	ingRepo.Create(ctx, db, ing2)

	stepRepo := NewStepRepo()
	step := &models.FormulaStep{ID: uuid.New(), FormulaID: f.ID, StepNo: 1, Name: "S"}
	stepRepo.Create(ctx, db, step)

	dosingRepo := NewDosingRepo()
	dosingRepo.Create(ctx, db, &models.FormulaDosingAction{ID: uuid.New(), StepID: step.ID, IngredientID: ing1.ID, DosingOrder: 1, UseRatio: 60})
	dosingRepo.Create(ctx, db, &models.FormulaDosingAction{ID: uuid.New(), StepID: step.ID, IngredientID: ing2.ID, DosingOrder: 2, UseRatio: 40})

	actions, err := dosingRepo.ListByStepID(ctx, db, step.ID)
	if err != nil {
		t.Fatalf("ListByStepID: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("len = %d, want 2", len(actions))
	}
}

func TestDosingRepo_ListByIngredientID(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	formulaRepo := NewFormulaRepo()
	f := makeSingleFormula()
	f.Parts = nil
	f.Steps = nil
	formulaRepo.Create(ctx, db, f)

	partRepo := NewPartRepo()
	part := &models.FormulaPart{ID: uuid.New(), FormulaID: f.ID, Name: models.PartMain, MixRatio: 100, SortOrder: 1}
	partRepo.Create(ctx, db, part)

	ingRepo := NewIngredientRepo()
	ing := &models.FormulaIngredient{ID: uuid.New(), PartID: part.ID, Material: "M", Percentage: 100, SortOrder: 1}
	ingRepo.Create(ctx, db, ing)

	stepRepo := NewStepRepo()
	step1 := &models.FormulaStep{ID: uuid.New(), FormulaID: f.ID, StepNo: 1, Name: "S1"}
	step2 := &models.FormulaStep{ID: uuid.New(), FormulaID: f.ID, StepNo: 2, Name: "S2"}
	stepRepo.Create(ctx, db, step1)
	stepRepo.Create(ctx, db, step2)

	dosingRepo := NewDosingRepo()
	dosingRepo.Create(ctx, db, &models.FormulaDosingAction{ID: uuid.New(), StepID: step1.ID, IngredientID: ing.ID, DosingOrder: 1, UseRatio: 50})
	dosingRepo.Create(ctx, db, &models.FormulaDosingAction{ID: uuid.New(), StepID: step2.ID, IngredientID: ing.ID, DosingOrder: 2, UseRatio: 50})

	actions, err := dosingRepo.ListByIngredientID(ctx, db, ing.ID)
	if err != nil {
		t.Fatalf("ListByIngredientID: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("len = %d, want 2", len(actions))
	}
}
