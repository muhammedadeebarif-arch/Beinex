import {
  ChangeDetectorRef,
  Component,
  EventEmitter,
  Input,
  OnInit,
  Output,
  inject,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import {
  Employee,
  CreateEmployeePayload,
  UpdateEmployeePayload,
} from '../../models/employee.model';
import { EmployeeService } from '../../services/employee.service';

@Component({
  selector: 'app-employee-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './employee-form.component.html',
  styleUrls: ['./employee-form.component.css'],
})
export class EmployeeFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly employeeService = inject(EmployeeService);
  private readonly cdr = inject(ChangeDetectorRef);

  @Input() employeeToEdit: Employee | null = null;
  @Output() formClosed = new EventEmitter<void>();
  @Output() formSaved = new EventEmitter<Employee>();

  form!: FormGroup;
  isSubmitting = false;
  backendError: string | null = null;
  fieldErrors: Record<string, string> = {};

  readonly departments = [
    'Engineering',
    'Finance',
    'Human Resources',
    'Marketing',
    'Operations',
    'Sales',
    'Consulting',
  ];

  ngOnInit(): void {
    this.initForm();
  }

  private initForm(): void {
    this.form = this.fb.group({
      name: [this.employeeToEdit?.name || '', [Validators.required, Validators.maxLength(100)]],
      email: [
        this.employeeToEdit?.email || '',
        [
          Validators.required,
          Validators.email,
          Validators.pattern('^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$'),
          Validators.maxLength(150),
        ],
      ],
      department: [
        this.employeeToEdit?.department || '',
        [Validators.required, Validators.maxLength(100)],
      ],
      designation: [
        this.employeeToEdit?.designation || '',
        [Validators.required, Validators.maxLength(100)],
      ],
      joining_date: [this.employeeToEdit?.joining_date || '', [Validators.required]],
      status: [this.employeeToEdit?.status || 'Active', [Validators.required]],
    });
  }

  get isEditMode(): boolean {
    return !!this.employeeToEdit;
  }

  onSubmit(): void {
    if (this.form.invalid || this.isSubmitting) {
      this.form.markAllAsTouched();
      return;
    }

    this.isSubmitting = true;
    this.backendError = null;
    this.fieldErrors = {};

    const formVal = this.form.value;
    const payload: CreateEmployeePayload = {
      name: formVal.name.trim(),
      email: formVal.email.trim().toLowerCase(),
      department: formVal.department,
      designation: formVal.designation.trim(),
      joining_date: formVal.joining_date,
      status: formVal.status,
    };

    if (this.isEditMode && this.employeeToEdit) {
      this.employeeService.updateEmployee(this.employeeToEdit.id, payload).subscribe({
        next: (updated) => {
          this.isSubmitting = false;
          this.formSaved.emit(updated);
        },
        error: (err) => this.handleError(err),
      });
    } else {
      this.employeeService.createEmployee(payload).subscribe({
        next: (created) => {
          this.isSubmitting = false;
          this.formSaved.emit(created);
        },
        error: (err) => this.handleError(err),
      });
    }
  }

  private handleError(err: any): void {
    this.isSubmitting = false;
    if (err.status === 409) {
      this.backendError = 'An employee with this email address already exists.';
      this.fieldErrors['email'] = 'This email is already in use by another employee.';
    } else if (err.status === 400 && err.error?.error?.fields) {
      this.backendError =
        err.error.error.message || 'Validation error. Please check the fields below.';
      this.fieldErrors = err.error.error.fields;
    } else {
      this.backendError = err.error?.error?.message || 'A server error occurred. Please try again.';
    }
    this.cdr.markForCheck();
  }

  onCancel(): void {
    this.formClosed.emit();
  }
}
