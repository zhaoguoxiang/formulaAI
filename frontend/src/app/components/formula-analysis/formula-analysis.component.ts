import { Component, OnInit, OnDestroy, inject, DestroyRef, ChangeDetectionStrategy } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { BaseChartDirective } from 'ng2-charts';
import { ChartData, ChartOptions } from 'chart.js';
import { MatCard, MatCardContent, MatCardHeader, MatCardTitle } from '@angular/material/card';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { MatIcon } from '@angular/material/icon';
import { FormulaApiService } from '../../services/formula-api.service';
import { ChartAnalysis, DosingMethod } from '../../types/analysis.types';
import { forkJoin, catchError, of } from 'rxjs';

interface ChartState {
  loading: boolean;
  error: string | null;
  data: ChartAnalysis | null;
}

interface DosingState {
  loading: boolean;
  error: string | null;
  data: DosingMethod[] | null;
}

@Component({
  selector: 'app-formula-analysis',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    BaseChartDirective,
    MatCard,
    MatCardContent,
    MatCardHeader,
    MatCardTitle,
    MatProgressSpinner,
    MatTableModule,
    MatIcon,
  ],
  templateUrl: './formula-analysis.component.html',
  styleUrl: './formula-analysis.component.scss',
})
export class FormulaAnalysisComponent implements OnInit, OnDestroy {
  private readonly api = inject(FormulaApiService);
  private readonly destroyRef = inject(DestroyRef);

  modeRatio: ChartState = { loading: true, error: null, data: null };
  ingredientFreq: ChartState = { loading: true, error: null, data: null };
  stepDistribution: ChartState = { loading: true, error: null, data: null };
  dosingMethods: DosingState = { loading: true, error: null, data: null };

  dosingColumns: string[] = ['name', 'count'];

  modeRatioChartData: ChartData<'pie'> = { labels: [], datasets: [] };
  modeRatioChartOptions: ChartOptions<'pie'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { position: 'bottom', labels: { padding: 20, usePointStyle: true } },
    },
  };

  ingredientFreqChartData: ChartData<'bar'> = { labels: [], datasets: [] };
  ingredientFreqChartOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    indexAxis: 'y',
    plugins: { legend: { display: false } },
    scales: {
      x: { ticks: { stepSize: 1 }, title: { display: true, text: '使用次数' } },
      y: { title: { display: true, text: '原料' } },
    },
  };

  stepDistributionChartData: ChartData<'bar'> = { labels: [], datasets: [] };
  stepDistributionChartOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { title: { display: true, text: '步骤数范围' } },
      y: { ticks: { stepSize: 1 }, title: { display: true, text: '配方数' } },
    },
  };

  private readonly chartColors = {
    pie: [
      'var(--mat-sys-color-primary, #6750A4)',
      'var(--mat-sys-color-secondary, #625B71)',
      'var(--mat-sys-color-tertiary, #7D5260)',
      'var(--mat-sys-color-primary-container, #EADDFF)',
      'var(--mat-sys-color-secondary-container, #E8DEF8)',
    ],
    bar: 'var(--mat-sys-color-primary, #6750A4)',
  };

  ngOnInit(): void {
    this.loadAll();
  }

  ngOnDestroy(): void {}

  loadAll(): void {
    forkJoin([
      this.api.getAnalysis('component-mode-ratio').pipe(catchError(() => of(null))),
      this.api.getAnalysis('ingredient-distribution').pipe(catchError(() => of(null))),
      this.api.getAnalysis('step-count-distribution').pipe(catchError(() => of(null))),
      this.api.getAnalysis('dosing-method-stats').pipe(catchError(() => of(null))),
    ]).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: ([modeRatioRes, ingredientFreqRes, stepDistRes, dosingRes]) => {
        this.modeRatio = modeRatioRes === null
          ? { loading: false, error: '数据加载失败，请稍后重试', data: null }
          : this.handleChartState(modeRatioRes, 'pie');
        this.ingredientFreq = ingredientFreqRes === null
          ? { loading: false, error: '数据加载失败，请稍后重试', data: null }
          : this.handleChartState(ingredientFreqRes, 'bar-horizontal');
        this.stepDistribution = stepDistRes === null
          ? { loading: false, error: '数据加载失败，请稍后重试', data: null }
          : this.handleChartState(stepDistRes, 'bar-vertical');
        this.dosingMethods = dosingRes === null
          ? { loading: false, error: '数据加载失败，请稍后重试', data: null }
          : this.buildDosingState(dosingRes);
      },
    });
  }

  private handleChartState(res: unknown, kind: 'pie' | 'bar-horizontal' | 'bar-vertical'): ChartState {
    const data = res as ChartAnalysis;
    if (!data?.labels || !data?.values || data.labels.length === 0) {
      return { loading: false, error: null, data: null };
    }

    if (kind === 'pie') {
      this.modeRatioChartData = {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: this.chartColors.pie.slice(0, data.labels.length),
          borderColor: 'var(--mat-sys-color-surface, #FEF7FF)',
          borderWidth: 2,
        }],
      };
    } else if (kind === 'bar-horizontal') {
      this.ingredientFreqChartData = {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: this.chartColors.bar,
          borderColor: 'var(--mat-sys-color-on-primary-container, #21005D)',
          borderWidth: 1,
          borderRadius: 4,
        }],
      };
    } else {
      this.stepDistributionChartData = {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: data.values.map((_, i) => this.chartColors.pie[i % this.chartColors.pie.length]),
          borderColor: 'var(--mat-sys-color-on-primary-container, #21005D)',
          borderWidth: 1,
          borderRadius: 4,
        }],
      };
    }
    return { loading: false, error: null, data };
  }

  private buildDosingState(res: unknown): DosingState {
    const methods = (res as { methods?: DosingMethod[] })?.methods ?? null;
    if (!methods || methods.length === 0) {
      return { loading: false, error: null, data: null };
    }
    return {
      loading: false,
      error: null,
      data: [...methods].sort((a, b) => b.count - a.count),
    };
  }
}
