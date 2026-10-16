import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Subject, debounceTime, distinctUntilChanged } from 'rxjs';
import { Employee, EmployeeFilter, PaginationMeta } from '../../models/employee.model';
import { EmployeeService } from '../../services/employee.service';
import { EmployeeFormComponent } from '../../components/employee-form/employee-form.component';
import { DeleteModalComponent } from '../../components/delete-modal/delete-modal.component';

@Component({
  selector: 'app-employee-list',
  standalone: true,
  imports: [CommonModule, FormsModule, EmployeeFormComponent, DeleteModalComponent],
  templateUrl: './employee-list.component.html',
  styleUrls: ['./employee-list.component.css'],
})
export class EmployeeListComponent implements OnInit {
  private readonly employeeService = inject(EmployeeService);

  // State signals
  employees = signal<Employee[]>([]);
  pagination = signal<PaginationMeta>({
    current_page: 1,
    limit: 10,
    total_records: 0,
    total_pages: 0,
  });
  isLoading = signal<boolean>(false);
  errorMessage = signal<string | null>(null);
  toastMessage = signal<string | null>(null);

  // Search & Filters
  searchTerm = '';
  selectedDepartment = '';
  selectedStatus = '';
  currentPage = 1;
  pageSize = 10;

  // Modals state
  isFormOpen = false;
  employeeToEdit: Employee | null = null;
  employeeToDelete: Employee | null = null;

  readonly departments = [
    'Engineering',
    'Finance',
    'Human Resources',
    'Marketing',
    'Operations',
    'Sales',
    'Consulting',
  ];

  private readonly searchSubject = new Subject<string>();

  ngOnInit(): void {
    // Setup debounced search input
    this.searchSubject.pipe(debounceTime(350), distinctUntilChanged()).subscribe((term) => {
      this.searchTerm = term;
      this.currentPage = 1;
      this.loadEmployees();
    });

    this.loadEmployees();
  }

  onSearchChange(term: string): void {
    this.searchSubject.next(term);
  }

  onFilterChange(): void {
    this.currentPage = 1;
    this.loadEmployees();
  }

  clearFilters(): void {
    this.searchTerm = '';
    this.selectedDepartment = '';
    this.selectedStatus = '';
    this.currentPage = 1;
    this.loadEmployees();
  }

  hasActiveFilters(): boolean {
    return !!(this.searchTerm.trim() || this.selectedDepartment || this.selectedStatus);
  }

  loadEmployees(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    const filter: EmployeeFilter = {
      search: this.searchTerm,
      department: this.selectedDepartment,
      status: this.selectedStatus,
      page: this.currentPage,
      limit: this.pageSize,
    };

    this.employeeService.getEmployees(filter).subscribe({
      next: (res) => {
        this.isLoading.set(false);
        this.employees.set(res.data);
        this.pagination.set(res.pagination);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(
          err.error?.error?.message ||
            'Failed to load employees from the server. Please ensure the backend is running.',
        );
      },
    });
  }

  // Pagination actions
  goToPage(page: number): void {
    if (page >= 1 && page <= this.pagination().total_pages && page !== this.currentPage) {
      this.currentPage = page;
      this.loadEmployees();
    }
  }

  onPageSizeChange(newSize: number): void {
    this.pageSize = newSize;
    this.currentPage = 1;
    this.loadEmployees();
  }

  // CRUD actions
  openAddModal(): void {
    this.employeeToEdit = null;
    this.isFormOpen = true;
  }

  openEditModal(emp: Employee): void {
    this.employeeToEdit = emp;
    this.isFormOpen = true;
  }

  closeFormModal(): void {
    this.isFormOpen = false;
    this.employeeToEdit = null;
  }

  onFormSaved(savedEmp: Employee): void {
    const wasEditing = !!this.employeeToEdit;
    this.closeFormModal();
    this.showToast(
      wasEditing
        ? `Employee "${savedEmp.name}" updated successfully.`
        : `Employee "${savedEmp.name}" onboarded successfully.`,
    );
    this.loadEmployees();
  }

  openDeleteModal(emp: Employee): void {
    this.employeeToDelete = emp;
  }

  closeDeleteModal(): void {
    this.employeeToDelete = null;
  }

  onDeleteConfirmed(id: number): void {
    const empName = this.employeeToDelete?.name || 'Employee';
    this.closeDeleteModal();
    this.showToast(`${empName} has been permanently deleted.`);

    // If current page is now empty and not page 1, step back one page
    if (this.employees().length === 1 && this.currentPage > 1) {
      this.currentPage--;
    }
    this.loadEmployees();
  }

  private showToast(msg: string): void {
    this.toastMessage.set(msg);
    setTimeout(() => {
      this.toastMessage.set(null);
    }, 4000);
  }
}
