package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"ems-backend/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound       = errors.New("employee not found")
	ErrDuplicateEmail = errors.New("an employee with this email already exists")
)

// EmployeeRepository defines the interface for employee persistence.
type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) (*model.Employee, error)
	GetByID(ctx context.Context, id int64) (*model.Employee, error)
	Update(ctx context.Context, emp *model.Employee) (*model.Employee, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter model.EmployeeFilter) ([]model.Employee, int64, error)
}

type postgresEmployeeRepository struct {
	db *sql.DB
}

// NewPostgresEmployeeRepository returns a new instance of EmployeeRepository.
func NewPostgresEmployeeRepository(db *sql.DB) EmployeeRepository {
	return &postgresEmployeeRepository{db: db}
}

func (r *postgresEmployeeRepository) Create(ctx context.Context, emp *model.Employee) (*model.Employee, error) {
	query := `
		INSERT INTO employees (name, email, department, designation, joining_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, name, email, department, designation, TO_CHAR(joining_date, 'YYYY-MM-DD'), status, created_at, updated_at;
	`

	var created model.Employee
	err := r.db.QueryRowContext(ctx, query,
		emp.Name,
		emp.Email,
		emp.Department,
		emp.Designation,
		emp.JoiningDate,
		emp.Status,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Email,
		&created.Department,
		&created.Designation,
		&created.JoiningDate,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	return &created, nil
}

func (r *postgresEmployeeRepository) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	query := `
		SELECT id, name, email, department, designation, TO_CHAR(joining_date, 'YYYY-MM-DD'), status, created_at, updated_at
		FROM employees
		WHERE id = $1;
	`

	var emp model.Employee
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&emp.ID,
		&emp.Name,
		&emp.Email,
		&emp.Department,
		&emp.Designation,
		&emp.JoiningDate,
		&emp.Status,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get employee by id: %w", err)
	}

	return &emp, nil
}

func (r *postgresEmployeeRepository) Update(ctx context.Context, emp *model.Employee) (*model.Employee, error) {
	query := `
		UPDATE employees
		SET name = $1, email = $2, department = $3, designation = $4, joining_date = $5, status = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING id, name, email, department, designation, TO_CHAR(joining_date, 'YYYY-MM-DD'), status, created_at, updated_at;
	`

	var updated model.Employee
	err := r.db.QueryRowContext(ctx, query,
		emp.Name,
		emp.Email,
		emp.Department,
		emp.Designation,
		emp.JoiningDate,
		emp.Status,
		emp.ID,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Email,
		&updated.Department,
		&updated.Designation,
		&updated.JoiningDate,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	return &updated, nil
}

func (r *postgresEmployeeRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM employees WHERE id = $1;`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *postgresEmployeeRepository) List(ctx context.Context, filter model.EmployeeFilter) ([]model.Employee, int64, error) {
	var whereClauses []string
	var args []any
	argIndex := 1
	var exactID int64

	if search := strings.TrimSpace(filter.Search); search != "" {
		cleaned := strings.TrimSpace(search)
		lower := strings.ToLower(cleaned)
		if strings.HasPrefix(lower, "id:") || strings.HasPrefix(lower, "id=") || strings.HasPrefix(lower, "id ") {
			cleaned = strings.TrimSpace(cleaned[3:])
		}
		cleaned = strings.TrimPrefix(cleaned, "#")
		cleaned = strings.TrimSpace(cleaned)

		if idVal, err := strconv.ParseInt(cleaned, 10, 64); err == nil && idVal > 0 {
			exactID = idVal
			whereClauses = append(whereClauses, fmt.Sprintf("(id = $%d OR CAST(id AS TEXT) ILIKE $%d OR name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex+1, argIndex+1, argIndex+1))
			args = append(args, idVal, "%"+cleaned+"%")
			argIndex += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("(CAST(id AS TEXT) ILIKE $%d OR name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex, argIndex))
			args = append(args, "%"+search+"%")
			argIndex++
		}
	}

	if dept := strings.TrimSpace(filter.Department); dept != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("department = $%d", argIndex))
		args = append(args, dept)
		argIndex++
	}

	if status := strings.TrimSpace(filter.Status); status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Get total matching count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees %s;", whereSQL)
	var totalRecords int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	// 2. Fetch paginated records
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	orderClause := "ORDER BY id DESC"
	if exactID > 0 {
		orderClause = fmt.Sprintf("ORDER BY (CASE WHEN id = %d THEN 0 ELSE 1 END), id DESC", exactID)
	}

	listQuery := fmt.Sprintf(`
		SELECT id, name, email, department, designation, TO_CHAR(joining_date, 'YYYY-MM-DD'), status, created_at, updated_at
		FROM employees
		%s
		%s
		LIMIT $%d OFFSET $%d;
	`, whereSQL, orderClause, argIndex, argIndex+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employees: %w", err)
	}
	defer rows.Close()

	employees := make([]model.Employee, 0)
	for rows.Next() {
		var emp model.Employee
		if err := rows.Scan(
			&emp.ID,
			&emp.Name,
			&emp.Email,
			&emp.Department,
			&emp.Designation,
			&emp.JoiningDate,
			&emp.Status,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee row: %w", err)
		}
		employees = append(employees, emp)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during row iteration: %w", err)
	}

	return employees, totalRecords, nil
}
