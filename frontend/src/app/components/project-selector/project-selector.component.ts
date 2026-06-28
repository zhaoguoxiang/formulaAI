import { Component, inject, ChangeDetectionStrategy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';

import { ProjectStateService } from '../../services/project-state.service';
import { ProjectDialogComponent } from '../project-dialog/project-dialog.component';

@Component({
  selector: 'app-project-selector',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    FormsModule,
    MatSelectModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    MatDialogModule,
    MatDividerModule,
  ],
  templateUrl: './project-selector.component.html',
  styleUrl: './project-selector.component.scss',
})
export class ProjectSelectorComponent implements OnInit {
  readonly projectState = inject(ProjectStateService);
  private readonly dialog = inject(MatDialog);

  ngOnInit(): void {
    this.projectState.loadProjects();
  }

  onProjectChange(id: string): void {
    this.projectState.selectProject(id);
  }

  openManageDialog(): void {
    this.dialog.open(ProjectDialogComponent, {
      width: '520px',
      maxHeight: '80vh',
    });
  }
}
