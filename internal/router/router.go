package router


import (
	_ "github.com/Hoi-Trang-Huynh/rally-backend-api/api/docs" // docs package
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/handler"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/middleware"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/firebase"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/database"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func Setup(cfg *config.Config) *fiber.App {
	app := fiber.New()

	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	// initialize dependencies
	db := database.GetDB()
	userRepo := repository.NewUserRepository(db)
	fbClient := firebase.GetClient()
	authService := service.NewAuthService(fbClient, userRepo)
	authHandler := handler.NewAuthHandler(authService)
	
	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	
	// Routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/health", handler.HealthCheck)

	// Auth routes
	auth := v1.Group("/auth")
	auth.Post("/register", handler.RegisterOrLogin)
	return app
}
