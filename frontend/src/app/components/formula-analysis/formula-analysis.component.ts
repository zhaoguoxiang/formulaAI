import { Component, ChangeDetectionStrategy } from '@angular/core';
import { MatCard, MatCardContent, MatCardHeader, MatCardTitle } from '@angular/material/card';
import { MatIcon } from '@angular/material/icon';

@Component({
  selector: 'app-formula-analysis',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    MatCard,
    MatCardContent,
    MatCardHeader,
    MatCardTitle,
    MatIcon,
  ],
  templateUrl: './formula-analysis.component.html',
  styleUrl: './formula-analysis.component.scss',
})
export class FormulaAnalysisComponent {}
