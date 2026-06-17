import { Component, ViewChild, ChangeDetectionStrategy } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { MatToolbar } from '@angular/material/toolbar';
import { MatTabGroup, MatTab } from '@angular/material/tabs';
import { MatCard, MatCardContent } from '@angular/material/card';

import { FormulaAnalysisComponent } from './components/formula-analysis/formula-analysis.component';
import { FormulaMatrixComponent } from './components/formula-matrix/formula-matrix.component';
import { FormulaEditorComponent } from './components/formula-editor/formula-editor.component';
import { TestOutlineComponent } from './components/test-outline/test-outline.component';
import { ChatPanelComponent } from './components/chat-panel/chat-panel.component';
import { Formula } from './types/formula.types';

@Component({
  selector: 'app-root',
  imports: [
    RouterOutlet,
    MatToolbar,
    MatTabGroup,
    MatTab,
    MatCard,
    MatCardContent,
    FormulaAnalysisComponent,
    FormulaMatrixComponent,
    FormulaEditorComponent,
    TestOutlineComponent,
    ChatPanelComponent,
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class App {
  title = 'FormulAI';

  @ViewChild(MatTabGroup) tabGroup?: MatTabGroup;

  onFormulaSaved(_formula: Formula): void {
    if (this.tabGroup) {
      this.tabGroup.selectedIndex = 0;
    }
  }

  onEditorCancelled(): void {
    if (this.tabGroup) {
      this.tabGroup.selectedIndex = 0;
    }
  }
}
