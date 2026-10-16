package model

import "time"

// AdminUser represents an administrator account in the database.
type AdminUser struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// AdminProfile represents the public-facing admin profile.
type AdminProfile struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToProfile converts an AdminUser to an AdminProfile.
func (u *AdminUser) ToProfile() AdminProfile {
	return AdminProfile{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// LoginRequest is the payload sent to POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response returned upon successful authentication.
type LoginResponse struct {
	Token string       `json:"token"`
	Admin AdminProfile `json:"admin"`
}
