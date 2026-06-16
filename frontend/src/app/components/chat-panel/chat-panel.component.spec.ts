import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideCopilotKit } from '@copilotkitnext/angular';
import { By } from '@angular/platform-browser';

import { ChatPanelComponent } from './chat-panel.component';

describe('ChatPanelComponent', () => {
  let component: ChatPanelComponent;
  let fixture: ComponentFixture<ChatPanelComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ChatPanelComponent, NoopAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideCopilotKit({
          runtimeUrl: '/api/copilotkit',
        }),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ChatPanelComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create the component', () => {
    expect(component).toBeTruthy();
  });

  it('should render a Material card', () => {
    const card = fixture.debugElement.query(By.css('mat-card'));
    expect(card).toBeTruthy();
  });

  it('should display "AI 配方助手" as the card title', () => {
    const cardTitle = fixture.debugElement.query(By.css('mat-card-title'));
    expect(cardTitle).toBeTruthy();
    expect(cardTitle.nativeElement.textContent).toContain('AI 配方助手');
  });

  it('should render a smart_toy icon in the header', () => {
    const icon = fixture.debugElement.query(By.css('mat-card-header mat-icon'));
    expect(icon).toBeTruthy();
    expect(icon.nativeElement.textContent).toContain('smart_toy');
  });

  it('should render the empty state when no messages exist', () => {
    const emptyState = fixture.debugElement.query(By.css('.empty-state'));
    expect(emptyState).toBeTruthy();

    const emptyTitle = emptyState.query(By.css('.empty-title'));
    expect(emptyTitle.nativeElement.textContent).toContain('AI 配方助手');
  });

  it('should render the input form field', () => {
    const formField = fixture.debugElement.query(By.css('mat-form-field'));
    expect(formField).toBeTruthy();
  });

  it('should render the send button', () => {
    const sendBtn = fixture.debugElement.query(By.css('.send-button'));
    expect(sendBtn).toBeTruthy();
  });

  it('should disable send button when input is empty', () => {
    component.inputText = '';
    fixture.detectChanges();

    const sendBtn = fixture.debugElement.query(By.css('.send-button'));
    expect(sendBtn.nativeElement.disabled).toBe(true);
  });

  it('should enable send button when input has text', () => {
    component.inputText = 'Hello';
    fixture.detectChanges();

    const sendBtn = fixture.debugElement.query(By.css('.send-button'));
    expect(sendBtn.nativeElement.disabled).toBe(false);
  });

  it('should render the input element inside mat-form-field', () => {
    const input = fixture.debugElement.query(By.css('input[matInput]'));
    expect(input).toBeTruthy();
    expect(input.nativeElement.placeholder).toContain('询问');
  });

  it('should show the connection status badge', () => {
    const badge = fixture.debugElement.query(By.css('.connection-badge'));
    expect(badge).toBeTruthy();
    expect(badge.nativeElement.textContent).toBeTruthy();
  });

  it('should add user message and clear input on send', () => {
    component.inputText = '测试消息';
    component.sendMessage();

    expect(component.messages().length).toBe(1);
    expect(component.messages()[0].role).toBe('user');
    expect(component.messages()[0].content).toBe('测试消息');
    expect(component.inputText).toBe('');
  });

  it('should display user messages in the template', () => {
    component.inputText = 'Hello';
    component.sendMessage();
    fixture.detectChanges();

    const userBubbles = fixture.debugElement.queryAll(By.css('.user-bubble'));
    expect(userBubbles.length).toBeGreaterThanOrEqual(1);

    const bubbleText = userBubbles[0].query(By.css('.bubble-text'));
    expect(bubbleText.nativeElement.textContent).toContain('Hello');
  });

  it('should show progress bar when isRunning is true', () => {
    component['isRunning'].set(true);
    fixture.detectChanges();

    const progressBar = fixture.debugElement.query(By.css('mat-progress-bar'));
    expect(progressBar).toBeTruthy();
  });

  it('should show typing indicator when isRunning is true', () => {
    component['isRunning'].set(true);
    fixture.detectChanges();

    const typingIndicator = fixture.debugElement.query(By.css('.typing-indicator'));
    expect(typingIndicator).toBeTruthy();
  });

  it('should not send empty messages', () => {
    component.inputText = '   ';
    component.sendMessage();

    expect(component.messages().length).toBe(0);
  });
});
