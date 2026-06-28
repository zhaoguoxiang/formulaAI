import { Injectable, inject, signal, effect, computed } from '@angular/core';
import { Project } from '../types/project.types';
import { ProjectApiService } from './project-api.service';

const STORAGE_KEY = 'formula_current_project_id';

/**
 * Manages the currently selected project (workspace).
 * Persists selection to localStorage and exposes a reactive signal.
 */
@Injectable({ providedIn: 'root' })
export class ProjectStateService {
  private readonly projectApi = inject(ProjectApiService);

  private readonly _currentProjectId = signal<string | null>(
    localStorage.getItem(STORAGE_KEY)
  );

  private readonly _projects = signal<Project[]>([]);
  private readonly _loaded = signal(false);

  readonly currentProjectId = this._currentProjectId.asReadonly();
  readonly projects = this._projects.asReadonly();
  readonly loaded = this._loaded.asReadonly();

  readonly currentProject = computed(() => {
    const id = this._currentProjectId();
    if (!id) return null;
    return this._projects().find(p => p.id === id) ?? null;
  });

  readonly hasProject = computed(() => this._currentProjectId() !== null);

  constructor() {
    // Persist project selection
    effect(() => {
      const id = this._currentProjectId();
      if (id) {
        localStorage.setItem(STORAGE_KEY, id);
      } else {
        localStorage.removeItem(STORAGE_KEY);
      }
    });
  }

  loadProjects(): void {
    this.projectApi.list().subscribe({
      next: (projects) => {
        this._projects.set(projects);
        this._loaded.set(true);

        // If current project no longer exists (deleted), clear selection
        const currentId = this._currentProjectId();
        if (currentId && !projects.find(p => p.id === currentId)) {
          this._currentProjectId.set(null);
        }
      },
    });
  }

  selectProject(id: string): void {
    this._currentProjectId.set(id);
  }

  clearSelection(): void {
    this._currentProjectId.set(null);
  }
}
