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

  it('should display "配方分析" as the hero title', () => {
    const heroTitle = fixture.debugElement.query(By.css('.hero-title'));
    expect(heroTitle.nativeElement.textContent).toContain('配方分析');
  });

  it('should render the analytics icon', () => {
    const icon = fixture.debugElement.query(By.css('.hero-icon'));
    expect(icon.nativeElement.textContent).toContain('analytics');
  });

  it('should indicate the feature uses AI chat', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('AI 配方助手');
  });

  it('should render three feature cards', () => {
    const cards = fixture.debugElement.queryAll(By.css('.feature-card'));
    expect(cards.length).toBe(3);
  });
});
