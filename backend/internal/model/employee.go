package model

import "time"

// Employee represents an employee record in the system.
type Employee struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Department  string    `json:"department"`
	Designation string    `json:"designation"`
	JoiningDate string    `json:"joining_date"` // YYYY-MM-DD
	Status      string    `json:"status"`       // Active | Inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateEmployeeRequest represents the payload for creating a new employee.
type CreateEmployeeRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	JoiningDate string `json:"joining_date"` // YYYY-MM-DD
	Status      string `json:"status"`       // Active | Inactive
}

// UpdateEmployeeRequest represents the payload for updating an existing employee.
type UpdateEmployeeRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	JoiningDate string `json:"joining_date"` // YYYY-MM-DD
	Status      string `json:"status"`       // Active | Inactive
}

// EmployeeFilter represents query parameters for filtering and paginating employees.
type EmployeeFilter struct {
	Search     string
	Department string
	Status     string
	Page       int
	Limit      int
}

// PaginationMeta provides pagination information.
type PaginationMeta struct {
	CurrentPage  int   `json:"current_page"`
	Limit        int   `json:"limit"`
	TotalRecords int64 `json:"total_records"`
	TotalPages   int   `json:"total_pages"`
}

// EmployeeListResponse represents the paginated list of employees.
type EmployeeListResponse struct {
	Data       []Employee     `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}
