import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./features/intro/pages/intro-page.component').then((m) => m.IntroPageComponent),
  },
  {
    path: 'employees',
    loadComponent: () =>
      import('./features/employees/pages/employee-list/employee-list.component').then(
        (m) => m.EmployeeListComponent,
      ),
  },
  {
    path: 'login',
    redirectTo: '',
  },
  {
    path: '**',
    redirectTo: '',
  },
];
