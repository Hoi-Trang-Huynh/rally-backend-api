package main

import (
	"log"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/database"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/firebase"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/router"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/version"
)

// @title Rally Backend API
// @version 1.0
// @description API documentation for Rally backend service
// @contact.name Rally
// @BasePath /api/v1
func main() {
	log.Printf("Starting Rally Backend API %s", version.Version)
	log.Printf("Commit SHA: %s, Build Time: %s", version.CommitSHA, version.BuildTime)

	cfg := config.Load() // load envs, firebase, etc.
	if err := database.InitializeDatabase(cfg.Database); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer database.CloseDatabase()
	firebase.MustInitialize(cfg.Firebase.CredentialsPath)
	app := router.Setup(cfg)
	log.Fatal(app.Listen(":" + cfg.Server.Port))
}
