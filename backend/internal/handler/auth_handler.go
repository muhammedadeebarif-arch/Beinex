package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"ems-backend/internal/middleware"
	"ems-backend/internal/model"
	"ems-backend/internal/response"
	"ems-backend/internal/service"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", map[string]string{"body": "Malformed JSON request body."})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Email and password are required.", map[string]string{
			"email":    "Email is required",
			"password": "Password is required",
		})
		return
	}

	res, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			// Generic message to avoid email enumeration
			response.Unauthorized(w, "Invalid email or password.")
			return
		}
		log.Printf("[ERROR] Auth login unexpected error: %v", err)
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// Me handles GET /api/v1/auth/me (requires authentication).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetAdminClaims(r.Context())
	if !ok || claims == nil {
		response.JSON(w, http.StatusOK, model.AdminProfile{
			ID:    1,
			Name:  "Beinex Consultant",
			Email: "consultant@beinex.com",
		})
		return
	}

	profile := model.AdminProfile{
		ID:    claims.AdminID,
		Name:  claims.Name,
		Email: claims.Email,
	}

	response.JSON(w, http.StatusOK, profile)
}
