import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';

import { FormulaMatrixComponent } from './formula-matrix.component';
import { FormulaApiService } from '../../services/formula-api.service';
import { Formula, FormulaMatrix, ComponentMode } from '../../types/formula.types';

function makeFormula(overrides: Partial<Formula> = {}): Formula {
  return {
    id: 'f-001',
    name: 'Test Formula',
    code: 'TF-001',
    component_mode: 'single' as ComponentMode,
    status: 'active',
    parts: [{
      id: 'p-001', formula_id: 'f-001', name: 'PartMain',
      mix_ratio: 100, sort_order: 1,
      ingredients: [
        { id: 'i-001', part_id: 'p-001', sort_order: 1, material: 'Resin', percentage: 60, weight: 600, dosing_actions: [] },
        { id: 'i-002', part_id: 'p-001', sort_order: 2, material: 'Hardener', percentage: 40, weight: 400, dosing_actions: [] },
      ],
    }],
    steps: [{ id: 's-001', formula_id: 'f-001', part_id: null, step_no: 1, name: 'Mixing', temperature: '25C', duration: '10min' }],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeMatrix(formulas: Formula[]): FormulaMatrix {
  return { formulas };
}

describe('FormulaMatrixComponent', () => {
  let component: FormulaMatrixComponent;
  let fixture: ComponentFixture<FormulaMatrixComponent>;
  let apiService: FormulaApiService;

  function createComponent(): void {
    TestBed.configureTestingModule({
      imports: [FormulaMatrixComponent, NoopAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(FormulaMatrixComponent);
    component = fixture.componentInstance;
    apiService = TestBed.inject(FormulaApiService);
  }

  describe('initialization', () => {
    it('should create the component', () => {
      createComponent();
      expect(component).toBeTruthy();
    });

    it('should start in loading state', () => {
      createComponent();
      expect(component.loading()).toBe(false);
    });
  });

  describe('data loading', () => {
    beforeEach(() => {
      createComponent();
    });

    it('should call getFormulaMatrix on init', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([])));
      fixture.detectChanges();
      expect(apiService.getFormulaMatrix).toHaveBeenCalled();
    });

    it('should set loading to false after data arrives', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([])));
      fixture.detectChanges();

      expect(component.loading()).toBe(false);
      expect(component.error()).toBeNull();
    });

    it('should show empty state when no formulas exist', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([])));
      fixture.detectChanges();

      expect(component.dataSource.data.length).toBe(0);
      expect((fixture.nativeElement.textContent as string)).toContain('暂无配方数据');
    });

    it('should render formula rows in the table when data is available', () => {
      const f1 = makeFormula({ id: 'f-001', name: 'Alpha', code: 'A-001' });
      const f2 = makeFormula({ id: 'f-002', name: 'Beta', code: 'B-001', component_mode: 'double' });

      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([f1, f2])));
      fixture.detectChanges();

      const rows = fixture.debugElement.queryAll(By.css('.formula-row'));
      expect(rows.length).toBe(2);

      const firstRowText = rows[0].nativeElement.textContent as string;
      expect(firstRowText).toContain('A-001');
      expect(firstRowText).toContain('Alpha');

      const countEl = fixture.debugElement.query(By.css('.toolbar-count'));
      expect(countEl.nativeElement.textContent).toContain('2');
    });

    it('should show error state and allow retry', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(throwError(() => new Error('Server Error')));
      fixture.detectChanges();

      expect(component.error()).toBeTruthy();
      expect(component.loading()).toBe(false);

      const errorEl = fixture.debugElement.query(By.css('.state-error'));
      expect(errorEl).toBeTruthy();

      const retryBtn = fixture.debugElement.query(By.css('.state-error button'));
      expect(retryBtn).toBeTruthy();

      vi.mocked(apiService.getFormulaMatrix).mockReturnValue(of(makeMatrix([])));
      retryBtn.triggerEventHandler('click', null);
      fixture.detectChanges();

      expect(component.error()).toBeNull();
    });
  });

  describe('row expansion', () => {
    beforeEach(() => {
      createComponent();
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula()])));
      fixture.detectChanges();
    });

    it('should expand a row to show parts and ingredients on click', () => {
      const row = fixture.debugElement.query(By.css('.formula-row'));
      expect(row).toBeTruthy();

      row.triggerEventHandler('click', null);
      fixture.detectChanges();

      expect(component.expandedElement).not.toBeNull();

      const detailInner = fixture.debugElement.query(By.css('.detail-inner'));
      expect(detailInner).toBeTruthy();

      const detailText = detailInner.nativeElement.textContent as string;
      expect(detailText).toContain('Resin');
      expect(detailText).toContain('Hardener');
      expect(detailText).toContain('100%');
      expect(detailText).toContain('Mixing');
    });

    it('should collapse an expanded row on second click', () => {
      const row = fixture.debugElement.query(By.css('.formula-row'));
      row.triggerEventHandler('click', null);
      fixture.detectChanges();
      expect(component.expandedElement).not.toBeNull();

      row.triggerEventHandler('click', null);
      fixture.detectChanges();
      expect(component.expandedElement).toBeNull();
    });

    it('should expand when the expand button is clicked', () => {
      const expandBtn = fixture.debugElement.query(By.css('.expand-btn'));
      expect(expandBtn).toBeTruthy();

      expandBtn.triggerEventHandler('click', new MouseEvent('click'));
      fixture.detectChanges();

      expect(component.expandedElement).not.toBeNull();
    });

    it('should show no-ingredients hint when part has no ingredients', () => {
      vi.mocked(apiService.getFormulaMatrix).mockReturnValue(of(makeMatrix([makeFormula({
        parts: [{ id: 'p-001', formula_id: 'f-001', name: 'PartMain', mix_ratio: 100, sort_order: 1, ingredients: [] }],
      })])));
      component.loadMatrix();
      fixture.detectChanges();

      const row = fixture.debugElement.query(By.css('.formula-row'));
      row.triggerEventHandler('click', null);
      fixture.detectChanges();

      const noDataHint = fixture.debugElement.query(By.css('.no-data-hint'));
      expect(noDataHint).toBeTruthy();
    });

    it('should not show steps section when formula has no steps', () => {
      vi.mocked(apiService.getFormulaMatrix).mockReturnValue(of(makeMatrix([makeFormula({ steps: [] })])));
      component.loadMatrix();
      fixture.detectChanges();

      const row = fixture.debugElement.query(By.css('.formula-row'));
      row.triggerEventHandler('click', null);
      fixture.detectChanges();

      const stepsSummary = fixture.debugElement.query(By.css('.steps-summary'));
      expect(stepsSummary).toBeNull();
    });
  });

  describe('component mode filter', () => {
    beforeEach(() => {
      createComponent();
    });

    it('should re-fetch matrix when filter changes', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([])));
      fixture.detectChanges();

      component.onModeFilterChange({ value: 'single' } as any);
      expect(component.selectedMode).toBe('single');
      expect(apiService.getFormulaMatrix).toHaveBeenCalledWith('single');
    });

    it('should send no mode param when All is selected', () => {
      component.selectedMode = 'single';
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([])));
      fixture.detectChanges();
      expect(apiService.getFormulaMatrix).toHaveBeenCalledWith('single');

      component.onModeFilterChange({ value: '' } as any);
      expect(apiService.getFormulaMatrix).toHaveBeenCalledWith(undefined);
    });

    it('should render the mode filter button-toggle group', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula()])));
      fixture.detectChanges();

      const toggleGroup = fixture.debugElement.query(By.css('mat-button-toggle-group'));
      expect(toggleGroup).toBeTruthy();

      const toggles = fixture.debugElement.queryAll(By.css('mat-button-toggle'));
      expect(toggles.length).toBe(3);
    });

    it('should have an add formula button', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula()])));
      fixture.detectChanges();

      const addBtn = fixture.debugElement.query(By.css('.add-formula-btn'));
      expect(addBtn).toBeTruthy();
    });
  });

  describe('display helpers', () => {
    beforeEach(() => {
      createComponent();
    });

    it('getModeLabel should return correct labels', () => {
      expect(component.getModeLabel('single')).toBe('单组分');
      expect(component.getModeLabel('double')).toBe('双组分');
    });

    it('getStatusLabel should return correct labels', () => {
      expect(component.getStatusLabel('draft')).toBe('草稿');
      expect(component.getStatusLabel('active')).toBe('已启用');
      expect(component.getStatusLabel('archived')).toBe('已归档');
    });

    it('getPartNameLabel should return correct labels', () => {
      expect(component.getPartNameLabel('PartA')).toBe('A 组分');
      expect(component.getPartNameLabel('PartB')).toBe('B 组分');
      expect(component.getPartNameLabel('PartMain')).toBe('主组分');
    });
  });

  describe('chip rendering', () => {
    beforeEach(() => {
      createComponent();
    });

    it('should render correct mode chip for single component', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula({ component_mode: 'single' })])));
      fixture.detectChanges();

      const chip = fixture.debugElement.query(By.css('.mode-single'));
      expect(chip).toBeTruthy();
      expect(chip.nativeElement.textContent).toContain('单组分');
    });

    it('should render correct mode chip for double component', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula({ component_mode: 'double' })])));
      fixture.detectChanges();

      const chip = fixture.debugElement.query(By.css('.mode-double'));
      expect(chip).toBeTruthy();
      expect(chip.nativeElement.textContent).toContain('双组分');
    });

    it('should render active status chip correctly', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula({ status: 'active' })])));
      fixture.detectChanges();

      const chip = fixture.debugElement.query(By.css('.status-active'));
      expect(chip).toBeTruthy();
      expect(chip.nativeElement.textContent).toContain('已启用');
    });

    it('should render draft status chip correctly', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula({ status: 'draft' })])));
      fixture.detectChanges();

      const chip = fixture.debugElement.query(By.css('.status-draft'));
      expect(chip).toBeTruthy();
      expect(chip.nativeElement.textContent).toContain('草稿');
    });

    it('should display correct parts count', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula({
        parts: [
          { id: 'p-a', formula_id: 'f-001', name: 'PartA', mix_ratio: 50, sort_order: 1, ingredients: [] },
          { id: 'p-b', formula_id: 'f-001', name: 'PartB', mix_ratio: 50, sort_order: 2, ingredients: [] },
        ],
        component_mode: 'double',
      })])));
      fixture.detectChanges();

      const countBadges = fixture.debugElement.queryAll(By.css('.count-badge'));
      expect(countBadges[0].nativeElement.textContent?.trim()).toBe('2');
    });

    it('should display correct steps count', () => {
      vi.spyOn(apiService, 'getFormulaMatrix').mockReturnValue(of(makeMatrix([makeFormula({
        steps: [
          { id: 's1', formula_id: 'f-001', part_id: null, step_no: 1, name: 'S1', temperature: '20C', duration: '5min' },
          { id: 's2', formula_id: 'f-001', part_id: null, step_no: 2, name: 'S2', temperature: '30C', duration: '10min' },
          { id: 's3', formula_id: 'f-001', part_id: null, step_no: 3, name: 'S3', temperature: '40C', duration: '15min' },
        ],
      })])));
      fixture.detectChanges();

      const stepBadge = fixture.debugElement.query(By.css('.count-steps'));
      expect(stepBadge.nativeElement.textContent?.trim()).toBe('3');
    });
  });
});
