import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, HostListener, signal, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  ReactiveFormsModule, FormBuilder, FormGroup, FormArray, FormControl,
  Validators, AbstractControl,
} from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { CdkDragDrop, CdkDrag, CdkDragHandle, CdkDropList, DragDropModule } from '@angular/cdk/drag-drop';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Subscription } from 'rxjs';

import { FormulaApiService } from '../../services/formula-api.service';
import { Formula, FormulaStep, ComponentMode, FormulaStatus } from '../../types/formula.types';
import { getModeLabel, getStatusLabel, getPartNameLabel } from '../../utils/formula-labels';
import { extractErrorMessage } from '../../utils/error.utils';

@Component({
  selector: 'app-formula-editor',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule, ReactiveFormsModule,
    MatFormFieldModule, MatInputModule, MatSelectModule, MatExpansionModule,
    MatButtonModule, MatIconModule, MatTableModule,
    CdkDrag, CdkDragHandle, CdkDropList, DragDropModule,
    MatCardModule, MatProgressBarModule, MatSnackBarModule, MatTooltipModule,
  ],
  templateUrl: './formula-editor.component.html',
  styleUrl: './formula-editor.component.scss',
})
export class FormulaEditorComponent implements OnInit, OnDestroy {
  @Input() formulaId: string | null = null;
  @Input() isMaterialMode = false;
  @Output() saved = new EventEmitter<Formula>();
  @Output() cancelled = new EventEmitter<void>();

  form!: FormGroup;
  readonly loading = signal(false);
  isEditMode = false;
  savedCode = '';

  readonly componentModes: ComponentMode[] = ['single', 'double'];
  readonly statuses: FormulaStatus[] = ['draft', 'active', 'archived'];
  readonly displayedStepColumns: string[] = ['drag', 'expand', 'step_no', 'name', 'part', 'description', 'instrument', 'actions'];
  readonly expandedSteps = signal<Set<number>>(new Set());
  readonly stepData = signal<AbstractControl[]>([]);
  readonly stepNameOptions: string[] = ['投料', '搅拌', '加热', '过辊', '冷却'];
  readonly categoryOptions: string[] = ['填料', '树脂', '助剂', '固化剂', '色浆', '其他'];

  readonly stepParameterTemplates: Record<string, { name: string; unit?: string }[]> = {
    '投料': [],
    '搅拌': [
      { name: '搅拌时间' }, { name: '时间单位', unit: 'min' },
      { name: '搅拌转速' }, { name: '转速单位', unit: 'rpm' },
      { name: '搅拌温度' }, { name: '温度单位', unit: '°C' },
    ],
  };

  readonly stepFieldTemplates: Record<string, { label: string; paramName: string; placeholder?: string; unit?: string }[]> = {
    '搅拌': [
      { label: '搅拌时间', paramName: '搅拌时间', placeholder: '例如: 30', unit: 'min' },
      { label: '搅拌转速', paramName: '搅拌转速', placeholder: '例如: 1200', unit: 'rpm' },
      { label: '搅拌温度', paramName: '搅拌温度', placeholder: '例如: 80', unit: '°C' },
    ],
  };

  private subs = new Subscription();

  constructor(
    private readonly fb: FormBuilder,
    private readonly api: FormulaApiService,
    private readonly snackBar: MatSnackBar,
  ) {}

  ngOnInit(): void {
    this.buildForm();
    this.syncPartsForMode(this.form.get('component_mode')!.value);
    this.subs.add(this.form.get('component_mode')!.valueChanges.subscribe((mode) => this.syncPartsForMode(mode)));
    if (this.formulaId) {
      this.isEditMode = true;
      this.loadFormula(this.formulaId);
    }
  }

  ngOnDestroy(): void { this.subs.unsubscribe(); }

  private buildForm(): void {
    this.form = this.fb.group({
      name: ['', [Validators.required, Validators.maxLength(200)]],
      component_mode: [{ value: 'single', disabled: this.isMaterialMode }, Validators.required],
      status: ['draft', Validators.required],
      labels: [[]],
      parts: this.fb.array([]),
      steps: this.fb.array([]),
    });
  }

  private syncPartsForMode(mode: string): void {
    while (this.parts.length) this.parts.removeAt(0);
    if (mode === 'double') {
      this.parts.push(this.fb.group({ name: ['PartA'], batch_no: [''], categories: this.fb.array([]) }));
      this.parts.push(this.fb.group({ name: ['PartB'], batch_no: [''], categories: this.fb.array([]) }));
    } else {
      this.parts.push(this.fb.group({ name: ['PartMain'], batch_no: [''], categories: this.fb.array([]) }));
    }
  }

  partCategories(partIdx: number): FormArray { return this.parts.at(partIdx).get('categories') as FormArray; }
  addPartCategory(partIdx: number): void { this.partCategories(partIdx).push(this.fb.group({ name: ['', [Validators.required, Validators.maxLength(100)]], ingredients: this.fb.array([]) })); }
  removePartCategory(partIdx: number, catIdx: number): void { this.partCategories(partIdx).removeAt(catIdx); }
  partIngredients(partIdx: number, catIdx: number): FormArray { return this.partCategories(partIdx).at(catIdx).get('ingredients') as FormArray; }
  addPartIngredient(partIdx: number, catIdx: number): void { this.partIngredients(partIdx, catIdx).push(this.fb.group({ material: ['', [Validators.required, Validators.maxLength(200)]], weight: [0], unit: [''], batch_no: [''] })); }
  removePartIngredient(partIdx: number, catIdx: number, ingIdx: number): void { this.partIngredients(partIdx, catIdx).removeAt(ingIdx); }
  getIngControl(ing: AbstractControl, field: string): FormControl { return ing.get(field) as FormControl; }

  get parts(): FormArray { return this.form.get('parts') as FormArray; }
  get steps(): FormArray { return this.form.get('steps') as FormArray; }

  partName(partIdx: number): string { return this.parts.at(partIdx).get('name')?.value ?? ''; }
  isDoubleMode(): boolean { return this.form.get('component_mode')?.value === 'double'; }
  getStepControl(step: AbstractControl, field: string): FormControl { return step.get(field) as FormControl; }
  getPartsControl(part: AbstractControl, field: string): FormControl { return part.get(field) as FormControl; }
  getCatControl(cat: AbstractControl, field: string): FormControl { return cat.get(field) as FormControl; }
  getMatControl(mat: AbstractControl, field: string): FormControl { return mat.get(field) as FormControl; }

  get partOptions(): { index: number; label: string }[] {
    return this.parts.controls.map((ctrl, idx) => ({
      index: idx, label: this.partName(idx) === 'PartMain' ? '主组分' : ctrl.get('name')?.value ?? `组分 ${idx + 1}`,
    }));
  }

  // ─── Steps ───

  addStep(): void {
    this.steps.push(this.makeStepGroup());
    this.renumberAndRefresh();
  }

  insertStepAfter(index: number): void {
    this.steps.insert(index + 1, this.makeStepGroup());
    this.renumberAndRefresh();
  }

  private makeStepGroup(): FormGroup {
    return this.fb.group({
      step_no: [this.steps.length + 1],
      name: ['', [Validators.required, Validators.maxLength(200)]],
      description: ['', Validators.maxLength(500)],
      instrument_name: ['', Validators.maxLength(100)],
      temperature: ['', Validators.maxLength(50)],
      duration: ['', Validators.maxLength(50)],
      part_ref: [-1, Validators.min(-1)],
      categories: this.fb.array([]),
      parameters: this.fb.array([]),
      feed_category: [''],
      feed_material: [''],
      feed_weight: [0],
      feed_unit: ['g'],
      feed_batch_no: [''],
    });
  }

  removeStep(index: number): void { this.steps.removeAt(index); this.renumberAndRefresh(); }

  dropStep(event: CdkDragDrop<AbstractControl[]>): void {
    if (event.previousIndex === event.currentIndex) return;
    const moved = this.steps.at(event.previousIndex);
    this.steps.removeAt(event.previousIndex);
    this.steps.insert(event.currentIndex, moved);
    this.expandedSteps.set(new Set());
    this.renumberAndRefresh();
  }

  private renumberAndRefresh(): void {
    this.steps.controls.forEach((ctrl, i) => ctrl.get('step_no')?.setValue(i + 1, { emitEvent: false }));
    this.stepData.set([...this.steps.controls]);
  }

  isStepExpanded(idx: number): boolean { return this.expandedSteps().has(idx); }
  toggleStepExpanded(idx: number): void { this.expandedSteps.update(s => { s = new Set(s); s.has(idx) ? s.delete(idx) : s.add(idx); return s; }); }

  collapseAllSteps(): void {
    const s = new Set<number>();
    for (let i = 0; i < this.steps.length; i++) s.add(i);
    this.expandedSteps.set(s);
  }

  expandAllSteps(): void {
    this.expandedSteps.set(new Set());
  }

  isAllStepsCollapsed(): boolean {
    return this.expandedSteps().size === this.steps.length && this.steps.length > 0;
  }

  stepName(stepIdx: number): string { return this.steps.at(stepIdx).get('name')?.value || ''; }

  isTemplateStep(stepIdx: number): boolean { return this.stepName(stepIdx) === '投料'; }

  getFeedControl(stepIdx: number, field: string): FormControl { return this.steps.at(stepIdx).get('feed_' + field) as FormControl; }

  partCategoryOptions(stepIdx: number): string[] {
    const partRef = this.steps.at(stepIdx).get('part_ref')?.value as number;
    const names: string[] = [];
    const partsToScan = partRef >= 0 ? [partRef] : [];
    for (const pi of partsToScan) {
      if (pi >= this.parts.length) continue;
      const cats = this.partCategories(pi);
      for (let j = 0; j < cats.length; j++) {
        const name = cats.at(j).get('name')?.value;
        if (name && !names.includes(name)) names.push(name);
      }
    }
    return names;
  }

  // ─── Step parameters ───
  stepParameters(stepIdx: number): FormArray { return this.steps.at(stepIdx).get('parameters') as FormArray; }
  addStepParameter(idx: number): void { this.stepParameters(idx).push(this.fb.group({ name: ['', [Validators.required, Validators.maxLength(100)]], value: ['', Validators.maxLength(200)], unit: ['', Validators.maxLength(50)] })); }
  removeStepParameter(idx: number, pIdx: number): void { this.stepParameters(idx).removeAt(pIdx); }
  getStepParamControl(param: AbstractControl, field: string): FormControl { return param.get(field) as FormControl; }

  getStepFieldControl(stepIdx: number, paramName: string): FormControl {
    const params = this.stepParameters(stepIdx);
    for (let i = 0; i < params.length; i++) {
      if (params.at(i).get('name')?.value === paramName) return params.at(i).get('value') as FormControl;
    }
    return new FormControl('');
  }

  onStepNameChange(stepIdx: number): void {
    const name = this.steps.at(stepIdx).get('name')?.value;
    const params = this.stepParameters(stepIdx);
    while (params.length) params.removeAt(0);
    const tpl = this.stepParameterTemplates[name];
    if (!tpl) return;
    tpl.forEach(t => params.push(this.fb.group({ name: [t.name, [Validators.required, Validators.maxLength(100)]], value: ['', Validators.maxLength(200)], unit: [t.unit || '', Validators.maxLength(50)] })));
  }

  // ─── Step material categories ───
  stepCategories(stepIdx: number): FormArray { return this.steps.at(stepIdx).get('categories') as FormArray; }
  addStepCategory(stepIdx: number): void { this.stepCategories(stepIdx).push(this.fb.group({ name: ['', [Validators.required, Validators.maxLength(100)]], materials: this.fb.array([]) })); }
  removeStepCategory(stepIdx: number, catIdx: number): void { this.stepCategories(stepIdx).removeAt(catIdx); }
  stepMaterials(stepIdx: number, catIdx: number): FormArray { return this.stepCategories(stepIdx).at(catIdx).get('materials') as FormArray; }
  addStepMaterial(stepIdx: number, catIdx: number): void { this.stepMaterials(stepIdx, catIdx).push(this.fb.group({ material: ['', [Validators.required, Validators.maxLength(200)]], percentage: [0], weight: [0], batch_no: [''], unit: [''] })); }
  removeStepMaterial(stepIdx: number, catIdx: number, matIdx: number): void { this.stepMaterials(stepIdx, catIdx).removeAt(matIdx); }

  // ─── Load ───
  private loadFormula(id: string): void {
    this.loading.set(true);
    this.subs.add(this.api.getFormula(id).subscribe({
      next: (f) => { this.populateForm(f); this.loading.set(false); },
      error: (err) => { this.loading.set(false); this.snackBar.open(`加载失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 }); },
    }));
  }

  private populateForm(formula: Formula): void {
    this.savedCode = formula.code;
    this.form.patchValue({
      name: formula.name, component_mode: formula.component_mode, status: formula.status,
      labels: formula.labels || [],
    });
    while (this.parts.length) this.parts.removeAt(0);
    while (this.steps.length) this.steps.removeAt(0);

    formula.parts.forEach(p => {
      const catsArr = this.fb.array((p.categories || []).map(cat => this.fb.group({
        name: [cat.name, [Validators.required, Validators.maxLength(100)]],
        ingredients: this.fb.array((cat.ingredients || []).map(ing => this.fb.group({
          material: [ing.material, [Validators.required, Validators.maxLength(200)]],
          weight: [ing.weight], unit: [ing.unit || ''], batch_no: [ing.batch_no || ''],
        }))),
      })));
      this.parts.push(this.fb.group({ name: [p.name], batch_no: [p.batch_no || ''], categories: catsArr }));
    });

    formula.steps.forEach(step => {
      const catsArr = this.fb.array((step.categories || []).map(cat => this.fb.group({
        name: [cat.name, [Validators.required, Validators.maxLength(100)]],
        materials: this.fb.array((cat.materials || []).map(m => this.fb.group({
          material: [m.material, [Validators.required, Validators.maxLength(200)]],
          percentage: [m.percentage], weight: [m.weight], batch_no: [m.batch_no || ''], unit: [m.unit || ''],
        }))),
      })));
      const paramsArr = this.fb.array((step.parameters || []).map(p => this.fb.group({
        name: [p.name, [Validators.required, Validators.maxLength(100)]],
        value: [p.value, Validators.maxLength(200)], unit: [p.unit, Validators.maxLength(50)],
      })));
      const isFeed = step.name === '投料';
      const firstCat = step.categories?.[0];
      const firstMat = firstCat?.materials?.[0];
      this.steps.push(this.fb.group({
        step_no: [step.step_no], name: [step.name, [Validators.required, Validators.maxLength(200)]],
        description: [step.description || '', Validators.maxLength(500)],
        instrument_name: [step.instrument_name || '', Validators.maxLength(100)],
        temperature: [step.temperature || '', Validators.maxLength(50)],
        duration: [step.duration || '', Validators.maxLength(50)],
        part_ref: [this.findPartRef(formula.parts, step.part_id), Validators.min(-1)],
        categories: catsArr, parameters: paramsArr,
        feed_category: [isFeed ? (firstCat?.name || '') : ''],
        feed_material: [isFeed ? (firstMat?.material || '') : ''],
        feed_weight: [isFeed ? (firstMat?.weight ?? 0) : 0],
        feed_unit: [isFeed ? (firstMat?.unit || '') : ''],
        feed_batch_no: [isFeed ? (firstMat?.batch_no || '') : ''],
      }));
    });
    this.stepData.set([...this.steps.controls]);
  }

  private findPartRef(parts: { id: string }[], partId: string | null): number { return partId ? parts.findIndex(p => p.id === partId) : -1; }

  // ─── Submit ───
  onSubmit(): void {
    if (this.form.invalid) { this.form.markAllAsTouched(); this.snackBar.open('请修正表单错误后再提交', '关闭', { duration: 3000 }); return; }
    this.loading.set(true);
    const payload = this.buildPayload();
    const req$ = this.isEditMode ? this.api.updateFormula(this.formulaId!, payload) : this.api.createFormula(payload);
    this.subs.add(req$.subscribe({
      next: (f) => { this.loading.set(false); this.saved.emit(f); },
      error: (err) => { this.loading.set(false); this.snackBar.open(`保存失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 }); },
    }));
  }

  onCancel(): void { this.cancelled.emit(); }

  private buildPayload(): Record<string, unknown> {
    const raw = this.form.getRawValue();
    return {
      name: raw.name, code: this.savedCode || undefined,
      component_mode: raw.component_mode, status: raw.status,
      formula_type: this.isMaterialMode ? 'material' : 'formula',
      labels: raw.labels || [],
      parts: (raw.parts as Array<Record<string, unknown>> || []).map((p, i) => ({
        name: p['name'], sort_order: i,
        batch_no: p['batch_no'] || '',
        categories: (p['categories'] as Array<Record<string, unknown>> || []).map((cat, ci) => ({
          name: cat['name'], sort_order: ci,
          ingredients: (cat['ingredients'] as Array<Record<string, unknown>> || []).map((ing, ii) => ({
            material: ing['material'], percentage: ing['weight'] ?? 0, weight: ing['weight'] ?? 0,
            batch_no: ing['batch_no'] || '', unit: ing['unit'] || '', sort_order: ii,
          })),
        })),
      })),
      steps: (raw.steps as Array<Record<string, unknown>> || []).map((s, i) => ({
        step_no: s['step_no'], name: s['name'], description: s['description'] || '',
        instrument_name: s['instrument_name'] || '', temperature: s['temperature'] || '', duration: s['duration'] || '',
        sort_order: i,
        part_id: (s['part_ref'] as number) >= 0 ? (raw.parts as Array<Record<string, unknown>>)[s['part_ref'] as number]?.['id'] : null,
        categories: s['name'] === '投料'
          ? (s['feed_category'] ? [{ name: s['feed_category'], sort_order: 0, materials: [{ material: s['feed_material'] || '', weight: s['feed_weight'] ?? 0, batch_no: s['feed_batch_no'] || '', unit: s['feed_unit'] || '', percentage: 100, sort_order: 0 }] }] : [])
          : (s['categories'] as Array<Record<string, unknown>> || []).map((cat, ci) => ({
            name: cat['name'], sort_order: ci,
            materials: (cat['materials'] as Array<Record<string, unknown>> || []).map((mat, mi) => ({
              material: mat['material'], percentage: mat['weight'] ?? 0, weight: mat['weight'] ?? 0,
              batch_no: mat['batch_no'] || '', unit: mat['unit'] || '', sort_order: mi,
            })),
          })),
        parameters: (s['parameters'] as Array<Record<string, unknown>> || []).map((p, pi) => ({
          name: p['name'], value: p['value'] || '', unit: p['unit'] || '', sort_order: pi,
        })),
      })),
    };
  }

  getModeLabel = getModeLabel; getStatusLabel = getStatusLabel; getPartNameLabel = getPartNameLabel;
  get stepOptions(): { index: number; label: string }[] {
    return this.steps.controls.map((c, i) => ({ index: i, label: `步骤 ${c.get('step_no')?.value ?? i + 1}: ${c.get('name')?.value || '(未命名)'}` }));
  }

  @HostListener('window:beforeunload', ['$event'])
  unloadHandler(event: BeforeUnloadEvent): void {
    if (this.form?.dirty) {
      event.preventDefault();
      event.returnValue = '';
    }
  }
}
