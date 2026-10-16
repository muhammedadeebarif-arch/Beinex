import { ChangeDetectorRef, Component, EventEmitter, Input, Output, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Employee } from '../../models/employee.model';
import { EmployeeService } from '../../services/employee.service';

@Component({
  selector: 'app-delete-modal',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './delete-modal.component.html',
  styleUrls: ['./delete-modal.component.css'],
})
export class DeleteModalComponent {
  private readonly employeeService = inject(EmployeeService);
  private readonly cdr = inject(ChangeDetectorRef);

  @Input({ required: true }) employee!: Employee;
  @Output() cancelled = new EventEmitter<void>();
  @Output() confirmed = new EventEmitter<number>();

  isDeleting = false;
  errorMessage: string | null = null;

  onConfirm(): void {
    if (this.isDeleting) return;

    this.isDeleting = true;
    this.errorMessage = null;

    this.employeeService.deleteEmployee(this.employee.id).subscribe({
      next: () => {
        this.isDeleting = false;
        this.confirmed.emit(this.employee.id);
      },
      error: (err) => {
        this.isDeleting = false;
        this.errorMessage =
          err.error?.error?.message || 'Failed to delete employee. Please try again.';
        this.cdr.markForCheck();
      },
    });
  }

  onCancel(): void {
    if (!this.isDeleting) {
      this.cancelled.emit();
    }
  }
}
