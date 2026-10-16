export type EmployeeStatus = 'Active' | 'Inactive';

export interface Employee {
  id: number;
  name: string;
  email: string;
  department: string;
  designation: string;
  joining_date: string; // YYYY-MM-DD
  status: EmployeeStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateEmployeePayload {
  name: string;
  email: string;
  department: string;
  designation: string;
  joining_date: string;
  status: EmployeeStatus;
}

export interface UpdateEmployeePayload {
  name: string;
  email: string;
  department: string;
  designation: string;
  joining_date: string;
  status: EmployeeStatus;
}

export interface EmployeeFilter {
  search?: string;
  department?: string;
  status?: string;
  page: number;
  limit: number;
}

export interface PaginationMeta {
  current_page: number;
  limit: number;
  total_records: number;
  total_pages: number;
}

export interface EmployeeListResponse {
  data: Employee[];
  pagination: PaginationMeta;
}

export interface ApiErrorDetail {
  code: string;
  message: string;
  fields?: Record<string, string>;
}

export interface ApiErrorResponse {
  error: ApiErrorDetail;
}
