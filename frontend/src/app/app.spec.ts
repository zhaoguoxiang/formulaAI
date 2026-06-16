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
    const toolbar = compiled.querySelector('.app-toolbar');
    expect(toolbar).toBeTruthy();
    const title = compiled.querySelector('.toolbar-title');
    expect(title?.textContent).toContain('FormulAI');
  });

  it('should render the split layout with workspace and chat panels', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const workspacePanel = compiled.querySelector('.workspace-panel');
    expect(workspacePanel).toBeTruthy();

    const chatPanel = compiled.querySelector('.chat-panel');
    expect(chatPanel).toBeTruthy();
  });

  it('should render workspace tab group with 3 tabs', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const tabGroup = compiled.querySelector('mat-tab-group');
    expect(tabGroup).toBeTruthy();

    // mat-tab creates role="tab" elements
    const tabElements = compiled.querySelectorAll('[role="tab"]');
    expect(tabElements.length).toBe(3);
  });

  it('should render the chat panel component', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const chatPanel = compiled.querySelector('app-chat-panel');
    expect(chatPanel).toBeTruthy();
  });

  it('should render workspace card with modern styling', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;

    const workspaceCard = compiled.querySelector('.workspace-card');
    expect(workspaceCard).toBeTruthy();
  });
});
