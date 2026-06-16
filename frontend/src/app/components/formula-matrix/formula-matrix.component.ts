import { Component, OnInit } from '@angular/core';
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
import {
  Formula,
  FormulaPart,
  FormulaIngredient,
  ComponentMode,
  FormulaStatus,
} from '../../types/formula.types';

/**
 * Cross-tabulation matrix table showing formulas with expandable rows
 * that reveal nested parts, ingredients, and steps.
 *
 * Filterable by component_mode (All / Single / Double).
 */
@Component({
  selector: 'app-formula-matrix',
  standalone: true,
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
  styleUrl: './formula-matrix.component.css',
  animations: [
    trigger('detailExpand', [
      state('collapsed, void', style({ height: '0px', minHeight: '0', visibility: 'hidden', overflow: 'hidden' })),
      state('expanded', style({ height: '*' })),
      transition('expanded <=> collapsed', animate('225ms cubic-bezier(0.4, 0.0, 0.2, 1)')),
    ]),
  ],
})
export class FormulaMatrixComponent implements OnInit {
  /* ── Table configuration ── */
  readonly displayedColumns: string[] = [
    'expand',
    'code',
    'name',
    'component_mode',
    'parts_count',
    'steps_count',
    'status',
  ];

  dataSource = new MatTableDataSource<Formula>([]);

  /** The currently expanded formula (only one row expanded at a time). */
  expandedElement: Formula | null = null;

  /* ── Filter ── */
  readonly modeOptions: { value: string; label: string }[] = [
    { value: '', label: '所有配方' },
    { value: 'single', label: '单组分' },
    { value: 'double', label: '双组分' },
  ];

  selectedMode = '';

  /* ── State ── */
  loading = false;
  error: string | null = null;

  constructor(private readonly formulaApi: FormulaApiService) {}

  ngOnInit(): void {
    this.loadMatrix();
  }

  /* ── Data loading ── */

  loadMatrix(): void {
    this.loading = true;
    this.error = null;
    this.expandedElement = null;

    const mode = this.selectedMode || undefined;
    this.formulaApi.getFormulaMatrix(mode).subscribe({
      next: (matrix) => {
        this.dataSource.data = matrix.formulas ?? [];
        this.loading = false;
      },
      error: (err) => {
        // eslint-disable-next-line no-console
        console.error('[FormulaMatrix] Failed to load matrix', err);
        this.error = err?.message ?? '加载配方数据失败，请检查后端服务';
        this.loading = false;
        this.dataSource.data = [];
      },
    });
  }

  /* ── Filter ── */

  onModeFilterChange(event: MatButtonToggleChange): void {
    this.selectedMode = event.value;
    this.loadMatrix();
  }

  /* ── Row expansion ── */

  toggleRow(formula: Formula): void {
    // Close if clicking the same row; otherwise open the new one
    this.expandedElement = this.expandedElement === formula ? null : formula;
  }

  isExpanded(formula: Formula): boolean {
    return this.expandedElement === formula;
  }

  /* ── Display helpers ── */

  getModeLabel(mode: ComponentMode): string {
    return mode === 'single' ? '单组分' : '双组分';
  }

  getModeClass(mode: ComponentMode): string {
    return mode === 'single' ? 'mode-single' : 'mode-double';
  }

  getStatusLabel(status: FormulaStatus): string {
    const map: Record<FormulaStatus, string> = {
      draft: '草稿',
      active: '已启用',
      archived: '已归档',
    };
    return map[status] ?? status;
  }

  getPartNameLabel(name: string): string {
    const map: Record<string, string> = {
      PartA: 'A 组分',
      PartB: 'B 组分',
      PartMain: '主组分',
    };
    return map[name] ?? name;
  }

  getPartNameClass(name: string): string {
    const map: Record<string, string> = {
      PartA: 'part-a',
      PartB: 'part-b',
      PartMain: 'part-main',
    };
    return map[name] ?? '';
  }

  getIngredientCount(parts: FormulaPart[]): number {
    return parts.reduce((sum, p) => sum + (p.ingredients?.length ?? 0), 0);
  }
}
