package handler

import (
	"context"
	"log"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type InviteLinkHandler struct {
	inviteLinkService *service.InviteLinkService
}

func NewInviteLinkHandler(inviteLinkService *service.InviteLinkService) *InviteLinkHandler {
	return &InviteLinkHandler{
		inviteLinkService: inviteLinkService,
	}
}

// CreateInviteLink godoc
// @Summary Create an invite link/QR token
// @Description Creates a new invite link token for a rally. The creator must be the rally owner or editor.
// @Tags Invite Links
// @ID createInviteLink
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.CreateInviteLinkRequest true "Invite Link Details"
// @Success 201 {object} model.InviteLinkResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request body or parameters"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Router /rallies/{id}/invite-links [post]
func (h *InviteLinkHandler) CreateInviteLink(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	if rallyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally ID is required",
		})
	}

	user := c.Locals("user").(*model.User)
	callerParticipant := c.Locals("rallyParticipant").(*model.RallyParticipant)

	var req model.CreateInviteLinkRequest
	// All fields are optional, so allow empty body
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: "Invalid request body: " + err.Error(),
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.inviteLinkService.CreateInviteLink(ctx, user, callerParticipant, rallyID, &req)
	if err != nil {
		switch err.Error() {
		case "only owners can create links for owner/editor roles":
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to create invite link",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetActiveInviteLinks godoc
// @Summary List active invite links
// @Description Retrieves all active invite links for a rally. Must be an owner or editor.
// @Tags Invite Links
// @ID getActiveInviteLinks
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {array} model.InviteLinkResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Router /rallies/{id}/invite-links [get]
func (h *InviteLinkHandler) GetActiveInviteLinks(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	if rallyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.inviteLinkService.GetActiveInviteLinks(ctx, rallyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to list invite links",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// DeactivateInviteLink godoc
// @Summary Revoke an invite link
// @Description Deactivates an invite link token so it can no longer be used.
// @Tags Invite Links
// @ID deactivateInviteLink
// @Produce json
// @Param id path string true "Rally ID"
// @Param token path string true "Invite Link Token"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 204 "No Content"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Not Found"
// @Router /rallies/{id}/invite-links/{token} [delete]
func (h *InviteLinkHandler) DeactivateInviteLink(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	token := c.Params("token")
	if rallyID == "" || token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally ID and Token are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := h.inviteLinkService.DeactivateInviteLink(ctx, rallyID, token)
	if err != nil {
		switch err.Error() {
		case "link not found or already inactive", "link does not belong to this rally":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to deactivate invite link",
			})
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// JoinViaLink godoc
// @Summary Accept an invite link and join a rally
// @Description Accepts a QR code / invite link token and immediately joins the rally with "joined" status. Takes the higher role when an existing in-app invitation exists.
// @Tags Invite Links
// @ID joinViaLink
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.JoinViaLinkRequest true "Token Details"
// @Success 200 {object} model.JoinViaLinkResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request or link expired/used up"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Router /rallies/join-via-link [post]
func (h *InviteLinkHandler) JoinViaLink(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	var req model.JoinViaLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Token is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.inviteLinkService.JoinViaLink(ctx, user, req.Token)
	if err != nil {
		switch err.Error() {
		case "invalid or expired invite link", "invite link has expired", "invite link has reached its maximum number of uses":
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to join via link",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// PreviewInviteLink godoc
// @Summary Preview an invite link
// @Description Get details about an invite link for a preview card (rally info, owner info, role offered). Requires authentication.
// @Tags Invite Links
// @ID previewInviteLink
// @Accept json
// @Produce json
// @Param token path string true "Invite link token (UUID)"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.InviteLinkPreviewResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Link or rally not found"
// @Failure 400 {object} model.ErrorResponse "Link expired or reached limit"
// @Router /rallies/invite-links/{token}/preview [get]
func (h *InviteLinkHandler) PreviewInviteLink(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Token is required",
		})
	}

	user := c.Locals("user").(*model.User)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.inviteLinkService.PreviewInviteLink(ctx, user, token)
	if err != nil {
		log.Printf("[PreviewInviteLink] ERROR for token %s: %v\n", token, err)

		switch err.Error() {
		case "link is invalid or inactive", "rally not found or inactive", "rally owner not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "link is expired", "link has reached its maximum uses":
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to preview invite link",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
