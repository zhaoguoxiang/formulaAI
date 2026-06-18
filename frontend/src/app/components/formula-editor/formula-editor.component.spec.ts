import { TestBed, ComponentFixture } from '@angular/core/testing';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';

import { FormulaEditorComponent } from './formula-editor.component';
import { Formula } from '../../types/formula.types';

const mockFormula: Formula = {
  id: 'f1',
  name: 'Test Formula',
  code: 'TF-001',
  component_mode: 'single',
  status: 'draft',
  parts: [
    {
      id: 'p1',
      formula_id: 'f1',
      name: 'PartMain',
      mix_ratio: 100,
      sort_order: 0,
      ingredients: [
        {
          id: 'i1',
          part_id: 'p1',
          sort_order: 0,
          material: 'Resin A',
          percentage: 60,
          weight: 600,
          dosing_actions: [
            {
              id: 'd1',
              step_id: 's1',
              ingredient_id: 'i1',
              dosing_order: 1,
              use_ratio: 100,
              dosing_method: 'manual',
            },
          ],
        },
        {
          id: 'i2',
          part_id: 'p1',
          sort_order: 1,
          material: 'Hardener B',
          percentage: 40,
          weight: 400,
          dosing_actions: [],
        },
      ],
    },
  ],
  steps: [
    {
      id: 's1',
      formula_id: 'f1',
      part_id: null,
      step_no: 1,
      name: 'Mix',
      temperature: '25°C',
      duration: '10min',
    },
  ],
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-01T00:00:00Z',
};

function setupComponent(fixture: ComponentFixture<FormulaEditorComponent>, formulaId: string | null): void {
  const component = fixture.componentInstance;
  component.formulaId = formulaId;
  if (formulaId) {
    component.isEditMode = true;
  }
}

describe('FormulaEditorComponent', () => {
  let fixture: ComponentFixture<FormulaEditorComponent>;
  let component: FormulaEditorComponent;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FormulaEditorComponent],
      providers: [
        provideAnimations(),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(FormulaEditorComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  // ─── Component Creation ───

  it('should create the component', () => {
    expect(component).toBeTruthy();
  });

  // ─── Create Mode: Initial Form State ───

  it('should build form with default values in create mode', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();

    expect(component.form).toBeTruthy();
    expect(component.isEditMode).toBe(false);

    const raw = component.form.getRawValue();
    expect(raw.name).toBe('');
    expect(raw.component_mode).toBe('single');
    expect(raw.status).toBe('draft');
  });

  it('should seed one MAIN part in single mode', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();

    expect(component.parts.length).toBe(1);
    const part = component.parts.at(0).getRawValue();
    expect(part.name).toBe('PartMain');
    expect(part.mix_ratio).toBe(100);
  });

  it('should seed two parts (A/B) in double mode', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();

    component.form.get('component_mode')?.setValue('double');
    fixture.detectChanges();

    expect(component.parts.length).toBe(2);
    expect(component.parts.at(0).get('name')?.value).toBe('PartA');
    expect(component.parts.at(1).get('name')?.value).toBe('PartB');
  });

  // ─── Dynamic Add/Remove: Ingredients ───

  it('should add an ingredient to a part', async () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    await fixture.whenStable();

    component.addIngredient(0);
    fixture.detectChanges();
    await fixture.whenStable();

    expect(component.ingredients(0).length).toBe(1);
  });

  it('should remove an ingredient from a part', async () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    await fixture.whenStable();
    component.addIngredient(0);
    component.addIngredient(0);
    fixture.detectChanges();
    await fixture.whenStable();

    component.removeIngredient(0, 0);
    fixture.detectChanges();
    await fixture.whenStable();

    expect(component.ingredients(0).length).toBe(1);
  });

  // ─── Dynamic Add/Remove: Steps ───

  it('should add a step', async () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    await fixture.whenStable();

    component.addStep();
    fixture.detectChanges();
    await fixture.whenStable();

    expect(component.steps.length).toBe(1);
  });

  it('should renumber steps after removal', async () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    await fixture.whenStable();
    component.addStep();
    component.addStep();
    fixture.detectChanges();
    await fixture.whenStable();

    component.removeStep(0);
    fixture.detectChanges();
    await fixture.whenStable();

    expect(component.steps.length).toBe(1);
    expect(component.steps.at(0).get('step_no')?.value).toBe(1);
  });

  // ─── Dynamic Add/Remove: Dosing Actions ───

  it('should add a dosing action to an ingredient', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    component.addIngredient(0);

    component.addDosingAction(0, 0);
    expect(component.dosingActions(0, 0).length).toBe(1);
  });

  it('should remove a dosing action', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    component.addIngredient(0);
    component.addDosingAction(0, 0);
    component.addDosingAction(0, 0);

    component.removeDosingAction(0, 0, 0);
    expect(component.dosingActions(0, 0).length).toBe(1);
  });

  // ─── Form Validation ───

  it('should mark form as invalid when required fields are empty', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    component.form.get('name')?.markAsTouched();
    expect(component.form.invalid).toBe(true);
  });

  it('should mark form as valid when all required fields are filled', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    component.form.patchValue({ name: 'Valid' });
    fixture.detectChanges();

    expect(component.form.valid).toBe(true);
  });

  // ─── Submit: Create ───

  it('should call createFormula API on submit in create mode', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();

    let emitted: Formula | undefined;
    component.saved.subscribe((f) => (emitted = f));

    component.form.patchValue({ name: 'New Formula' });
    component.addStep();
    fixture.detectChanges();
    component.steps.at(0).patchValue({ name: 'Mix', temperature: '25°C', duration: '10min' });
    component.ingredients(0).clear();
    component.addIngredient(0);
    component.addIngredient(0);
    fixture.detectChanges();

    component.onSubmit();

    const req = httpMock.expectOne('/api/formulas');
    expect(req.request.method).toBe('POST');
    req.flush({ ...mockFormula, name: 'New Formula' });

    expect(emitted).toBeTruthy();
    expect(emitted?.name).toBe('New Formula');
  });

  // ─── Submit: Update (edit mode) ───

  it('should call updateFormula API on submit in edit mode', () => {
    setupComponent(fixture, 'f1');
    fixture.detectChanges();

    const getReq = httpMock.expectOne('/api/formulas/f1');
    getReq.flush(mockFormula);
    fixture.detectChanges();

    let emitted: Formula | undefined;
    component.saved.subscribe((f) => (emitted = f));

    component.onSubmit();

    const putReq = httpMock.expectOne('/api/formulas/f1');
    expect(putReq.request.method).toBe('PUT');
    putReq.flush({ ...mockFormula, name: 'Test Formula' });

    expect(emitted).toBeTruthy();
  });

  // ─── Edit Mode: Data Loading ───

  it('should load formula data in edit mode', () => {
    setupComponent(fixture, 'f1');
    fixture.detectChanges();

    const req = httpMock.expectOne('/api/formulas/f1');
    expect(req.request.method).toBe('GET');
    req.flush(mockFormula);
    fixture.detectChanges();

    expect(component.isEditMode).toBe(true);
    expect(component.form.get('name')?.value).toBe('Test Formula');
    expect(component.parts.length).toBe(1);
    expect(component.ingredients(0).length).toBe(2);
    expect(component.steps.length).toBe(1);
  });

  // ─── Cancel ───

  it('should emit cancelled event on cancel', () => {
    setupComponent(fixture, null);
    fixture.detectChanges();

    let cancelled = false;
    component.cancelled.subscribe(() => (cancelled = true));

    component.onCancel();

    expect(cancelled).toBe(true);
  });

  // ─── Step Options ───

  it('should return step options for dosing action dropdown', async () => {
    setupComponent(fixture, null);
    fixture.detectChanges();
    await fixture.whenStable();
    component.addStep();
    fixture.detectChanges();
    await fixture.whenStable();
    component.steps.at(0).patchValue({ step_no: 1, name: 'Mixing' });
    component.addStep();
    fixture.detectChanges();
    await fixture.whenStable();
    component.steps.at(1).patchValue({ step_no: 2, name: 'Curing' });
    fixture.detectChanges();
    await fixture.whenStable();

    const options = component.stepOptions;
    expect(options.length).toBe(2);
    expect(options[0].label).toContain('步骤 1');
    expect(options[0].label).toContain('Mixing');
  });
});
