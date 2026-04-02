package routes

import (
	"github.com/M4yankkkk/findash/internal/config"
	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/handlers"
	"github.com/M4yankkkk/findash/internal/middleware"
	"github.com/M4yankkkk/findash/internal/repository"
	"github.com/M4yankkkk/findash/internal/services"
	"github.com/gin-gonic/gin"
)

// Setup wires all dependencies and registers all routes on the given Gin engine.
// This is the single place that owns the full dependency graph.
func Setup(r *gin.Engine, db *database.DB, cfg *config.Config) {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, auditRepo, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)

	// ── Public routes (no auth required) ──────────────────────────────────────
	public := r.Group("/api/v1")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
	}

	// ── Protected routes (JWT required) ───────────────────────────────────────
	protected := r.Group("/api/v1")
	protected.Use(middleware.Authenticate(cfg.JWTSecret))
	{
		// Phase 3: user and entry routes will be added here
		// Phase 4: analytics routes will be added here
	}

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "finance-dashboard"})
	})
}
