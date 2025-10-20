package router

import (
	_ "github.com/Hoi-Trang-Huynh/rally-backend-api/api/docs" // docs package
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/handler"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/middleware"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger" // swagger handler
)

func Setup(cfg *config.Config) *fiber.App {
	app := fiber.New()

	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/health", handler.HealthCheck)

	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	return app
}
