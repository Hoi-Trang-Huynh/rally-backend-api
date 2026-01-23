package handler

import (
	"strconv"
	"strings"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
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
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/feedback [post]
func (h *FeedbackHandler) CreateFeedback(c *fiber.Ctx) error {
	var req model.CreateFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Basic validation
	if req.Username == "" || req.Comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and Comment are required",
		})
	}

	if len(req.AttachmentURLs) > 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 3 attachments allowed",
		})
	}

	feedback, err := h.service.SubmitFeedback(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create feedback",
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
// @Param page_size query int false "Items per page (default: 20)"
// @Param username query string false "Filter by username"
// @Param categories query string false "Filter by categories (comma-separated)"
// @Success 200 {object} model.FeedbackListResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/feedback [get]
func (h *FeedbackHandler) GetFeedbackList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	username := c.Query("username")

	var categories []string
	catsQuery := c.Query("categories")
	if catsQuery != "" {
		categories = strings.Split(catsQuery, ",")
	}

	response, err := h.service.ListFeedbacks(c.Context(), page, pageSize, username, categories)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch feedbacks",
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
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/feedback/{id}/resolve [patch]
func (h *FeedbackHandler) UpdateFeedbackStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Feedback ID is required",
		})
	}

	var req model.UpdateFeedbackStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.service.ResolveFeedback(c.Context(), id, req.Resolved)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update feedback status",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Feedback status updated successfully",
		"resolved": req.Resolved,
	})
}
