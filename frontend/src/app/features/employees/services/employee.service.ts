import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../../environments/environment';
import {
  Employee,
  CreateEmployeePayload,
  UpdateEmployeePayload,
  EmployeeFilter,
  EmployeeListResponse,
} from '../models/employee.model';

@Injectable({
  providedIn: 'root',
})
export class EmployeeService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiUrl}/employees`;

  getEmployees(filter: EmployeeFilter): Observable<EmployeeListResponse> {
    let params = new HttpParams()
      .set('page', filter.page.toString())
      .set('limit', filter.limit.toString());

    if (filter.search && filter.search.trim()) {
      params = params.set('search', filter.search.trim());
    }
    if (filter.department && filter.department.trim()) {
      params = params.set('department', filter.department.trim());
    }
    if (filter.status && filter.status.trim()) {
      params = params.set('status', filter.status.trim());
    }

    return this.http.get<EmployeeListResponse>(this.baseUrl, { params });
  }

  getEmployee(id: number): Observable<Employee> {
    return this.http.get<Employee>(`${this.baseUrl}/${id}`);
  }

  createEmployee(payload: CreateEmployeePayload): Observable<Employee> {
    return this.http.post<Employee>(this.baseUrl, payload);
  }

  updateEmployee(id: number, payload: UpdateEmployeePayload): Observable<Employee> {
    return this.http.put<Employee>(`${this.baseUrl}/${id}`, payload);
  }

  deleteEmployee(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
