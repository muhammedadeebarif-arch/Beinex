# Employee Management System — Beinex Technical Assessment

A production-grade, enterprise Employee Management System built from scratch for the Beinex Consulting technical recruitment assessment.

## Mandatory Technology Stack

- **Frontend:** Angular 19+ (Standalone Components, TypeScript, Reactive Forms, Signals, Modern Responsive CSS)
- **Backend:** Go 1.27+ (Clean Architecture: Handler → Service → Repository → PostgreSQL)
- **Database:** PostgreSQL 16 (Parameterized SQL, transactions, connection pooling, case-insensitive email uniqueness constraints)
- **API:** RESTful JSON HTTP APIs under `/api/v1/employees` and `/health`

---

## Project Structure

```
employee-management-system/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                  # Application entry point with graceful shutdown
│   ├── internal/
│   │   ├── config/                      # Environment-based configuration loader
│   │   ├── database/                    # Connection pool & migration runner
│   │   ├── handler/                     # HTTP Handlers (REST controllers)
│   │   ├── middleware/                  # Logging, Recoverer, CORS, BodyLimit
│   │   ├── model/                       # Data models & DTO definitions
│   │   ├── repository/                  # Parameterized PostgreSQL data access
│   │   ├── response/                    # Standardized API response & error format
│   │   ├── service/                     # Business logic, email normalization & validation
│   │   └── validation/                  # Request payload validation rules
│   ├── migrations/
│   │   └── 001_create_employees_table.sql
│   ├── tests/
│   │   └── integration_test.go          # Full CRUD integration tests with PostgreSQL
│   ├── .env.example
│   ├── .gitignore
│   ├── Dockerfile
│   └── README.md
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── features/
│   │   │   │   └── employees/
│   │   │   │       ├── components/
│   │   │   │       │   ├── employee-form/      # Reactive Form modal (Add/Edit)
│   │   │   │       │   └── delete-modal/       # Accessible delete confirmation dialog
│   │   │   │       ├── models/
│   │   │   │       │   └── employee.model.ts   # Strongly-typed TypeScript interfaces
│   │   │   │       ├── pages/
│   │   │   │       │   └── employee-list/      # Main dashboard page with search & filters
│   │   │   │       └── services/
│   │   │   │           └── employee.service.ts # Angular HttpClient service
│   │   │   ├── app.config.ts
│   │   │   ├── app.routes.ts
│   │   │   ├── app.html
│   │   │   ├── app.css
│   │   │   └── app.ts
│   │   ├── environments/
│   │   │   ├── environment.ts
│   │   │   └── environment.development.ts
│   │   ├── index.html
│   │   └── styles.css                   # Enterprise design system CSS
│   ├── angular.json
│   ├── package.json
│   ├── Dockerfile
│   └── README.md
├── database/
│   ├── migrations/
│   │   └── 001_create_employees_table.sql
│   └── seed.sql                         # Explicit development-only seed data
├── docs/
│   ├── api_documentation.md             # Complete REST API specification with examples
│   └── architecture.md                  # Architectural decisions & design patterns
├── docker-compose.yml                   # Optional multi-container orchestration
├── .gitignore
└── README.md
```

---

## Prerequisites

1. **Go:** Version 1.22+ (Installed: `go version go1.27.0`)
2. **Node.js & npm:** Node 20+ (Installed: `v24.21.0`), npm 10+ (Installed: `11.19.0`)
3. **Angular CLI:** Version 19+ (Installed: `22.2.3`)
4. **PostgreSQL:** Version 14+ (Installed: `PostgreSQL 16.15`)

---

## Quick Start: Single HTTP URL (`http://localhost:8080`)

The application is engineered to run both the Angular frontend and Go backend seamlessly behind a single HTTP URL at **`http://localhost:8080`**.

### Step 1: Start PostgreSQL Database

```powershell
# In Terminal 1: Start PostgreSQL server cluster
& "d:\adeeb\database\pgsql\bin\postgres.exe" -D "d:\adeeb\database\data"
```

*(If setting up for the first time)* create the database and seed records:
```powershell
# Create database
& "d:\adeeb\database\pgsql\bin\createdb.exe" -U postgres employee_db

# Seed initial records
& "d:\adeeb\database\pgsql\bin\psql.exe" -U postgres -d employee_db -f "d:\adeeb\database\seed.sql"
```

### Step 2: Build the Angular Frontend

```powershell
# In Terminal 2: Compile the production Angular application
cd d:\adeeb\frontend
npm run build
```
This outputs the compiled Angular distribution to `frontend/dist/frontend/browser`.

### Step 3: Run the Combined Go Application

```powershell
# In Terminal 2 (or a new terminal): Run the combined Go server
cd d:\adeeb\backend
go run cmd/api/main.go
```

The Go server automatically:
1. Verifies connectivity to PostgreSQL.
2. Applies any pending database migrations.
3. Serves the compiled Angular Single Page Application and static assets on root `/` and all client-side routes.
4. Handles all REST API requests on `/api/v1/employees` and `/health`.

👉 **Open in your browser:** **[http://localhost:8080](http://localhost:8080)**

---

### Alternative: Separate Development Mode (Optional)

If you wish to run the Angular dev server with Hot Module Reloading separately:
1. Run backend on port `8080`: `cd d:\adeeb\backend; go run cmd/api/main.go`
2. Run Angular dev server on port `4200`: `cd d:\adeeb\frontend; npm start`
3. The Angular dev server automatically proxies `/api` requests to `http://localhost:8080` via `proxy.conf.json`.

---

## Running Tests

### Backend Unit & Integration Tests

All backend tests execute against the PostgreSQL database instance:
```powershell
cd d:\adeeb\backend
go test -v ./...
```

**Verified Test Results:**
- `TestValidateCreateEmployee`: PASS (Required fields, email formatting, date formats, status enum)
- `TestNormalizeEmail`: PASS (Whitespace trimming, case-folding)
- `TestEmployeeService_Create`: PASS (Service validation & normalization)
- `TestEmployeeService_PaginationCalculation`: PASS (Pagination boundaries)
- `TestEmployee_FullCRUDIntegration`: PASS
  - Create valid employee (201 Created)
  - Case-insensitive duplicate email rejection (409 Conflict)
  - Comprehensive field-level validation errors (400 Bad Request)
  - Get employee by ID (200 OK & 404 Not Found)
  - Update employee record (200 OK)
  - Search & filter by department/status (200 OK)
  - Permanent delete (204 No Content & subsequent 404)
  - Health check endpoint (200 OK)

### Frontend Production Build

```powershell
cd d:\adeeb\frontend
npm run build
```
Generates an optimized production bundle with zero errors.

---

## API Summary

| Method | Endpoint | Description | Status Codes |
|---|---|---|---|
| `GET` | `/health` | Application & DB health check | `200`, `503` |
| `POST` | `/api/v1/employees` | Create employee | `201`, `400`, `409` |
| `GET` | `/api/v1/employees` | List employees (search, filters, pagination) | `200` |
| `GET` | `/api/v1/employees/{id}` | Get employee by ID | `200`, `400`, `404` |
| `PUT` | `/api/v1/employees/{id}` | Update employee | `200`, `400`, `404`, `409` |
| `DELETE` | `/api/v1/employees/{id}` | Delete employee | `204`, `400`, `404` |

Refer to [docs/api_documentation.md](file:///d:/adeeb/docs/api_documentation.md) for full request/response examples.

---

## Architecture Highlights

- **Handler → Service → Repository Pattern:** Clean separation of concerns.
- **SQL Injection Prevention:** Every query uses parameterized statements (`$1, $2, ...`).
- **Concurrent Duplicate Email Prevention:** Enforced at both the service layer (case normalization) and PostgreSQL layer (`UNIQUE INDEX idx_employees_email_unique ON employees (LOWER(email))`).
- **Graceful Shutdown:** Intercepts `SIGINT`/`SIGTERM` to allow active HTTP requests and database transactions to finish cleanly.
- **Enterprise Design System:** Custom CSS tokens, accessible labels, loading spinners, debounce-driven search, and responsive cards.

Refer to [docs/architecture.md](file:///d:/adeeb/docs/architecture.md) for detailed architectural design patterns.
