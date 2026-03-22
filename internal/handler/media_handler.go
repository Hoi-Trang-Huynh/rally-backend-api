package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	uploader    *utils.CloudinaryUploader
	userService *service.UserService
}

func NewMediaHandler(uploader *utils.CloudinaryUploader, userService *service.UserService) *MediaHandler {
	return &MediaHandler{
		uploader:    uploader,
		userService: userService,
	}
}

type VerifyAvatarRequest struct {
	PublicID  string `json:"public_id"`
	AvatarUrl string `json:"avatar_url"`
}

// VerifyAvatar godoc
// @Summary Verify and update user avatar
// @Description Verify uploaded avatar and update user profile. Deletes image if update fails.
// @Tags Media
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body VerifyAvatarRequest true "Avatar details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /media/verify-avatar [post]
func (h *MediaHandler) VerifyAvatar(c *fiber.Ctx) error {
	idToken := c.Locals("idToken").(string)

	var req VerifyAvatarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	if req.PublicID == "" || req.AvatarUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "public_id and avatar_url are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.userService.GetUserProfileByToken(ctx, idToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Invalid token or user not found",
		})
	}

	updateReq := &model.ProfileUpdateRequest{
		AvatarUrl: &req.AvatarUrl,
	}

	_, err = h.userService.UpdateUserProfile(ctx, user.ID.Hex(), updateReq)
	if err != nil {
		delCtx, delCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer delCancel()

		_ = h.uploader.DeleteImage(delCtx, req.PublicID)

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Avatar updated successfully",
		"url":     req.AvatarUrl,
	})
}

// UploadSignatureRequest defines the allowed parameters for signature generation
type UploadSignatureRequest struct {
	Folder string `json:"folder"`
	UserID string `json:"user_id"`
}

// GetUploadSignature godoc
// @Summary Get Cloudinary upload signature
// @Description Generate a signature for uploading media to Cloudinary
// @Tags Media
// @Accept json
// @Produce json
// @Param request body UploadSignatureRequest true "Upload parameters"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} model.ErrorResponse
// @Router /media/sign [post]
func (h *MediaHandler) GetUploadSignature(c *fiber.Ctx) error {
	var req UploadSignatureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	timestamp := time.Now().Unix()
	params := map[string]interface{}{
		"timestamp": timestamp,
	}

	if req.Folder != "" {
		params["folder"] = req.Folder
	}

	if req.UserID != "" {
		params["public_id"] = req.UserID
	}

	signature, err := h.uploader.GenerateUploadSignature(params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to generate signature",
		})
	}

	return c.JSON(fiber.Map{
		"signature":  signature,
		"timestamp":  timestamp,
		"api_key":    h.uploader.GetAPIKey(),
		"cloud_name": h.uploader.GetCloudName(),
		"public_id":  req.UserID,
	})
}
