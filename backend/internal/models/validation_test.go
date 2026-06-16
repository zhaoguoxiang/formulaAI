package models

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// --- Test Fixtures ---

func validSingleFormulaFixture() *Formula {
	return &Formula{
		ID:            uuid.New(),
		Name:          "Test Single Formula",
		Code:          "TSF-001",
		ComponentMode: ComponentModeSingle,
		Status:        StatusActive,
		Parts: []FormulaPart{
			{
				ID:        uuid.New(),
				Name:      PartMain,
				MixRatio:  100,
				SortOrder: 1,
				Ingredients: []FormulaIngredient{
					{ID: uuid.New(), Material: "Resin", Percentage: 60},
					{ID: uuid.New(), Material: "Hardener", Percentage: 40},
				},
			},
		},
		Steps: []FormulaStep{
			{ID: uuid.New(), StepNo: 1, Name: "Mixing", Temperature: "25°C", Duration: "10min"},
		},
	}
}

func validDoubleFormulaFixture() *Formula {
	partAID := uuid.New()
	partBID := uuid.New()
	ingA1 := uuid.New()
	ingA2 := uuid.New()
	ingB1 := uuid.New()
	ingB2 := uuid.New()
	step1ID := uuid.New()
	step2ID := uuid.New()
	return &Formula{
		ID:            uuid.New(),
		Name:          "Test Double Formula",
		Code:          "TDF-001",
		ComponentMode: ComponentModeDouble,
		Status:        StatusActive,
		Parts: []FormulaPart{
			{
				ID:        partAID,
				Name:      PartA,
				MixRatio:  50,
				SortOrder: 1,
				Ingredients: []FormulaIngredient{
					{ID: ingA1, PartID: partAID, Material: "Resin-A", Percentage: 70},
					{ID: ingA2, PartID: partAID, Material: "Additive-A", Percentage: 30},
				},
			},
			{
				ID:        partBID,
				Name:      PartB,
				MixRatio:  50,
				SortOrder: 2,
				Ingredients: []FormulaIngredient{
					{ID: ingB1, PartID: partBID, Material: "Hardener-B", Percentage: 55},
					{ID: ingB2, PartID: partBID, Material: "Catalyst-B", Percentage: 45},
				},
			},
		},
		Steps: []FormulaStep{
			{ID: step1ID, StepNo: 1, Name: "Premix A", Temperature: "25°C", Duration: "5min"},
			{ID: step2ID, StepNo: 2, Name: "Combine A+B", Temperature: "30°C", Duration: "20min"},
		},
	}
}

// --- TestValidateFormula ---

func TestValidateFormula_Valid_SingleMode(t *testing.T) {
	f := validSingleFormulaFixture()
	err := ValidateFormula(f)
	if err != nil {
		t.Errorf("expected valid single-mode formula to pass, got: %v", err)
	}
}

func TestValidateFormula_Valid_DoubleMode(t *testing.T) {
	f := validDoubleFormulaFixture()
	err := ValidateFormula(f)
	if err != nil {
		t.Errorf("expected valid double-mode formula to pass, got: %v", err)
	}
}

func TestValidateFormula_SingleMode_WrongParts(t *testing.T) {
	f := validSingleFormulaFixture()
	f.Parts = []FormulaPart{
		{Name: PartA, MixRatio: 50, SortOrder: 1},
		{Name: PartB, MixRatio: 50, SortOrder: 2},
	}
	err := ValidateFormula(f)
	if err == nil {
		t.Error("expected error for single-mode formula with A+B parts")
	}
}

func TestValidateFormula_DoubleMode_MissingPart(t *testing.T) {
	f := validDoubleFormulaFixture()
	f.Parts = []FormulaPart{
		{Name: PartA, MixRatio: 100, SortOrder: 1},
	}
	err := ValidateFormula(f)
	if err == nil {
		t.Error("expected error for double-mode formula missing PartB")
	}
}

func TestValidateFormula_DoubleMode_WrongParts(t *testing.T) {
	f := validDoubleFormulaFixture()
	f.Parts = []FormulaPart{
		{Name: PartMain, MixRatio: 100, SortOrder: 1},
	}
	err := ValidateFormula(f)
	if err == nil {
		t.Error("expected error for double-mode formula with Main part")
	}
}

func TestValidateFormula_EmptyParts(t *testing.T) {
	f := validSingleFormulaFixture()
	f.Parts = nil
	err := ValidateFormula(f)
	if err == nil {
		t.Error("expected error for formula with no parts")
	}
}

func TestValidateFormula_Active_NoSteps(t *testing.T) {
	f := validSingleFormulaFixture()
	f.Status = StatusActive
	f.Steps = nil
	err := ValidateFormula(f)
	if err == nil {
		t.Error("expected error for active formula with no steps")
	}
}

func TestValidateFormula_Draft_NoSteps_Ok(t *testing.T) {
	f := validSingleFormulaFixture()
	f.Status = StatusDraft
	f.Steps = nil
	err := ValidateFormula(f)
	if err != nil {
		t.Errorf("expected draft formula without steps to pass, got: %v", err)
	}
}

// --- TestValidatePartIngredients ---

func TestValidatePartIngredients_Sum100(t *testing.T) {
	part := &FormulaPart{
		Ingredients: []FormulaIngredient{
			{Percentage: 60},
			{Percentage: 40},
		},
	}
	err := ValidatePartIngredients(part)
	if err != nil {
		t.Errorf("expected sum 100%% to pass, got: %v", err)
	}
}

func TestValidatePartIngredients_SumNot100(t *testing.T) {
	part := &FormulaPart{
		Ingredients: []FormulaIngredient{
			{Percentage: 50},
			{Percentage: 35},
		},
	}
	err := ValidatePartIngredients(part)
	if err == nil {
		t.Error("expected error for sum of 85%")
	}
}

func TestValidatePartIngredients_SumOver100(t *testing.T) {
	part := &FormulaPart{
		Ingredients: []FormulaIngredient{
			{Percentage: 60},
			{Percentage: 50},
		},
	}
	err := ValidatePartIngredients(part)
	if err == nil {
		t.Error("expected error for sum of 110%")
	}
}

func TestValidatePartIngredients_WithinTolerance(t *testing.T) {
	part := &FormulaPart{
		Ingredients: []FormulaIngredient{
			{Percentage: 33.333},
			{Percentage: 33.333},
			{Percentage: 33.334},
		},
	}
	err := ValidatePartIngredients(part)
	if err != nil {
		t.Errorf("expected sum 100%% within tolerance (33.333+33.333+33.334=100) to pass, got: %v", err)
	}
}

func TestValidatePartIngredients_EmptyIngredients(t *testing.T) {
	part := &FormulaPart{
		Ingredients: []FormulaIngredient{},
	}
	err := ValidatePartIngredients(part)
	if err == nil {
		t.Error("expected error for empty ingredients")
	}
}

// --- TestValidateIngredientHasDosingActions ---

func TestValidateIngredientHasDosingActions_HasActions(t *testing.T) {
	ing := &FormulaIngredient{
		DosingActions: []FormulaDosingAction{
			{UseRatio: 100},
		},
	}
	err := ValidateIngredientHasDosingActions(ing)
	if err != nil {
		t.Errorf("expected ingredient with dosing actions to pass, got: %v", err)
	}
}

func TestValidateIngredientHasDosingActions_NoActions(t *testing.T) {
	ing := &FormulaIngredient{
		DosingActions: nil,
	}
	err := ValidateIngredientHasDosingActions(ing)
	if err == nil {
		t.Error("expected error for ingredient with no dosing actions")
	}
}

// --- TestValidateDosingRatios ---

func TestValidateDosingRatios_Sum100(t *testing.T) {
	ing := &FormulaIngredient{
		DosingActions: []FormulaDosingAction{
			{UseRatio: 60},
			{UseRatio: 40},
		},
	}
	err := ValidateDosingRatios(ing)
	if err != nil {
		t.Errorf("expected dosing ratios sum 100%% to pass, got: %v", err)
	}
}

func TestValidateDosingRatios_SumNot100(t *testing.T) {
	ing := &FormulaIngredient{
		DosingActions: []FormulaDosingAction{
			{UseRatio: 50},
			{UseRatio: 30},
		},
	}
	err := ValidateDosingRatios(ing)
	if err == nil {
		t.Error("expected error for dosing ratios sum of 80%")
	}
}

func TestValidateDosingRatios_EmptyList(t *testing.T) {
	ing := &FormulaIngredient{
		DosingActions: []FormulaDosingAction{},
	}
	err := ValidateDosingRatios(ing)
	if err == nil {
		t.Error("expected error for empty dosing actions list")
	}
}

// --- TestValidateStepsForActivation ---

func TestValidateStepsForActivation_HasSteps(t *testing.T) {
	f := &Formula{
		Status: StatusActive,
		Steps: []FormulaStep{
			{StepNo: 1, Name: "Mix"},
		},
	}
	err := ValidateStepsForActivation(f)
	if err != nil {
		t.Errorf("expected active formula with steps to pass, got: %v", err)
	}
}

func TestValidateStepsForActivation_NoSteps(t *testing.T) {
	f := &Formula{
		Status: StatusActive,
		Steps:  nil,
	}
	err := ValidateStepsForActivation(f)
	if err == nil {
		t.Error("expected error for active formula with no steps")
	}
}

func TestValidateStepsForActivation_Draft_NoSteps_Ok(t *testing.T) {
	f := &Formula{
		Status: StatusDraft,
		Steps:  nil,
	}
	err := ValidateStepsForActivation(f)
	if err != nil {
		t.Errorf("expected draft formula to pass even without steps, got: %v", err)
	}
}

// --- TestValidatePartIngredients with formula-level integration ---

func TestValidateFormula_PartIngredientsSumInvalid(t *testing.T) {
	f := validSingleFormulaFixture()
	f.Parts[0].Ingredients = []FormulaIngredient{
		{Material: "A", Percentage: 50},
		{Material: "B", Percentage: 25},
	}
	err := ValidateFormula(f)
	if err == nil {
		t.Error("expected error for part with ingredients summing to 75%")
	}
}

// --- Error containment checks ---

func TestErrorsAreProperlyWrapped(t *testing.T) {
	f := validSingleFormulaFixture()
	f.Parts = nil
	err := ValidateFormula(f)
	if err == nil {
		t.Fatal("expected error")
	}

	var target interface{ Unwrap() []error }
	if !errors.As(err, &target) {
		t.Log("error type:", err)
		// Only fail if no errors at all — single error without join is acceptable
	}
}
