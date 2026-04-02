// @title           Finance Dashboard API
// @version         1.0
// @description     A role-based finance dashboard backend built with Go and Gin.

// @contact.name    Mayank
// @contact.email   mayank@example.com

// @license.name    MIT

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token as: Bearer <token>

package main

import (
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/M4yankkkk/findash/internal/config"
	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/M4yankkkk/findash/docs" // injected by swag init
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config error: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("❌ Database error: %v", err)
	}
	defer db.Close()

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		log.Fatalf("❌ Migration error: %v", err)
	}

	r := gin.Default()

	// CORS — allows the React frontend (port 5173) to call the API
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.Setup(r, db, cfg)

	log.Printf("🚀 Server running on port %s", cfg.Port)
	log.Printf("📖 Swagger UI at http://localhost:%s/swagger/index.html", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
