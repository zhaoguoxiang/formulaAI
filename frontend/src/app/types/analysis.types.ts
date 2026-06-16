/** Response shape for chart-based analysis endpoints (mode-ratio, ingredient-distribution, step-distribution). */
export interface ChartAnalysis {
  labels: string[];
  values: number[];
}

/** Response shape for the dosing-methods analysis endpoint. */
export interface DosingMethod {
  name: string;
  count: number;
}

export interface DosingMethodAnalysis {
  methods: DosingMethod[];
}
