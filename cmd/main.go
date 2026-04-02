package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/M4yankkkk/findash/internal/config"
	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/routes"
	"github.com/gin-gonic/gin"
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
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		log.Fatalf("❌ Migration error: %v", err)
	}

	r := gin.Default()
	routes.Setup(r, db, cfg)

	log.Printf("🚀 Server running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
