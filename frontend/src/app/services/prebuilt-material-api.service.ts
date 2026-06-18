import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Formula } from '../types/formula.types';

@Injectable({ providedIn: 'root' })
export class PrebuiltMaterialApiService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = '/api/formulas';

  getPrebuiltMaterials(): Observable<Formula[]> {
    return this.http.get<Formula[]>(`${this.baseUrl}?formula_type=material`);
  }

  getPrebuiltMaterial(id: string): Observable<Formula> {
    return this.http.get<Formula>(`${this.baseUrl}/${id}`);
  }

  createPrebuiltMaterial(data: Partial<Formula>): Observable<Formula> {
    return this.http.post<Formula>(this.baseUrl, { ...data, component_mode: 'single', formula_type: 'material' });
  }

  updatePrebuiltMaterial(id: string, data: Partial<Formula>): Observable<Formula> {
    return this.http.put<Formula>(`${this.baseUrl}/${id}`, data);
  }

  deletePrebuiltMaterial(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
