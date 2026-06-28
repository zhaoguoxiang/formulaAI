import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Project, CreateProjectRequest, UpdateProjectRequest } from '../types/project.types';

/**
 * Service for project (workspace) CRUD operations.
 * NOTE: These endpoints do NOT carry the X-Project-Id header
 * since they operate at the global level.
 */
@Injectable({ providedIn: 'root' })
export class ProjectApiService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = '/api/projects';

  list(): Observable<Project[]> {
    return this.http.get<Project[]>(this.baseUrl);
  }

  get(id: string): Observable<Project> {
    return this.http.get<Project>(`${this.baseUrl}/${id}`);
  }

  create(data: CreateProjectRequest): Observable<Project> {
    return this.http.post<Project>(this.baseUrl, data);
  }

  update(id: string, data: UpdateProjectRequest): Observable<Project> {
    return this.http.put<Project>(`${this.baseUrl}/${id}`, data);
  }

  delete(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.baseUrl}/${id}`);
  }
}
