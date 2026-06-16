// ─── Enums / Union Types matching Go domain models ───

/** How components are organized in a formula (mirrors Go ComponentMode). */
export type ComponentMode = 'single' | 'double';

/** Lifecycle state of a formula (mirrors Go Status). */
export type FormulaStatus = 'draft' | 'active' | 'archived';

/** Standard part role in a formula (mirrors Go PartName). */
export type PartName = 'PartA' | 'PartB' | 'PartMain';

// ─── Domain Interfaces ───

/** Dosing action that ties a step to an ingredient (mirrors Go FormulaDosingAction). */
export interface FormulaDosingAction {
  id: string;
  step_id: string;
  ingredient_id: string;
  dosing_order: number;
  use_ratio: number;
  dosing_method: string;
}

/** Material and its percentage in a formula part (mirrors Go FormulaIngredient). */
export interface FormulaIngredient {
  id: string;
  part_id: string;
  sort_order: number;
  material: string;
  percentage: number;
  weight: number;
  dosing_actions: FormulaDosingAction[];
}

/** Process step in a formula (mirrors Go FormulaStep). PartID is nullable. */
export interface FormulaStep {
  id: string;
  formula_id: string;
  part_id: string | null;
  step_no: number;
  name: string;
  temperature: string;
  duration: string;
}

/** Component part (A, B, or Main) of a formula (mirrors Go FormulaPart). */
export interface FormulaPart {
  id: string;
  formula_id: string;
  name: PartName;
  mix_ratio: number;
  sort_order: number;
  ingredients: FormulaIngredient[];
}

/** Root domain entity representing a complete formula (mirrors Go Formula). */
export interface Formula {
  id: string;
  name: string;
  code: string;
  component_mode: ComponentMode;
  status: FormulaStatus;
  parts: FormulaPart[];
  steps: FormulaStep[];
  created_at: string;
  updated_at: string;
}

/** Matrix view aggregating multiple formulas (mirrors formula matrix endpoint). */
export interface FormulaMatrix {
  formulas: Formula[];
}
