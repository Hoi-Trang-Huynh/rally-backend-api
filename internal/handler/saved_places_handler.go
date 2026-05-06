package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

// SavedPlacesHandler handles saved-places HTTP requests.
type SavedPlacesHandler struct {
	savedPlacesService *service.SavedPlacesService
}

// NewSavedPlacesHandler creates a SavedPlacesHandler.
func NewSavedPlacesHandler(savedPlacesService *service.SavedPlacesService) *SavedPlacesHandler {
	return &SavedPlacesHandler{savedPlacesService: savedPlacesService}
}

// GetSavedPlaces godoc
// @Summary Get saved places
// @Description Returns the authenticated user's bookmarked places
// @Tags SavedPlaces
// @ID getSavedPlaces
// @Produce json
// @Success 200 {object} model.SavedPlacesResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /saved-places [get]
func (h *SavedPlacesHandler) GetSavedPlaces(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "user not resolved",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	places, err := h.savedPlacesService.GetSavedPlaces(ctx, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "failed to fetch saved places",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.SavedPlacesResponse{Places: places})
}

// SavePlace godoc
// @Summary Save a place
// @Description Bookmarks a place by fetching its details from Google and persisting a snapshot
// @Tags SavedPlaces
// @ID savePlace
// @Accept json
// @Produce json
// @Param body body model.SavePlaceRequest true "Place to save"
// @Success 201 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /saved-places [post]
func (h *SavedPlacesHandler) SavePlace(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "user not resolved",
		})
	}

	var req model.SavePlaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid request body",
		})
	}
	if req.PlaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "placeId is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.savedPlacesService.SavePlace(ctx, user.ID, req.PlaceID); err != nil {
		switch err.Error() {
		case "place not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "place not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "failed to save place",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "place saved"})
}

// RemovePlace godoc
// @Summary Remove a saved place
// @Description Removes a bookmarked place for the authenticated user
// @Tags SavedPlaces
// @ID removePlace
// @Produce json
// @Param placeId path string true "Google Place ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /saved-places/{placeId} [delete]
func (h *SavedPlacesHandler) RemovePlace(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Message: "user not resolved",
		})
	}

	placeID := c.Params("placeId")
	if placeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "placeId is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.savedPlacesService.RemovePlace(ctx, user.ID, placeID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "failed to remove saved place",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "place removed"})
}
