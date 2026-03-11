package handler

import (
	"context"
	"strings"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type FeedbackHandler struct {
	service *service.FeedbackService
}

func NewFeedbackHandler(service *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{service: service}
}

// CreateFeedback creates a new feedback entry
// @Summary Create a new feedback
// @Description Submit user feedback with optional attachments (max 3) and categories
// @Tags Feedback
// @Accept json
// @Produce json
// @Param request body model.CreateFeedbackRequest true "Feedback content"
// @Success 201 {object} model.Feedback
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/feedback [post]
func (h *FeedbackHandler) CreateFeedback(c *fiber.Ctx) error {
	var req model.CreateFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	if req.Username == "" || req.Comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Username and Comment are required",
		})
	}

	if len(req.AttachmentURLs) > 3 {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Maximum 3 attachments allowed",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feedback, err := h.service.SubmitFeedback(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to create feedback",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(feedback)
}

// GetFeedbackList retrieves a list of feedbacks
// @Summary List feedbacks
// @Description Get paginated list of feedbacks with optional filtering by username and categories
// @Tags Feedback
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Items per page (default: 20)"
// @Param username query string false "Filter by username"
// @Param categories query string false "Filter by categories (comma-separated)"
// @Success 200 {object} model.FeedbackListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/feedback [get]
func (h *FeedbackHandler) GetFeedbackList(c *fiber.Ctx) error {
	page, pageSize := utils.ClampPagination(c.QueryInt("page", 1), c.QueryInt("pageSize", 20), 50)
	username := c.Query("username")

	var categories []string
	catsQuery := c.Query("categories")
	if catsQuery != "" {
		categories = strings.Split(catsQuery, ",")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.service.ListFeedbacks(ctx, page, pageSize, username, categories)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to fetch feedbacks",
		})
	}

	return c.JSON(response)
}

// UpdateFeedbackStatus updates the resolved status of a feedback
// @Summary Update feedback status
// @Description Mark a feedback item as resolved or unresolved
// @Tags Feedback
// @Accept json
// @Produce json
// @Param id path string true "Feedback ID"
// @Param request body model.UpdateFeedbackStatusRequest true "Status update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/feedback/{id}/resolve [patch]
func (h *FeedbackHandler) UpdateFeedbackStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Feedback ID is required",
		})
	}

	var req model.UpdateFeedbackStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := h.service.ResolveFeedback(ctx, id, req.Resolved)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to update feedback status",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Feedback status updated successfully",
		"resolved": req.Resolved,
	})
}
