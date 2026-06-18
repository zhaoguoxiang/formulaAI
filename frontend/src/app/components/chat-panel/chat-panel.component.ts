import { Component, inject, signal, viewChild, ElementRef, DestroyRef, ChangeDetectionStrategy, afterNextRender } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';

import { MatCard, MatCardContent, MatCardHeader, MatCardTitle } from '@angular/material/card';
import { MatIcon } from '@angular/material/icon';
import { MatFormField, MatLabel } from '@angular/material/form-field';
import { MatInput } from '@angular/material/input';
import { MatIconButton } from '@angular/material/button';
import { MatProgressBar } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSnackBar } from '@angular/material/snack-bar';

import { MarkdownPipe } from '../../utils/markdown.pipe';

interface ChatMessage {
  role: 'user' | 'assistant' | 'tool';
  content: string;
  toolName?: string;
  timestamp: number;
}

const SUGGESTIONS = [
  '分析一下填料对硬度的影响',
  '最近5个配方有什么趋势？',
  '对比当前配方与其他配方的差异',
];

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
    MatTooltipModule,
    MarkdownPipe,
  ],
  templateUrl: './chat-panel.component.html',
  styleUrl: './chat-panel.component.scss',
})
export class ChatPanelComponent {
  private readonly destroyRef = inject(DestroyRef);
  private readonly snackBar = inject(MatSnackBar);
  private readonly pythonChatUrl = 'http://localhost:5050/api/chat';

  readonly messages = signal<ChatMessage[]>([]);
  inputText = '';
  readonly isRunning = signal(false);
  readonly isStreaming = signal(false);
  readonly analysisId = signal<string | null>(null);
  readonly userScrolledUp = signal(false);
  readonly suggestions = SUGGESTIONS;

  private readonly messagesContainer = viewChild<ElementRef>('messagesContainer');
  private readonly textareaRef = viewChild<ElementRef>('chatTextarea');
  private abortController: AbortController | null = null;

  constructor() {
    afterNextRender(() => {
      if (this.textareaRef()) {
        this.autoResize(this.textareaRef()!.nativeElement);
      }
    });
  }

  sendMessage(text?: string): void {
    const content = (text ?? this.inputText).trim();
    if (!content || this.isRunning()) return;

    this.messages.update((msgs) => [...msgs, { role: 'user', content, timestamp: Date.now() }]);
    this.inputText = '';
    this.isRunning.set(true);
    this.isStreaming.set(true);
    this.analysisId.set(null);
    this.scheduleScroll();

    const payload = {
      messages: this.messages()
        .filter((m) => m.role !== 'tool')
        .map((m) => ({ role: m.role, content: m.content })),
    };

    this.abortController = new AbortController();
    this.destroyRef.onDestroy(() => this.abortController?.abort());

    this.doStreamFetch(payload);
  }

  stopGeneration(): void {
    this.abortController?.abort();
    this.abortController = null;
    this.isRunning.set(false);
    this.isStreaming.set(false);
  }

  clearChat(): void {
    this.abortController?.abort();
    this.abortController = null;
    this.messages.set([]);
    this.inputText = '';
    this.isRunning.set(false);
    this.isStreaming.set(false);
    this.analysisId.set(null);
  }

  copyMessage(content: string): void {
    navigator.clipboard.writeText(content).then(() => {
      this.snackBar.open('已复制到剪贴板', '', { duration: 2000 });
    }).catch(() => {
      this.snackBar.open('复制失败', '', { duration: 2000 });
    });
  }

  relativeTime(ts: number): string {
    const diff = Date.now() - ts;
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return '刚刚';
    if (mins < 60) return `${mins} 分钟前`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours} 小时前`;
    return new Date(ts).toLocaleDateString('zh-CN');
  }

  autoResize(el: HTMLTextAreaElement): void {
    el.style.height = 'auto';
    el.style.height = Math.min(el.scrollHeight, 120) + 'px';
  }

  onTextareaInput(event: Event): void {
    const el = event.target as HTMLTextAreaElement;
    this.autoResize(el);
  }

  onMessagesScroll(): void {
    const el = this.messagesContainer()?.nativeElement;
    if (!el) return;
    const threshold = 60;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
    this.userScrolledUp.set(!atBottom);
  }

  private async doStreamFetch(payload: {
    messages: { role: string; content: string }[];
  }): Promise<void> {
    try {
      const response = await fetch(this.pythonChatUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: this.abortController!.signal,
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const reader = response.body!.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let streamingIndex = -1;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const dataStr = line.slice(6).trim();
          if (dataStr === '[DONE]') continue;

          try {
            const event = JSON.parse(dataStr);

            switch (event.type) {
              case 'tool_call':
                this.messages.update((msgs) => [
                  ...msgs,
                  { role: 'tool', content: event.name, toolName: event.name, timestamp: Date.now() },
                ]);
                this.scheduleScroll();
                break;

              case 'text':
                if (streamingIndex === -1) {
                  streamingIndex = this.messages().length;
                  this.messages.update((msgs) => [
                    ...msgs,
                    { role: 'assistant', content: event.content, timestamp: Date.now() },
                  ]);
                } else {
                  this.messages.update((msgs) => {
                    const copy = [...msgs];
                    if (
                      copy[streamingIndex] &&
                      copy[streamingIndex].role === 'assistant'
                    ) {
                      copy[streamingIndex] = {
                        ...copy[streamingIndex],
                        content: copy[streamingIndex].content + event.content,
                      };
                    }
                    return copy;
                  });
                }
                this.isStreaming.set(false);
                this.scheduleScroll();
                break;

              case 'done':
                if (event.analysis_id) {
                  this.analysisId.set(event.analysis_id);
                }
                streamingIndex = -1;
                this.isRunning.set(false);
                this.isStreaming.set(false);
                this.abortController = null;
                this.scheduleScroll();
                return;

              case 'error':
                this.messages.update((msgs) => [
                  ...msgs,
                  {
                    role: 'assistant',
                    content:
                      event.message || '分析服务返回错误，请稍后重试。',
                    timestamp: Date.now(),
                  },
                ]);
                streamingIndex = -1;
                this.isRunning.set(false);
                this.isStreaming.set(false);
                this.abortController = null;
                this.scheduleScroll();
                return;
            }
          } catch {
            // Skip unparseable SSE data lines
          }
        }
      }

      streamingIndex = -1;
      this.isRunning.set(false);
      this.isStreaming.set(false);
      this.abortController = null;
    } catch (err: unknown) {
      if (err instanceof DOMException && err.name === 'AbortError') return;
      console.error('[ChatPanel] SSE failed', err);
      this.messages.update((msgs) => [
        ...msgs,
        {
          role: 'assistant',
          content: '连接分析服务失败，请确认 Python 侧车已启动。',
          timestamp: Date.now(),
        },
      ]);
      this.isRunning.set(false);
      this.isStreaming.set(false);
      this.abortController = null;
      this.scheduleScroll();
    }
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
      if (el && !this.userScrolledUp()) {
        el.scrollTop = el.scrollHeight;
      }
    });
  }

  scrollToBottom(): void {
    const el = this.messagesContainer()?.nativeElement;
    if (el) {
      el.scrollTop = el.scrollHeight;
      this.userScrolledUp.set(false);
    }
  }
}
