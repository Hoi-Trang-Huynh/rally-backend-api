package middleware

import (
	"strings"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/gofiber/fiber/v2"
)

// AuthRequired is a middleware that extracts and validates the Bearer token
// from the Authorization header. On success, the token is stored in
// c.Locals("idToken") for downstream handlers.
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Authorization header is required",
			})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Invalid authorization format. Use 'Bearer <token>'",
			})
		}

		c.Locals("idToken", authHeader[7:])
		return c.Next()
	}
}
