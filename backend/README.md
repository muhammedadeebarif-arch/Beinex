# Go Backend — Employee Management System

## Architectural Pattern
This backend adheres strictly to the **Handler → Service → Repository → PostgreSQL** architecture:
- `cmd/api/main.go`: Application initialization, config loading, migration execution, and graceful shutdown.
- `internal/config`: Loads settings from `.env` or system environment variables.
- `internal/database`: PostgreSQL connection pool setup using `jackc/pgx/v5`.
- `internal/handler`: HTTP request routing (Go 1.22+ ServeMux), REST JSON serialization, and SPAHandler for serving the compiled Angular frontend.
- `internal/service`: Business validation, email normalization, and pagination metadata calculation.
- `internal/repository`: Parameterized PostgreSQL queries with SQL state error translation.
- `internal/middleware`: Request logging, panic recovery, CORS, and 1MB body limit protection.
- `internal/response`: Unified API response structures.

## Running Backend
```powershell
go run cmd/api/main.go
```

## Running Tests
```powershell
go test -v ./...
```
