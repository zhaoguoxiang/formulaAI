import { Component, OnInit, inject, signal, DestroyRef, ChangeDetectionStrategy, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { animate, state, style, transition, trigger } from '@angular/animations';
import { FormsModule } from '@angular/forms';

import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatSortModule, MatSort } from '@angular/material/sort';
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

import { Formula, FormulaPart } from '../../types/formula.types';
import { getStatusLabel } from '../../utils/formula-labels';
import { extractErrorMessage } from '../../utils/error.utils';
import { PrebuiltMaterialApiService } from '../../services/prebuilt-material-api.service';
import { FormulaEditorComponent } from '../formula-editor/formula-editor.component';

@Component({
  selector: 'app-prebuilt-material',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    FormsModule,
    MatTableModule,
    MatSortModule,
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
  ],
  templateUrl: './prebuilt-material.component.html',
  styleUrl: './prebuilt-material.component.scss',
  animations: [
    trigger('detailExpand', [
      state('collapsed, void', style({ height: '0px', minHeight: '0', visibility: 'hidden', overflow: 'hidden' })),
      state('expanded', style({ height: '*' })),
      transition('expanded <=> collapsed', animate('225ms cubic-bezier(0.4, 0.0, 0.2, 1)')),
    ]),
  ],
})
export class PrebuiltMaterialComponent implements OnInit {
  private readonly api = inject(PrebuiltMaterialApiService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);

  private readonly _sort = viewChild(MatSort);

  readonly displayedColumns: string[] = [
    'expand', 'code', 'name', 'labels', 'steps_count', 'status', 'actions',
  ];

  dataSource = new MatTableDataSource<Formula>([]);
  expandedElement: Formula | null = null;

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  searchText = '';

  ngOnInit(): void {
    this.loadMaterials();
  }

  ngAfterViewInit(): void {
    const sort = this._sort();
    if (sort) {
      this.dataSource.sort = sort;
      this.dataSource.sortingDataAccessor = (data: Formula, sortHeaderId: string) => {
        switch (sortHeaderId) {
          case 'code': return data.code;
          case 'name': return data.name;
          case 'status': return getStatusLabel(data.status);
          case 'steps_count': return data.steps?.length ?? 0;
          default: return '';
        }
      };
    }
  }

  onSearchChange(value: string): void {
    this.dataSource.filter = value.trim().toLowerCase();
  }

  loadMaterials(): void {
    this.loading.set(true);
    this.error.set(null);
    this.expandedElement = null;

    this.api.getPrebuiltMaterials()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (materials) => {
          this.dataSource.data = materials ?? [];
          this.dataSource.filterPredicate = (data: Formula, filter: string) => {
            const search = filter.toLowerCase();
            return (
              data.name.toLowerCase().includes(search) ||
              data.code.toLowerCase().includes(search) ||
              data.labels.some(l => l.toLowerCase().includes(search)) ||
              getStatusLabel(data.status).includes(search)
            );
          };
          this.loading.set(false);
        },
        error: (err) => {
          console.error('[PrebuiltMaterial] Failed to load', err);
          this.error.set(extractErrorMessage(err, '加载预制物料失败，请检查后端服务'));
          this.loading.set(false);
          this.dataSource.data = [];
        },
      });
  }

  openNewDialog(): void {
    const dialogRef = this.dialog.open(FormulaEditorComponent, {
      width: '90vw',
      maxWidth: '90vw',
      maxHeight: '92vh',
      disableClose: true,
      autoFocus: false,
    });
    dialogRef.componentInstance.isMaterialMode = true;

    const sub = dialogRef.componentInstance.saved.subscribe(() => {
      dialogRef.close();
      this.loadMaterials();
    });
    const cancelSub = dialogRef.componentInstance.cancelled.subscribe(() => {
      dialogRef.close();
    });
    dialogRef.afterClosed().subscribe(() => {
      sub.unsubscribe();
      cancelSub.unsubscribe();
    });
  }

  toggleRow(material: Formula): void {
    this.expandedElement = this.expandedElement === material ? null : material;
  }

  isExpanded(material: Formula): boolean {
    return this.expandedElement === material;
  }

  openEditDialog(material: Formula, event: Event): void {
    event.stopPropagation();
    const dialogRef = this.dialog.open(FormulaEditorComponent, {
      width: '90vw',
      maxWidth: '90vw',
      maxHeight: '92vh',
      disableClose: true,
      autoFocus: false,
    });
    dialogRef.componentInstance.formulaId = material.id;
    dialogRef.componentInstance.isMaterialMode = true;

    const sub = dialogRef.componentInstance.saved.subscribe(() => {
      dialogRef.close();
      this.loadMaterials();
    });
    const cancelSub = dialogRef.componentInstance.cancelled.subscribe(() => {
      dialogRef.close();
    });
    dialogRef.afterClosed().subscribe(() => {
      sub.unsubscribe();
      cancelSub.unsubscribe();
    });
  }

  deleteMaterial(material: Formula, event: Event): void {
    event.stopPropagation();
    const snackRef = this.snackBar.open(
      `确定要删除预制物料「${material.name}」吗？`,
      '确认删除',
      { duration: 8000, politeness: 'assertive' }
    );
    snackRef.onAction().subscribe(() => {
      this.loading.set(true);
      this.api.deletePrebuiltMaterial(material.id)
        .pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe({
          next: () => {
            this.snackBar.open(`已删除预制物料「${material.name}」`, '关闭', { duration: 3000 });
            this.loadMaterials();
          },
          error: (err) => {
            this.loading.set(false);
            this.snackBar.open(`删除失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
          },
        });
    });
  }

  cloneMaterial(material: Formula, event: Event): void {
    event.stopPropagation();
    this.loading.set(true);
    this.api.getPrebuiltMaterial(material.id)
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
          cloneData['formula_type'] = 'material';
          this.stripNestedIds(cloneData);
          this.api.createPrebuiltMaterial(cloneData as Partial<Formula>)
            .pipe(takeUntilDestroyed(this.destroyRef))
            .subscribe({
              next: () => {
                this.loading.set(false);
                this.snackBar.open(`已克隆预制物料「${full.name}」`, '关闭', { duration: 3000 });
                this.loadMaterials();
              },
              error: (err) => {
                this.loading.set(false);
                this.snackBar.open(`克隆失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
              },
            });
        },
        error: (err) => {
          this.loading.set(false);
          this.snackBar.open(`获取物料失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
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

  flatMaterials(part: FormulaPart): { material: string; percentage: number; weight: number; unit: string; batch_no: string }[] {
    const result: { material: string; percentage: number; weight: number; unit: string; batch_no: string }[] = [];
    for (const cat of part.categories ?? []) {
      for (const ing of cat.ingredients ?? []) {
        result.push({ material: ing.material, percentage: ing.percentage, weight: ing.weight, unit: ing.unit, batch_no: ing.batch_no });
      }
    }
    return result;
  }

  getStatusLabel = getStatusLabel;
}
