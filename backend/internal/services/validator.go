package services

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"formula-ai-system/backend/internal/models"
)

const percentageTolerance = 0.001

type ValidationError struct {
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

func (ve *ValidationError) HasErrors() bool   { return len(ve.Errors) > 0 }
func (ve *ValidationError) HasWarnings() bool { return len(ve.Warnings) > 0 }
func (ve *ValidationError) IsBlocking() bool  { return ve.HasErrors() }
func (ve *ValidationError) Error() string {
	var parts []string
	if len(ve.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("validation errors (%d): %s", len(ve.Errors), strings.Join(ve.Errors, "; ")))
	}
	if len(ve.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("validation warnings (%d): %s", len(ve.Warnings), strings.Join(ve.Warnings, "; ")))
	}
	return strings.Join(parts, " | ")
}

func ValidateAndPrepare(f *models.Formula) error {
	if f == nil {
		return &ValidationError{Errors: []string{"formula is nil"}}
	}

	result := &ValidationError{}

	if err := validatePartsMatchMode(f); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// Validate materials in 投料 steps
	for si := range f.Steps {
		if err := validateStepMaterials(&f.Steps[si]); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("step %d %q: %s", f.Steps[si].StepNo, f.Steps[si].Name, err.Error()))
		}
	}

	if err := validateStepsForActivation(f); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	if f.Status == models.StatusDraft && len(f.Steps) == 0 {
		result.Warnings = append(result.Warnings, "draft formula has no steps; add at least one step before activating")
	}

	if result.HasErrors() || result.HasWarnings() {
		return result
	}
	return nil
}

func validatePartsMatchMode(f *models.Formula) error {
	if len(f.Parts) == 0 {
		return errors.New("formula must have at least one part")
	}
	switch f.ComponentMode {
	case models.ComponentModeSingle:
		if len(f.Parts) != 1 || f.Parts[0].Name != models.PartMain {
			return fmt.Errorf("single component mode requires exactly one part named %q, got %d part(s)", models.PartMain, len(f.Parts))
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
				return fmt.Errorf("double component mode only allows %q and %q parts, got %q", models.PartA, models.PartB, p.Name)
			}
		}
		if !hasA {
			return fmt.Errorf("double component mode requires a %q part", models.PartA)
		}
		if !hasB {
			return fmt.Errorf("double component mode requires a %q part", models.PartB)
		}
	}
	return nil
}

func validateStepMaterials(step *models.FormulaStep) error {
	if len(step.Categories) == 0 {
		return nil // Not a 投料 step
	}
	var sum float64
	for ci := range step.Categories {
		for _, mat := range step.Categories[ci].Materials {
			sum += mat.Percentage
		}
	}
	if sum > 0 && math.Abs(sum-100.0) > percentageTolerance {
		return fmt.Errorf("material percentages sum to %.4f%%, expected 100%%", sum)
	}
	return nil
}

func validateStepsForActivation(f *models.Formula) error {
	if f.Status == models.StatusActive && len(f.Steps) == 0 {
		return errors.New("active formula must have at least one step")
	}
	return nil
}
