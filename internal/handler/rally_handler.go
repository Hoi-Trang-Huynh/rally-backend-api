package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
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
// @Description Create a new rally. The authenticated user becomes the owner. Description supports rich text JSON.
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
	user := c.Locals("user").(*model.User)

	var req model.CreateRallyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload: " + err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Rally name is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.rallyService.CreateRally(ctx, user, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "Failed to create rally",
		})
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

	var req model.UpdateRallyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.rallyService.UpdateRally(ctx, rallyID, &req)
	if err != nil {
		switch err.Error() {
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

// GetRalliesList godoc
// @Summary Get user's rallies list
// @Description Get a paginated, filtered and sorted list of rallies where the specified user is a participant (with joined status). Returns only essential fields for list views.
// @Tags User
// @ID getUserRalliesList
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Param name query string false "Filter by rally name (case-insensitive partial match)"
// @Param status query string false "Filter by status (draft, active, inactive, completed, archived)"
// @Param sort query string false "Sort order for start date (asc or desc)" default(asc)
// @Param page query int false "Page number (starts from 1)" default(1)
// @Param pageSize query int false "Number of items per page" default(20)
// @Success 200 {object} model.RalliesListResponse
// @Failure 400 {object} model.ErrorResponse "Invalid request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/{id}/rallies [get]
func (h *RallyHandler) GetRalliesList(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "User ID is required",
		})
	}

	nameFilter := c.Query("name", "")
	statusFilter := c.Query("status", "")
	sortOrder := c.Query("sort", "asc")

	page, pageSize := utils.ClampPagination(c.QueryInt("page", 1), c.QueryInt("pageSize", 20), 50)

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	if statusFilter != "" {
		validStatuses := map[string]bool{
			"draft": true, "active": true, "inactive": true,
			"completed": true, "archived": true,
		}
		if !validStatuses[statusFilter] {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: "Invalid status filter. Must be one of: draft, active, inactive, completed, archived",
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.rallyService.GetRalliesList(ctx, userID, nameFilter, statusFilter, sortOrder, page, pageSize)
	if err != nil {
		switch err.Error() {
		case "user not found", "invalid user ID":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to get rallies list",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetRally godoc
// @Summary Get rally details
// @Description Get detailed information about a rally. Requires user to be a joined participant.
// @Tags Rally
// @ID getRally
// @Accept json
// @Produce json
// @Param id path string true "Rally ID"
// @Param Authorization header string true "Bearer Firebase ID Token"
// @Success 200 {object} model.RallyJoinResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 403 {object} model.ErrorResponse "Forbidden"
// @Failure 404 {object} model.ErrorResponse "Rally not found"
// @Router /rallies/{id} [get]
func (h *RallyHandler) GetRally(c *fiber.Ctx) error {
	rallyID := c.Params("id")
	participant := c.Locals("rallyParticipant").(*model.RallyParticipant)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := h.rallyService.GetRally(ctx, participant, rallyID)
	if err != nil {
		switch err.Error() {
		case "unauthorized: you must be joined or invited to view this rally":
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		case "rally not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to get rally details",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
