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
      sort_order: 0,
      categories: [
        {
          id: 'c1',
          part_id: 'p1',
          name: '树脂',
          sort_order: 0,
          ingredients: [
            {
              id: 'i1',
              category_id: 'c1',
              material: 'Resin A',
              percentage: 60,
              weight: 600,
              batch_no: '',
              unit: 'g',
              sort_order: 0,
            },
          ],
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
      name: '搅拌',
      description: '',
      instrument_name: 'Mixer',
      temperature: '25°C',
      duration: '10min',
      categories: [],
      parameters: [
        { id: 'sp1', step_id: 's1', name: '搅拌时间', value: '30', unit: 'min', sort_order: 0 },
      ],
    },
  ],
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-01T00:00:00Z',
};

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

  it('should create the component', () => {
    expect(component).toBeTruthy();
  });

  it('should build form with default values in create mode', () => {
    fixture.detectChanges();

    expect(component.form).toBeTruthy();
    expect(component.isEditMode).toBe(false);

    const raw = component.form.getRawValue();
    expect(raw.name).toBe('');
    expect(raw.component_mode).toBe('single');
    expect(raw.status).toBe('draft');
  });

  it('should seed one MAIN part in single mode', () => {
    fixture.detectChanges();

    expect(component.parts.length).toBe(1);
    const part = component.parts.at(0).getRawValue();
    expect(part.name).toBe('PartMain');
  });

  it('should seed two parts (A/B) in double mode', () => {
    fixture.detectChanges();

    component.form.get('component_mode')?.setValue('double');
    fixture.detectChanges();

    expect(component.parts.length).toBe(2);
    expect(component.parts.at(0).get('name')?.value).toBe('PartA');
    expect(component.parts.at(1).get('name')?.value).toBe('PartB');
  });

  it('should add a step', () => {
    fixture.detectChanges();

    component.addStep();
    fixture.detectChanges();

    expect(component.steps.length).toBe(1);
  });

  it('should renumber steps after removal', () => {
    fixture.detectChanges();
    component.addStep();
    component.addStep();
    fixture.detectChanges();

    component.removeStep(0);
    fixture.detectChanges();

    expect(component.steps.length).toBe(1);
    expect(component.steps.at(0).get('step_no')?.value).toBe(1);
  });

  it('should add a part ingredient', () => {
    fixture.detectChanges();
    component.addPartCategory(0);
    fixture.detectChanges();

    component.addPartIngredient(0, 0);
    fixture.detectChanges();

    expect(component.partIngredients(0, 0).length).toBe(1);
  });

  it('should add a step material', () => {
    fixture.detectChanges();
    component.addStep();
    component.addStepCategory(0);
    fixture.detectChanges();

    component.addStepMaterial(0, 0);
    fixture.detectChanges();

    expect(component.stepMaterials(0, 0).length).toBe(1);
  });

  it('should mark form as invalid when required fields are empty', () => {
    fixture.detectChanges();
    component.form.get('name')?.markAsTouched();
    expect(component.form.invalid).toBe(true);
  });

  it('should mark form as valid when all required fields are filled', () => {
    fixture.detectChanges();
    component.form.patchValue({ name: 'Valid' });
    fixture.detectChanges();

    expect(component.form.valid).toBe(true);
  });

  it('should call createFormula API on submit in create mode', () => {
    fixture.detectChanges();

    let emitted: Formula | undefined;
    component.saved.subscribe((f) => (emitted = f));

    component.form.patchValue({ name: 'New Formula' });
    component.addStep();
    fixture.detectChanges();
    component.steps.at(0).patchValue({ name: '搅拌', temperature: '25°C', duration: '10min' });
    fixture.detectChanges();

    component.onSubmit();

    const req = httpMock.expectOne('/api/formulas');
    expect(req.request.method).toBe('POST');
    req.flush({ ...mockFormula, name: 'New Formula' });

    expect(emitted).toBeTruthy();
    expect(emitted?.name).toBe('New Formula');
  });

  it('should call updateFormula API on submit in edit mode', () => {
    component.formulaId = 'f1';
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

  it('should load formula data in edit mode', () => {
    component.formulaId = 'f1';
    fixture.detectChanges();

    const req = httpMock.expectOne('/api/formulas/f1');
    expect(req.request.method).toBe('GET');
    req.flush(mockFormula);
    fixture.detectChanges();

    expect(component.isEditMode).toBe(true);
    expect(component.form.get('name')?.value).toBe('Test Formula');
    expect(component.parts.length).toBe(1);
    expect(component.steps.length).toBe(1);
  });

  it('should emit cancelled event on cancel', () => {
    fixture.detectChanges();

    let cancelled = false;
    component.cancelled.subscribe(() => (cancelled = true));

    component.onCancel();

    expect(cancelled).toBe(true);
  });
});
