import { TestBed } from '@angular/core/testing';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { provideCopilotKit } from '@copilotkitnext/angular';
import { App } from './app';

describe('App', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [
        provideAnimations(),
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        provideCopilotKit({
          runtimeUrl: '/api/copilotkit',
        }),
      ],
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(App);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });

  it('should render the toolbar with app title', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    const toolbar = compiled.querySelector('mat-toolbar');
    expect(toolbar).toBeTruthy();
    const title = compiled.querySelector('.app-title');
    expect(title?.textContent).toContain('FormulAI');
  });

  it('should render the split layout with main content and sidebar', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const mainPanel = compiled.querySelector('.app-main');
    expect(mainPanel).toBeTruthy();

    const sidebar = compiled.querySelector('.app-sidebar');
    expect(sidebar).toBeTruthy();
  });

  it('should render tab group with 4 tabs', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const tabGroup = compiled.querySelector('mat-tab-group');
    expect(tabGroup).toBeTruthy();

    const tabElements = compiled.querySelectorAll('[role="tab"]');
    expect(tabElements.length).toBe(4);
  });

  it('should render the chat panel component', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const chatPanel = compiled.querySelector('app-chat-panel');
    expect(chatPanel).toBeTruthy();
  });

  it('should render the main card container', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const mainCard = compiled.querySelector('mat-card');
    expect(mainCard).toBeTruthy();
  });
});
