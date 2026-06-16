package services

import (
	"strings"
	"testing"

	"formula-ai-system/backend/internal/models"

	"github.com/google/uuid"
)

// ============================================================================
// Test fixtures
// ============================================================================

func validSingleFormula() *models.Formula {
	partID := uuid.New()
	ing1ID := uuid.New()
	ing2ID := uuid.New()
	stepID := uuid.New()

	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Test Single",
		Code:          "TS-001",
		ComponentMode: models.ComponentModeSingle,
		Status:        models.StatusActive,
		Parts: []models.FormulaPart{
			{
				ID:        partID,
				Name:      models.PartMain,
				MixRatio:  100,
				SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{
						ID: ing1ID, PartID: partID, Material: "Resin",
						Percentage: 60, SortOrder: 1,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: stepID, IngredientID: ing1ID, DosingOrder: 1, UseRatio: 100, DosingMethod: "Pour"},
						},
					},
					{
						ID: ing2ID, PartID: partID, Material: "Hardener",
						Percentage: 40, SortOrder: 2,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: stepID, IngredientID: ing2ID, DosingOrder: 2, UseRatio: 100, DosingMethod: "Pour"},
						},
					},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: stepID, StepNo: 1, Name: "Mix", Temperature: "25°C", Duration: "10min"},
		},
	}
}

func validDoubleFormula() *models.Formula {
	partAID := uuid.New()
	partBID := uuid.New()
	ingA1ID := uuid.New()
	ingA2ID := uuid.New()
	ingB1ID := uuid.New()
	ingB2ID := uuid.New()
	step1ID := uuid.New()
	step2ID := uuid.New()

	return &models.Formula{
		ID:            uuid.New(),
		Name:          "Test Double",
		Code:          "TD-001",
		ComponentMode: models.ComponentModeDouble,
		Status:        models.StatusActive,
		Parts: []models.FormulaPart{
			{
				ID: partAID, Name: models.PartA, MixRatio: 50, SortOrder: 1,
				Ingredients: []models.FormulaIngredient{
					{
						ID: ingA1ID, PartID: partAID, Material: "Resin-A",
						Percentage: 70, SortOrder: 1,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step1ID, IngredientID: ingA1ID, DosingOrder: 1, UseRatio: 100, DosingMethod: "Pour"},
						},
					},
					{
						ID: ingA2ID, PartID: partAID, Material: "Additive-A",
						Percentage: 30, SortOrder: 2,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step1ID, IngredientID: ingA2ID, DosingOrder: 2, UseRatio: 100, DosingMethod: "Spray"},
						},
					},
				},
			},
			{
				ID: partBID, Name: models.PartB, MixRatio: 50, SortOrder: 2,
				Ingredients: []models.FormulaIngredient{
					{
						ID: ingB1ID, PartID: partBID, Material: "Hardener-B",
						Percentage: 55, SortOrder: 1,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step2ID, IngredientID: ingB1ID, DosingOrder: 1, UseRatio: 100, DosingMethod: "Pour"},
						},
					},
					{
						ID: ingB2ID, PartID: partBID, Material: "Catalyst-B",
						Percentage: 45, SortOrder: 2,
						DosingActions: []models.FormulaDosingAction{
							{ID: uuid.New(), StepID: step2ID, IngredientID: ingB2ID, DosingOrder: 2, UseRatio: 100, DosingMethod: "Inject"},
						},
					},
				},
			},
		},
		Steps: []models.FormulaStep{
			{ID: step1ID, StepNo: 1, Name: "Premix A", Temperature: "25°C", Duration: "5min"},
			{ID: step2ID, StepNo: 2, Name: "Combine A+B", Temperature: "30°C", Duration: "20min"},
		},
	}
}

// ============================================================================
// ValidationError tests
// ============================================================================

func TestValidationError_Empty(t *testing.T) {
	ve := &ValidationError{}
	if ve.HasErrors() {
		t.Error("empty ValidationError should not have errors")
	}
	if ve.HasWarnings() {
		t.Error("empty ValidationError should not have warnings")
	}
	if ve.IsBlocking() {
		t.Error("empty ValidationError should not be blocking")
	}
}

func TestValidationError_ErrorsOnly(t *testing.T) {
	ve := &ValidationError{Errors: []string{"bad thing"}}
	if !ve.HasErrors() {
		t.Error("should have errors")
	}
	if ve.HasWarnings() {
		t.Error("should not have warnings")
	}
	if !ve.IsBlocking() {
		t.Error("errors should make it blocking")
	}
}

func TestValidationError_WarningsOnly(t *testing.T) {
	ve := &ValidationError{Warnings: []string{"heads up"}}
	if ve.HasErrors() {
		t.Error("should not have errors")
	}
	if !ve.HasWarnings() {
		t.Error("should have warnings")
	}
	if ve.IsBlocking() {
		t.Error("warnings alone should not be blocking")
	}
}

func TestValidationError_ErrorFormatting(t *testing.T) {
	ve := &ValidationError{
		Errors:   []string{"e1", "e2"},
		Warnings: []string{"w1"},
	}
	msg := ve.Error()
	if !strings.Contains(msg, "e1") || !strings.Contains(msg, "e2") {
		t.Errorf("error message should contain all errors, got: %s", msg)
	}
	if !strings.Contains(msg, "w1") {
		t.Errorf("error message should contain all warnings, got: %s", msg)
	}
}

// ============================================================================
// validatePartsMatchMode tests
// ============================================================================

func TestValidatePartsMatchMode_ValidSingle(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeSingle,
		Parts:         []models.FormulaPart{{Name: models.PartMain}},
	}
	if err := validatePartsMatchMode(f); err != nil {
		t.Errorf("valid single-mode should pass, got: %v", err)
	}
}

func TestValidatePartsMatchMode_ValidDouble(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeDouble,
		Parts: []models.FormulaPart{
			{Name: models.PartA},
			{Name: models.PartB},
		},
	}
	if err := validatePartsMatchMode(f); err != nil {
		t.Errorf("valid double-mode should pass, got: %v", err)
	}
}

func TestValidatePartsMatchMode_SingleWithAB(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeSingle,
		Parts: []models.FormulaPart{
			{Name: models.PartA},
			{Name: models.PartB},
		},
	}
	if err := validatePartsMatchMode(f); err == nil {
		t.Error("single mode with A+B parts should fail")
	}
}

func TestValidatePartsMatchMode_SingleWithPartA(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeSingle,
		Parts:         []models.FormulaPart{{Name: models.PartA}},
	}
	if err := validatePartsMatchMode(f); err == nil {
		t.Error("single mode with a PartA (not MAIN) should fail")
	}
}

func TestValidatePartsMatchMode_DoubleMissingB(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeDouble,
		Parts:         []models.FormulaPart{{Name: models.PartA}},
	}
	if err := validatePartsMatchMode(f); err == nil {
		t.Error("double mode without PartB should fail")
	}
}

func TestValidatePartsMatchMode_DoubleWithMain(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeDouble,
		Parts:         []models.FormulaPart{{Name: models.PartMain}},
	}
	if err := validatePartsMatchMode(f); err == nil {
		t.Error("double mode with PartMain should fail")
	}
}

func TestValidatePartsMatchMode_EmptyParts(t *testing.T) {
	f := &models.Formula{
		ComponentMode: models.ComponentModeSingle,
		Parts:         nil,
	}
	if err := validatePartsMatchMode(f); err == nil {
		t.Error("formula with no parts should fail")
	}
}

// ============================================================================
// validatePartIngredients tests
// ============================================================================

func TestValidatePartIngredients_Sum100(t *testing.T) {
	part := &models.FormulaPart{
		Ingredients: []models.FormulaIngredient{
			{Percentage: 60},
			{Percentage: 40},
		},
	}
	if err := validatePartIngredients(part); err != nil {
		t.Errorf("sum 100%% should pass, got: %v", err)
	}
}

func TestValidatePartIngredients_SumUnder100(t *testing.T) {
	part := &models.FormulaPart{
		Ingredients: []models.FormulaIngredient{
			{Percentage: 50},
			{Percentage: 35},
		},
	}
	if err := validatePartIngredients(part); err == nil {
		t.Error("sum 85%% should fail")
	}
}

func TestValidatePartIngredients_SumOver100(t *testing.T) {
	part := &models.FormulaPart{
		Ingredients: []models.FormulaIngredient{
			{Percentage: 60},
			{Percentage: 50},
		},
	}
	if err := validatePartIngredients(part); err == nil {
		t.Error("sum 110%% should fail")
	}
}

func TestValidatePartIngredients_WithinTolerance(t *testing.T) {
	// 33.333 + 33.333 + 33.334 = 100.000 → |100-100| = 0 ≤ 0.001 ✓
	part := &models.FormulaPart{
		Ingredients: []models.FormulaIngredient{
			{Percentage: 33.333},
			{Percentage: 33.333},
			{Percentage: 33.334},
		},
	}
	if err := validatePartIngredients(part); err != nil {
		t.Errorf("sum within tolerance should pass, got: %v", err)
	}
}

func TestValidatePartIngredients_OutsideTolerance(t *testing.T) {
	// 33.33 + 33.33 + 33.33 = 99.99 → |99.99-100| = 0.01 > 0.001 ✗
	part := &models.FormulaPart{
		Ingredients: []models.FormulaIngredient{
			{Percentage: 33.33},
			{Percentage: 33.33},
			{Percentage: 33.33},
		},
	}
	if err := validatePartIngredients(part); err == nil {
		t.Error("sum 99.99%% (|diff|=0.01 > tolerance 0.001) should fail")
	}
}

func TestValidatePartIngredients_EmptyIngredients(t *testing.T) {
	part := &models.FormulaPart{
		Ingredients: []models.FormulaIngredient{},
	}
	if err := validatePartIngredients(part); err == nil {
		t.Error("part with no ingredients should fail")
	}
}

func TestValidatePartIngredients_NilPart(t *testing.T) {
	if err := validatePartIngredients(nil); err == nil {
		t.Error("nil part should fail")
	}
}

// ============================================================================
// validateDosingCompleteness tests
// ============================================================================

func TestValidateDosingCompleteness_AllHaveDosing(t *testing.T) {
	f := validSingleFormula()
	if err := validateDosingCompleteness(f); err != nil {
		t.Errorf("formula with all ingredients having dosing should pass, got: %v", err)
	}
}

func TestValidateDosingCompleteness_OneMissingDosing(t *testing.T) {
	f := validSingleFormula()
	f.Parts[0].Ingredients[0].DosingActions = nil
	if err := validateDosingCompleteness(f); err == nil {
		t.Error("formula with ingredient missing dosing action should fail")
	}
}

func TestValidateDosingCompleteness_MultipleMissingDosing(t *testing.T) {
	f := validDoubleFormula()
	f.Parts[0].Ingredients[0].DosingActions = nil
	f.Parts[1].Ingredients[0].DosingActions = nil
	if err := validateDosingCompleteness(f); err == nil {
		t.Error("formula with multiple ingredients missing dosing should fail")
	}
}

// ============================================================================
// validateDosingRatios tests
// ============================================================================

func TestValidateDosingRatios_Sum100(t *testing.T) {
	ing := &models.FormulaIngredient{
		DosingActions: []models.FormulaDosingAction{
			{UseRatio: 60},
			{UseRatio: 40},
		},
	}
	if err := validateDosingRatios(ing); err != nil {
		t.Errorf("dosing ratios sum 100%% should pass, got: %v", err)
	}
}

func TestValidateDosingRatios_SumUnder100(t *testing.T) {
	ing := &models.FormulaIngredient{
		DosingActions: []models.FormulaDosingAction{
			{UseRatio: 50},
			{UseRatio: 30},
		},
	}
	if err := validateDosingRatios(ing); err == nil {
		t.Error("dosing ratios sum 80%% should fail")
	}
}

func TestValidateDosingRatios_SumOver100(t *testing.T) {
	ing := &models.FormulaIngredient{
		DosingActions: []models.FormulaDosingAction{
			{UseRatio: 70},
			{UseRatio: 60},
		},
	}
	if err := validateDosingRatios(ing); err == nil {
		t.Error("dosing ratios sum 130%% should fail")
	}
}

func TestValidateDosingRatios_WithinTolerance(t *testing.T) {
	// 33.333 + 33.333 + 33.334 = 100.000 → within tolerance
	ing := &models.FormulaIngredient{
		DosingActions: []models.FormulaDosingAction{
			{UseRatio: 33.333},
			{UseRatio: 33.333},
			{UseRatio: 33.334},
		},
	}
	if err := validateDosingRatios(ing); err != nil {
		t.Errorf("dosing ratios within tolerance should pass, got: %v", err)
	}
}

func TestValidateDosingRatios_OutsideTolerance(t *testing.T) {
	// 33.33 + 33.33 + 33.33 = 99.99 → |99.99-100| = 0.01 > 0.001 ✗
	ing := &models.FormulaIngredient{
		DosingActions: []models.FormulaDosingAction{
			{UseRatio: 33.33},
			{UseRatio: 33.33},
			{UseRatio: 33.33},
		},
	}
	if err := validateDosingRatios(ing); err == nil {
		t.Error("dosing ratios 99.99%% outside tolerance should fail")
	}
}

func TestValidateDosingRatios_NoActions(t *testing.T) {
	ing := &models.FormulaIngredient{
		DosingActions: []models.FormulaDosingAction{},
	}
	if err := validateDosingRatios(ing); err == nil {
		t.Error("ingredient with no dosing actions should fail ratio validation")
	}
}

func TestValidateDosingRatios_NilIngredient(t *testing.T) {
	if err := validateDosingRatios(nil); err == nil {
		t.Error("nil ingredient should fail")
	}
}

// ============================================================================
// validateStepsForActivation tests
// ============================================================================

func TestValidateStepsForActivation_ActiveWithSteps(t *testing.T) {
	f := &models.Formula{
		Status: models.StatusActive,
		Steps:  []models.FormulaStep{{StepNo: 1, Name: "Mix"}},
	}
	if err := validateStepsForActivation(f); err != nil {
		t.Errorf("active formula with steps should pass, got: %v", err)
	}
}

func TestValidateStepsForActivation_ActiveNoSteps(t *testing.T) {
	f := &models.Formula{
		Status: models.StatusActive,
		Steps:  nil,
	}
	if err := validateStepsForActivation(f); err == nil {
		t.Error("active formula without steps should fail")
	}
}

func TestValidateStepsForActivation_DraftNoSteps(t *testing.T) {
	f := &models.Formula{
		Status: models.StatusDraft,
		Steps:  nil,
	}
	if err := validateStepsForActivation(f); err != nil {
		t.Errorf("draft formula without steps should pass, got: %v", err)
	}
}

func TestValidateStepsForActivation_ArchivedNoSteps(t *testing.T) {
	f := &models.Formula{
		Status: models.StatusArchived,
		Steps:  nil,
	}
	if err := validateStepsForActivation(f); err != nil {
		t.Errorf("archived formula without steps should pass, got: %v", err)
	}
}

func TestValidateStepsForActivation_NilFormula(t *testing.T) {
	if err := validateStepsForActivation(nil); err == nil {
		t.Error("nil formula should fail")
	}
}

// ============================================================================
// ValidateAndPrepare integration tests
// ============================================================================

func TestValidateAndPrepare_ValidSingle(t *testing.T) {
	f := validSingleFormula()
	if err := ValidateAndPrepare(f); err != nil {
		t.Errorf("valid single formula should pass, got: %v", err)
	}
}

func TestValidateAndPrepare_ValidDouble(t *testing.T) {
	f := validDoubleFormula()
	if err := ValidateAndPrepare(f); err != nil {
		t.Errorf("valid double formula should pass, got: %v", err)
	}
}

func TestValidateAndPrepare_NilFormula(t *testing.T) {
	err := ValidateAndPrepare(nil)
	if err == nil {
		t.Fatal("nil formula should return error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if !ve.HasErrors() {
		t.Error("nil formula should produce errors")
	}
}

func TestValidateAndPrepare_SingleModeWrongParts(t *testing.T) {
	f := validSingleFormula()
	f.Parts = []models.FormulaPart{
		{Name: models.PartA, MixRatio: 50},
		{Name: models.PartB, MixRatio: 50},
	}
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected error for wrong parts")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if !ve.HasErrors() {
		t.Error("should have blocking errors")
	}
}

func TestValidateAndPrepare_IngredientsSumInvalid(t *testing.T) {
	f := validSingleFormula()
	f.Parts[0].Ingredients = []models.FormulaIngredient{
		{Material: "A", Percentage: 50},
		{Material: "B", Percentage: 25},
	}
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected error for invalid ingredient sum")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if !ve.HasErrors() {
		t.Error("should have blocking errors")
	}
	if !strings.Contains(ve.Error(), "sum") {
		t.Errorf("error message should mention sum, got: %s", ve.Error())
	}
}

func TestValidateAndPrepare_DosingCompletenessMissing(t *testing.T) {
	f := validSingleFormula()
	f.Parts[0].Ingredients[0].DosingActions = nil
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected error for missing dosing action")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if !ve.HasErrors() {
		t.Error("should have blocking errors")
	}
	if !strings.Contains(ve.Error(), "dosing") {
		t.Errorf("error message should mention dosing, got: %s", ve.Error())
	}
}

func TestValidateAndPrepare_ActiveNoSteps(t *testing.T) {
	f := validSingleFormula()
	f.Status = models.StatusActive
	f.Steps = nil
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected error for active formula with no steps")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if !ve.HasErrors() {
		t.Error("should have blocking errors")
	}
}

func TestValidateAndPrepare_MultipleErrorsAggregated(t *testing.T) {
	f := validSingleFormula()
	// Cause 3 errors: wrong parts mode + bad ingredient sum + missing dosing
	f.ComponentMode = models.ComponentModeDouble
	f.Parts = []models.FormulaPart{
		{
			Name: models.PartA, MixRatio: 50,
			Ingredients: []models.FormulaIngredient{
				{Material: "X", Percentage: 50, DosingActions: nil},
			},
		},
	}
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected aggregated errors")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) < 3 {
		t.Errorf("expected ≥3 aggregated errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidateAndPrepare_WarningOnly_DraftNoSteps(t *testing.T) {
	f := validSingleFormula()
	f.Status = models.StatusDraft
	f.Steps = nil
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected warnings to be returned")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if ve.HasErrors() {
		t.Error("draft without steps should not produce errors")
	}
	if !ve.HasWarnings() {
		t.Error("draft without steps should produce a warning")
	}
	if ve.IsBlocking() {
		t.Error("warnings alone should not block save")
	}
}

func TestValidateAndPrepare_ErrorContainsActualAndExpected(t *testing.T) {
	f := validSingleFormula()
	f.Parts[0].Ingredients = []models.FormulaIngredient{
		{Material: "A", Percentage: 95},
	}
	err := ValidateAndPrepare(f)
	if err == nil {
		t.Fatal("expected error for 95% sum")
	}
	msg := err.Error()
	if !strings.Contains(msg, "95") {
		t.Errorf("error should contain actual sum 95, got: %s", msg)
	}
	if !strings.Contains(msg, "100") {
		t.Errorf("error should contain expected 100%%, got: %s", msg)
	}
}

func TestValidateAndPrepare_ValidDoubleFormula(t *testing.T) {
	f := validDoubleFormula()
	if err := ValidateAndPrepare(f); err != nil {
		t.Errorf("valid double formula should pass, got: %v", err)
	}
}
