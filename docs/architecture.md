# System Architecture — Employee Management System

## Architectural Overview

The Employee Management System follows an enterprise-grade **layered clean architecture** with strict separation of concerns, high testability, and zero unneeded bloat:

```
[ Angular 19+ SPA ]
        │  (HTTP / JSON over REST)
        ▼
[ HTTP Middleware ]
  ├─ BodyLimit (1MB)
  ├─ CORS Protection
  ├─ Panic Recovery
  └─ Request Logging
        │
        ▼
[ Handlers (Delivery Layer) ]
  ├─ HTTP parameter decoding
  ├─ Route dispatching (Go 1.22+ ServeMux)
  └─ Unified JSON response & error formatting
        │
        ▼
[ Service (Business Logic Layer) ]
  ├─ Email normalization (trim & lowercase)
  ├─ Business rule validation
  └─ Pagination metadata calculations
        │
        ▼
[ Repository (Data Access Layer) ]
  ├─ Parameterized SQL queries (No SQL injection)
  ├─ Transaction management
  ├─ PostgreSQL error code translation (e.g., 23505 -> Conflict)
  └─ pgx/v5 connection pool management
        │
        ▼
[ PostgreSQL 16 Database ]
  ├─ BIGSERIAL Primary Key
  ├─ Case-insensitive UNIQUE index on LOWER(email)
  ├─ Status CHECK constraint ('Active', 'Inactive')
  └─ Performance indexes on department, status, created_at
```

---

## Key Design Decisions

### 1. Handler → Service → Repository Separation
- **Handler**: Reads JSON requests, parses URL query and path parameters (`r.PathValue`), calls services, and encodes standard responses.
- **Service**: Implements business constraints, orchestrates validation, normalizes inputs, and maps domain results.
- **Repository**: Directly interacts with `*sql.DB`, ensuring all database operations use parameterized statements (`$1, $2, ...`) and context timeouts.

### 2. Case-Insensitive Email Uniqueness
To prevent duplicate emails even under high concurrency:
- Applications trim and convert emails to lower-case before inserting or querying.
- PostgreSQL enforces a functional unique index:
  ```sql
  CREATE UNIQUE INDEX idx_employees_email_unique ON employees (LOWER(email));
  ```
- If two concurrent requests attempt to create an employee with `user@domain.com` and `USER@domain.com`, PostgreSQL immediately rejects the second with SQL state `23505`, which the repository maps cleanly to `ErrDuplicateEmail` and HTTP `409 Conflict`.

### 3. Frontend Reactive Forms & Signals
- Angular standalone architecture without legacy NgModules.
- Reactive forms provide immediate, synchronous client-side validation for required fields, email formats, and date formats.
- Server-side validation errors are mapped directly back to the relevant form field for visual clarity.
- Angular signals (`signal<Employee[]>`, `signal<PaginationMeta>`) provide fast, fine-grained UI reactivity.

### 4. Database Connection Pool Settings
- Connection pooling is configured through environment variables:
  - `DB_MAX_OPEN_CONNS=25`: Prevents socket exhaustion.
  - `DB_MAX_IDLE_CONNS=10`: Keeps healthy warm connections.
  - `DB_CONN_MAX_LIFETIME_MINS=15`: Rotates stale database sockets.

### 5. Single HTTP URL Unified Delivery (SPA Fallback Architecture)
The Go HTTP server serves the compiled Angular Single Page Application (`frontend/dist/frontend/browser`) directly alongside the REST API on a single port (`8080`).
- Exact routes like `GET /health` and `/api/v1/employees` take precedence and execute their respective Go handlers.
- Static assets (`.js`, `.css`, `.png`, `.ico`) are served directly from the distribution directory with proper MIME types.
- Missing static asset files return HTTP 404 rather than HTML, preventing browser script/stylesheet execution errors.
- Any API route (`/api/*`) that is not found returns a JSON 404 error envelope (`{"error":{"code":"NOT_FOUND"}}`) and is never intercepted by the SPA fallback.
- Client-side navigation routes (e.g. `/`, `/employees`) serve `index.html` with `Cache-Control: no-cache` so Angular router initializes the appropriate view.
