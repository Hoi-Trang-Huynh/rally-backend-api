package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type FollowHandler struct {
	followService *service.FollowService
}

func NewFollowHandler(followService *service.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// FollowUser godoc
// @Summary Follow a user
// @Description Follow another user by their ID
// @Tags Follow
// @ID followUser
// @Accept json
// @Produce json
// @Param id path string true "User ID to follow"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.FollowResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID or cannot follow yourself"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /user/{id}/follow [post]
func (h *FollowHandler) FollowUser(c *fiber.Ctx) error {
	targetUserID := c.Params("id")
	if targetUserID == "" {
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Follow the user
	response, err := h.followService.FollowUser(ctx, idToken, targetUserID)
	if err != nil {
		switch err.Error() {
		case "invalid or expired token":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "cannot follow yourself":
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "target user not found", "current user not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to follow user",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// UnfollowUser godoc
// @Summary Unfollow a user
// @Description Unfollow a user by their ID
// @Tags Follow
// @ID unfollowUser
// @Accept json
// @Produce json
// @Param id path string true "User ID to unfollow"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.FollowResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /user/{id}/follow [delete]
func (h *FollowHandler) UnfollowUser(c *fiber.Ctx) error {
	targetUserID := c.Params("id")
	if targetUserID == "" {
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Unfollow the user
	response, err := h.followService.UnfollowUser(ctx, idToken, targetUserID)
	if err != nil {
		switch err.Error() {
		case "invalid or expired token":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "current user not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to unfollow user",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetFollowStatus godoc
// @Summary Check follow status
// @Description Check if the authenticated user follows the target user
// @Tags Follow
// @ID getFollowStatus
// @Accept json
// @Produce json
// @Param id path string true "User ID to check"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.FollowStatusResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Router /user/{id}/follow/status [get]
func (h *FollowHandler) GetFollowStatus(c *fiber.Ctx) error {
	targetUserID := c.Params("id")
	if targetUserID == "" {
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check follow status
	response, err := h.followService.IsFollowing(ctx, idToken, targetUserID)
	if err != nil {
		switch err.Error() {
		case "invalid or expired token":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "current user not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to check follow status",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetUserPublicProfile godoc
// @Summary Get user public profile
// @Description Get the public profile of a user including follow counts
// @Tags User Profile
// @ID getUserPublicProfile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.UserPublicProfileResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /user/{id}/profile [get]
func (h *FollowHandler) GetUserPublicProfile(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get public profile
	response, err := h.followService.GetUserPublicProfile(ctx, userID)
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

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetFollowersList godoc
// @Summary Get user's followers list
// @Description Get a paginated list of users who follow the specified user
// @Tags Follow
// @ID getFollowersList
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Number of results per page (default: 20, max: 50)"
// @Success 200 {object} model.FollowListResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID"
// @Router /user/{id}/followers [get]
func (h *FollowHandler) GetFollowersList(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
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

	// Get followers list
	response, err := h.followService.GetFollowersList(ctx, userID, page, pageSize)
	if err != nil {
		if err.Error() == "invalid user ID" {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to get followers list",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetFollowingList godoc
// @Summary Get user's following list
// @Description Get a paginated list of users that the specified user follows
// @Tags Follow
// @ID getFollowingList
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Number of results per page (default: 20, max: 50)"
// @Success 200 {object} model.FollowListResponse
// @Failure 400 {object} model.ErrorResponse "Invalid user ID"
// @Router /user/{id}/following [get]
func (h *FollowHandler) GetFollowingList(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
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

	// Get following list
	response, err := h.followService.GetFollowingList(ctx, userID, page, pageSize)
	if err != nil {
		if err.Error() == "invalid user ID" {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to get following list",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
