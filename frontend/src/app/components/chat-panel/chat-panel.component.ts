import { Component, inject, signal, computed, viewChild, ElementRef, DestroyRef, ChangeDetectionStrategy } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
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

interface ChatMessage {
  role: 'user' | 'assistant' | 'tool';
  content: string;
  toolName?: string;
}

@Component({
  selector: 'app-chat-panel',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
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
  styleUrl: './chat-panel.component.scss',
})
export class ChatPanelComponent {
  private readonly http = inject(HttpClient);
  readonly copilotKit = inject(CopilotKit);
  private readonly destroyRef = inject(DestroyRef);

  readonly messages = signal<ChatMessage[]>([]);
  inputText = '';
  readonly isRunning = signal(false);

  private readonly messagesContainer = viewChild<ElementRef>('messagesContainer');

  protected readonly connectionStatus = computed(() => {
    const status = this.copilotKit.runtimeConnectionStatus();
    return status === 'connected' ? '已连接' : status === 'connecting' ? '连接中...' : '未连接';
  });

  protected readonly runtimeUrl = computed(() => this.copilotKit.runtimeUrl());

  sendMessage(): void {
    const text = this.inputText.trim();
    if (!text || this.isRunning()) return;

    this.messages.update((msgs) => [...msgs, { role: 'user', content: text }]);
    this.inputText = '';
    this.isRunning.set(true);
    this.scheduleScroll();

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
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (response) => {
          const responseMessages = response.messages ?? [];
          for (const msg of responseMessages) {
            if (msg.role === 'assistant') {
              if (msg.tool_calls && msg.tool_calls.length > 0) {
                for (const tc of msg.tool_calls) {
                  this.messages.update((msgs) => [
                    ...msgs,
                    { role: 'tool', content: `${tc.function.name}`, toolName: tc.function.name },
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
          this.scheduleScroll();
        },
        error: (err) => {
          console.error('[ChatPanel] Send failed', err);
          this.messages.update((msgs) => [
            ...msgs,
            { role: 'assistant', content: '连接AI服务失败，请确认后端服务已启动。' },
          ]);
          this.isRunning.set(false);
          this.scheduleScroll();
        },
      });
  }

  onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.sendMessage();
    }
  }

  private scheduleScroll(): void {
    requestAnimationFrame(() => {
      const el = this.messagesContainer()?.nativeElement;
      if (el) {
        el.scrollTop = el.scrollHeight;
      }
    });
  }
}
