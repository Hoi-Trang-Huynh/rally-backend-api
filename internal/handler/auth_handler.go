package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Hoi-Trang-Huynh/rally-backend-api-2/internal/model"
)

// Login godoc
// @Summary Login a user via Firebase
// @Description Accepts a Firebase ID token and returns user info
// @Tags Authentication
// @ID Login
// @Accept json
// @Produce json
// @Param request body model.FirebaseAuthRequest true "Firebase authentication payload"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} model.ErrorResponse "Invalid or expired token"
// @Router /auth/register [post]
func Login(c *fiber.Ctx) error {
	var req model.FirebaseAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.LoginResponse{
		Message: "User loged in successfully",
		User: &model.UserResponse{
			UserID: "mock-uuid-1234",
			Email:  "user@example.com",
		},
	})
}

