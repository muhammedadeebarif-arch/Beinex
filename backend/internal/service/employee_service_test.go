package service

import (
	"context"
	"testing"
	"time"

	"ems-backend/internal/model"
	"ems-backend/internal/repository"
)

type mockRepository struct {
	employees []model.Employee
	createErr error
}

func (m *mockRepository) Create(ctx context.Context, emp *model.Employee) (*model.Employee, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	newEmp := *emp
	newEmp.ID = int64(len(m.employees) + 1)
	newEmp.CreatedAt = time.Now()
	newEmp.UpdatedAt = time.Now()
	m.employees = append(m.employees, newEmp)
	return &newEmp, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	for _, e := range m.employees {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepository) Update(ctx context.Context, emp *model.Employee) (*model.Employee, error) {
	for i, e := range m.employees {
		if e.ID == emp.ID {
			updated := *emp
			updated.UpdatedAt = time.Now()
			m.employees[i] = updated
			return &updated, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	for i, e := range m.employees {
		if e.ID == id {
			m.employees = append(m.employees[:i], m.employees[i+1:]...)
			return nil
		}
	}
	return repository.ErrNotFound
}

func (m *mockRepository) List(ctx context.Context, filter model.EmployeeFilter) ([]model.Employee, int64, error) {
	return m.employees, int64(len(m.employees)), nil
}

func TestEmployeeService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Rejects invalid request without hitting repository", func(t *testing.T) {
		repo := &mockRepository{}
		svc := NewEmployeeService(repo)

		_, err := svc.Create(ctx, &model.CreateEmployeeRequest{
			Name: "", // missing
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var valErr *ValidationError
		if !errorAs(err, &valErr) {
			t.Fatalf("expected ValidationError, got %T", err)
		}
	})

	t.Run("Successfully creates employee and normalizes email", func(t *testing.T) {
		repo := &mockRepository{}
		svc := NewEmployeeService(repo)

		created, err := svc.Create(ctx, &model.CreateEmployeeRequest{
			Name:        "John Doe",
			Email:       "  JOHN.DOE@example.com  ",
			Department:  "Operations",
			Designation: "Manager",
			JoiningDate: "2024-02-01",
			Status:      "Active",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if created.Email != "john.doe@example.com" {
			t.Errorf("expected normalized email, got %s", created.Email)
		}
	})

	t.Run("Propagates duplicate email error", func(t *testing.T) {
		repo := &mockRepository{createErr: repository.ErrDuplicateEmail}
		svc := NewEmployeeService(repo)

		_, err := svc.Create(ctx, &model.CreateEmployeeRequest{
			Name:        "Duplicate User",
			Email:       "dup@example.com",
			Department:  "HR",
			Designation: "Recruiter",
			JoiningDate: "2024-01-01",
			Status:      "Active",
		})
		if err != ErrDuplicateEmail {
			t.Errorf("expected ErrDuplicateEmail, got %v", err)
		}
	})
}

func TestEmployeeService_PaginationCalculation(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepository{
		employees: make([]model.Employee, 25), // 25 employees
	}
	svc := NewEmployeeService(repo)

	resp, err := svc.List(ctx, model.EmployeeFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Pagination.TotalRecords != 25 {
		t.Errorf("expected 25 total records, got %d", resp.Pagination.TotalRecords)
	}
	if resp.Pagination.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", resp.Pagination.TotalPages)
	}
}

func errorAs(err error, target any) bool {
	// simple helper
	if valErr, ok := err.(*ValidationError); ok {
		if ptr, ok := target.(**ValidationError); ok {
			*ptr = valErr
			return true
		}
	}
	return false
}
