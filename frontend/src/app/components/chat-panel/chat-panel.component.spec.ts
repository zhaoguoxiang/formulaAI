import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { By } from '@angular/platform-browser';

import { ChatPanelComponent } from './chat-panel.component';

describe('ChatPanelComponent', () => {
  let component: ChatPanelComponent;
  let fixture: ComponentFixture<ChatPanelComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ChatPanelComponent, NoopAnimationsModule],
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

  it('should render the empty state when no messages exist', () => {
    const emptyState = fixture.debugElement.query(By.css('.empty-state'));
    expect(emptyState).toBeTruthy();

    const emptyTitle = emptyState.query(By.css('.empty-title'));
    expect(emptyTitle.nativeElement.textContent).toContain('AI 配方助手');
  });

  it('should render the textarea input', () => {
    const textarea = fixture.debugElement.query(By.css('.chat-textarea'));
    expect(textarea).toBeTruthy();
  });

  it('should render the send button', () => {
    const sendBtn = fixture.debugElement.query(By.css('.send-btn'));
    expect(sendBtn).toBeTruthy();
  });

  it('should disable send button when input is empty', () => {
    component.inputText = '';
    fixture.detectChanges();

    const sendBtn = fixture.debugElement.query(By.css('.send-btn'));
    expect(sendBtn.nativeElement.disabled).toBe(true);
  });

  it('should enable send button when input has text', () => {
    component.inputText = 'Hello';
    fixture.detectChanges();

    const sendBtn = fixture.debugElement.query(By.css('.send-btn'));
    expect(sendBtn.nativeElement.disabled).toBe(false);
  });

  it('should add user message and clear input on send', () => {
    component.inputText = '测试消息';
    component.sendMessage();

    expect(component.messages().length).toBe(1);
    expect(component.messages()[0].role).toBe('user');
    expect(component.messages()[0].content).toBe('测试消息');
    expect(component.inputText).toBe('');
  });

  it('should not send empty messages', () => {
    component.inputText = '   ';
    component.sendMessage();

    expect(component.messages().length).toBe(0);
  });

  it('should clear messages on clearChat', () => {
    component.inputText = 'Hello';
    component.sendMessage();

    component.clearChat();
    expect(component.messages().length).toBe(0);
    expect(component.inputText).toBe('');
  });
});
