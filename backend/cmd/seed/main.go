package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const defaultBaseURL = "http://localhost:8080"

// ─────────────────────────────────────────────────────────────────────────────
// API payload types (subset of models package for HTTP seeding)
// ─────────────────────────────────────────────────────────────────────────────

// formulaPayload is the JSON body sent to POST /api/formulas.
type formulaPayload struct {
	Name          string        `json:"name"`
	Code          string        `json:"code"`
	ComponentMode string        `json:"component_mode"`
	Status        string        `json:"status"`
	Parts         []partPayload `json:"parts"`
	Steps         []stepPayload `json:"steps"`
}

type partPayload struct {
	Name        string              `json:"name"`
	MixRatio    float64             `json:"mix_ratio"`
	SortOrder   int                 `json:"sort_order"`
	Ingredients []ingredientPayload `json:"ingredients"`
}

type ingredientPayload struct {
	SortOrder     int                   `json:"sort_order"`
	Material      string                `json:"material"`
	Percentage    float64               `json:"percentage"`
	DosingActions []dosingActionPayload `json:"dosing_actions"`
}

type dosingActionPayload struct {
	StepID       string  `json:"step_id"`
	DosingOrder  int     `json:"dosing_order"`
	UseRatio     float64 `json:"use_ratio"`
	DosingMethod string  `json:"dosing_method"`
}

type stepPayload struct {
	ID          string `json:"id"` // client-generated UUID so dosing actions can reference it
	StepNo      int    `json:"step_no"`
	Name        string `json:"name"`
	Temperature string `json:"temperature"`
	Duration    string `json:"duration"`
}

// formulaSummary is the minimal shape used when reading GET /api/formulas to
// check which codes already exist (idempotency guard).
type formulaSummary struct {
	Code string `json:"code"`
}

// ─────────────────────────────────────────────────────────────────────────────
// main
// ─────────────────────────────────────────────────────────────────────────────

func main() {
	baseURL := getEnv("API_BASE_URL", defaultBaseURL)

	// 1. Health check – make sure the API is reachable.
	if err := checkHealth(baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "API not reachable at %s: %v\n", baseURL, err)
		fmt.Fprintln(os.Stderr, "Make sure the backend server is running before seeding.")
		os.Exit(1)
	}

	// 2. Fetch existing formula codes for idempotency.
	existingCodes, err := fetchExistingCodes(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch existing formulas: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Existing formulas in system: %d\n", len(existingCodes))

	// 3. Build the three seed formulas.
	seeds := buildSeedFormulas()

	// 4. Create formulas whose code does not already exist.
	created := 0
	for _, f := range seeds {
		if existingCodes[f.Code] {
			fmt.Printf("SKIP: %s (%s) – already exists\n", f.Code, f.Name)
			continue
		}
		if err := createFormula(baseURL, f); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %s – %v\n", f.Code, err)
			continue
		}
		fmt.Printf("OK:   %s (%s)\n", f.Code, f.Name)
		created++
	}

	// 5. Final verification.
	finalCodes, err := fetchExistingCodes(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Verification fetch failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n─── Verification ───")
	fmt.Printf("Total formulas in system: %d\n", len(finalCodes))
	if len(finalCodes) >= 3 {
		fmt.Println("✓ PASS: At least 3 formulas present.")
	} else {
		fmt.Printf("✗ FAIL: Expected ≥ 3 formulas, got %d\n", len(finalCodes))
		os.Exit(1)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HTTP helpers
// ─────────────────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func checkHealth(baseURL string) error {
	resp, err := http.Get(baseURL + "/api/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned %d", resp.StatusCode)
	}
	return nil
}

// fetchExistingCodes returns a set of formula codes that already exist in the
// system (used for idempotency: skip seeding a code that is already present).
func fetchExistingCodes(baseURL string) (map[string]bool, error) {
	resp, err := http.Get(baseURL + "/api/formulas")
	if err != nil {
		return nil, fmt.Errorf("GET /api/formulas: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/formulas returned %d", resp.StatusCode)
	}

	var formulas []formulaSummary
	if err := json.NewDecoder(resp.Body).Decode(&formulas); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	codes := make(map[string]bool, len(formulas))
	for _, f := range formulas {
		codes[f.Code] = true
	}
	return codes, nil
}

// createFormula POSTs a formula payload to the API. Returns nil on 201 Created.
func createFormula(baseURL string, f formulaPayload) error {
	body, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := http.Post(
		baseURL+"/api/formulas",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("POST /api/formulas: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("server returned %d: %v", resp.StatusCode, errBody)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Seed data
// ─────────────────────────────────────────────────────────────────────────────

func buildSeedFormulas() []formulaPayload {
	// ── Formula 1: Single-component Silicone ──────────────────────────────
	s1Steps := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	f1 := formulaPayload{
		Name:          "Silicone Sealant Formula",
		Code:          "FML-SL-001",
		ComponentMode: "single",
		Status:        "draft",
		Parts: []partPayload{
			{
				Name:      "PartMain",
				MixRatio:  100,
				SortOrder: 1,
				Ingredients: []ingredientPayload{
					{SortOrder: 1, Material: "Silicone Base Polymer", Percentage: 60,
						DosingActions: []dosingActionPayload{
							{StepID: s1Steps[0].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "metered"},
						}},
					{SortOrder: 2, Material: "Crosslinker", Percentage: 15,
						DosingActions: []dosingActionPayload{
							{StepID: s1Steps[0].String(), DosingOrder: 2, UseRatio: 100, DosingMethod: "metered"},
						}},
					{SortOrder: 3, Material: "Filler (Calcium Carbonate)", Percentage: 20,
						DosingActions: []dosingActionPayload{
							{StepID: s1Steps[1].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "gravimetric"},
						}},
					{SortOrder: 4, Material: "Catalyst", Percentage: 3,
						DosingActions: []dosingActionPayload{
							{StepID: s1Steps[0].String(), DosingOrder: 3, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 5, Material: "Adhesion Promoter", Percentage: 1,
						DosingActions: []dosingActionPayload{
							{StepID: s1Steps[0].String(), DosingOrder: 4, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 6, Material: "Plasticizer", Percentage: 1,
						DosingActions: []dosingActionPayload{
							{StepID: s1Steps[1].String(), DosingOrder: 2, UseRatio: 100, DosingMethod: "gravimetric"},
						}},
				},
			},
		},
		Steps: []stepPayload{
			{ID: s1Steps[0].String(), StepNo: 1, Name: "Raw Material Charging", Temperature: "25°C", Duration: "30 min"},
			{ID: s1Steps[1].String(), StepNo: 2, Name: "Mixing & Dispersion", Temperature: "40°C", Duration: "60 min"},
			{ID: s1Steps[2].String(), StepNo: 3, Name: "Deaeration", Temperature: "25°C", Duration: "20 min"},
		},
	}

	// ── Formula 2: Double-component Epoxy ─────────────────────────────────
	eSteps := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	f2 := formulaPayload{
		Name:          "Two-Component Epoxy Resin",
		Code:          "FML-EP-001",
		ComponentMode: "double",
		Status:        "draft",
		Parts: []partPayload{
			{
				Name:      "PartA",
				MixRatio:  100,
				SortOrder: 1,
				Ingredients: []ingredientPayload{
					{SortOrder: 1, Material: "Bisphenol A Epoxy Resin", Percentage: 80,
						DosingActions: []dosingActionPayload{
							{StepID: eSteps[0].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "metered"},
						}},
					{SortOrder: 2, Material: "Reactive Diluent", Percentage: 20,
						DosingActions: []dosingActionPayload{
							{StepID: eSteps[0].String(), DosingOrder: 2, UseRatio: 100, DosingMethod: "metered"},
						}},
				},
			},
			{
				Name:      "PartB",
				MixRatio:  50,
				SortOrder: 2,
				Ingredients: []ingredientPayload{
					{SortOrder: 1, Material: "Polyamine Hardener", Percentage: 70,
						DosingActions: []dosingActionPayload{
							{StepID: eSteps[1].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "metered"},
						}},
					{SortOrder: 2, Material: "Accelerator", Percentage: 15,
						DosingActions: []dosingActionPayload{
							{StepID: eSteps[1].String(), DosingOrder: 2, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 3, Material: "Flexibilizer", Percentage: 15,
						DosingActions: []dosingActionPayload{
							{StepID: eSteps[1].String(), DosingOrder: 3, UseRatio: 100, DosingMethod: "manual"},
						}},
				},
			},
		},
		Steps: []stepPayload{
			{ID: eSteps[0].String(), StepNo: 1, Name: "Part A Preparation (Resin Mixing)", Temperature: "40°C", Duration: "45 min"},
			{ID: eSteps[1].String(), StepNo: 2, Name: "Part B Preparation (Hardener Blend)", Temperature: "25°C", Duration: "30 min"},
			{ID: eSteps[2].String(), StepNo: 3, Name: "Combined Mixing", Temperature: "25°C", Duration: "15 min"},
			{ID: eSteps[3].String(), StepNo: 4, Name: "Vacuum Degassing", Temperature: "25°C", Duration: "10 min"},
		},
	}

	// ── Formula 3: Single-component Coating ───────────────────────────────
	cSteps := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	f3 := formulaPayload{
		Name:          "Water-Based Acrylic Coating",
		Code:          "FML-CT-001",
		ComponentMode: "single",
		Status:        "draft",
		Parts: []partPayload{
			{
				Name:      "PartMain",
				MixRatio:  100,
				SortOrder: 1,
				Ingredients: []ingredientPayload{
					{SortOrder: 1, Material: "Acrylic Emulsion", Percentage: 50,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[0].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "metered"},
						}},
					{SortOrder: 2, Material: "Titanium Dioxide", Percentage: 20,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[1].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "gravimetric"},
						}},
					{SortOrder: 3, Material: "Calcium Carbonate Filler", Percentage: 15,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[1].String(), DosingOrder: 2, UseRatio: 100, DosingMethod: "gravimetric"},
						}},
					{SortOrder: 4, Material: "Coalescing Agent", Percentage: 5,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[0].String(), DosingOrder: 2, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 5, Material: "Thickener", Percentage: 3,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[2].String(), DosingOrder: 1, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 6, Material: "Defoamer", Percentage: 1,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[0].String(), DosingOrder: 3, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 7, Material: "Biocide", Percentage: 1,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[0].String(), DosingOrder: 4, UseRatio: 100, DosingMethod: "manual"},
						}},
					{SortOrder: 8, Material: "Water", Percentage: 5,
						DosingActions: []dosingActionPayload{
							{StepID: cSteps[0].String(), DosingOrder: 5, UseRatio: 100, DosingMethod: "metered"},
						}},
				},
			},
		},
		Steps: []stepPayload{
			{ID: cSteps[0].String(), StepNo: 1, Name: "Pre-Mix Liquid Components", Temperature: "25°C", Duration: "20 min"},
			{ID: cSteps[1].String(), StepNo: 2, Name: "Pigment Dispersion", Temperature: "35°C", Duration: "45 min"},
			{ID: cSteps[2].String(), StepNo: 3, Name: "Let-Down & Adjustment", Temperature: "25°C", Duration: "15 min"},
			{ID: cSteps[3].String(), StepNo: 4, Name: "Filtration", Temperature: "25°C", Duration: "10 min"},
		},
	}

	return []formulaPayload{f1, f2, f3}
}
