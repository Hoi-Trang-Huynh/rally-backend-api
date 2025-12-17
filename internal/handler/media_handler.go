package handler

import (
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	uploader *utils.CloudinaryUploader
}

func NewMediaHandler(uploader *utils.CloudinaryUploader) *MediaHandler {
	return &MediaHandler{
		uploader: uploader,
	}
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
// @Failure 400 {object} map[string]string
// @Router /media/sign [post]
func (h *MediaHandler) GetUploadSignature(c *fiber.Ctx) error {
	var req UploadSignatureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Construct strict parameters map
	timestamp := time.Now().Unix()
	params := map[string]interface{}{
		"timestamp":     timestamp,
		"max_file_size": 10 * 1024 * 1024,
	}

	if req.Folder != "" {
		params["folder"] = req.Folder
	}

	signature, err := h.uploader.GenerateUploadSignature(params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate signature",
		})
	}

	return c.JSON(fiber.Map{
		"signature":  signature,
		"timestamp":  timestamp,
		"api_key":    h.uploader.GetAPIKey(),
		"cloud_name": h.uploader.GetCloudName(),
	})
}
