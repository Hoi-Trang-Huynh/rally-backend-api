package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// CreateEvent godoc
// @Summary Create a new event in a rally
// @Description Create a new event within a rally. Requires owner or editor role.
// @Tags Event
// @ID createEvent
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.CreateEventRequest true "Event creation payload"
// @Success 201 {object} model.EventResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Rally not found"
// @Router /rallies/{id}/events [post]
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	user := c.Locals("user").(*model.User)

	var req model.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Event name is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.eventService.CreateEvent(ctx, user, rallyID, &req)
	if err != nil {
		switch err.Error() {
		case "rally not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to create event",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// UpdateEvent godoc
// @Summary Update an event
// @Description Update event details. Requires owner or editor role in the event's rally.
// @Tags Event
// @ID updateEvent
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.UpdateEventRequest true "Event update payload"
// @Success 200 {object} model.EventResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Event not found"
// @Router /events/{id} [put]
func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Event ID is required",
		})
	}

	user := c.Locals("user").(*model.User)

	var req model.UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.eventService.UpdateEvent(ctx, user, eventID, &req)
	if err != nil {
		switch err.Error() {
		case "unauthorized: insufficient permissions", "unauthorized: not a participant of this rally", "unauthorized: participant status is not active (must be joined)":
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "event not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to update event",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
