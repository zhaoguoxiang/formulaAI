import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { FormulaApiService } from './formula-api.service';
import { Formula, FormulaList } from '../types/formula.types';

describe('FormulaApiService', () => {
  let service: FormulaApiService;
  let httpMock: HttpTestingController;

  const mockFormula: Formula = {
    id: 'f1',
    name: 'Test Formula',
    code: 'TF-001',
    component_mode: 'single',
    status: 'draft',
    formula_type: 'formula',
    labels: [],
    parts: [],
    steps: [],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  };

  const mockFormulaList: FormulaList = {
    formulas: [mockFormula],
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [FormulaApiService, provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(FormulaApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  // ─── getFormula ───

  it('getFormula(id) should GET /api/formulas/:id and return a Formula', () => {
    service.getFormula('f1').subscribe((formula) => {
      expect(formula.id).toBe('f1');
      expect(formula.name).toBe('Test Formula');
    });

    const req = httpMock.expectOne('/api/formulas/f1');
    expect(req.request.method).toBe('GET');
    req.flush(mockFormula);
  });

  // ─── createFormula ───

  it('createFormula(data) should POST /api/formulas with body and return created Formula', () => {
    const payload: Partial<Formula> = { name: 'New', code: 'N-001', component_mode: 'double' };

    service.createFormula(payload).subscribe((formula) => {
      expect(formula.id).toBe('f2');
      expect(formula.name).toBe('New');
    });

    const req = httpMock.expectOne('/api/formulas');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(payload);
    req.flush({ ...mockFormula, id: 'f2', name: 'New', code: 'N-001', component_mode: 'double' });
  });

  // ─── updateFormula ───

  it('updateFormula(id, data) should PUT /api/formulas/:id with body and return updated Formula', () => {
    const payload: Partial<Formula> = { name: 'Updated', status: 'active' };

    service.updateFormula('f1', payload).subscribe((formula) => {
      expect(formula.name).toBe('Updated');
      expect(formula.status).toBe('active');
    });

    const req = httpMock.expectOne('/api/formulas/f1');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual(payload);
    req.flush({ ...mockFormula, name: 'Updated', status: 'active' });
  });

  // ─── getFormulaList ───

  it('getFormulaList() should GET /api/formulas/list and return FormulaList', () => {
    service.getFormulaList().subscribe((list) => {
      expect(list.formulas.length).toBe(1);
      expect(list.formulas[0].id).toBe('f1');
    });

    const req = httpMock.expectOne('/api/formulas/list');
    expect(req.request.method).toBe('GET');
    req.flush(mockFormulaList);
  });

  it('getFormulaList("double") should add component_mode query param', () => {
    service.getFormulaList('double').subscribe((list) => {
      expect(list.formulas.length).toBe(1);
    });

    const req = httpMock.expectOne(
      (r) => r.url === '/api/formulas/list' && r.params.get('component_mode') === 'double',
    );
    expect(req.request.method).toBe('GET');
    req.flush(mockFormulaList);
  });
});
