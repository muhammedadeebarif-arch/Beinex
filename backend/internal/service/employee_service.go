package service

import (
	"context"
	"errors"
	"math"

	"ems-backend/internal/model"
	"ems-backend/internal/repository"
	"ems-backend/internal/validation"
)

var (
	ErrNotFound       = repository.ErrNotFound
	ErrDuplicateEmail = repository.ErrDuplicateEmail
	ErrInvalidID      = errors.New("employee id must be a positive integer")
)

// ValidationError represents validation failures for request fields.
type ValidationError struct {
	Fields map[string]string
}

func (v *ValidationError) Error() string {
	return "validation failed"
}

// EmployeeService outlines business operations on employees.
type EmployeeService interface {
	Create(ctx context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error)
	GetByID(ctx context.Context, id int64) (*model.Employee, error)
	Update(ctx context.Context, id int64, req *model.UpdateEmployeeRequest) (*model.Employee, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter model.EmployeeFilter) (*model.EmployeeListResponse, error)
}

type employeeService struct {
	repo repository.EmployeeRepository
}

// NewEmployeeService creates a new EmployeeService.
func NewEmployeeService(repo repository.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) Create(ctx context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error) {
	if fieldErrors := validation.ValidateCreateEmployee(req); len(fieldErrors) > 0 {
		return nil, &ValidationError{Fields: fieldErrors}
	}

	emp := &model.Employee{
		Name:        req.Name,
		Email:       validation.NormalizeEmail(req.Email),
		Department:  req.Department,
		Designation: req.Designation,
		JoiningDate: req.JoiningDate,
		Status:      req.Status,
	}

	return s.repo.Create(ctx, emp)
}

func (s *employeeService) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *employeeService) Update(ctx context.Context, id int64, req *model.UpdateEmployeeRequest) (*model.Employee, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	if fieldErrors := validation.ValidateUpdateEmployee(req); len(fieldErrors) > 0 {
		return nil, &ValidationError{Fields: fieldErrors}
	}

	emp := &model.Employee{
		ID:          id,
		Name:        req.Name,
		Email:       validation.NormalizeEmail(req.Email),
		Department:  req.Department,
		Designation: req.Designation,
		JoiningDate: req.JoiningDate,
		Status:      req.Status,
	}

	return s.repo.Update(ctx, emp)
}

func (s *employeeService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}

func (s *employeeService) List(ctx context.Context, filter model.EmployeeFilter) (*model.EmployeeListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	employees, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	return &model.EmployeeListResponse{
		Data: employees,
		Pagination: model.PaginationMeta{
			CurrentPage:  filter.Page,
			Limit:        filter.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}
