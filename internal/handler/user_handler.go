package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile information for a specific user
// @Tags User Profile
// @ID getUserProfile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.ProfileResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /users/{id}/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user profile
	user, err := h.userService.GetUserProfile(ctx, userID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to get user profile",
		})
	}

	// Convert to response format
	profile := h.userService.ConvertToProfileResponse(user)
	return c.Status(fiber.StatusOK).JSON(profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update profile information for the authenticated user
// @Tags User Profile
// @ID updateUserProfile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.ProfileUpdateRequest true "Profile update payload"
// @Success 200 {object} model.ProfileResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request or user ID"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /users/{id}/profile [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
		})
	}

	// Get Firebase token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Authorization header is required",
		})
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Invalid authorization format. Use 'Bearer <token>'",
		})
	}
	idToken := authHeader[7:]

	// Parse request body
	var req model.ProfileUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate user ownership
	if err := h.userService.ValidateUserOwnership(ctx, idToken, userID); err != nil {
		if err.Error() == "unauthorized: cannot modify another user's profile" {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Update user profile
	updatedUser, err := h.userService.UpdateUserProfile(ctx, userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to update user profile",
		})
	}

	// Convert to response format
	profile := h.userService.ConvertToProfileResponse(updatedUser)
	return c.Status(fiber.StatusOK).JSON(profile)
}

// GetMyProfile godoc
// @Summary Get current user's profile
// @Description Get the profile information for the currently authenticated user
// @Tags User Profile
// @ID getMyProfile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.ProfileResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /users/me/profile [get]
func (h *UserHandler) GetMyProfile(c *fiber.Ctx) error {
	// Get Firebase token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Authorization header is required",
		})
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Invalid authorization format. Use 'Bearer <token>'",
		})
	}
	idToken := authHeader[7:]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user by Firebase token
	user, err := h.userService.GetUserProfileByToken(ctx, idToken)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Convert to response format
	profile := h.userService.ConvertToProfileResponse(user)
	return c.Status(fiber.StatusOK).JSON(profile)
}

// GetMyProfileDetails godoc
// @Summary Get current user's profile details
// @Description Get detailed profile information for the profile page view (bio, etc.)
// @Tags User Profile
// @ID getMyProfileDetails
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.ProfileDetailsResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /user/me/profile/details [get]
func (h *UserHandler) GetMyProfileDetails(c *fiber.Ctx) error {
	// Get Firebase token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Authorization header is required",
		})
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Invalid authorization format. Use 'Bearer <token>'",
		})
	}
	idToken := authHeader[7:]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user by Firebase token
	user, err := h.userService.GetUserProfileByToken(ctx, idToken)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Convert to profile details response format
	details := h.userService.ConvertToProfileDetailsResponse(user)
	return c.Status(fiber.StatusOK).JSON(details)
}

// SearchUsers godoc
// @Summary Search users
// @Description Search for users by username, first name, or last name with pagination
// @Tags User Search
// @ID searchUsers
// @Accept json
// @Produce json
// @Param q query string true "Search query (matches username, first name, or last name)"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Number of results per page (default: 20, max: 50)"
// @Success 200 {object} model.UserSearchResponse
// @Failure 400 {object} model.ErrorResponse "Invalid query parameters"
// @Router /user/search [get]
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	// Get search query
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Search query 'q' is required",
		})
	}

	// Get pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Search users
	response, err := h.userService.SearchUsers(ctx, query, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to search users",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
