package handler

import (
	"context"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterOrLogin godoc
// @Summary Register or login a user via Firebase
// @Description Accepts a Firebase ID token and returns user info (registers if new)
// @Tags Authentication
// @ID registerOrLogin
// @Accept json
// @Produce json
// @Param request body model.FirebaseAuthRequest true "Firebase authentication payload"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse "Invalid or expired token"
// @Router /auth/register [post]
// RegisterOrLogin godoc
// @Summary Register or login a user via Firebase
// @Description Accepts a Firebase ID token and returns user info (registers if new)
// @Tags Authentication
// @ID registerOrLogin
// @Accept json
// @Produce json
// @Param request body model.FirebaseAuthRequest true "Firebase authentication payload"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse "Invalid or expired token"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterOrLogin(c *fiber.Ctx) error {
	var req model.FirebaseAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	// Validate token is not empty
	if req.IDToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Firebase ID token is required",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify Firebase token and get/create user
	user, isNewUser, err := h.authService.RegisterOrLogin(ctx, req.IDToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
		})
	}

	message := "User logged in successfully"
	if isNewUser {
		message = "User registered successfully"
	}

	return c.Status(fiber.StatusOK).JSON(model.RegisterResponse{
		Message: message,
		User: &model.UserResponse{
			UserID: user.UserID,
			Email:  user.Email,
		},
		Onboarding: isNewUser,
	})
}
  
// Login godoc
// @Summary Login a user via Firebase
// @Description Accepts a Firebase ID token and returns user info
// @Tags Authentication
// @ID login
// @Accept json
// @Produce json
// @Param request body model.FirebaseAuthRequest true "Firebase authentication payload"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} model.ErrorResponse "Invalid or expired token"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.FirebaseAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	// Validate token
	if req.IDToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Firebase ID token is required",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify token
	user, err := h.authService.Login(ctx, req.IDToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.LoginResponse{
		Message: "User logged in successfully",
		User: &model.UserResponse{
			UserID: user.UserID,
			Email:  user.Email,
		},
	})
}