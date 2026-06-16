import { Component, Input, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  ReactiveFormsModule,
  FormBuilder,
  FormGroup,
  FormArray,
  FormControl,
  Validators,
  AbstractControl,
} from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Subscription, combineLatest, map, startWith } from 'rxjs';

import { FormulaApiService } from '../../services/formula-api.service';
import {
  Formula,
  FormulaPart,
  FormulaIngredient,
  FormulaStep,
  FormulaDosingAction,
  ComponentMode,
  FormulaStatus,
  PartName,
} from '../../types/formula.types';

// ─── Helper ───

function generateTempId(): string {
  return 'tmp_' + Math.random().toString(36).substring(2, 10);
}

// ─── Component ───

@Component({
  selector: 'app-formula-editor',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatExpansionModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatProgressBarModule,
    MatSnackBarModule,
    MatDividerModule,
    MatTooltipModule,
  ],
  templateUrl: './formula-editor.component.html',
  styleUrl: './formula-editor.component.css',
})
export class FormulaEditorComponent implements OnInit, OnDestroy {
  @Input() formulaId: string | null = null;
  @Output() saved = new EventEmitter<Formula>();
  @Output() cancelled = new EventEmitter<void>();

  form!: FormGroup;
  loading = false;
  isEditMode = false;

  readonly componentModes: ComponentMode[] = ['single', 'double'];
  readonly statuses: FormulaStatus[] = ['draft', 'active', 'archived'];
  readonly partNames: PartName[] = ['PartA', 'PartB', 'PartMain'];

  // Percentage sum streams per part → observable of number
  percentageSums: Map<number, number> = new Map();

  private subs = new Subscription();

  constructor(
    private readonly fb: FormBuilder,
    private readonly api: FormulaApiService,
    private readonly snackBar: MatSnackBar,
  ) {}

  // ─── Lifecycle ───

  ngOnInit(): void {
    this.buildForm();

    if (this.formulaId) {
      this.isEditMode = true;
      this.loadFormula(this.formulaId);
    } else {
      // Create mode: seed one default MAIN part with two empty ingredients
      this.addPart('PartMain', 100);
    }
  }

  ngOnDestroy(): void {
    this.subs.unsubscribe();
  }

  // ─── Form Construction ───

  private buildForm(): void {
    this.form = this.fb.group({
      name: ['', [Validators.required, Validators.maxLength(200)]],
      code: ['', [Validators.required, Validators.maxLength(100)]],
      component_mode: ['single', Validators.required],
      status: ['draft', Validators.required],
      parts: this.fb.array([]),
      steps: this.fb.array([]),
    });
  }

  // ─── Getters for FormArrays ───

  get parts(): FormArray {
    return this.form.get('parts') as FormArray;
  }

  get steps(): FormArray {
    return this.form.get('steps') as FormArray;
  }

  ingredients(partIndex: number): FormArray {
    return this.parts.at(partIndex).get('ingredients') as FormArray;
  }

  dosingActions(partIndex: number, ingredientIndex: number): FormArray {
    return this.ingredients(partIndex).at(ingredientIndex).get('dosing_actions') as FormArray;
  }

  stepControls(): AbstractControl[] {
    return this.steps.controls;
  }

  // ─── Add / Remove ───

  addPart(name: PartName = 'PartMain', mixRatio = 100): void {
    const partGroup = this.fb.group({
      name: [name, Validators.required],
      mix_ratio: [mixRatio, [Validators.required, Validators.min(0), Validators.max(100)]],
      sort_order: [this.parts.length],
      ingredients: this.fb.array([]),
    });

    const partIdx = this.parts.length;
    this.parts.push(partGroup);

    // Track percentage sum for this part
    const ingArr = partGroup.get('ingredients') as FormArray;
    this.subs.add(
      ingArr.valueChanges.subscribe(() => {
        this.recalcPercentageSum(partIdx, ingArr);
      }),
    );

    // Initial calculation
    this.recalcPercentageSum(partIdx, ingArr);
  }

  removePart(index: number): void {
    this.parts.removeAt(index);
    this.percentageSums.delete(index);
    // Re-index percentage sums
    const newMap = new Map<number, number>();
    this.percentageSums.forEach((val, key) => {
      if (key > index) newMap.set(key - 1, val);
      else if (key < index) newMap.set(key, val);
    });
    this.percentageSums = newMap;
  }

  addIngredient(partIndex: number): void {
    const ingGroup = this.fb.group({
      material: ['', [Validators.required, Validators.maxLength(200)]],
      percentage: [0, [Validators.required, Validators.min(0), Validators.max(100)]],
      weight: [0, [Validators.min(0)]],
      dosing_actions: this.fb.array([]),
    });
    const ingArr = this.ingredients(partIndex);
    ingArr.push(ingGroup);
  }

  removeIngredient(partIndex: number, ingredientIndex: number): void {
    this.ingredients(partIndex).removeAt(ingredientIndex);
  }

  addStep(): void {
    const stepGroup = this.fb.group({
      step_no: [this.steps.length + 1, [Validators.required, Validators.min(1)]],
      name: ['', [Validators.required, Validators.maxLength(200)]],
      temperature: ['', Validators.maxLength(50)],
      duration: ['', Validators.maxLength(50)],
    });
    this.steps.push(stepGroup);
  }

  removeStep(index: number): void {
    this.steps.removeAt(index);
    // Re-number remaining steps
    this.steps.controls.forEach((ctrl, i) => {
      ctrl.get('step_no')?.setValue(i + 1, { emitEvent: false });
    });
  }

  addDosingAction(partIndex: number, ingredientIndex: number): void {
    const daGroup = this.fb.group({
      // Store step index reference (-1 = unselected)
      step_ref: [-1, Validators.min(-1)],
      use_ratio: [100, [Validators.required, Validators.min(0), Validators.max(100)]],
      dosing_order: [1, [Validators.required, Validators.min(1)]],
      dosing_method: ['', Validators.maxLength(100)],
    });
    this.dosingActions(partIndex, ingredientIndex).push(daGroup);
  }

  removeDosingAction(partIndex: number, ingredientIndex: number, daIndex: number): void {
    this.dosingActions(partIndex, ingredientIndex).removeAt(daIndex);
  }

  // ─── Percentage Sum ───

  private recalcPercentageSum(partIndex: number, ingArr: FormArray): void {
    const sum = ingArr.controls.reduce((acc, ctrl) => {
      const pct = ctrl.get('percentage')?.value ?? 0;
      return acc + Number(pct);
    }, 0);
    this.percentageSums.set(partIndex, sum);
  }

  getPercentageSum(partIndex: number): number {
    return this.percentageSums.get(partIndex) ?? 0;
  }

  isPercentageValid(partIndex: number): boolean {
    const sum = this.getPercentageSum(partIndex);
    return Math.abs(sum - 100) <= 0.1;
  }

  hasIngredients(partIndex: number): boolean {
    return this.ingredients(partIndex).length > 0;
  }

  // ─── Data Loading (Edit Mode) ───

  private loadFormula(id: string): void {
    this.loading = true;
    this.api.getFormula(id).subscribe({
      next: (formula) => {
        this.populateForm(formula);
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.snackBar.open(`加载配方失败: ${err.message || '未知错误'}`, '关闭', { duration: 5000 });
      },
    });
  }

  private populateForm(formula: Formula): void {
    this.form.patchValue({
      name: formula.name,
      code: formula.code,
      component_mode: formula.component_mode,
      status: formula.status,
    });

    // Clear defaults added in create mode
    while (this.parts.length) this.parts.removeAt(0);
    while (this.steps.length) this.steps.removeAt(0);

    // Populate parts with ingredients and dosing actions
    formula.parts.forEach((part, pIdx) => {
      const partGroup = this.fb.group({
        name: [part.name, Validators.required],
        mix_ratio: [part.mix_ratio, [Validators.required, Validators.min(0), Validators.max(100)]],
        sort_order: [part.sort_order],
        ingredients: this.fb.array([]),
      });

      const ingArr = partGroup.get('ingredients') as FormArray;
      part.ingredients.forEach((ing) => {
        const daArr = this.fb.array(
          (ing.dosing_actions || []).map((da) =>
            this.fb.group({
              // Find matching step index
              step_ref: [this.findStepRef(formula.steps, da.step_id), Validators.min(-1)],
              use_ratio: [da.use_ratio, [Validators.required, Validators.min(0), Validators.max(100)]],
              dosing_order: [da.dosing_order, [Validators.required, Validators.min(1)]],
              dosing_method: [da.dosing_method || '', Validators.maxLength(100)],
            }),
          ),
        );

        ingArr.push(
          this.fb.group({
            material: [ing.material, [Validators.required, Validators.maxLength(200)]],
            percentage: [ing.percentage, [Validators.required, Validators.min(0), Validators.max(100)]],
            weight: [ing.weight, [Validators.min(0)]],
            dosing_actions: daArr,
          }),
        );
      });

      this.parts.push(partGroup);
      this.subs.add(
        ingArr.valueChanges.subscribe(() => this.recalcPercentageSum(pIdx, ingArr)),
      );
      this.recalcPercentageSum(pIdx, ingArr);
    });

    // Populate steps
    formula.steps.forEach((step) => {
      this.steps.push(
        this.fb.group({
          step_no: [step.step_no, [Validators.required, Validators.min(1)]],
          name: [step.name, [Validators.required, Validators.maxLength(200)]],
          temperature: [step.temperature || '', Validators.maxLength(50)],
          duration: [step.duration || '', Validators.maxLength(50)],
        }),
      );
    });
  }

  private findStepRef(steps: FormulaStep[], stepId: string): number {
    const idx = steps.findIndex((s) => s.id === stepId);
    return idx >= 0 ? idx : -1;
  }

  // ─── Submit ───

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      this.snackBar.open('请修正表单中的错误后再提交', '关闭', { duration: 3000 });
      return;
    }

    // Additional validation: check percentage sums
    for (let i = 0; i < this.parts.length; i++) {
      if (!this.isPercentageValid(i)) {
        this.snackBar.open(`配料百分比总和必须为100%（当前: ${this.getPercentageSum(i).toFixed(1)}%）`, '关闭', {
          duration: 5000,
        });
        return;
      }
    }

    this.loading = true;
    const payload = this.buildPayload();

    const request$ = this.isEditMode
      ? this.api.updateFormula(this.formulaId!, payload)
      : this.api.createFormula(payload);

    request$.subscribe({
      next: (formula) => {
        this.loading = false;
        this.snackBar.open(
          this.isEditMode ? '配方更新成功' : '配方创建成功',
          '关闭',
          { duration: 3000 },
        );
        this.saved.emit(formula);
      },
      error: (err) => {
        this.loading = false;
        const msg = err?.error?.message || err?.message || '保存失败';
        this.snackBar.open(`保存失败: ${msg}`, '关闭', { duration: 5000 });
      },
    });
  }

  onCancel(): void {
    this.cancelled.emit();
  }

  private buildPayload(): Partial<Formula> {
    const raw = this.form.getRawValue();

    const mappedParts: any[] = raw.parts.map((p: any, pIdx: number) => ({
      name: p.name,
      mix_ratio: p.mix_ratio,
      sort_order: pIdx,
      ingredients: (p.ingredients || []).map((ing: any) => ({
        material: ing.material,
        percentage: ing.percentage,
        weight: ing.weight,
        sort_order: 0,
        dosing_actions: (ing.dosing_actions || [])
          .filter((da: any) => da.step_ref >= 0)
          .map((da: any) => ({
            dosing_order: da.dosing_order,
            use_ratio: da.use_ratio,
            dosing_method: da.dosing_method,
            _step_no: raw.steps[da.step_ref]?.step_no ?? null,
          })),
      })),
    }));

    const mappedSteps: any[] = raw.steps.map((s: any, sIdx: number) => ({
      step_no: s.step_no,
      name: s.name,
      temperature: s.temperature,
      duration: s.duration,
      sort_order: sIdx,
    }));

    return {
      name: raw.name,
      code: raw.code,
      component_mode: raw.component_mode,
      status: raw.status,
      parts: mappedParts,
      steps: mappedSteps,
    } as Partial<Formula>;
  }

  // ─── Template Helper Methods ───

  /** Get the name FormControl for a part group. */
  getPartNameControl(part: AbstractControl): FormControl {
    return part.get('name') as FormControl;
  }

  /** Get the mix_ratio FormControl for a part group. */
  getPartMixRatioControl(part: AbstractControl): FormControl {
    return part.get('mix_ratio') as FormControl;
  }

  /** Get a specific FormControl from an ingredient group. */
  getIngControl(ing: AbstractControl, field: string): FormControl {
    return ing.get(field) as FormControl;
  }

  /** Get a specific FormControl from a dosing action group. */
  getDaControl(da: AbstractControl, field: string): FormControl {
    return da.get(field) as FormControl;
  }

  /** Get a specific FormControl from a step group. */
  getStepControl(step: AbstractControl, field: string): FormControl {
    return step.get(field) as FormControl;
  }

  /** Returns the step numbers as an array for mat-select options. */
  get stepOptions(): { index: number; label: string }[] {
    return this.steps.controls.map((ctrl, idx) => ({
      index: idx,
      label: `步骤 ${ctrl.get('step_no')?.value ?? idx + 1}: ${ctrl.get('name')?.value || '(未命名)'}`,
    }));
  }

  /** Determines default part arrangement based on component mode. */
  get expectedParts(): PartName[] {
    return this.form?.get('component_mode')?.value === 'double'
      ? ['PartA', 'PartB']
      : ['PartMain'];
  }
}
