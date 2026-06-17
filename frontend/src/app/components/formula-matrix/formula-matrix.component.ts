import { Component, OnInit, OnDestroy, inject, DestroyRef, ChangeDetectionStrategy } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { animate, state, style, transition, trigger } from '@angular/animations';

import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatButtonToggleModule, MatButtonToggleChange } from '@angular/material/button-toggle';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';

import { FormulaApiService } from '../../services/formula-api.service';
import { Formula, FormulaPart, ComponentMode, FormulaStatus } from '../../types/formula.types';
import { getModeClass, getModeLabel, getStatusLabel, getPartNameLabel, getPartNameClass } from '../../utils/formula-labels';
import { extractErrorMessage } from '../../utils/error.utils';

@Component({
  selector: 'app-formula-matrix',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    MatTableModule,
    MatButtonToggleModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatButtonModule,
    MatCardModule,
    MatChipsModule,
    MatDividerModule,
    MatTooltipModule,
  ],
  templateUrl: './formula-matrix.component.html',
  styleUrl: './formula-matrix.component.scss',
  animations: [
    trigger('detailExpand', [
      state('collapsed, void', style({ height: '0px', minHeight: '0', visibility: 'hidden', overflow: 'hidden' })),
      state('expanded', style({ height: '*' })),
      transition('expanded <=> collapsed', animate('225ms cubic-bezier(0.4, 0.0, 0.2, 1)')),
    ]),
  ],
})
export class FormulaMatrixComponent implements OnInit, OnDestroy {
  private readonly formulaApi = inject(FormulaApiService);
  private readonly destroyRef = inject(DestroyRef);

  readonly displayedColumns: string[] = [
    'expand', 'code', 'name', 'component_mode', 'parts_count', 'steps_count', 'status',
  ];

  dataSource = new MatTableDataSource<Formula>([]);
  expandedElement: Formula | null = null;

  readonly modeOptions: { value: string; label: string }[] = [
    { value: '', label: '所有配方' },
    { value: 'single', label: '单组分' },
    { value: 'double', label: '双组分' },
  ];

  selectedMode = '';
  loading = false;
  error: string | null = null;

  ngOnInit(): void {
    this.loadMatrix();
  }

  ngOnDestroy(): void {}

  loadMatrix(): void {
    this.loading = true;
    this.error = null;
    this.expandedElement = null;

    const mode = this.selectedMode || undefined;
    this.formulaApi.getFormulaMatrix(mode)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (matrix) => {
          this.dataSource.data = matrix.formulas ?? [];
          this.loading = false;
        },
        error: (err) => {
          console.error('[FormulaMatrix] Failed to load matrix', err);
          this.error = extractErrorMessage(err, '加载配方数据失败，请检查后端服务');
          this.loading = false;
          this.dataSource.data = [];
        },
      });
  }

  onModeFilterChange(event: MatButtonToggleChange): void {
    this.selectedMode = event.value;
    this.loadMatrix();
  }

  toggleRow(formula: Formula): void {
    this.expandedElement = this.expandedElement === formula ? null : formula;
  }

  isExpanded(formula: Formula): boolean {
    return this.expandedElement === formula;
  }

  getModeLabel = getModeLabel;
  getStatusLabel = getStatusLabel;
  getPartNameLabel = getPartNameLabel;
  getModeClass = getModeClass;
  getPartNameClass = getPartNameClass;

  getIngredientCount(parts: FormulaPart[]): number {
    return parts.reduce((sum, p) => sum + (p.ingredients?.length ?? 0), 0);
  }
}
