package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type RallyParticipantHandler struct {
	participantService *service.RallyParticipantService
}

func NewRallyParticipantHandler(participantService *service.RallyParticipantService) *RallyParticipantHandler {
	return &RallyParticipantHandler{
		participantService: participantService,
	}
}

// InviteParticipant godoc
// @Summary Invite a user to a rally
// @Description Invite a user to join a rally. Requires owner or editor role.
// @Tags Rally Participants
// @ID inviteParticipant
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.InviteParticipantRequest true "Invite payload"
// @Success 201 {object} model.RallyParticipantResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request or user already a participant"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Rally or user not found"
// @Router /rallies/{id}/participants [post]
func (h *RallyParticipantHandler) InviteParticipant(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	if rallyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally ID is required",
		})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Authorization header is required",
		})
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Invalid authorization format. Use 'Bearer <token>'",
		})
	}
	idToken := authHeader[7:]

	var req model.InviteParticipantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.participantService.InviteParticipant(ctx, idToken, rallyID, &req)
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
		case "rally not found", "target user not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "user is already a participant":
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to invite participant",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// UpdateParticipant godoc
// @Summary Update a participant's role or status
// @Description Update participant details. Role changes require owner. Status changes allowed for the participant themselves.
// @Tags Rally Participants
// @ID updateParticipant
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param participantId path string true "Participant ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.UpdateParticipantRequest true "Participant update payload"
// @Success 200 {object} model.RallyParticipantResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Participant not found"
// @Router /rallies/{id}/participants/{participantId} [put]
func (h *RallyParticipantHandler) UpdateParticipant(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	if rallyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally ID is required",
		})
	}

	participantID := c.Params("participantId")
	if participantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Participant ID is required",
		})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Authorization header is required",
		})
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "Invalid authorization format. Use 'Bearer <token>'",
		})
	}
	idToken := authHeader[7:]

	var req model.UpdateParticipantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.participantService.UpdateParticipant(ctx, idToken, rallyID, participantID, &req)
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
		case "participant not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to update participant",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
