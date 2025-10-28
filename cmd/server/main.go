package main

import (
	"log"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/router"
)

// @title Rally Backend API
// @version 1.0
// @description API documentation for Rally backend service
// @contact.name Rally
// @BasePath /api/v1
func main() {
	cfg := config.Load() // load envs, firebase, etc.
	app := router.Setup(cfg)
	log.Fatal(app.Listen(":" + cfg.Port))
}
