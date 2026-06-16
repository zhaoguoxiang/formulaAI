package services

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"formula-ai-system/backend/internal/models"
)

const percentageTolerance = 0.001

// ---------------------------------------------------------------------------
// ValidationError – structured result that distinguishes blocking errors
// from advisory warnings.
// ---------------------------------------------------------------------------

// ValidationError collects all issues found during formula validation.
// Errors block a save; warnings are advisory and do not block.
type ValidationError struct {
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// HasErrors returns true when at least one blocking error is present.
func (ve *ValidationError) HasErrors() bool { return len(ve.Errors) > 0 }

// HasWarnings returns true when at least one advisory warning is present.
func (ve *ValidationError) HasWarnings() bool { return len(ve.Warnings) > 0 }

// IsBlocking returns true when the formula must not be saved.
func (ve *ValidationError) IsBlocking() bool { return ve.HasErrors() }

// Error implements the error interface and renders both errors and warnings.
func (ve *ValidationError) Error() string {
	var parts []string
	if len(ve.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("validation errors (%d): %s",
			len(ve.Errors), strings.Join(ve.Errors, "; ")))
	}
	if len(ve.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("validation warnings (%d): %s",
			len(ve.Warnings), strings.Join(ve.Warnings, "; ")))
	}
	return strings.Join(parts, " | ")
}

// ---------------------------------------------------------------------------
// Master entry point
// ---------------------------------------------------------------------------

// ValidateAndPrepare runs all business-level validation rules against a
// formula. It returns nil when the formula is fully valid (no errors, no
// warnings).  When issues exist it returns a *ValidationError whose
// .Errors slice contains blocking problems and .Warnings contains advisory
// notes that do not prevent saving.
//
// This service layer wraps the basic field-level checks in
// internal/models/validation.go with additional business context and
// aggregates all issues into a single structured result.
func ValidateAndPrepare(f *models.Formula) error {
	if f == nil {
		return &ValidationError{
			Errors: []string{"formula is nil"},
		}
	}

	result := &ValidationError{}

	// 1 ─ Parts match the declared component mode.
	if err := validatePartsMatchMode(f); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// 2 ─ Every part's ingredient percentages must sum to 100 %.
	for i := range f.Parts {
		if err := validatePartIngredients(&f.Parts[i]); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("part %q: %s", f.Parts[i].Name, err.Error()))
		}
	}

	// 3 ─ Every ingredient must be referenced by at least one dosing action.
	if err := validateDosingCompleteness(f); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// 4 ─ Every ingredient's dosing use-ratios must sum to 100 %.
	for pi := range f.Parts {
		for ii := range f.Parts[pi].Ingredients {
			if err := validateDosingRatios(&f.Parts[pi].Ingredients[ii]); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("ingredient %q (part %q): %s",
						f.Parts[pi].Ingredients[ii].Material,
						f.Parts[pi].Name,
						err.Error()))
			}
		}
	}

	// 5 ─ Active formulas must have at least one step.
	if err := validateStepsForActivation(f); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// ── Warnings (advisory, do NOT block save) ──

	if f.Status == models.StatusDraft && len(f.Steps) == 0 {
		result.Warnings = append(result.Warnings,
			"draft formula has no steps; add at least one step before activating")
	}

	if result.HasErrors() || result.HasWarnings() {
		return result
	}
	return nil
}

// ---------------------------------------------------------------------------
// Rule 1 – Parts must match component mode
// ---------------------------------------------------------------------------

// validatePartsMatchMode checks that the formula's parts are consistent with
// its ComponentMode.
//
//	Single → exactly 1 part named PartMain.
//	Double → must contain both PartA and PartB (no other part names).
func validatePartsMatchMode(f *models.Formula) error {
	if len(f.Parts) == 0 {
		return errors.New("formula must have at least one part")
	}

	switch f.ComponentMode {
	case models.ComponentModeSingle:
		if len(f.Parts) != 1 || f.Parts[0].Name != models.PartMain {
			return fmt.Errorf("single component mode requires exactly one part named %q, got %d part(s)",
				models.PartMain, len(f.Parts))
		}
	case models.ComponentModeDouble:
		hasA, hasB := false, false
		for _, p := range f.Parts {
			switch p.Name {
			case models.PartA:
				hasA = true
			case models.PartB:
				hasB = true
			default:
				return fmt.Errorf("double component mode only allows %q and %q parts, got %q",
					models.PartA, models.PartB, p.Name)
			}
		}
		if !hasA {
			return fmt.Errorf("double component mode requires a %q part", models.PartA)
		}
		if !hasB {
			return fmt.Errorf("double component mode requires a %q part", models.PartB)
		}
	default:
		return fmt.Errorf("unknown component mode: %q", f.ComponentMode)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Rule 2 – Ingredient percentages sum to 100 %
// ---------------------------------------------------------------------------

// validatePartIngredients checks that the percentages of all ingredients
// within a single part sum to 100 % (tolerance ±0.001).
func validatePartIngredients(part *models.FormulaPart) error {
	if part == nil {
		return errors.New("part is nil")
	}
	if len(part.Ingredients) == 0 {
		return errors.New("part must have at least one ingredient")
	}

	var sum float64
	for _, ing := range part.Ingredients {
		sum += ing.Percentage
	}

	if math.Abs(sum-100.0) > percentageTolerance {
		return fmt.Errorf("ingredient percentages sum to %.4f%%, expected 100%% (tolerance: ±%.3f%%)",
			sum, percentageTolerance)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Rule 3 – Every ingredient must have at least one dosing action
// ---------------------------------------------------------------------------

// validateDosingCompleteness verifies that every ingredient across all parts
// is referenced by at least one FormulaDosingAction.  It returns a single
// error listing all offending ingredients.
func validateDosingCompleteness(f *models.Formula) error {
	var missing []string

	for pi := range f.Parts {
		for ii := range f.Parts[pi].Ingredients {
			ing := &f.Parts[pi].Ingredients[ii]
			if len(ing.DosingActions) == 0 {
				missing = append(missing,
					fmt.Sprintf("%q in part %q", ing.Material, f.Parts[pi].Name))
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("the following ingredients have no dosing actions: %s",
			strings.Join(missing, ", "))
	}

	return nil
}

// ---------------------------------------------------------------------------
// Rule 4 – Dosing use-ratios sum to 100 %
// ---------------------------------------------------------------------------

// validateDosingRatios checks that the UseRatio values of all dosing actions
// belonging to a single ingredient sum to 100 % (tolerance ±0.001).
func validateDosingRatios(ingredient *models.FormulaIngredient) error {
	if ingredient == nil {
		return errors.New("ingredient is nil")
	}
	if len(ingredient.DosingActions) == 0 {
		return errors.New("ingredient must have at least one dosing action to validate ratios")
	}

	var sum float64
	for _, da := range ingredient.DosingActions {
		sum += da.UseRatio
	}

	if math.Abs(sum-100.0) > percentageTolerance {
		return fmt.Errorf("dosing use_ratios sum to %.4f%%, expected 100%% (tolerance: ±%.3f%%)",
			sum, percentageTolerance)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Rule 5 – Active formulas require at least one step
// ---------------------------------------------------------------------------

// validateStepsForActivation ensures that a formula with Status == Active
// has at least one FormulaStep.  Draft and Archived formulas are allowed to
// have zero steps.
func validateStepsForActivation(f *models.Formula) error {
	if f == nil {
		return errors.New("formula is nil")
	}

	if f.Status == models.StatusActive && len(f.Steps) == 0 {
		return errors.New("active formula must have at least one step")
	}

	return nil
}
