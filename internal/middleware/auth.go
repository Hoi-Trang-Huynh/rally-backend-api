package middleware

import (
	"context"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/gofiber/fiber/v2"
)

// AuthRequired extracts the Bearer token from the Authorization header and
// verifies it with Firebase. On success, the verified *auth.Token is stored
// in c.Locals("authToken") for downstream handlers.
//
// This is the single place where ID tokens are verified — downstream
// middleware and services work with the verified token or resolved user only.
func AuthRequired(firebaseAuth *auth.Client) fiber.Handler {
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

		idToken := strings.TrimSpace(authHeader[7:])
		if idToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Authorization token is required",
			})
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		token, err := firebaseAuth.VerifyIDToken(ctx, idToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Invalid or expired token",
			})
		}

		c.Locals("authToken", token)
		return c.Next()
	}
}

// firebaseUserInfoFromToken maps the standard Firebase token claims to a
// model.FirebaseUserInfo used for just-in-time user provisioning.
func firebaseUserInfoFromToken(token *auth.Token) model.FirebaseUserInfo {
	info := model.FirebaseUserInfo{UID: token.UID}

	if email, ok := token.Claims["email"].(string); ok {
		info.Email = email
	}
	if verified, ok := token.Claims["email_verified"].(bool); ok {
		info.EmailVerified = verified
	}
	if name, ok := token.Claims["name"].(string); ok {
		info.DisplayName = name
	}
	if picture, ok := token.Claims["picture"].(string); ok {
		info.PictureURL = picture
	}

	return info
}
