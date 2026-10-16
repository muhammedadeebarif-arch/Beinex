package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ems-backend/internal/config"
	"ems-backend/internal/database"
	"ems-backend/internal/handler"
	"ems-backend/internal/model"
	"ems-backend/internal/repository"
	"ems-backend/internal/response"
	"ems-backend/internal/service"
)

func setupTestApp(t *testing.T) (*database.DB, http.Handler) {
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
		t.Fatalf("failed to connect to PostgreSQL (is Postgres running?): %v", err)
	}

	// Apply migration
	migrationPath := filepath.Join("..", "migrations", "001_create_employees_table.sql")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.ApplyMigrations(ctx, migrationPath); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Clean employees table before running test suite
	_, err = db.Pool.ExecContext(ctx, "TRUNCATE TABLE employees RESTART IDENTITY;")
	if err != nil {
		t.Fatalf("failed to truncate employees table: %v", err)
	}

	empRepo := repository.NewPostgresEmployeeRepository(db.Pool)
	empSvc := service.NewEmployeeService(empRepo)
	empHandler := handler.NewEmployeeHandler(empSvc)
	healthHandler := handler.NewHealthHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.HealthCheck)
	mux.HandleFunc("POST /api/v1/employees", empHandler.Create)
	mux.HandleFunc("GET /api/v1/employees", empHandler.List)
	mux.HandleFunc("GET /api/v1/employees/{id}", empHandler.GetByID)
	mux.HandleFunc("PUT /api/v1/employees/{id}", empHandler.Update)
	mux.HandleFunc("DELETE /api/v1/employees/{id}", empHandler.Delete)

	return db, mux
}

func TestEmployee_FullCRUDIntegration(t *testing.T) {
	db, router := setupTestApp(t)
	defer db.Close()

	var createdEmp model.Employee
	var emp2 model.Employee

	// 1. Create Employee
	t.Run("Create valid employee", func(t *testing.T) {
		body := map[string]any{
			"name":         "John Doe",
			"email":        "John.Doe@example.com",
			"department":   "Engineering",
			"designation":  "Senior Backend Engineer",
			"joining_date": "2024-03-01",
			"status":       "Active",
		}
		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &createdEmp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if createdEmp.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if createdEmp.Email != "john.doe@example.com" {
			t.Errorf("expected normalized lowercase email, got %s", createdEmp.Email)
		}
	})

	// 2. Reject duplicate email (Conflict 409)
	t.Run("Reject duplicate email case-insensitively", func(t *testing.T) {
		body := map[string]any{
			"name":         "Another John",
			"email":        "JOHN.DOE@example.com", // same email different case
			"department":   "Finance",
			"designation":  "Analyst",
			"joining_date": "2024-04-01",
			"status":       "Active",
		}
		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 3. Reject invalid inputs (Bad Request 400)
	t.Run("Reject invalid inputs with clear field validation", func(t *testing.T) {
		body := map[string]any{
			"name":         "",
			"email":        "not-an-email",
			"department":   "",
			"designation":  "",
			"joining_date": "invalid-date",
			"status":       "Archived",
		}
		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
		}

		var errResp response.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("failed to unmarshal error response: %v", err)
		}

		if errResp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("expected code VALIDATION_ERROR, got %s", errResp.Error.Code)
		}
		if len(errResp.Error.Fields) < 5 {
			t.Errorf("expected field errors for all invalid fields, got %v", errResp.Error.Fields)
		}
	})

	// 4. Get by ID
	t.Run("Get employee by ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/employees/%d", createdEmp.ID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var emp model.Employee
		if err := json.Unmarshal(rec.Body.Bytes(), &emp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if emp.ID != createdEmp.ID || emp.Name != createdEmp.Name {
			t.Errorf("got %v, want %v", emp, createdEmp)
		}
	})

	// 5. Update employee
	t.Run("Update employee details", func(t *testing.T) {
		updateBody := map[string]any{
			"name":         "Johnathan Doe",
			"email":        "johnathan.doe@example.com",
			"department":   "Engineering",
			"designation":  "Lead Backend Engineer",
			"joining_date": "2024-03-01",
			"status":       "Inactive",
		}
		data, _ := json.Marshal(updateBody)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/employees/%d", createdEmp.ID), bytes.NewReader(data))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var updated model.Employee
		if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if updated.Name != "Johnathan Doe" || updated.Status != "Inactive" {
			t.Errorf("unexpected updated data: %v", updated)
		}
	})

	// 6. List with filtering, search, and pagination
	t.Run("List with filtering and search", func(t *testing.T) {
		// Create a second employee
		emp2Body := map[string]any{
			"name":         "Sarah Connor",
			"email":        "sarah@skynet.com",
			"department":   "Security",
			"designation":  "Security Officer",
			"joining_date": "2024-05-10",
			"status":       "Active",
		}
		data2, _ := json.Marshal(emp2Body)
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(data2))
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusCreated {
			t.Fatalf("failed to create second employee: %s", rec2.Body.String())
		}
		_ = json.Unmarshal(rec2.Body.Bytes(), &emp2)

		// Filter by search="sarah"
		reqSearch := httptest.NewRequest(http.MethodGet, "/api/v1/employees?search=sarah&page=1&limit=10", nil)
		recSearch := httptest.NewRecorder()
		router.ServeHTTP(recSearch, reqSearch)

		var searchResp model.EmployeeListResponse
		_ = json.Unmarshal(recSearch.Body.Bytes(), &searchResp)

		if searchResp.Pagination.TotalRecords != 1 || len(searchResp.Data) != 1 {
			t.Errorf("expected 1 search match, got %d", searchResp.Pagination.TotalRecords)
		}
		// Combined filter: department="Security" AND status="Active"
		reqCombined := httptest.NewRequest(http.MethodGet, "/api/v1/employees?department=Security&status=Active", nil)
		recCombined := httptest.NewRecorder()
		router.ServeHTTP(recCombined, reqCombined)
		var combinedResp model.EmployeeListResponse
		_ = json.Unmarshal(recCombined.Body.Bytes(), &combinedResp)
		if combinedResp.Pagination.TotalRecords != 1 || combinedResp.Data[0].Email != "sarah@skynet.com" {
			t.Errorf("expected 1 match for combined filter, got %d", combinedResp.Pagination.TotalRecords)
		}

		// Search matching email domain
		reqEmailSearch := httptest.NewRequest(http.MethodGet, "/api/v1/employees?search=skynet.com", nil)
		recEmailSearch := httptest.NewRecorder()
		router.ServeHTTP(recEmailSearch, reqEmailSearch)
		var emailSearchResp model.EmployeeListResponse
		_ = json.Unmarshal(recEmailSearch.Body.Bytes(), &emailSearchResp)
		if emailSearchResp.Pagination.TotalRecords != 1 || emailSearchResp.Data[0].Name != "Sarah Connor" {
			t.Errorf("expected Sarah Connor for email search, got %v", emailSearchResp.Data)
		}

		// Search matching employee ID (numeric)
		reqIDSearch := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/employees?search=%d", createdEmp.ID), nil)
		recIDSearch := httptest.NewRecorder()
		router.ServeHTTP(recIDSearch, reqIDSearch)
		var idSearchResp model.EmployeeListResponse
		_ = json.Unmarshal(recIDSearch.Body.Bytes(), &idSearchResp)
		if idSearchResp.Pagination.TotalRecords < 1 || idSearchResp.Data[0].ID != createdEmp.ID {
			t.Errorf("expected employee with ID %d as top result, got %v", createdEmp.ID, idSearchResp.Data)
		}

		// Search matching employee ID with '#' prefix (e.g. "#1")
		reqHashIDSearch := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/employees?search=%%23%d", createdEmp.ID), nil)
		recHashIDSearch := httptest.NewRecorder()
		router.ServeHTTP(recHashIDSearch, reqHashIDSearch)
		var hashIDSearchResp model.EmployeeListResponse
		_ = json.Unmarshal(recHashIDSearch.Body.Bytes(), &hashIDSearchResp)
		if hashIDSearchResp.Pagination.TotalRecords < 1 || hashIDSearchResp.Data[0].ID != createdEmp.ID {
			t.Errorf("expected employee with ID %d for # search, got %v", createdEmp.ID, hashIDSearchResp.Data)
		}

		// Combined filter that yields 0 matches
		reqZero := httptest.NewRequest(http.MethodGet, "/api/v1/employees?department=Security&status=Inactive", nil)
		recZero := httptest.NewRecorder()
		router.ServeHTTP(recZero, reqZero)
		var zeroResp model.EmployeeListResponse
		_ = json.Unmarshal(recZero.Body.Bytes(), &zeroResp)
		if zeroResp.Pagination.TotalRecords != 0 || len(zeroResp.Data) != 0 {
			t.Errorf("expected 0 matches, got %d", zeroResp.Pagination.TotalRecords)
		}
	})

	// 7. Reject duplicate email on update (case-insensitively)
	t.Run("Reject duplicate email on update case-insensitively", func(t *testing.T) {
		body := map[string]any{
			"name":         "Sarah Connor",
			"email":        "JOHNATHAN.DOE@EXAMPLE.COM", // conflict with createdEmp.Email
			"department":   "Security",
			"designation":  "Director",
			"joining_date": "2024-05-10",
			"status":       "Active",
		}
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/employees/%d", emp2.ID), bytes.NewReader(data))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict on update with duplicate email, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 8. Concurrent duplicate employee inserts
	t.Run("Concurrent duplicate employee inserts protect database integrity", func(t *testing.T) {
		const numGoroutines = 8
		email := "concurrent.race.test@example.com"
		var successCount, conflictCount int64
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				body := map[string]any{
					"name":         fmt.Sprintf("Concurrent User %d", idx),
					"email":        email,
					"department":   "Engineering",
					"designation":  "Load Tester",
					"joining_date": "2024-06-01",
					"status":       "Active",
				}
				data, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(data))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				if rec.Code == http.StatusCreated {
					atomic.AddInt64(&successCount, 1)
				} else if rec.Code == http.StatusConflict {
					atomic.AddInt64(&conflictCount, 1)
				}
			}(i)
		}
		wg.Wait()

		if successCount != 1 {
			t.Errorf("expected exactly 1 successful insert among concurrent requests, got %d", successCount)
		}
		if conflictCount != int64(numGoroutines-1) {
			t.Errorf("expected %d conflicts among concurrent requests, got %d", numGoroutines-1, conflictCount)
		}
	})

	// 9. Error handling for malformed IDs & non-existent IDs
	t.Run("Reject malformed ID with 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for malformed ID, got %d", rec.Code)
		}
	})

	t.Run("Return 404 for non-existent employee", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/999999", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for non-existent employee, got %d", rec.Code)
		}
	})

	// 8. Delete employee
	t.Run("Delete employee", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/employees/%d", createdEmp.ID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status 204 No Content, got %d", rec.Code)
		}

		// Verify deletion
		reqVerify := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/employees/%d", createdEmp.ID), nil)
		recVerify := httptest.NewRecorder()
		router.ServeHTTP(recVerify, reqVerify)

		if recVerify.Code != http.StatusNotFound {
			t.Errorf("expected status 404 after deletion, got %d", recVerify.Code)
		}
	})

	// 8. Health check
	t.Run("Health check reports UP with Postgres", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
