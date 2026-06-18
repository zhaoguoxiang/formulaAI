import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

@Component({
  selector: 'app-version-note-dialog',
  standalone: true,
  imports: [CommonModule, FormsModule, MatDialogModule, MatButtonModule, MatFormFieldModule, MatInputModule],
  templateUrl: './version-note-dialog.component.html',
  styleUrl: './version-note-dialog.component.scss',
})
export class VersionNoteDialogComponent {
  private readonly dialogRef = inject(MatDialogRef<VersionNoteDialogComponent>);

  note = '';

  confirm(): void {
    this.dialogRef.close(this.note.trim());
  }

  cancel(): void {
    this.dialogRef.close(null);
  }
}
