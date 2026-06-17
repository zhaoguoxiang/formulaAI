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
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';

import { TestOutlineApiService } from '../../services/test-outline-api.service';
import { TestOutline } from '../../types/test-outline.types';
import { extractErrorMessage } from '../../utils/error.utils';
import { VditorFieldComponent } from '../milkdown-field/milkdown-field.component';

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
    MatSnackBarModule,
    VditorFieldComponent,
  ],
  templateUrl: './test-outline.component.html',
  styleUrl: './test-outline.component.scss',
})
export class TestOutlineComponent implements OnInit, OnDestroy {
  private readonly api = inject(TestOutlineApiService);
  private readonly snackBar = inject(MatSnackBar);
  private readonly destroyRef = inject(DestroyRef);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly saving = signal(false);

  readonly editOutlineVersion = signal(0);
  readonly collapsedItems = signal<Set<number>>(new Set());

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
            this.populateForm(data[0]);
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
  }

  saveOutline(): void {
    if (!this.editForm) return;

    this.saving.set(true);

    const payload = {
      name: '测试大纲',
      items: this.editForm.items.map((item, itemIdx) => ({
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

    const request$ = this.editForm.id
      ? this.api.saveVersion(this.editForm.id, payload)
      : this.api.createOutline(payload);

    request$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (saved) => {
        this.saving.set(false);
        const newVersion = saved.version ?? (this.editOutlineVersion() + 1);
        this.editOutlineVersion.set(newVersion);
        if (saved.id) this.editForm!.id = saved.id;
        this.snackBar.open(`已保存为 v${newVersion}`, '关闭', { duration: 3000 });
      },
      error: (err) => {
        this.saving.set(false);
        const msg = extractErrorMessage(err, '保存失败');
        this.snackBar.open(`保存失败: ${msg}`, '关闭', { duration: 5000 });
      },
    });
  }

  addItem(): void {
    this.editForm.items.push({ id: '', name: '', indicators: [] });
  }

  removeItem(idx: number): void {
    this.editForm.items.splice(idx, 1);
  }

  addIndicator(itemIdx: number): void {
    this.editForm.items[itemIdx].indicators.push(this.emptyIndicator());
  }

  removeIndicator(itemIdx: number, indicatorIdx: number): void {
    this.editForm.items[itemIdx].indicators.splice(indicatorIdx, 1);
  }

  isItemCollapsed(idx: number): boolean {
    return this.collapsedItems().has(idx);
  }

  toggleItemCollapse(idx: number): void {
    this.collapsedItems.update((s) => {
      s = new Set(s);
      if (s.has(idx)) {
        s.delete(idx);
      } else {
        s.add(idx);
      }
      return s;
    });
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
