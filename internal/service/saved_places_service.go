package service

import (
	"context"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SavedPlacesService handles bookmark persistence backed by the Places proxy.
type SavedPlacesService struct {
	savedPlaceRepo repository.SavedPlaceRepository
	placesService  *PlacesService
}

// NewSavedPlacesService creates a SavedPlacesService.
func NewSavedPlacesService(savedPlaceRepo repository.SavedPlaceRepository, placesService *PlacesService) *SavedPlacesService {
	return &SavedPlacesService{
		savedPlaceRepo: savedPlaceRepo,
		placesService:  placesService,
	}
}

// GetSavedPlaces returns all places bookmarked by the user.
func (s *SavedPlacesService) GetSavedPlaces(ctx context.Context, userID primitive.ObjectID) ([]model.PlaceResult, error) {
	places, err := s.savedPlaceRepo.GetSavedPlaces(ctx, userID)
	if err != nil {
		return nil, err
	}
	if places == nil {
		return []model.PlaceResult{}, nil
	}
	return places, nil
}

// SavePlace fetches place details from Google and persists a snapshot for the user.
func (s *SavedPlacesService) SavePlace(ctx context.Context, userID primitive.ObjectID, placeID string) error {
	place, err := s.placesService.GetPlaceDetails(ctx, placeID)
	if err != nil {
		return err
	}
	return s.savedPlaceRepo.SavePlace(ctx, userID, *place)
}

// RemovePlace deletes a user's saved place.
func (s *SavedPlacesService) RemovePlace(ctx context.Context, userID primitive.ObjectID, placeID string) error {
	return s.savedPlaceRepo.RemovePlace(ctx, userID, placeID)
}
