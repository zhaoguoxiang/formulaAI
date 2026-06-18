import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { TestOutline, VersionSummary } from '../types/test-outline.types';

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

  /** Fetch a specific test outline by ID. */
  getOutline(id: string): Observable<TestOutline> {
    return this.http.get<TestOutline>(`${this.baseUrl}/${id}`);
  }

  /** List all versions of an outline by its ID. */
  getVersions(id: string): Observable<VersionSummary[]> {
    return this.http.get<VersionSummary[]>(`${this.baseUrl}/${id}/versions`);
  }

  /** Create a new test outline (version 1). */
  createOutline(data: { name: string; items: Array<Record<string, unknown>> }): Observable<TestOutline> {
    return this.http.post<TestOutline>(this.baseUrl, data);
  }

  /** Save a new version of an existing outline (auto-increments version). */
  saveVersion(id: string, data: { name: string; items: Array<Record<string, unknown>> }): Observable<TestOutline> {
    return this.http.put<TestOutline>(`${this.baseUrl}/${id}`, data);
  }
}
