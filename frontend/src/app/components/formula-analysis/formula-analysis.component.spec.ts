import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { FormulaAnalysisComponent } from './formula-analysis.component';

describe('FormulaAnalysisComponent', () => {
  let component: FormulaAnalysisComponent;
  let fixture: ComponentFixture<FormulaAnalysisComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FormulaAnalysisComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(FormulaAnalysisComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create the component', () => {
    expect(component).toBeTruthy();
  });

  it('should render the placeholder card', () => {
    const card = fixture.debugElement.query(By.css('mat-card'));
    expect(card).toBeTruthy();
  });

  it('should display "配方分析" as the card title', () => {
    const cardTitle = fixture.debugElement.query(By.css('mat-card-title'));
    expect(cardTitle.nativeElement.textContent).toContain('配方分析');
  });

  it('should show the construction icon', () => {
    const icon = fixture.debugElement.query(By.css('.placeholder-icon'));
    expect(icon.nativeElement.textContent).toContain('construction');
  });

  it('should indicate the feature is under construction', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('分析功能建设中');
    expect(compiled.textContent).toContain('AI 助手');
  });
});
