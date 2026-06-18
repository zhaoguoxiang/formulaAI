import { Component, OnInit, OnDestroy, signal, inject, DestroyRef, ChangeDetectionStrategy } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { CdkDragDrop, DragDropModule, moveItemInArray } from '@angular/cdk/drag-drop';

import { TestOutlineApiService } from '../../services/test-outline-api.service';
import { TestOutline, VersionSummary } from '../../types/test-outline.types';
import { extractErrorMessage } from '../../utils/error.utils';
import { VditorFieldComponent } from '../milkdown-field/milkdown-field.component';
import { VersionNoteDialogComponent } from '../version-note-dialog/version-note-dialog.component';

interface EditFormIndicator {
  id: string;
  name: string;
  unit: string;
  min_value: number | null;
  max_value: number | null;
  sample_prep_method: string;
  test_method: string;
  test_condition: string;
}

interface EditFormItem {
  id: string;
  name: string;
  indicators: EditFormIndicator[];
}

interface EditFormData {
  id: string;
  items: EditFormItem[];
}

@Component({
  selector: 'app-test-outline',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    MatTooltipModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatSnackBarModule,
    MatDialogModule,
    VditorFieldComponent,
    VersionNoteDialogComponent,
    DragDropModule,
  ],
  templateUrl: './test-outline.component.html',
  styleUrl: './test-outline.component.scss',
})
export class TestOutlineComponent implements OnInit, OnDestroy {
  private readonly api = inject(TestOutlineApiService);
  private readonly snackBar = inject(MatSnackBar);
  private readonly dialog = inject(MatDialog);
  private readonly destroyRef = inject(DestroyRef);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly saving = signal(false);

  readonly editOutlineVersion = signal(0);
  readonly collapsedItems = signal<Set<number>>(new Set());
  readonly collapsedIndicators = signal<Set<string>>(new Set());
  readonly versions = signal<VersionSummary[]>([]);
  readonly selectedVersionId = signal('');
  readonly indicatorSearch = signal('');
  readonly dirty = signal(false);

  private initialFormJson = '';
  private isViewingHistory = false;

  editForm: EditFormData = this.emptyForm();

  ngOnInit(): void {
    this.loadOutline();
  }

  ngOnDestroy(): void {}

  loadOutline(): void {
    this.loading.set(true);
    this.error.set(null);

    this.api.getOutlines()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (data) => {
          if (data.length > 0) {
            const outline = data[0];
            this.populateForm(outline);
            this.loadVersions(outline.id);
          }
          this.loading.set(false);
        },
        error: (err) => {
          console.error('Failed to load test outline', err);
          this.error.set(extractErrorMessage(err, '加载测试大纲失败'));
          this.loading.set(false);
        },
      });
  }

  private loadVersions(outlineId: string): void {
    this.api.getVersions(outlineId)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (list) => {
          this.versions.set(list);
          this.selectedVersionId.set(outlineId);
        },
        error: (err) => {
          console.error('Failed to load versions', err);
        },
      });
  }

  onVersionChange(versionId: string): void {
    if (!versionId || versionId === this.selectedVersionId()) return;
    this.loading.set(true);
    this.isViewingHistory = true;
    this.api.getOutline(versionId)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (outline) => {
          this.populateForm(outline);
          this.selectedVersionId.set(versionId);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.snackBar.open(`加载版本失败: ${extractErrorMessage(err)}`, '关闭', { duration: 5000 });
        },
      });
  }

  private populateForm(outline: TestOutline): void {
    this.editOutlineVersion.set(outline.version);
    this.editForm = {
      id: outline.id,
      items: outline.items.map((item) => ({
        id: item.id,
        name: item.name,
        indicators: item.indicators.map((ind) => ({
          id: ind.id,
          name: ind.name,
          unit: ind.unit ?? '',
          min_value: ind.min_value ?? null,
          max_value: ind.max_value ?? null,
          sample_prep_method: ind.sample_prep_method ?? '',
          test_method: ind.test_method ?? '',
          test_condition: ind.test_condition ?? '',
        })),
      })),
    };
    this.initialFormJson = JSON.stringify(this.editForm);
    this.dirty.set(false);
    this.collapsedItems.set(new Set());
    this.collapsedIndicators.set(new Set());
  }

  markDirty(): void {
    if (this.isViewingHistory) return;
    const currentJson = JSON.stringify(this.editForm);
    this.dirty.set(currentJson !== this.initialFormJson);
  }

  saveOutline(): void {
    if (!this.editForm) return;

    const dialogRef = this.dialog.open(VersionNoteDialogComponent, {
      width: '440px',
      disableClose: true,
    });

    dialogRef.afterClosed()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((note: string | null) => {
        if (note === null) return;

        this.saving.set(true);

        const payload = {
          name: '测试大纲',
          version_note: note,
          items: this.editForm!.items.map((item, itemIdx) => ({
            id: item.id,
            name: item.name.trim(),
            sort_order: itemIdx,
            indicators: item.indicators.map((ind, indIdx) => ({
              id: ind.id,
              name: ind.name.trim(),
              unit: ind.unit || undefined,
              min_value: ind.min_value ?? undefined,
              max_value: ind.max_value ?? undefined,
              sample_prep_method: ind.sample_prep_method || undefined,
              test_method: ind.test_method || undefined,
              test_condition: ind.test_condition || undefined,
              sort_order: indIdx,
            })),
          })),
        };

        const request$ = this.editForm!.id
          ? this.api.saveVersion(this.editForm!.id, payload)
          : this.api.createOutline(payload);

        request$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
          next: (saved) => {
            this.saving.set(false);
            const newVersion = saved.version ?? (this.editOutlineVersion() + 1);
            this.editOutlineVersion.set(newVersion);
            if (saved.id) {
              this.editForm!.id = saved.id;
              this.selectedVersionId.set(saved.id);
              this.loadVersions(saved.id);
            }
            this.isViewingHistory = false;
            this.initialFormJson = JSON.stringify(this.editForm);
            this.dirty.set(false);
            this.snackBar.open(`已保存为 v${newVersion}`, '关闭', { duration: 3000 });
          },
          error: (err) => {
            this.saving.set(false);
            const msg = extractErrorMessage(err, '保存失败');
            this.snackBar.open(`保存失败: ${msg}`, '关闭', { duration: 5000 });
          },
        });
      });
  }

  addItem(): void {
    this.editForm.items.push({ id: '', name: '', indicators: [] });
    this.markDirty();
  }

  removeItem(idx: number): void {
    this.editForm.items.splice(idx, 1);
    this.markDirty();
  }

  addIndicator(itemIdx: number): void {
    this.editForm.items[itemIdx].indicators.push(this.emptyIndicator());
    this.markDirty();
  }

  removeIndicator(itemIdx: number, indicatorIdx: number): void {
    this.editForm.items[itemIdx].indicators.splice(indicatorIdx, 1);
    this.markDirty();
  }

  dropIndicator(event: CdkDragDrop<EditFormIndicator[]>, itemIdx: number): void {
    moveItemInArray(
      this.editForm.items[itemIdx].indicators,
      event.previousIndex,
      event.currentIndex,
    );
    this.markDirty();
  }

  /* Collapse / Expand */
  isItemCollapsed(idx: number): boolean {
    return this.collapsedItems().has(idx);
  }

  toggleItemCollapse(idx: number): void {
    this.collapsedItems.update((s) => {
      s = new Set(s);
      if (s.has(idx)) { s.delete(idx); } else { s.add(idx); }
      return s;
    });
  }

  collapseAllItems(): void {
    const s = new Set<number>();
    for (let i = 0; i < this.editForm.items.length; i++) s.add(i);
    this.collapsedItems.set(s);
  }

  expandAllItems(): void {
    this.collapsedItems.set(new Set());
  }

  isAllItemsCollapsed(): boolean {
    return this.collapsedItems().size === this.editForm.items.length && this.editForm.items.length > 0;
  }

  isIndicatorCollapsed(itemIdx: number, indIdx: number): boolean {
    return this.collapsedIndicators().has(this.indicatorKey(itemIdx, indIdx));
  }

  toggleIndicatorCollapse(itemIdx: number, indIdx: number): void {
    const key = this.indicatorKey(itemIdx, indIdx);
    this.collapsedIndicators.update((s) => {
      s = new Set(s);
      if (s.has(key)) { s.delete(key); } else { s.add(key); }
      return s;
    });
  }

  collapseAllIndicators(): void {
    const s = new Set<string>();
    for (let i = 0; i < this.editForm.items.length; i++) {
      for (let j = 0; j < this.editForm.items[i].indicators.length; j++) {
        s.add(this.indicatorKey(i, j));
      }
    }
    this.collapsedIndicators.set(s);
  }

  expandAllIndicators(): void {
    this.collapsedIndicators.set(new Set());
  }

  isAllIndicatorsCollapsed(): boolean {
    let total = 0;
    for (const item of this.editForm.items) total += item.indicators.length;
    return this.collapsedIndicators().size === total && total > 0;
  }

  indicatorKey(itemIdx: number, indIdx: number): string {
    return `${itemIdx}:${indIdx}`;
  }

  /* Indicator search */
  indicatorMatchesSearch(itemIdx: number, indIdx: number): boolean {
    const search = this.indicatorSearch().trim().toLowerCase();
    if (!search) return true;
    const ind = this.editForm.items[itemIdx]?.indicators[indIdx];
    if (!ind) return false;
    return (
      ind.name.toLowerCase().includes(search) ||
      ind.unit.toLowerCase().includes(search) ||
      (ind.test_condition ?? '').toLowerCase().includes(search)
    );
  }

  itemHasMatchingIndicator(itemIdx: number): boolean {
    const search = this.indicatorSearch().trim().toLowerCase();
    if (!search) return true;
    const item = this.editForm.items[itemIdx];
    if (!item) return false;
    return item.indicators.some((ind) =>
      ind.name.toLowerCase().includes(search) ||
      ind.unit.toLowerCase().includes(search) ||
      (ind.test_condition ?? '').toLowerCase().includes(search)
    );
  }

  private emptyForm(): EditFormData {
    return { id: '', items: [] };
  }

  private emptyIndicator(): EditFormIndicator {
    return {
      id: '', name: '', unit: '', min_value: null, max_value: null,
      sample_prep_method: '', test_method: '', test_condition: '',
    };
  }
}
