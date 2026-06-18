export type ComponentMode = 'single' | 'double';
export type FormulaStatus = 'draft' | 'active' | 'archived';
export type FormulaType = 'formula' | 'material';
export type PartName = 'PartA' | 'PartB' | 'PartMain';

export interface FormulaStepMaterial {
  id: string;
  category_id: string;
  material: string;
  percentage: number;
  weight: number;
  batch_no: string;
  unit: string;
  sort_order: number;
}

export interface FormulaStepMaterialCategory {
  id: string;
  step_id: string;
  name: string;
  sort_order: number;
  materials: FormulaStepMaterial[];
}

export interface FormulaStepParameter {
  id: string;
  step_id: string;
  name: string;
  value: string;
  unit: string;
  sort_order: number;
}

export interface FormulaStep {
  id: string;
  formula_id: string;
  part_id: string | null;
  step_no: number;
  name: string;
  description: string;
  instrument_name: string;
  temperature: string;
  duration: string;
  categories: FormulaStepMaterialCategory[];
  parameters: FormulaStepParameter[];
}

export interface FormulaPart {
  id: string;
  formula_id: string;
  name: PartName;
  sort_order: number;
  material_id: string | null;
  batch_no: string;
  categories: FormulaIngredientCategory[];
}

export interface FormulaIngredientCategory {
  id: string;
  part_id: string;
  name: string;
  sort_order: number;
  ingredients: FormulaIngredient[];
}

export interface FormulaIngredient {
  id: string;
  category_id: string;
  material: string;
  percentage: number;
  weight: number;
  batch_no: string;
  unit: string;
  sort_order: number;
}

export interface Formula {
  id: string;
  name: string;
  code: string;
  component_mode: ComponentMode;
  status: FormulaStatus;
  formula_type: FormulaType;
  labels: string[];
  parts: FormulaPart[];
  steps: FormulaStep[];
  created_at: string;
  updated_at: string;
}

export interface FormulaMatrix {
  formulas: Formula[];
}



