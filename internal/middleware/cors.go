package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*", // Allow all origins (change in prod)
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Authorization",
		AllowCredentials: true,
	})
}

// Note: In production, we should restrict AllowOrigins to specific domains for security reasons.
// The above configuration allows all origins, which is suitable for development but should be limited in production environments.
