package main

import (
	"log"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/router"
)

func main() {
	cfg := config.Load() // load envs, firebase, etc.
	app := router.Setup(cfg)
	log.Fatal(app.Listen(":" + cfg.Port))
}
