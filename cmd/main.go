package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/M4yankkkk/findash/internal/config"
	"github.com/M4yankkkk/findash/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from .env / environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config error: %v", err)
	}

	// Set Gin mode before creating the engine
	gin.SetMode(cfg.GinMode)

	// Connect to PostgreSQL
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("❌ Database error: %v", err)
	}
	defer db.Close()

	// Run SQL migrations from the /migrations directory
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		log.Fatalf("❌ Migration error: %v", err)
	}

	// Bootstrap Gin router
	// Routes and middleware will be wired here in Phase 2
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "finance-dashboard",
		})
	})

	log.Printf("🚀 Server running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
