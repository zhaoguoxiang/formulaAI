import { Component, OnInit, OnDestroy, inject, signal, DestroyRef, ChangeDetectionStrategy, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { animate, state, style, transition, trigger } from '@angular/animations';

import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatSortModule, MatSort, Sort } from '@angular/material/sort';
import { MatButtonToggleModule, MatButtonToggleChange } from '@angular/material/button-toggle';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { FormsModule } from '@angular/forms';

import { FormulaApiService } from '../../services/formula-api.service';
import { Formula, FormulaPart, FormulaIngredientCategory, FormulaIngredient } from '../../types/formula.types';
import { getModeClass, getModeLabel, getStatusLabel, getPartNameLabel, getPartNameClass } from '../../utils/formula-labels';
import { extractErrorMessage } from '../../utils/error.utils';
import { FormulaEditorComponent } from '../formula-editor/formula-editor.component';

@Component({
  selector: 'app-formula-matrix',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    MatTableModule,
    MatSortModule,
    MatButtonToggleModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatButtonModule,
    MatCardModule,
    MatChipsModule,
    MatDividerModule,
    MatTooltipModule,
    MatFormFieldModule,
    MatInputModule,
    MatDialogModule,
    MatSnackBarModule,
    FormsModule,
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
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);

  private readonly _sort = viewChild(MatSort);

  readonly displayedColumns: string[] = [
    'expand', 'code', 'name', 'component_mode', 'parts_count', 'steps_count', 'status', 'actions',
  ];

  dataSource = new MatTableDataSource<Formula>([]);
  expandedElement: Formula | null = null;

  readonly modeOptions: { value: string; label: string }[] = [
    { value: '', label: '所有配方' },
    { value: 'single', label: '单组分' },
    { value: 'double', label: '双组分' },
  ];

  selectedMode = '';
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  searchText = '';

  ngOnInit(): void {
    this.loadMatrix();
  }

  ngOnDestroy(): void {}

  ngAfterViewInit(): void {
    const sort = this._sort();
    if (sort) {
      this.dataSource.sort = sort;
      this.dataSource.sortingDataAccessor = (data: Formula, sortHeaderId: string) => {
        switch (sortHeaderId) {
          case 'code': return data.code;
          case 'name': return data.name;
          case 'component_mode': return getModeLabel(data.component_mode);
          case 'status': return getStatusLabel(data.status);
          case 'parts_count': return data.parts?.length ?? 0;
          case 'steps_count': return data.steps?.length ?? 0;
          default: return '';
        }
      };
    }
  }

  onSearchChange(value: string): void {
    this.dataSource.filter = value.trim().toLowerCase();
  }

  loadMatrix(): void {
    this.loading.set(true);
    this.error.set(null);
    this.expandedElement = null;

    const mode = this.selectedMode || undefined;
    this.formulaApi.getFormulaMatrix(mode)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (matrix) => {
          this.dataSource.data = matrix.formulas ?? [];
          this.dataSource.filterPredicate = (data: Formula, filter: string) => {
            const search = filter.toLowerCase();
            return (
              data.name.toLowerCase().includes(search) ||
              data.code.toLowerCase().includes(search) ||
              getModeLabel(data.component_mode).includes(search) ||
              getStatusLabel(data.status).includes(search)
            );
          };
          this.loading.set(false);
        },
        error: (err) => {
          console.error('[FormulaMatrix] Failed to load matrix', err);
          this.error.set(extractErrorMessage(err, '加载配方数据失败，请检查后端服务'));
          this.loading.set(false);
          this.dataSource.data = [];
        },
      });
  }

  openNewFormulaDialog(): void {
    const dialogRef = this.dialog.open(FormulaEditorComponent, {
      width: '90vw',
      maxWidth: '90vw',
      maxHeight: '92vh',
      disableClose: true,
      autoFocus: false,
    });

    const sub = dialogRef.componentInstance.saved.subscribe(() => {
      dialogRef.close();
      this.loadMatrix();
    });

    dialogRef.componentInstance.cancelled.subscribe(() => {
      dialogRef.close();
    });

    dialogRef.afterClosed().subscribe(() => {
      sub.unsubscribe();
    });
  }

  onModeFilterChange(event: MatButtonToggleChange): void {
    this.selectedMode = event.value;
    this.searchText = '';
    this.dataSource.filter = '';
    this.loadMatrix();
  }

  toggleRow(formula: Formula): void {
    this.expandedElement = this.expandedElement === formula ? null : formula;
  }

  isExpanded(formula: Formula): boolean {
    return this.expandedElement === formula;
  }

  openEditDialog(formula: Formula, event: Event): void {
    event.stopPropagation();
    const dialogRef = this.dialog.open(FormulaEditorComponent, {
      width: '90vw',
      maxWidth: '90vw',
      maxHeight: '92vh',
      disableClose: true,
      autoFocus: false,
    });
    dialogRef.componentInstance.formulaId = formula.id;

    const sub = dialogRef.componentInstance.saved.subscribe(() => {
      dialogRef.close();
      this.loadMatrix();
    });

    dialogRef.componentInstance.cancelled.subscribe(() => {
      dialogRef.close();
    });

    dialogRef.afterClosed().subscribe(() => {
      sub.unsubscribe();
    });
  }

  deleteFormula(formula: Formula, event: Event): void {
    event.stopPropagation();
    const snackRef = this.snackBar.open(
      `确定要删除配方「${formula.name}」吗？`,
      '确认删除',
      { duration: 8000, politeness: 'assertive' }
    );
    snackRef.onAction().subscribe(() => {
      this.loading.set(true);
      this.formulaApi.deleteFormula(formula.id)
        .pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe({
          next: () => {
            this.snackBar.open(`已删除配方「${formula.name}」`, '关闭', { duration: 3000 });
            this.loadMatrix();
          },
          error: (err) => {
            this.loading.set(false);
            this.snackBar.open(`删除失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
          },
        });
    });
  }

  cloneFormula(formula: Formula, event: Event): void {
    event.stopPropagation();
    this.loading.set(true);
    this.formulaApi.getFormula(formula.id)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (full) => {
          const cloneData = JSON.parse(JSON.stringify(full)) as Record<string, unknown>;
          delete cloneData['id'];
          delete cloneData['code'];
          delete cloneData['created_at'];
          delete cloneData['updated_at'];
          cloneData['name'] = full.name;
          cloneData['status'] = 'draft';
          this.stripNestedIds(cloneData);
          this.formulaApi.createFormula(cloneData as Partial<Formula>)
            .pipe(takeUntilDestroyed(this.destroyRef))
            .subscribe({
              next: () => {
                this.loading.set(false);
                this.snackBar.open(`已克隆配方「${full.name}」→「${cloneData['name']}」`, '关闭', { duration: 3000 });
                this.loadMatrix();
              },
              error: (err) => {
                this.loading.set(false);
                this.snackBar.open(`克隆失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
              },
            });
        },
        error: (err) => {
          this.loading.set(false);
          this.snackBar.open(`获取配方失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
        },
      });
  }

  private stripNestedIds(data: Record<string, unknown>): void {
    for (const part of (data['parts'] as Array<Record<string, unknown>> || [])) {
      delete part['id'];
      delete part['formula_id'];
      for (const cat of (part['categories'] as Array<Record<string, unknown>> || [])) {
        delete cat['id'];
        for (const ing of (cat['ingredients'] as Array<Record<string, unknown>> || [])) {
          delete ing['id'];
          delete ing['category_id'];
        }
      }
    }
    for (const step of (data['steps'] as Array<Record<string, unknown>> || [])) {
      delete step['id'];
      delete step['formula_id'];
      for (const cat of (step['categories'] as Array<Record<string, unknown>> || [])) {
        delete cat['id'];
        delete cat['step_id'];
        for (const mat of (cat['materials'] as Array<Record<string, unknown>> || [])) {
          delete mat['id'];
          delete mat['category_id'];
        }
      }
      for (const param of (step['parameters'] as Array<Record<string, unknown>> || [])) {
        delete param['id'];
        delete param['step_id'];
      }
    }
  }

  totalIngredientCount(parts: FormulaPart[]): number {
    let count = 0;
    for (const part of parts) {
      for (const cat of part.categories ?? []) {
        count += cat.ingredients?.length ?? 0;
      }
    }
    return count;
  }

  flatIngredients(part: FormulaPart): FormulaIngredient[] {
    const result: FormulaIngredient[] = [];
    for (const cat of part.categories ?? []) {
      for (const ing of cat.ingredients ?? []) {
        result.push(ing);
      }
    }
    return result;
  }

  get sort(): MatSort | null {
    return this._sort() ?? null;
  }

  getModeLabel = getModeLabel;
  getStatusLabel = getStatusLabel;
  getPartNameLabel = getPartNameLabel;
  getModeClass = getModeClass;
  getPartNameClass = getPartNameClass;
}
