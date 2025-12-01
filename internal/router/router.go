package router

import (
	_ "github.com/Hoi-Trang-Huynh/rally-backend-api/api/docs"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/handler"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/database"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/firebase"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/middleware"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
    fiberSwagger "github.com/swaggo/fiber-swagger"
	fb "firebase.google.com/go/v4"
)

func Setup(cfg *config.Config) *fiber.App {
	db := database.GetDB()
	userRepo := repository.NewUserRepository(db)

	fbApp := firebase.GetClient()

	app, err := SetupWithDeps(userRepo, fbApp)
	if err != nil {
		panic(err)
	}

	return app
}

func SetupWithDeps(
	userRepo repository.UserRepository,
	fbApp *fb.App,
) (*fiber.App, error) {

	app := fiber.New()

	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	authService, err := service.NewAuthService(fbApp, userRepo)
	if err != nil {
		return nil, err
	}

	userService, err := service.NewUserService(fbApp, userRepo)
	if err != nil {
		return nil, err
	}

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/health", handler.HealthCheck)

	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.RegisterOrLogin)
	auth.Post("/login", authHandler.Login)

	users := v1.Group("/user")
	users.Get("/me/profile", userHandler.GetMyProfile)
	users.Get("/:id/profile", userHandler.GetProfile)
	users.Put("/:id/profile", userHandler.UpdateProfile)

	return app, nil
}
