import { Component, inject, signal, computed, viewChild, ElementRef, AfterViewChecked, DestroyRef } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CopilotKit } from '@copilotkitnext/angular';

import { MatCard, MatCardContent, MatCardHeader, MatCardTitle } from '@angular/material/card';
import { MatIcon } from '@angular/material/icon';
import { MatFormField, MatLabel } from '@angular/material/form-field';
import { MatInput } from '@angular/material/input';
import { MatIconButton } from '@angular/material/button';
import { MatProgressBar } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';

/** A single chat message, mirroring the AG-UI message shape. */
interface ChatMessage {
  role: 'user' | 'assistant' | 'tool';
  content: string;
  toolName?: string;
}

/**
 * Chat panel providing an AI formula assistant interface.
 *
 * Uses CopilotKit's configuration (runtimeUrl: /api/copilotkit)
 * for LLM communication. UI built with Angular Material components
 * for visual consistency with the rest of the app.
 */
@Component({
  selector: 'app-chat-panel',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCard,
    MatCardContent,
    MatCardHeader,
    MatCardTitle,
    MatIcon,
    MatFormField,
    MatLabel,
    MatInput,
    MatIconButton,
    MatProgressBar,
    MatChipsModule,
  ],
  templateUrl: './chat-panel.component.html',
  styleUrl: './chat-panel.component.css',
})
export class ChatPanelComponent implements AfterViewChecked {
  private readonly http = inject(HttpClient);
  readonly copilotKit = inject(CopilotKit);
  private readonly destroyRef = inject(DestroyRef);

  /* ── Messages ── */
  readonly messages = signal<ChatMessage[]>([]);

  /* ── Input ── */
  inputText = '';

  /* ── State ── */
  readonly isRunning = signal(false);

  /* ── Scroll ── */
  private readonly messagesContainer = viewChild<ElementRef>('messagesContainer');
  private shouldScrollToBottom = false;

  /* ── Computed ── */
  protected readonly connectionStatus = computed(() => {
    const status = this.copilotKit.runtimeConnectionStatus();
    return status === 'connected' ? '已连接' : status === 'connecting' ? '连接中...' : '未连接';
  });

  protected readonly runtimeUrl = computed(() => this.copilotKit.runtimeUrl());

  /* ── Send message to CopilotKit runtime ── */
  sendMessage(): void {
    const text = this.inputText.trim();
    if (!text || this.isRunning()) return;

    // Add user message
    this.messages.update((msgs) => [...msgs, { role: 'user', content: text }]);
    this.inputText = '';
    this.isRunning.set(true);
    this.scrollToBottom();

    // Send to CopilotKit runtime
    const payload = {
      messages: this.messages()
        .filter((m) => m.role !== 'tool')
        .map((m) => ({
          role: m.role,
          content: m.content,
        })),
    };

    this.http
      .post<{
        messages?: { role: string; content: string; tool_calls?: { function: { name: string; arguments: string } }[] }[];
      }>(this.copilotKit.runtimeUrl(), payload, {
        headers: { 'Content-Type': 'application/json' },
      })
      .subscribe({
        next: (response) => {
          const responseMessages = response.messages ?? [];
          for (const msg of responseMessages) {
            if (msg.role === 'assistant') {
              // Display tool call chips when the LLM invokes a tool
              if (msg.tool_calls && msg.tool_calls.length > 0) {
                for (const tc of msg.tool_calls) {
                  this.messages.update((msgs) => [
                    ...msgs,
                    {
                      role: 'tool',
                      content: `${tc.function.name}`,
                      toolName: tc.function.name,
                    },
                  ]);
                }
              }
              if (msg.content) {
                this.messages.update((msgs) => [
                  ...msgs,
                  { role: 'assistant', content: msg.content },
                ]);
              }
            }
          }
          this.isRunning.set(false);
          this.scrollToBottom();
        },
        error: (err) => {
          // eslint-disable-next-line no-console
          console.error('[ChatPanel] Send failed', err);
          this.messages.update((msgs) => [
            ...msgs,
            {
              role: 'assistant',
              content: '连接AI服务失败，请确认后端服务已启动。',
            },
          ]);
          this.isRunning.set(false);
          this.scrollToBottom();
        },
      });
  }

  /* ── Keyboard handler ── */
  onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.sendMessage();
    }
  }

  /* ── Auto-scroll ── */
  private scrollToBottom(): void {
    this.shouldScrollToBottom = true;
  }

  ngAfterViewChecked(): void {
    if (this.shouldScrollToBottom) {
      this.shouldScrollToBottom = false;
      const el = this.messagesContainer()?.nativeElement;
      if (el) {
        setTimeout(() => {
          el.scrollTop = el.scrollHeight;
        }, 0);
      }
    }
  }
}
