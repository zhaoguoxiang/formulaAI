import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { TestOutline } from '../types/test-outline.types';

@Injectable({
  providedIn: 'root',
})
export class TestOutlineApiService {
  private readonly baseUrl = '/api/test-outlines';

  constructor(private readonly http: HttpClient) {}

  /** Fetch all active test outlines with nested items and indicators. */
  getOutlines(): Observable<TestOutline[]> {
    return this.http.get<TestOutline[]>(this.baseUrl);
  }

  /** Fetch a single test outline by ID. */
  getOutline(id: string): Observable<TestOutline> {
    return this.http.get<TestOutline>(`${this.baseUrl}/${id}`);
  }

  /** Create a new test outline (version 1). */
  createOutline(data: { name: string; items: Array<Record<string, unknown>> }): Observable<TestOutline> {
    return this.http.post<TestOutline>(this.baseUrl, data);
  }

  /** Save a new version of an existing outline (auto-increments version). */
  saveVersion(id: string, data: { name: string; items: Array<Record<string, unknown>> }): Observable<TestOutline> {
    return this.http.put<TestOutline>(`${this.baseUrl}/${id}`, data);
  }

  /** Archive a test outline (soft delete). */
  archiveOutline(id: string): Observable<{ status: string }> {
    return this.http.put<{ status: string }>(`${this.baseUrl}/${id}/archive`, {});
  }

  /** List all versions of a test outline. */
  listVersions(id: string): Observable<TestOutline[]> {
    return this.http.get<TestOutline[]>(`${this.baseUrl}/${id}/versions`);
  }

  /** Activate a specific version. */
  activateVersion(id: string): Observable<{ status: string }> {
    return this.http.put<{ status: string }>(`${this.baseUrl}/${id}/activate`, {});
  }
}
