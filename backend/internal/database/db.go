package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ems-backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps standard sql.DB pool.
type DB struct {
	Pool *sql.DB
}

// Connect initializes the PostgreSQL connection pool.
func Connect(cfg *config.Config) (*DB, error) {
	pool, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	pool.SetMaxOpenConns(cfg.DBMaxOpenConns)
	pool.SetMaxIdleConns(cfg.DBMaxIdleConns)
	pool.SetConnMaxLifetime(cfg.DBConnMaxLife)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Pool.Close()
}

// ApplyMigrations executes SQL migration scripts from a file or directory.
func (db *DB) ApplyMigrations(ctx context.Context, migrationPath string) error {
	info, err := os.Stat(migrationPath)
	if err != nil {
		return fmt.Errorf("could not stat migration path %s: %w", migrationPath, err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(migrationPath)
		if err != nil {
			return fmt.Errorf("could not read migration directory: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
				files = append(files, filepath.Join(migrationPath, e.Name()))
			}
		}
		sort.Strings(files)
	} else {
		files = []string{migrationPath}
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("could not read migration file %s: %w", file, err)
		}

		tx, err := db.Pool.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("could not begin migration transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("could not execute migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("could not commit migration %s: %w", file, err)
		}
	}

	return nil
}
