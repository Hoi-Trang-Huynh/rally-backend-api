package router

import (
	"fmt"
	_ "github.com/Hoi-Trang-Huynh/rally-backend-api/api/docs" // docs package
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/handler"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/middleware"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/firebase"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/database"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func Setup(cfg *config.Config) *fiber.App {
    db := database.GetDB()
    if db == nil {
        panic("failed to get database connection")
    }

	userRepo := repository.NewUserRepository(db)
    
    fbClient := firebase.GetClient()
    if fbClient == nil {
        return nil, fmt.Errorf("failed to get firebase client")
    }
    
    return SetupWithDeps(userRepo, fbClient)
}

func SetupWithDeps(
    userRepo repository.UserRepository,
    fbClient *firebase.App,
) (*fiber.App, error) {
    app := fiber.New()

    app.Use(middleware.Logger())
    app.Use(middleware.CORS())

    authService, err := service.NewAuthService(fbClient, userRepo)
    if err != nil {
        return nil, fmt.Errorf("failed to create auth service: %w", err)
    }

    authHandler := handler.NewAuthHandler(authService)

    // Swagger route
    app.Get("/swagger/*", fiberSwagger.WrapHandler)

    // Routes
    api := app.Group("/api")
    v1 := api.Group("/v1")

    v1.Get("/health", handler.HealthCheck)

    // Auth routes
    auth := v1.Group("/auth")
    auth.Post("/register", authHandler.RegisterOrLogin)

    return app, nil
}
