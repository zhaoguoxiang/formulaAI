import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'editor',
    loadComponent: () =>
      import('./components/formula-editor/formula-editor.component').then(
        (m) => m.FormulaEditorComponent,
      ),
  },
  {
    path: 'editor/:id',
    loadComponent: () =>
      import('./components/formula-editor/formula-editor.component').then(
        (m) => m.FormulaEditorComponent,
      ),
  },
];
