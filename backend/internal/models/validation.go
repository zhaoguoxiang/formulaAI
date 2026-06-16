package models

import (
	"errors"
	"fmt"
	"math"
)

// percentageTolerance defines the acceptable deviation from 100%.
const percentageTolerance = 0.001

// ValidateFormula runs all domain validation rules against a formula.
// It returns the first error found or nil if the formula is valid.
func ValidateFormula(f *Formula) error {
	if f == nil {
		return errors.New("formula is nil")
	}

	if err := validatePartsMatchMode(f); err != nil {
		return err
	}

	for i := range f.Parts {
		if err := ValidatePartIngredients(&f.Parts[i]); err != nil {
			return fmt.Errorf("part %s: %w", f.Parts[i].Name, err)
		}
	}

	if err := ValidateStepsForActivation(f); err != nil {
		return err
	}

	return nil
}

// validatePartsMatchMode checks that the formula's parts match its component mode.
func validatePartsMatchMode(f *Formula) error {
	if len(f.Parts) == 0 {
		return errors.New("formula must have at least one part")
	}

	switch f.ComponentMode {
	case ComponentModeSingle:
		if len(f.Parts) != 1 || f.Parts[0].Name != PartMain {
			return fmt.Errorf("single component mode requires exactly one part named %q, got %d part(s)", PartMain, len(f.Parts))
		}
	case ComponentModeDouble:
		hasA, hasB := false, false
		for _, p := range f.Parts {
			switch p.Name {
			case PartA:
				hasA = true
			case PartB:
				hasB = true
			default:
				return fmt.Errorf("double component mode only allows %q and %q parts, got %q", PartA, PartB, p.Name)
			}
		}
		if !hasA {
			return fmt.Errorf("double component mode requires a %q part", PartA)
		}
		if !hasB {
			return fmt.Errorf("double component mode requires a %q part", PartB)
		}
	default:
		return fmt.Errorf("unknown component mode: %q", f.ComponentMode)
	}

	return nil
}

// ValidatePartIngredients checks that ingredient percentages sum to 100%
// within percentageTolerance.
func ValidatePartIngredients(part *FormulaPart) error {
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
		return fmt.Errorf("ingredient percentages sum to %.4f%%, expected 100%% (tolerance: ±%.3f%%)", sum, percentageTolerance)
	}

	return nil
}

// ValidateIngredientHasDosingActions checks that an ingredient has at least
// one dosing action associated with it.
func ValidateIngredientHasDosingActions(ingredient *FormulaIngredient) error {
	if ingredient == nil {
		return errors.New("ingredient is nil")
	}

	if len(ingredient.DosingActions) == 0 {
		return fmt.Errorf("ingredient %q must have at least one dosing action", ingredient.Material)
	}

	return nil
}

// ValidateDosingRatios checks that an ingredient's dosing action use_ratios
// sum to 100% within percentageTolerance.
func ValidateDosingRatios(ingredient *FormulaIngredient) error {
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
		return fmt.Errorf("dosing action use_ratios sum to %.4f%%, expected 100%% (tolerance: ±%.3f%%)", sum, percentageTolerance)
	}

	return nil
}

// ValidateStepsForActivation checks that an active formula has at least one step.
// Draft and archived formulas are allowed to have no steps.
func ValidateStepsForActivation(f *Formula) error {
	if f == nil {
		return errors.New("formula is nil")
	}

	if f.Status == StatusActive && len(f.Steps) == 0 {
		return errors.New("active formula must have at least one step")
	}

	return nil
}
