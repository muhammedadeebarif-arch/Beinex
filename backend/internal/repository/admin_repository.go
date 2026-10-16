package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ems-backend/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAdminNotFound       = errors.New("admin user not found")
	ErrAdminDuplicateEmail = errors.New("admin with this email already exists")
)

// AdminRepository defines persistence operations for admin accounts.
type AdminRepository interface {
	GetByEmail(ctx context.Context, email string) (*model.AdminUser, error)
	Create(ctx context.Context, admin *model.AdminUser) (*model.AdminUser, error)
	Count(ctx context.Context) (int64, error)
}

type postgresAdminRepository struct {
	db *sql.DB
}

// NewPostgresAdminRepository constructs a new AdminRepository instance.
func NewPostgresAdminRepository(db *sql.DB) AdminRepository {
	return &postgresAdminRepository{db: db}
}

func (r *postgresAdminRepository) GetByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM admin_users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1;
	`
	var admin model.AdminUser
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(email)).Scan(
		&admin.ID,
		&admin.Name,
		&admin.Email,
		&admin.PasswordHash,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to query admin user: %w", err)
	}

	return &admin, nil
}

func (r *postgresAdminRepository) Create(ctx context.Context, admin *model.AdminUser) (*model.AdminUser, error) {
	query := `
		INSERT INTO admin_users (name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, name, email, password_hash, created_at, updated_at;
	`
	var created model.AdminUser
	err := r.db.QueryRowContext(ctx, query,
		strings.TrimSpace(admin.Name),
		strings.TrimSpace(admin.Email),
		admin.PasswordHash,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Email,
		&created.PasswordHash,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAdminDuplicateEmail
		}
		return nil, fmt.Errorf("failed to insert admin user: %w", err)
	}

	return &created, nil
}

func (r *postgresAdminRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM admin_users;`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admin users: %w", err)
	}
	return count, nil
}
