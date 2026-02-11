package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// CreateActivity godoc
// @Summary Create a new activity in an event
// @Description Create a new activity within an event. Requires owner or editor role in the event's rally.
// @Tags Activity
// @ID createActivity
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.CreateActivityRequest true "Activity creation payload"
// @Success 201 {object} model.ActivityResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Event not found"
// @Router /events/{id}/activities [post]
func (h *ActivityHandler) CreateActivity(c *fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Event ID is required",
		})
	}

	idToken := c.Locals("idToken").(string)

	var req model.CreateActivityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Activity name is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.activityService.CreateActivity(ctx, idToken, eventID, &req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired token", "user not found":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "unauthorized: insufficient permissions", "unauthorized: not a participant of this rally":
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "event not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to create activity",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// UpdateActivity godoc
// @Summary Update an activity
// @Description Update activity details. Requires owner or editor role in the activity's rally.
// @Tags Activity
// @ID updateActivity
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.UpdateActivityRequest true "Activity update payload"
// @Success 200 {object} model.ActivityResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Activity not found"
// @Router /activities/{id} [put]
func (h *ActivityHandler) UpdateActivity(c *fiber.Ctx) error {
	activityID := c.Params("id")
	if activityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Activity ID is required",
		})
	}

	idToken := c.Locals("idToken").(string)

	var req model.UpdateActivityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.activityService.UpdateActivity(ctx, idToken, activityID, &req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired token", "user not found":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "unauthorized: insufficient permissions", "unauthorized: not a participant of this rally":
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "activity not found", "event not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to update activity",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
