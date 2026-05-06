package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type PlacesHandler struct {
	placesService *service.PlacesService
}

func NewPlacesHandler(placesService *service.PlacesService) *PlacesHandler {
	return &PlacesHandler{placesService: placesService}
}

// NearbySearch godoc
// @Summary Search nearby places
// @Description Returns places near a given location, proxied from Google Places API
// @Tags Places
// @ID nearbySearch
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param type query string true "Place type (restaurant, lodging, bar, amusement_park, tourist_attraction)"
// @Param maxCount query int false "Max results (default 10, max 20)"
// @Success 200 {object} model.NearbyPlacesResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /places/nearby [get]
func (h *PlacesHandler) NearbySearch(c *fiber.Ctx) error {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	placeType := c.Query("type")

	if latStr == "" || lngStr == "" || placeType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "lat, lng, and type query parameters are required",
		})
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid lat value",
		})
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid lng value",
		})
	}

	maxCount := 10
	if mc := c.Query("maxCount"); mc != "" {
		if v, err := strconv.Atoi(mc); err == nil && v > 0 {
			if v > 20 {
				v = 20
			}
			maxCount = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	places, err := h.placesService.NearbySearch(ctx, lat, lng, placeType, maxCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "failed to fetch nearby places",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.NearbyPlacesResponse{Places: places})
}

// GetPlaceDetails godoc
// @Summary Get place details
// @Description Returns full details for a place, proxied from Google Places API
// @Tags Places
// @ID getPlaceDetails
// @Produce json
// @Param placeId path string true "Google Place ID"
// @Success 200 {object} model.PlaceResult
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /places/{placeId} [get]
func (h *PlacesHandler) GetPlaceDetails(c *fiber.Ctx) error {
	placeID := c.Params("placeId")
	if placeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "placeId is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	place, err := h.placesService.GetPlaceDetails(ctx, placeID)
	if err != nil {
		switch err.Error() {
		case "place not found":
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Message: "place not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "failed to fetch place details",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(place)
}
