import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, ChangeDetectionStrategy } from '@angular/core';
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
import { Subscription } from 'rxjs';

import { FormulaApiService } from '../../services/formula-api.service';
import {
  Formula, FormulaPart, FormulaStep, ComponentMode, FormulaStatus,
} from '../../types/formula.types';
import { getModeLabel, getStatusLabel, getPartNameLabel } from '../../utils/formula-labels';
import { extractErrorMessage } from '../../utils/error.utils';

function generateTempId(): string {
  return 'tmp_' + Math.random().toString(36).substring(2, 10);
}

@Component({
  selector: 'app-formula-editor',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
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
  styleUrl: './formula-editor.component.scss',
})
export class FormulaEditorComponent implements OnInit, OnDestroy {
  @Input() formulaId: string | null = null;
  @Output() saved = new EventEmitter<Formula>();
  @Output() cancelled = new EventEmitter<void>();

  form!: FormGroup;
  loading = false;
  isEditMode = false;
  savedCode = '';

  readonly componentModes: ComponentMode[] = ['single', 'double'];
  readonly statuses: FormulaStatus[] = ['draft', 'active', 'archived'];

  private subs = new Subscription();

  constructor(
    private readonly fb: FormBuilder,
    private readonly api: FormulaApiService,
    private readonly snackBar: MatSnackBar,
  ) {}

  ngOnInit(): void {
    this.buildForm();
    this.syncPartsForMode(this.form.get('component_mode')!.value);

    this.subs.add(
      this.form.get('component_mode')!.valueChanges.subscribe((mode) => {
        this.syncPartsForMode(mode);
      }),
    );

    if (this.formulaId) {
      this.isEditMode = true;
      this.loadFormula(this.formulaId);
    }
  }

  ngOnDestroy(): void {
    this.subs.unsubscribe();
  }

  // ─── Form ───

  private buildForm(): void {
    this.form = this.fb.group({
      name: ['', [Validators.required, Validators.maxLength(200)]],
      component_mode: ['single', Validators.required],
      status: ['draft', Validators.required],
      parts: this.fb.array([]),
      steps: this.fb.array([]),
    });
  }

  /** Rebuild parts array based on the selected component_mode. */
  private syncPartsForMode(mode: string): void {
    while (this.parts.length) this.parts.removeAt(0);

    if (mode === 'double') {
      this.createPart('PartA', 50);
      this.createPart('PartB', 50);
    } else {
      this.createPart('PartMain', 100);
    }
  }

  private createPart(name: string, mixRatio: number): void {
    const partGroup = this.fb.group({
      name: [name],
      mix_ratio: [mixRatio, [Validators.required, Validators.min(0), Validators.max(100)]],
      ingredients: this.fb.array([]),
    });
    this.parts.push(partGroup);
  }

  // ─── Getters ───

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

  partName(partIndex: number): string {
    return this.parts.at(partIndex).get('name')?.value ?? '';
  }

  partLabel(partIndex: number): string {
    const n = this.partName(partIndex);
    if (n === 'PartMain') return '主组分';
    return n;
  }

  isDoubleMode(): boolean {
    return this.form.get('component_mode')?.value === 'double';
  }

  // ─── Ingredients ───

  addIngredient(partIndex: number): void {
    const ingGroup = this.fb.group({
      material: ['', [Validators.required, Validators.maxLength(200)]],
      weight: [0, [Validators.min(0)]],
      dosing_actions: this.fb.array([]),
    });
    this.ingredients(partIndex).push(ingGroup);
  }

  removeIngredient(partIndex: number, ingredientIndex: number): void {
    this.ingredients(partIndex).removeAt(ingredientIndex);
  }

  /** Auto-calculated percentage for an ingredient based on weights. */
  ingredientPercentage(partIndex: number, ingredientIndex: number): number {
    const ings = this.ingredients(partIndex);
    if (ings.length === 0) return 0;
    const totalWeight = ings.controls.reduce((sum, c) => sum + (c.get('weight')?.value || 0), 0);
    if (totalWeight === 0) return 0;
    const w = ings.at(ingredientIndex).get('weight')?.value || 0;
    return (w / totalWeight) * 100;
  }

  // ─── Steps ───

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
    this.steps.controls.forEach((ctrl, i) => {
      ctrl.get('step_no')?.setValue(i + 1, { emitEvent: false });
    });
  }

  // ─── Dosing Actions ───

  addDosingAction(partIndex: number, ingredientIndex: number): void {
    const daGroup = this.fb.group({
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

  // ─── Load (Edit) ───

  private loadFormula(id: string): void {
    this.loading = true;
    const sub = this.api.getFormula(id).subscribe({
      next: (formula) => {
        this.populateForm(formula);
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.snackBar.open(`加载配方失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
      },
    });
    this.subs.add(sub);
  }

  private populateForm(formula: Formula): void {
    this.savedCode = formula.code;
    this.form.patchValue({
      name: formula.name,
      component_mode: formula.component_mode,
      status: formula.status,
    });

    // Clear current parts and rebuild from data
    while (this.parts.length) this.parts.removeAt(0);
    while (this.steps.length) this.steps.removeAt(0);

    formula.parts.forEach((part) => {
      const partGroup = this.fb.group({
        name: [part.name],
        mix_ratio: [part.mix_ratio, [Validators.required, Validators.min(0), Validators.max(100)]],
        ingredients: this.fb.array([]),
      });

      const ingArr = partGroup.get('ingredients') as FormArray;
      part.ingredients.forEach((ing) => {
        const daArr = this.fb.array(
          (ing.dosing_actions || []).map((da) =>
            this.fb.group({
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
            weight: [ing.weight, [Validators.min(0)]],
            dosing_actions: daArr,
          }),
        );
      });

      this.parts.push(partGroup);
    });

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

    this.loading = true;
    const payload = this.buildPayload();

    const request$ = this.isEditMode
      ? this.api.updateFormula(this.formulaId!, payload)
      : this.api.createFormula(payload);

    const sub = request$.subscribe({
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
        this.snackBar.open(`保存失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
      },
    });
    this.subs.add(sub);
  }

  onCancel(): void {
    this.cancelled.emit();
  }

  private buildPayload(): Record<string, unknown> {
    const raw = this.form.getRawValue();

    const mappedParts = raw.parts.map((p: Record<string, unknown>, pIdx: number) => {
      const ingredients = (p['ingredients'] as Array<Record<string, unknown>> || []).map((ing) => {
        const dosing_actions = (ing['dosing_actions'] as Array<Record<string, unknown>> || [])
          .filter((da) => (da['step_ref'] as number) >= 0)
          .map((da) => ({
            dosing_order: da['dosing_order'],
            use_ratio: da['use_ratio'],
            dosing_method: da['dosing_method'],
            _step_no: (raw.steps as Array<Record<string, unknown>>)[da['step_ref'] as number]?.['step_no'] ?? null,
          }));
        return {
          material: ing['material'],
          percentage: ing['weight'] ?? 0,
          weight: ing['weight'] ?? 0,
          dosing_actions,
        };
      });
      return {
        name: p['name'],
        mix_ratio: p['mix_ratio'],
        sort_order: pIdx,
        ingredients,
      };
    });

    const mappedSteps = raw.steps.map((s: Record<string, unknown>, sIdx: number) => ({
      step_no: s['step_no'],
      name: s['name'],
      temperature: s['temperature'],
      duration: s['duration'],
      sort_order: sIdx,
    }));

    return {
      name: raw.name,
      component_mode: raw.component_mode,
      status: raw.status,
      parts: mappedParts,
      steps: mappedSteps,
      ...(this.savedCode ? { code: this.savedCode } : {}),
    };
  }

  // ─── Template helpers ───

  getPartMixRatioControl(part: AbstractControl): FormControl {
    return part.get('mix_ratio') as FormControl;
  }

  getIngControl(ing: AbstractControl, field: string): FormControl {
    return ing.get(field) as FormControl;
  }

  getDaControl(da: AbstractControl, field: string): FormControl {
    return da.get(field) as FormControl;
  }

  getStepControl(step: AbstractControl, field: string): FormControl {
    return step.get(field) as FormControl;
  }

  getModeLabel = getModeLabel;
  getStatusLabel = getStatusLabel;
  getPartNameLabel = getPartNameLabel;

  get stepOptions(): { index: number; label: string }[] {
    return this.steps.controls.map((ctrl, idx) => ({
      index: idx,
      label: `步骤 ${ctrl.get('step_no')?.value ?? idx + 1}: ${ctrl.get('name')?.value || '(未命名)'}`,
    }));
  }
}
