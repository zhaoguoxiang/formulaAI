import { Component, signal, effect, Renderer2, inject, ChangeDetectionStrategy } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { MatToolbar } from '@angular/material/toolbar';
import { MatTabGroup, MatTab } from '@angular/material/tabs';
import { MatCard, MatCardContent } from '@angular/material/card';
import { MatIconButton } from '@angular/material/button';
import { MatIcon } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';

import { FormulaAnalysisComponent } from './components/formula-analysis/formula-analysis.component';
import { FormulaListComponent } from './components/formula-list/formula-list.component';
import { PrebuiltMaterialComponent } from './components/prebuilt-material/prebuilt-material.component';
import { TestOutlineComponent } from './components/test-outline/test-outline.component';
import { ChatPanelComponent } from './components/chat-panel/chat-panel.component';

@Component({
  selector: 'app-root',
  imports: [
    RouterOutlet,
    MatToolbar,
    MatTabGroup,
    MatTab,
    MatCard,
    MatCardContent,
    MatIconButton,
    MatIcon,
    MatTooltipModule,
    PrebuiltMaterialComponent,
    FormulaAnalysisComponent,
    FormulaListComponent,
    TestOutlineComponent,
    ChatPanelComponent,
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class App {
  private readonly renderer = inject(Renderer2);

  readonly sidebarCollapsed = signal(true);
  readonly isDark = signal(false);

  constructor() {
    let stored: string | null = null;
    try {
      stored = localStorage.getItem('theme');
    } catch {
      // localStorage unavailable (e.g., private browsing)
    }
    if (stored === 'dark') {
      this.isDark.set(true);
    } else if (!stored) {
      this.isDark.set(window.matchMedia('(prefers-color-scheme: dark)').matches);
    }
    effect(() => {
      const dark = this.isDark();
      if (dark) {
        this.renderer.addClass(document.documentElement, 'dark-theme');
        try { localStorage.setItem('theme', 'dark'); } catch { /* noop */ }
      } else {
        this.renderer.removeClass(document.documentElement, 'dark-theme');
        try { localStorage.setItem('theme', 'light'); } catch { /* noop */ }
      }
    });
  }

  toggleDarkMode(): void {
    this.isDark.update(v => !v);
  }

  toggleSidebar(): void {
    this.sidebarCollapsed.update((v) => !v);
  }
}
