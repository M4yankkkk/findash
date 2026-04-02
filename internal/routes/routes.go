package routes

import (
	"github.com/M4yankkkk/findash/internal/config"
	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/handlers"
	"github.com/M4yankkkk/findash/internal/middleware"
	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/repository"
	"github.com/M4yankkkk/findash/internal/services"
	"github.com/gin-gonic/gin"
)

// Setup wires all dependencies and registers all routes on the given Gin engine.
func Setup(r *gin.Engine, db *database.DB, cfg *config.Config) {
	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	entryRepo := repository.NewEntryRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	authService := services.NewAuthService(userRepo, auditRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	userService := services.NewUserService(userRepo, auditRepo)
	entryService := services.NewEntryService(entryRepo, auditRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	entryHandler := handlers.NewEntryHandler(entryService)

	// ── Health check (no auth) ────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "finance-dashboard"})
	})

	api := r.Group("/api/v1")

	// ── Public routes ─────────────────────────────────────────────────────────
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// ── Protected routes (JWT required) ───────────────────────────────────────
	auth := api.Group("")
	auth.Use(middleware.Authenticate(cfg.JWTSecret))
	{
		// Users — admin only
		users := auth.Group("/users")
		users.Use(middleware.RequireAdmin())
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PATCH("/:id/role", userHandler.UpdateRole)
		}

		// Entries — all authenticated users
		entries := auth.Group("/entries")
		{
			// Create is restricted to manager and admin
			entries.POST("", middleware.RequireManagerOrAdmin(), entryHandler.CreateEntry)

			// List and get are open to all roles (service layer handles scoping)
			entries.GET("", entryHandler.ListEntries)
			entries.GET("/:id", entryHandler.GetEntry)

			// Update and delete: ownership enforced in service layer
			entries.PUT("/:id", middleware.RequireManagerOrAdmin(), entryHandler.UpdateEntry)
			entries.DELETE("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleManager), entryHandler.DeleteEntry)
		}

		// Phase 4: analytics routes will be added here
	}
}
