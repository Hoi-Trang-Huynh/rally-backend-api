package handler

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

// CollectionHandler handles collection HTTP requests.
type CollectionHandler struct {
	collectionService *service.CollectionService
}

// NewCollectionHandler creates a CollectionHandler.
func NewCollectionHandler(collectionService *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{collectionService: collectionService}
}

// GetCollections godoc
// @Summary Get user's collections
// @Description Returns all place collections owned by the authenticated user
// @Tags Collections
// @ID getCollections
// @Produce json
// @Success 200 {object} model.CollectionsResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /collections [get]
func (h *CollectionHandler) GetCollections(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{Message: "user not resolved"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cols, err := h.collectionService.GetUserCollections(ctx, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{Message: "failed to fetch collections"})
	}

	return c.Status(fiber.StatusOK).JSON(model.CollectionsResponse{Collections: cols})
}

// CreateCollection godoc
// @Summary Create a collection
// @Description Creates a new named place collection for the authenticated user
// @Tags Collections
// @ID createCollection
// @Accept json
// @Produce json
// @Param body body model.CreateCollectionRequest true "Collection to create"
// @Success 201 {object} model.CollectionResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /collections [post]
func (h *CollectionHandler) CreateCollection(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{Message: "user not resolved"})
	}

	var req model.CreateCollectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "name is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col, err := h.collectionService.CreateCollection(ctx, user.ID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{Message: "failed to create collection"})
	}

	return c.Status(fiber.StatusCreated).JSON(col)
}

// AddPlaceToCollection godoc
// @Summary Add a place to a collection
// @Description Adds a place ID to the specified collection
// @Tags Collections
// @ID addPlaceToCollection
// @Accept json
// @Produce json
// @Param id path string true "Collection ID"
// @Param body body model.AddPlaceToCollectionRequest true "Place to add"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /collections/{id}/places [post]
func (h *CollectionHandler) AddPlaceToCollection(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{Message: "user not resolved"})
	}

	collectionID := c.Params("id")
	if collectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "collection id is required"})
	}

	var req model.AddPlaceToCollectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "invalid request body"})
	}
	if req.PlaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "placeId is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.collectionService.AddPlace(ctx, user.ID, collectionID, req.PlaceID); err != nil {
		switch err.Error() {
		case "collection not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{Message: "collection not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{Message: "failed to add place"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "place added"})
}

// RemovePlaceFromCollection godoc
// @Summary Remove a place from a collection
// @Description Removes a place ID from the specified collection
// @Tags Collections
// @ID removePlaceFromCollection
// @Produce json
// @Param id path string true "Collection ID"
// @Param placeId path string true "Google Place ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /collections/{id}/places/{placeId} [delete]
func (h *CollectionHandler) RemovePlaceFromCollection(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{Message: "user not resolved"})
	}

	collectionID := c.Params("id")
	placeID := c.Params("placeId")
	if collectionID == "" || placeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "collection id and place id are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.collectionService.RemovePlace(ctx, user.ID, collectionID, placeID); err != nil {
		switch err.Error() {
		case "collection not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{Message: "collection not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{Message: "failed to remove place"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "place removed"})
}

// GetCollectionPlaces godoc
// @Summary Get places in a collection
// @Description Returns saved place details for all places in the specified collection
// @Tags Collections
// @ID getCollectionPlaces
// @Produce json
// @Param id path string true "Collection ID"
// @Success 200 {object} map[string][]model.PlaceResult
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /collections/{id}/places [get]
func (h *CollectionHandler) GetCollectionPlaces(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{Message: "user not resolved"})
	}

	collectionID := c.Params("id")
	if collectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "collection id is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	places, err := h.collectionService.GetCollectionPlaces(ctx, user.ID, collectionID)
	if err != nil {
		switch err.Error() {
		case "collection not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{Message: "collection not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{Message: "failed to fetch collection places"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"places": places})
}

// DeleteCollection godoc
// @Summary Delete a collection
// @Description Deletes a collection owned by the authenticated user
// @Tags Collections
// @ID deleteCollection
// @Produce json
// @Param id path string true "Collection ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /collections/{id} [delete]
func (h *CollectionHandler) DeleteCollection(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{Message: "user not resolved"})
	}

	collectionID := c.Params("id")
	if collectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{Message: "collection id is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.collectionService.DeleteCollection(ctx, user.ID, collectionID); err != nil {
		switch err.Error() {
		case "collection not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{Message: "collection not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{Message: "failed to delete collection"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "collection deleted"})
}
