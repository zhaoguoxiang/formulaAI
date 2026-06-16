import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
  Formula,
  FormulaMatrix,
} from '../types/formula.types';

@Injectable({
  providedIn: 'root',
})
export class FormulaApiService {
  private readonly baseUrl = '/api';

  constructor(private readonly http: HttpClient) {}

  /** Fetch all formulas. */
  getFormulas(): Observable<Formula[]> {
    return this.http.get<Formula[]>(`${this.baseUrl}/formulas`);
  }

  /** Fetch a single formula by ID. */
  getFormula(id: string): Observable<Formula> {
    return this.http.get<Formula>(`${this.baseUrl}/formulas/${id}`);
  }

  /** Create a new formula. */
  createFormula(data: Partial<Formula>): Observable<Formula> {
    return this.http.post<Formula>(`${this.baseUrl}/formulas`, data);
  }

  /** Update an existing formula. */
  updateFormula(id: string, data: Partial<Formula>): Observable<Formula> {
    return this.http.put<Formula>(`${this.baseUrl}/formulas/${id}`, data);
  }

  /** Delete a formula by ID. */
  deleteFormula(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/formulas/${id}`);
  }

  /** Fetch the formula matrix (cross-tab view), optionally filtered by component mode. */
  getFormulaMatrix(componentMode?: string): Observable<FormulaMatrix> {
    let params = new HttpParams();
    if (componentMode) {
      params = params.set('component_mode', componentMode);
    }
    return this.http.get<FormulaMatrix>(`${this.baseUrl}/formulas/matrix`, { params });
  }

  /** Fetch analysis data from a named endpoint. */
  getAnalysis(endpoint: string): Observable<unknown> {
    return this.http.get<unknown>(`${this.baseUrl}/analysis/${endpoint}`);
  }
}
