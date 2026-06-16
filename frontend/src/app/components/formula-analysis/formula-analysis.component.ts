import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { provideCharts, withDefaultRegisterables, BaseChartDirective } from 'ng2-charts';
import { ChartType, ChartData, ChartOptions } from 'chart.js';
import { MatCard, MatCardContent, MatCardHeader, MatCardTitle } from '@angular/material/card';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { MatIcon } from '@angular/material/icon';
import { FormulaApiService } from '../../services/formula-api.service';
import { ChartAnalysis, DosingMethod } from '../../types/analysis.types';
import { forkJoin, catchError, of, finalize, tap } from 'rxjs';

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
  providers: [provideCharts(withDefaultRegisterables())],
  templateUrl: './formula-analysis.component.html',
  styleUrl: './formula-analysis.component.css',
})
export class FormulaAnalysisComponent implements OnInit {
  private readonly api = inject(FormulaApiService);

  // ── Chart States ──

  modeRatio: ChartState = { loading: true, error: null, data: null };
  ingredientFreq: ChartState = { loading: true, error: null, data: null };
  stepDistribution: ChartState = { loading: true, error: null, data: null };
  dosingMethods: DosingState = { loading: true, error: null, data: null };

  // ── Table columns ──
  dosingColumns: string[] = ['name', 'count'];

  // ── Chart data bindings ──

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
    plugins: {
      legend: { display: false },
    },
    scales: {
      x: { ticks: { stepSize: 1 }, title: { display: true, text: '使用次数' } },
      y: { title: { display: true, text: '原料' } },
    },
  };

  stepDistributionChartData: ChartData<'bar'> = { labels: [], datasets: [] };
  stepDistributionChartOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
    },
    scales: {
      x: { title: { display: true, text: '步骤数范围' } },
      y: { ticks: { stepSize: 1 }, title: { display: true, text: '配方数' } },
    },
  };

  private readonly chartColors = {
    pie: [
      'var(--mat-sys-primary, #6750A4)',
      'var(--mat-sys-secondary, #625B71)',
      'var(--mat-sys-tertiary, #7D5260)',
      'var(--mat-sys-primary-container, #EADDFF)',
      'var(--mat-sys-secondary-container, #E8DEF8)',
    ],
    bar: 'var(--mat-sys-primary, #6750A4)',
  };

  ngOnInit(): void {
    this.loadAll();
  }

  /** Fire all 4 analysis requests, map responses into chart/table shapes. */
  loadAll(): void {
    forkJoin([
      this.api.getAnalysis('component-mode-ratio').pipe(
        tap((res) => this.setChartState(this.modeRatio, res, 'pie')),
        catchError((err) => { this.modeRatio = { loading: false, error: this.extractError(err), data: null }; return of(null); }),
      ),
      this.api.getAnalysis('ingredient-distribution').pipe(
        tap((res) => this.setChartState(this.ingredientFreq, res, 'bar-horizontal')),
        catchError((err) => { this.ingredientFreq = { loading: false, error: this.extractError(err), data: null }; return of(null); }),
      ),
      this.api.getAnalysis('step-count-distribution').pipe(
        tap((res) => this.setChartState(this.stepDistribution, res, 'bar-vertical')),
        catchError((err) => { this.stepDistribution = { loading: false, error: this.extractError(err), data: null }; return of(null); }),
      ),
      this.api.getAnalysis('dosing-method-stats').pipe(
        tap((res: any) => this.setDosingState(res)),
        catchError((err) => { this.dosingMethods = { loading: false, error: this.extractError(err), data: null }; return of(null); }),
      ),
    ]).pipe(
      finalize(() => {
        // finalize ensures loading completes even if some observables fall through catchError
      }),
    ).subscribe();
  }

  private setChartState(state: ChartState, res: unknown, kind: 'pie' | 'bar-horizontal' | 'bar-vertical'): void {
    const data = res as ChartAnalysis;
    if (!data || !data.labels || !data.values || data.labels.length === 0) {
      state.data = null;
      state.loading = false;
      state.error = null;
      return;
    }

    state.data = data;
    state.loading = false;
    state.error = null;

    if (kind === 'pie') {
      this.modeRatioChartData = {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: this.chartColors.pie.slice(0, data.labels.length),
          borderColor: 'var(--mat-sys-surface, #FEF7FF)',
          borderWidth: 2,
        }],
      };
    } else if (kind === 'bar-horizontal') {
      this.ingredientFreqChartData = {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: this.chartColors.bar,
          borderColor: 'var(--mat-sys-on-primary-container, #21005D)',
          borderWidth: 1,
          borderRadius: 4,
        }],
      };
    } else {
      this.stepDistributionChartData = {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: data.values.map(
            (_, i) => this.chartColors.pie[i % this.chartColors.pie.length],
          ),
          borderColor: 'var(--mat-sys-outline-variant, #CAC4D0)',
          borderWidth: 1,
          borderRadius: 4,
        }],
      };
    }
  }

  private setDosingState(res: unknown): void {
    const methods = (res as { methods?: DosingMethod[] })?.methods ?? null;
    if (!methods || methods.length === 0) {
      this.dosingMethods = { loading: false, error: null, data: null };
      return;
    }

    this.dosingMethods = {
      loading: false,
      error: null,
      data: [...methods].sort((a, b) => b.count - a.count),
    };
  }

  private extractError(err: any): string {
    if (typeof err === 'string') return err;
    return err?.message ?? err?.error?.message ?? '数据加载失败，请稍后重试';
  }
}
