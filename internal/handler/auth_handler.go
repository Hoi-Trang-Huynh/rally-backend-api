package handler

import (
    "github.com/gofiber/fiber/v2"
    // "github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
)

// @Summary Register or log in with email
// @Description Register a new user if email doesn't exist, or log in if it already exists
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.RegisterEmailRequest true "User registration payload"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /api/v1/auth/register/email [post]
func RegisterEmail(c *fiber.Ctx) error {
    // Placeholder logic
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "Registration successful (email)",
    })
}

// RegisterOAuth godoc
// @Summary Register or login using OAuth providers
// @Description Registers a user through Google, Facebook, or other OAuth providers using Firebase.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.RegisterOAuthRequest true "OAuth registration payload"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse "Invalid token or provider"
// @Router /api/v1/auth/register/oauth [post]
func RegisterOAuth(c *fiber.Ctx) error {
    // Placeholder logic
    return c.JSON(fiber.Map{
        "message": "Registration successful (OAuth)",
    })
}
