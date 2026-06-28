import { Component, inject, ChangeDetectionStrategy, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatListModule } from '@angular/material/list';
import { MatDividerModule } from '@angular/material/divider';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { Project, CreateProjectRequest } from '../../types/project.types';
import { ProjectApiService } from '../../services/project-api.service';
import { ProjectStateService } from '../../services/project-state.service';

@Component({
  selector: 'app-project-dialog',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    FormsModule,
    MatDialogModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatListModule,
    MatDividerModule,
    MatSnackBarModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './project-dialog.component.html',
  styleUrl: './project-dialog.component.scss',
})
export class ProjectDialogComponent implements OnInit {
  private readonly projectApi = inject(ProjectApiService);
  readonly projectState = inject(ProjectStateService);
  private readonly dialogRef = inject(MatDialogRef<ProjectDialogComponent>);
  private readonly snackBar = inject(MatSnackBar);

  readonly loading = signal(false);
  readonly creating = signal(false);

  // Create form
  newName = '';
  newDescription = '';

  ngOnInit(): void {
    this.projectState.loadProjects();
  }

  createProject(): void {
    const name = this.newName.trim();
    if (!name) return;

    this.creating.set(true);
    const data: CreateProjectRequest = {
      name,
      description: this.newDescription.trim(),
    };

    this.projectApi.create(data).subscribe({
      next: (project) => {
        this.projectState.loadProjects();
        this.projectState.selectProject(project.id);
        this.newName = '';
        this.newDescription = '';
        this.creating.set(false);
        this.snackBar.open(`工作空间"${project.name}"已创建`, '', { duration: 2000 });
      },
      error: () => {
        this.creating.set(false);
        this.snackBar.open('创建失败，请重试', '', { duration: 3000 });
      },
    });
  }

  deleteProject(project: Project): void {
    if (!confirm(`确定要删除工作空间"${project.name}"吗？\n\n该操作将删除此工作空间下的所有配方和测试大纲数据，且不可恢复。`)) {
      return;
    }

    this.loading.set(true);
    this.projectApi.delete(project.id).subscribe({
      next: () => {
        // If we deleted the currently selected project, clear selection
        if (this.projectState.currentProjectId() === project.id) {
          this.projectState.clearSelection();
        }
        this.projectState.loadProjects();
        this.loading.set(false);
        this.snackBar.open(`工作空间"${project.name}"已删除`, '', { duration: 2000 });
      },
      error: () => {
        this.loading.set(false);
        this.snackBar.open('删除失败，请重试', '', { duration: 3000 });
      },
    });
  }

  close(): void {
    this.dialogRef.close();
  }
}
