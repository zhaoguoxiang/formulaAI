import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { FormulaAnalysisComponent } from './formula-analysis.component';
import { FormulaApiService } from '../../services/formula-api.service';
import { of, throwError } from 'rxjs';

describe('FormulaAnalysisComponent', () => {
  let component: FormulaAnalysisComponent;
  let fixture: ComponentFixture<FormulaAnalysisComponent>;
  let apiService: FormulaApiService;

  const mockModeRatio = { labels: ['single', 'double'], values: [15, 7] };
  const mockIngredientFreq = { labels: ['水', '甘油', '乙醇', '丙二醇'], values: [20, 15, 10, 8] };
  const mockStepDist = { labels: ['1-2步', '3-5步', '5步以上'], values: [8, 10, 4] };
  const mockDosing = { methods: [{ name: '手动添加', count: 12 }, { name: '泵送', count: 8 }] };

  function createComponent(): void {
    TestBed.configureTestingModule({
      imports: [FormulaAnalysisComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(FormulaAnalysisComponent);
    component = fixture.componentInstance;
    apiService = TestBed.inject(FormulaApiService);
  }

  describe('initialization', () => {
    it('should create the component', () => {
      createComponent();
      expect(component).toBeTruthy();
    });

    it('should start all 4 sections in loading state', () => {
      createComponent();
      expect(component.modeRatio.loading).toBe(true);
      expect(component.ingredientFreq.loading).toBe(true);
      expect(component.stepDistribution.loading).toBe(true);
      expect(component.dosingMethods.loading).toBe(true);
    });
  });

  describe('data loading — success', () => {
    beforeEach(() => {
      createComponent();
      vi.spyOn(apiService, 'getAnalysis').mockImplementation((endpoint: string) => {
        switch (endpoint) {
          case 'component-mode-ratio': return of(mockModeRatio);
          case 'ingredient-distribution': return of(mockIngredientFreq);
          case 'step-count-distribution': return of(mockStepDist);
          case 'dosing-method-stats': return of(mockDosing);
          default: return of(null);
        }
      });
    });

    it('should call all 4 analysis endpoints on init', () => {
      fixture.detectChanges();
      expect(apiService.getAnalysis).toHaveBeenCalledWith('component-mode-ratio');
      expect(apiService.getAnalysis).toHaveBeenCalledWith('ingredient-distribution');
      expect(apiService.getAnalysis).toHaveBeenCalledWith('step-count-distribution');
      expect(apiService.getAnalysis).toHaveBeenCalledWith('dosing-method-stats');
    });

    it('should set mode ratio chart data from API', () => {
      fixture.detectChanges();
      expect(component.modeRatio.loading).toBe(false);
      expect(component.modeRatio.error).toBe(null);
      expect(component.modeRatio.data).toEqual(mockModeRatio);
      expect(component.modeRatioChartData.labels).toEqual(['single', 'double']);
      expect(component.modeRatioChartData.datasets?.[0]?.data).toEqual([15, 7]);
    });

    it('should set ingredient frequency chart data from API', () => {
      fixture.detectChanges();
      expect(component.ingredientFreq.loading).toBe(false);
      expect(component.ingredientFreq.data).toEqual(mockIngredientFreq);
      expect(component.ingredientFreqChartData.labels).toEqual(['水', '甘油', '乙醇', '丙二醇']);
      expect(component.ingredientFreqChartData.datasets?.[0]?.data).toEqual([20, 15, 10, 8]);
    });

    it('should set step distribution chart data from API', () => {
      fixture.detectChanges();
      expect(component.stepDistribution.loading).toBe(false);
      expect(component.stepDistribution.data).toEqual(mockStepDist);
      expect(component.stepDistributionChartData.labels).toEqual(['1-2步', '3-5步', '5步以上']);
      expect(component.stepDistributionChartData.datasets?.[0]?.data).toEqual([8, 10, 4]);
    });

    it('should set dosing methods table data sorted by count desc', () => {
      fixture.detectChanges();
      expect(component.dosingMethods.loading).toBe(false);
      expect(component.dosingMethods.data?.length).toBe(2);
      expect(component.dosingMethods.data?.[0]?.name).toBe('手动添加');
      expect(component.dosingMethods.data?.[0]?.count).toBe(12);
    });
  });

  describe('data loading — empty', () => {
    beforeEach(() => {
      createComponent();
      vi.spyOn(apiService, 'getAnalysis').mockReturnValue(of({ labels: [], values: [] }));
    });

    it('should show empty state when chart data has no labels', () => {
      fixture.detectChanges();
      expect(component.modeRatio.loading).toBe(false);
      expect(component.modeRatio.data).toBe(null);
      expect(component.modeRatio.error).toBe(null);
    });

    it('should show empty state when dosing methods array is empty', () => {
      vi.spyOn(apiService, 'getAnalysis').mockImplementation((endpoint: string) => {
        if (endpoint === 'dosing-method-stats') return of({ methods: [] });
        return of({ labels: ['x'], values: [1] });
      });
      fixture.detectChanges();
      expect(component.dosingMethods.data).toBe(null);
    });
  });

  describe('data loading — error', () => {
    beforeEach(() => {
      createComponent();
      vi.spyOn(apiService, 'getAnalysis').mockReturnValue(
        throwError(() => new Error('网络错误')),
      );
    });

    it('should set error state on API failure for all sections', () => {
      fixture.detectChanges();
      expect(component.modeRatio.loading).toBe(false);
      expect(component.modeRatio.data).toBe(null);
      expect(component.modeRatio.error).toBeTruthy();
      expect(component.dosingMethods.loading).toBe(false);
      expect(component.dosingMethods.data).toBe(null);
      expect(component.dosingMethods.error).toBeTruthy();
    });
  });

  describe('template rendering', () => {
    beforeEach(() => {
      createComponent();
      vi.spyOn(apiService, 'getAnalysis').mockImplementation((endpoint: string) => {
        switch (endpoint) {
          case 'component-mode-ratio': return of(mockModeRatio);
          case 'ingredient-distribution': return of(mockIngredientFreq);
          case 'step-count-distribution': return of(mockStepDist);
          case 'dosing-method-stats': return of(mockDosing);
          default: return of(null);
        }
      });
      fixture.detectChanges();
    });

    it('should render 4 chart cards', () => {
      const compiled = fixture.nativeElement as HTMLElement;
      const cards = compiled.querySelectorAll('mat-card');
      expect(cards.length).toBe(4);
    });
  });
});
