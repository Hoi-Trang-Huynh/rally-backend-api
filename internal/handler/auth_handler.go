package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Helper function to convert User to UserResponse
func convertToUserResponse(user *model.User) *model.UserResponse {
	return &model.UserResponse{
		ID:              user.ID.Hex(),
		Email:           user.Email,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		AvatarUrl:       user.AvatarUrl,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		IsOnboarding:    user.IsOnboarding,
	}
}

// Register godoc
// @Summary Register or login a user via Firebase
// @Description Idempotent: the user is provisioned from the verified Firebase token in the Authorization header; the optional body sets initial profile fields in the same request.
// @Tags Authentication
// @ID registerOrLogin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.RegisterRequest false "Optional initial profile fields"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request payload"
// @Failure 401 {object} model.ErrorResponse "Invalid or expired token"
// @Failure 409 {object} model.ErrorResponse "Username is already taken"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// The body is optional; legacy clients may post other fields (ignored).
	var req model.RegisterRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: "Invalid request payload",
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.authService.CompleteRegistration(ctx, user, &req)
	if err != nil {
		switch err.Error() {
		case "username is already taken":
			return c.Status(fiber.StatusConflict).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to register or login",
			})
		}
	}

	message := "User logged in successfully"
	if user.IsOnboarding {
		message = "User registered successfully"
	}

	return c.Status(fiber.StatusOK).JSON(model.RegisterResponse{
		Message: message,
		User:    convertToUserResponse(user),
	})
}

// CheckEmailAvailability godoc
// @Summary Check if email is available
// @Description Check if an email address is already registered
// @Tags Authentication
// @ID checkEmailAvailability
// @Produce json
// @Param email query string true "Email to check"
// @Success 200 {object} model.AvailabilityResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /auth/check-email [get]
func (h *AuthHandler) CheckEmailAvailability(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Email query parameter is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	available, err := h.authService.CheckEmailAvailability(ctx, email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to check email availability",
		})
	}

	message := "Email is available"
	if !available {
		message = "Email is already registered"
	}

	return c.Status(fiber.StatusOK).JSON(model.AvailabilityResponse{
		Available: available,
		Message:   message,
	})
}

// CheckUsernameAvailability godoc
// @Summary Check if username is available
// @Description Check if a username is already taken
// @Tags Authentication
// @ID checkUsernameAvailability
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} model.AvailabilityResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /auth/check-username [get]
func (h *AuthHandler) CheckUsernameAvailability(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Username query parameter is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	available, err := h.authService.CheckUsernameAvailability(ctx, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to check username availability",
		})
	}

	message := "Username is available"
	if !available {
		message = "Username is already taken"
	}

	return c.Status(fiber.StatusOK).JSON(model.AvailabilityResponse{
		Available: available,
		Message:   message,
	})
}
