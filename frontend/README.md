# Angular Frontend — Employee Management System

## Features
- **Standalone Component Architecture**: Uses modern Angular 19+ standalone components and signals.
- **Reactive Forms with Field-Level Validation**: Real-time validation for names, valid email patterns, date restrictions, and departments.
- **Enterprise Design System**: Modern responsive CSS without external component libraries, supporting desktop, tablet, and mobile.
- **Debounced Server-Side Search & Filtering**: Debounced search by name/email, department dropdown filter, and status filter.
- **Server-Driven Pagination**: Page size selection, previous/next controls, and total records calculation.
- **Interactive Modals**: Accessible modal dialogs for Employee Onboarding/Editing and permanent Deletion Confirmation.
- **Toast Notifications**: Automatic non-intrusive feedback upon successful operations.

## Running Locally
```powershell
npm install
npm start
```
Available at `http://localhost:4200/`.

## Production Build
```powershell
npm run build
```
Generates production bundle in `dist/frontend`.
