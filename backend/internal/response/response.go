package response

import (
	"encoding/json"
	"net/http"
)

// ErrorDetail describes an API error.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ErrorResponse wraps the error detail.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// JSON sends a JSON response with status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// Error sends an ErrorResponse.
func Error(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	JSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
	})
}

// Common error helpers:
func BadRequest(w http.ResponseWriter, message string, fields map[string]string) {
	Error(w, http.StatusBadRequest, "VALIDATION_ERROR", message, fields)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "DUPLICATE_RESOURCE", message, map[string]string{"email": message})
}

func InternalServerError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred. Please try again later.", nil)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message, nil)
}

func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, nil)
}
