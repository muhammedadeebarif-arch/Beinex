package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ems-backend/internal/config"
	"ems-backend/internal/database"
	"ems-backend/internal/handler"
	"ems-backend/internal/middleware"
	"ems-backend/internal/repository"
	"ems-backend/internal/service"
)

func main() {
	log.Println("Starting Employee Management System API...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fatal: could not load config: %v", err)
	}

	// 2. Connect to PostgreSQL
	log.Printf("Connecting to PostgreSQL at %s:%s (database: %s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Fatal: could not connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database pool: %v", err)
		}
	}()
	log.Println("Successfully connected to PostgreSQL.")

	// 3. Apply SQL migrations
	migrationPath := "migrations"
	if _, err := os.Stat(migrationPath); errors.Is(err, os.ErrNotExist) {
		// Try checking relative to parent if running from another directory
		migrationPath = filepath.Join("..", "migrations")
	}

	ctxMigrate, cancelMigrate := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelMigrate()

	if err := db.ApplyMigrations(ctxMigrate, migrationPath); err != nil {
		log.Fatalf("Fatal: database migration failed: %v", err)
	}
	log.Println("Database migrations verified and applied successfully.")

	// 4. Initialize Dependency Injection: Repository -> Service -> Handler
	empRepo := repository.NewPostgresEmployeeRepository(db.Pool)
	empService := service.NewEmployeeService(empRepo)
	empHandler := handler.NewEmployeeHandler(empService)

	adminRepo := repository.NewPostgresAdminRepository(db.Pool)
	authService := service.NewAuthService(adminRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	authHandler := handler.NewAuthHandler(authService)

	healthHandler := handler.NewHealthHandler(db)
	spaHandler := handler.NewSPAHandler(cfg.StaticDir)

	// Provision initial administrator if needed
	ctxProv, cancelProv := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelProv()
	if err := authService.ProvisionInitialAdmin(ctxProv, cfg.AdminName, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Printf("[WARN] Failed to verify initial administrator: %v", err)
	} else {
		log.Printf("Administrator provisioning verified for %s", cfg.AdminEmail)
	}

	// Rate limiter for login requests (10 attempts per minute)
	loginLimiter := middleware.NewLoginRateLimiter(10, 1*time.Minute)

	// 5. Register HTTP Routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler.HealthCheck)

	// Authentication endpoints (optional/available)
	mux.Handle("POST /api/v1/auth/login", loginLimiter.Limit(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("GET /api/v1/auth/me", authHandler.Me)

	// Employee REST endpoints (Direct access - no admin email/password gate)
	mux.HandleFunc("POST /api/v1/employees", empHandler.Create)
	mux.HandleFunc("GET /api/v1/employees", empHandler.List)
	mux.HandleFunc("GET /api/v1/employees/{id}", empHandler.GetByID)
	mux.HandleFunc("PUT /api/v1/employees/{id}", empHandler.Update)
	mux.HandleFunc("DELETE /api/v1/employees/{id}", empHandler.Delete)

	// Angular Frontend SPA & Static Assets Fallback
	mux.Handle("/", spaHandler)

	// 6. Chain Middlewares: BodyLimit -> CORS -> Recoverer -> Logger
	var h http.Handler = mux
	h = middleware.BodyLimit(1 << 20)(h) // 1 MB limit
	h = middleware.CORS(cfg.CORSAllowedOrig)(h)
	h = middleware.Recoverer(h)
	h = middleware.Logger(h)

	// 7. Setup HTTP Server with Timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Run Server in goroutine
	shutdownErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP Server is listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdownErr <- err
		}
	}()

	// 9. Graceful Shutdown on OS signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-shutdownErr:
		log.Fatalf("Fatal server failure: %v", err)
	case sig := <-stopChan:
		log.Printf("Received termination signal (%v). Initiating graceful shutdown...", sig)
	}

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
