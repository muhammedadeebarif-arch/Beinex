package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ems-backend/internal/model"
	"ems-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenInvalid        = errors.New("invalid authentication token")
)

// AdminClaims defines JWT token payload.
type AdminClaims struct {
	AdminID int64  `json:"admin_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	jwt.RegisteredClaims
}

// AuthService defines business operations for admin authentication.
type AuthService interface {
	Login(ctx context.Context, email, password string) (*model.LoginResponse, error)
	ValidateToken(tokenString string) (*AdminClaims, error)
	ProvisionInitialAdmin(ctx context.Context, name, email, password string) error
	GetAdminByID(ctx context.Context, id int64) (*model.AdminProfile, error)
}

type authService struct {
	adminRepo repository.AdminRepository
	jwtSecret []byte
	expiry    time.Duration
}

// NewAuthService creates a new AuthService.
func NewAuthService(adminRepo repository.AdminRepository, jwtSecret string, expiryHours int) AuthService {
	if expiryHours <= 0 {
		expiryHours = 24
	}
	return &authService{
		adminRepo: adminRepo,
		jwtSecret: []byte(jwtSecret),
		expiry:    time.Duration(expiryHours) * time.Hour,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*model.LoginResponse, error) {
	cleanEmail := strings.TrimSpace(email)
	cleanPass := strings.TrimSpace(password)

	if cleanEmail == "" || cleanPass == "" {
		return nil, ErrInvalidCredentials
	}

	admin, err := s.adminRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			// Generic message to prevent user enumeration
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to retrieve admin: %w", err)
	}

	// Verify bcrypt password hash
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(cleanPass)); err != nil {
		// Generic message to prevent disclosing password validity
		return nil, ErrInvalidCredentials
	}

	// Issue JWT token
	now := time.Now()
	claims := AdminClaims{
		AdminID: admin.ID,
		Email:   admin.Email,
		Name:    admin.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "beinex-ems",
			Subject:   fmt.Sprintf("%d", admin.ID),
			Audience:  jwt.ClaimStrings{"beinex-admin"},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &model.LoginResponse{
		Token: tokenString,
		Admin: admin.ToProfile(),
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func (s *authService) ProvisionInitialAdmin(ctx context.Context, name, email, password string) error {
	count, err := s.adminRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing admins: %w", err)
	}

	if count > 0 {
		return nil // Admin already provisioned
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	_, err = s.adminRepo.Create(ctx, &model.AdminUser{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return fmt.Errorf("failed to provision initial admin: %w", err)
	}

	return nil
}

func (s *authService) GetAdminByID(ctx context.Context, id int64) (*model.AdminProfile, error) {
	// For profile query, we can query by email or we can add a lightweight lookup if needed.
	// We'll support retrieving via repository if needed.
	return nil, nil
}
