package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type RallyHandler struct {
	rallyService *service.RallyService
}

func NewRallyHandler(rallyService *service.RallyService) *RallyHandler {
	return &RallyHandler{
		rallyService: rallyService,
	}
}

// CreateRally godoc
// @Summary Create a new rally
// @Description Create a new rally. The authenticated user becomes the owner. Can optionally invite participants.
// @Tags Rally
// @ID createRally
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.CreateRallyRequest true "Rally creation payload"
// @Success 201 {object} model.RallyResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request payload"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Router /rallies [post]
func (h *RallyHandler) CreateRally(c *fiber.Ctx) error {
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

	var req model.CreateRallyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally name is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.rallyService.CreateRally(ctx, idToken, &req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired token":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "user not found":
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to create rally",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// UpdateRally godoc
// @Summary Update a rally
// @Description Update rally details. Requires owner or editor role.
// @Tags Rally
// @ID updateRally
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param request body model.UpdateRallyRequest true "Rally update payload"
// @Success 200 {object} model.RallyResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Rally not found"
// @Router /rallies/{id} [put]
func (h *RallyHandler) UpdateRally(c *fiber.Ctx) error {
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

	var req model.UpdateRallyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.rallyService.UpdateRally(ctx, idToken, rallyID, &req)
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
		case "rally not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to update rally",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
