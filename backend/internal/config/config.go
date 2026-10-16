package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the application.
type Config struct {
	ServerPort      string
	CORSAllowedOrig string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBConnMaxLife   time.Duration
	StaticDir       string
	JWTSecret       string
	JWTExpirationHours int
	AdminEmail      string
	AdminPassword   string
	AdminName       string
}

// Load loads the configuration from environment variables or .env file.
func Load() (*Config, error) {
	// Attempt to load .env if present; ignore error if not found.
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		CORSAllowedOrig:    getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:4200"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "employee_db"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:     getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:     getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLife:      time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINS", 15)) * time.Minute,
		StaticDir:          resolveStaticDir(),
		JWTSecret:          getEnv("JWT_SECRET", "beinex-consulting-jwt-secure-signing-secret-key-2026"),
		JWTExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		AdminEmail:         getEnv("ADMIN_EMAIL", "admin@beinex.com"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", "BeinexAdmin@2026!"),
		AdminName:          getEnv("ADMIN_NAME", "Beinex Administrator"),
	}

	return cfg, nil
}

func resolveStaticDir() string {
	if custom := os.Getenv("STATIC_DIR"); custom != "" {
		return custom
	}

	candidates := []string{
		filepath.Join("..", "frontend", "dist", "frontend", "browser"),
		filepath.Join("frontend", "dist", "frontend", "browser"),
		filepath.Join(".", "dist", "frontend", "browser"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "index.html")); err == nil {
			return c
		}
	}

	return filepath.Join("..", "frontend", "dist", "frontend", "browser")
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
