package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ems-backend/internal/config"
	"ems-backend/internal/database"
	"ems-backend/internal/handler"
	"ems-backend/internal/middleware"
	"ems-backend/internal/model"
	"ems-backend/internal/repository"
	"ems-backend/internal/service"
)

func setupAuthTestApp(t *testing.T) (*database.DB, http.Handler, service.AuthService) {
	_ = os.Setenv("DB_HOST", "127.0.0.1")
	_ = os.Setenv("DB_PORT", "5432")
	_ = os.Setenv("DB_USER", "postgres")
	_ = os.Setenv("DB_PASSWORD", "")
	_ = os.Setenv("DB_NAME", "employee_test_db")
	_ = os.Setenv("DB_SSLMODE", "disable")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	migrationPath := filepath.Join("..", "migrations")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.ApplyMigrations(ctx, migrationPath); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Clean tables before test run
	_, _ = db.Pool.ExecContext(ctx, "TRUNCATE TABLE admin_users, employees RESTART IDENTITY;")

	adminRepo := repository.NewPostgresAdminRepository(db.Pool)
	authSvc := service.NewAuthService(adminRepo, "test-secret-key-12345", 1)
	authHandler := handler.NewAuthHandler(authSvc)

	empRepo := repository.NewPostgresEmployeeRepository(db.Pool)
	empSvc := service.NewEmployeeService(empRepo)
	empHandler := handler.NewEmployeeHandler(empSvc)

	// Provision test admin
	if err := authSvc.ProvisionInitialAdmin(ctx, "Test Admin", "admin@beinex.com", "Password@123"); err != nil {
		t.Fatalf("failed to provision admin: %v", err)
	}

	authGuard := middleware.RequireAuth(authSvc)
	limiter := middleware.NewLoginRateLimiter(5, 1*time.Minute)

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/auth/login", limiter.Limit(http.HandlerFunc(authHandler.Login)))
	mux.Handle("GET /api/v1/auth/me", authGuard(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /api/v1/employees", authGuard(http.HandlerFunc(empHandler.List)))

	return db, mux, authSvc
}

func TestAuth_Flow(t *testing.T) {
	db, router, _ := setupAuthTestApp(t)
	defer db.Close()

	var validToken string

	t.Run("Valid administrator login succeeds", func(t *testing.T) {
		payload := map[string]string{
			"email":    "admin@beinex.com",
			"password": "Password@123",
		}
		data, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var res model.LoginResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode login response: %v", err)
		}

		if res.Token == "" {
			t.Fatal("expected non-empty JWT token")
		}
		if res.Admin.Email != "admin@beinex.com" {
			t.Fatalf("expected email admin@beinex.com, got %s", res.Admin.Email)
		}

		validToken = res.Token
	})

	t.Run("Invalid credentials rejected with 401", func(t *testing.T) {
		payload := map[string]string{
			"email":    "admin@beinex.com",
			"password": "WrongPassword!",
		}
		data, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Nonexistent admin rejected with generic 401", func(t *testing.T) {
		payload := map[string]string{
			"email":    "nobody@beinex.com",
			"password": "WrongPassword!",
		}
		data, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("Unauthenticated access to protected /api/v1/employees returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Authenticated access with Bearer token succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Tampered token is rejected with 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
		req.Header.Set("Authorization", "Bearer "+validToken+"corrupted")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for corrupted token, got %d", rec.Code)
		}
	})

	t.Run("Rate limiter throttles excessive failed logins", func(t *testing.T) {
		payload := map[string]string{
			"email":    "hacker@beinex.com",
			"password": "badpassword",
		}
		data, _ := json.Marshal(payload)

		hit429 := false
		for i := 0; i < 7; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(data))
			req.RemoteAddr = "192.0.2.1:1234"
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusTooManyRequests {
				hit429 = true
				break
			}
		}

		if !hit429 {
			t.Fatal("expected rate limiter to trigger HTTP 429 TooManyRequests")
		}
	})
}
