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
	user := c.Locals("user").(*model.User)

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

	response, err := h.participantService.InviteParticipant(ctx, user, rallyID, &req)
	if err != nil {
		switch err.Error() {
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
	participantID := c.Params("participantId")
	if participantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Participant ID is required",
		})
	}

	user := c.Locals("user").(*model.User)
	callerParticipant := c.Locals("rallyParticipant").(*model.RallyParticipant)

	var req model.UpdateParticipantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.participantService.UpdateParticipant(ctx, user, callerParticipant, rallyID, participantID, &req)
	if err != nil {
		switch err.Error() {
		case "unauthorized: only owners can change roles", "unauthorized: insufficient permissions", "unauthorized: participant status is not active":
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

// GetParticipantsList godoc
// @Summary Get participants of a rally
// @Description Get a paginated list of participants in a rally. Requires user to be a joined participant.
// @Tags Rally Participants
// @ID getRallyParticipantsList
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param role query string false "Filter by role (owner, editor, participant)"
// @Param page query int false "Page number (starts from 1)" default(1)
// @Param pageSize query int false "Number of items per page" default(20)
// @Success 200 {object} model.ParticipantListResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Rally not found"
// @Router /rallies/{id}/participants [get]
func (h *RallyParticipantHandler) GetParticipantsList(c *fiber.Ctx) error {
	rallyID := c.Params("id")

	roleFilter := c.Query("role", "")
	if roleFilter != "" {
		validRoles := map[string]bool{
			string(model.ParticipantRoleOwner):       true,
			string(model.ParticipantRoleEditor):      true,
			string(model.ParticipantRoleParticipant): true,
		}
		if !validRoles[roleFilter] {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: "Invalid role filter. Must be one of: owner, editor, participant",
			})
		}
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.participantService.GetParticipantsList(ctx, rallyID, roleFilter, page, pageSize)
	if err != nil {
		switch err.Error() {
		case "invalid rally ID":
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to get participants list",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetInvitableFriends godoc
// @Summary Get friends who can be invited to a rally
// @Description Get a paginated list of the authenticated user's friends who are NOT already participants (any status) in the given rally. Supports search by name/username.
// @Tags Rally Participants
// @ID getInvitableFriends
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param q query string false "Search query (matches username, first name, or last name)"
// @Param page query int false "Page number (starts from 1)" default(1)
// @Param pageSize query int false "Number of items per page" default(20)
// @Success 200 {object} model.FriendListResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Rally not found"
// @Router /rallies/{id}/invitable-friends [get]
func (h *RallyParticipantHandler) GetInvitableFriends(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	user := c.Locals("user").(*model.User)

	query := c.Query("q", "")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.participantService.GetInvitableFriends(ctx, user, rallyID, query, page, pageSize)
	if err != nil {
		switch err.Error() {
		case "rally not found", "invalid rally ID":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to get invitable friends",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
