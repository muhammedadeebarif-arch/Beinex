package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"ems-backend/internal/model"
	"ems-backend/internal/response"
	"ems-backend/internal/service"
)

// EmployeeHandler handles employee-related HTTP requests.
type EmployeeHandler struct {
	svc service.EmployeeService
}

// NewEmployeeHandler creates a new EmployeeHandler.
func NewEmployeeHandler(svc service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

// Create handles POST /api/v1/employees
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", map[string]string{"body": "Malformed JSON body."})
		return
	}

	emp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			response.BadRequest(w, "Please correct the submitted fields.", valErr.Fields)
			return
		}
		if errors.Is(err, service.ErrDuplicateEmail) {
			response.Conflict(w, "An employee with this email address already exists.")
			return
		}
		log.Printf("[ERROR] Failed to create employee: %v", err)
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusCreated, emp)
}

// List handles GET /api/v1/employees
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	filter := model.EmployeeFilter{
		Search:     q.Get("search"),
		Department: q.Get("department"),
		Status:     q.Get("status"),
		Page:       page,
		Limit:      limit,
	}

	result, err := h.svc.List(r.Context(), filter)
	if err != nil {
		log.Printf("[ERROR] Failed to list employees: %v", err)
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetByID handles GET /api/v1/employees/{id}
func (h *EmployeeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Employee ID must be a positive integer.", map[string]string{"id": "Invalid ID parameter."})
		return
	}

	emp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "Employee not found.")
			return
		}
		if errors.Is(err, service.ErrInvalidID) {
			response.BadRequest(w, "Employee ID must be a positive integer.", map[string]string{"id": "Invalid ID parameter."})
			return
		}
		log.Printf("[ERROR] Failed to get employee %d: %v", id, err)
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, emp)
}

// Update handles PUT /api/v1/employees/{id}
func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Employee ID must be a positive integer.", map[string]string{"id": "Invalid ID parameter."})
		return
	}

	var req model.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", map[string]string{"body": "Malformed JSON body."})
		return
	}

	emp, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			response.BadRequest(w, "Please correct the submitted fields.", valErr.Fields)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "Employee not found.")
			return
		}
		if errors.Is(err, service.ErrDuplicateEmail) {
			response.Conflict(w, "An employee with this email address already exists.")
			return
		}
		if errors.Is(err, service.ErrInvalidID) {
			response.BadRequest(w, "Employee ID must be a positive integer.", map[string]string{"id": "Invalid ID parameter."})
			return
		}
		log.Printf("[ERROR] Failed to update employee %d: %v", id, err)
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, emp)
}

// Delete handles DELETE /api/v1/employees/{id}
func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Employee ID must be a positive integer.", map[string]string{"id": "Invalid ID parameter."})
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "Employee not found.")
			return
		}
		if errors.Is(err, service.ErrInvalidID) {
			response.BadRequest(w, "Employee ID must be a positive integer.", map[string]string{"id": "Invalid ID parameter."})
			return
		}
		log.Printf("[ERROR] Failed to delete employee %d: %v", id, err)
		response.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
