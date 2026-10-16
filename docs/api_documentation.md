# REST API Documentation — Employee Management System

Base URL: `http://localhost:8080` (or configured `SERVER_PORT`)

Prefix: `/api/v1`

---

## 1. Health Readiness Check

### `GET /health`
Verifies server liveness and PostgreSQL database readiness.

**Request:**
```http
GET /health HTTP/1.1
Host: localhost:8080
```

**Response (200 OK):**
```json
{
  "database": "UP",
  "status": "UP",
  "time": "2026-10-16T10:58:14Z"
}
```

---

## 2. List Employees

### `GET /api/v1/employees`
Returns a paginated list of employees with optional search and filters.

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `search` | string | No | `""` | Case-insensitive search matching `name` or `email` |
| `department` | string | No | `""` | Exact department filter (e.g., `Engineering`) |
| `status` | string | No | `""` | Status filter (`Active` or `Inactive`) |
| `page` | integer | No | `1` | Page number (1-based) |
| `limit` | integer | No | `10` | Records per page (Max: 100) |

**Request Example:**
```http
GET /api/v1/employees?search=arjun&department=Engineering&status=Active&page=1&limit=10 HTTP/1.1
Host: localhost:8080
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Arjun Nair",
      "email": "arjun.nair@beinex.com",
      "department": "Engineering",
      "designation": "Lead Cloud Architect",
      "joining_date": "2023-01-10",
      "status": "Active",
      "created_at": "2026-10-16T16:21:33.35231+05:30",
      "updated_at": "2026-10-16T16:21:33.35231+05:30"
    }
  ],
  "pagination": {
    "current_page": 1,
    "limit": 10,
    "total_records": 1,
    "total_pages": 1
  }
}
```

---

## 3. Get Employee By ID

### `GET /api/v1/employees/{id}`
Retrieves an individual employee record by ID.

**Request:**
```http
GET /api/v1/employees/1 HTTP/1.1
Host: localhost:8080
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Arjun Nair",
  "email": "arjun.nair@beinex.com",
  "department": "Engineering",
  "designation": "Lead Cloud Architect",
  "joining_date": "2023-01-10",
  "status": "Active",
  "created_at": "2026-10-16T16:21:33.35231+05:30",
  "updated_at": "2026-10-16T16:21:33.35231+05:30"
}
```

**Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Employee not found."
  }
}
```

**Response (400 Bad Request):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Employee ID must be a positive integer.",
    "fields": {
      "id": "Invalid ID parameter."
    }
  }
}
```

---

## 4. Create Employee

### `POST /api/v1/employees`
Creates a new employee record. Validates all required fields, date formats, status enum, and case-insensitive email uniqueness.

**Request:**
```http
POST /api/v1/employees HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "name": "Priya Sharma",
  "email": "priya.sharma@beinex.com",
  "department": "Consulting",
  "designation": "Senior Data Consultant",
  "joining_date": "2023-04-15",
  "status": "Active"
}
```

**Response (201 Created):**
```json
{
  "id": 2,
  "name": "Priya Sharma",
  "email": "priya.sharma@beinex.com",
  "department": "Consulting",
  "designation": "Senior Data Consultant",
  "joining_date": "2023-04-15",
  "status": "Active",
  "created_at": "2026-10-16T16:28:40.10012+05:30",
  "updated_at": "2026-10-16T16:28:40.10012+05:30"
}
```

**Response (409 Conflict - Duplicate Email):**
```json
{
  "error": {
    "code": "DUPLICATE_RESOURCE",
    "message": "An employee with this email address already exists.",
    "fields": {
      "email": "An employee with this email address already exists."
    }
  }
}
```

**Response (400 Bad Request - Validation Error):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Please correct the submitted fields.",
    "fields": {
      "email": "Please enter a valid email address.",
      "joining_date": "Joining date must be in YYYY-MM-DD format.",
      "name": "Employee name is required."
    }
  }
}
```

---

## 5. Update Employee

### `PUT /api/v1/employees/{id}`
Updates an existing employee.

**Request:**
```http
PUT /api/v1/employees/2 HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "name": "Priya Sharma",
  "email": "priya.sharma@beinex.com",
  "department": "Consulting",
  "designation": "Lead Consultant",
  "joining_date": "2023-04-15",
  "status": "Active"
}
```

**Response (200 OK):**
```json
{
  "id": 2,
  "name": "Priya Sharma",
  "email": "priya.sharma@beinex.com",
  "department": "Consulting",
  "designation": "Lead Consultant",
  "joining_date": "2023-04-15",
  "status": "Active",
  "created_at": "2026-10-16T16:28:40.10012+05:30",
  "updated_at": "2026-10-16T16:30:10.51230+05:30"
}
```

---

## 6. Delete Employee

### `DELETE /api/v1/employees/{id}`
Permanently deletes the employee record.

**Request:**
```http
DELETE /api/v1/employees/2 HTTP/1.1
Host: localhost:8080
```

**Response (204 No Content):**
Empty body.

**Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Employee not found."
  }
}
```
