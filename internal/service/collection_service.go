package service

import (
	"context"
	"errors"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionService handles business logic for place collections.
type CollectionService struct {
	collectionRepo  repository.CollectionRepository
	savedPlaceRepo  repository.SavedPlaceRepository
}

// NewCollectionService creates a CollectionService.
func NewCollectionService(collectionRepo repository.CollectionRepository, savedPlaceRepo repository.SavedPlaceRepository) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		savedPlaceRepo: savedPlaceRepo,
	}
}

// GetUserCollections returns all collections owned by the user, with cover image URLs populated
// from the user's saved places.
func (s *CollectionService) GetUserCollections(ctx context.Context, userID primitive.ObjectID) ([]model.CollectionResponse, error) {
	cols, err := s.collectionRepo.GetUserCollections(ctx, userID)
	if err != nil {
		return nil, err
	}

	savedMap, err := s.savedPlaceRepo.GetSavedPlacesMap(ctx, userID)
	if err != nil {
		savedMap = map[string]model.PlaceResult{}
	}

	result := make([]model.CollectionResponse, len(cols))
	for i, col := range cols {
		resp := toCollectionResponse(col)
		if len(col.PlaceIDs) > 0 {
			if place, ok := savedMap[col.PlaceIDs[0]]; ok {
				resp.CoverImageURL = place.ImageUrl
			}
		}
		result[i] = resp
	}
	return result, nil
}

// CreateCollection creates a new named collection for the user.
func (s *CollectionService) CreateCollection(ctx context.Context, userID primitive.ObjectID, name string) (*model.CollectionResponse, error) {
	col := &model.Collection{
		UserID: userID,
		Name:   name,
	}
	if err := s.collectionRepo.CreateCollection(ctx, col); err != nil {
		return nil, err
	}
	resp := toCollectionResponse(*col)
	return &resp, nil
}

// AddPlace adds a place to a collection owned by the user.
func (s *CollectionService) AddPlace(ctx context.Context, userID primitive.ObjectID, collectionID string, placeID string) error {
	id, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return errors.New("collection not found")
	}
	col, err := s.collectionRepo.GetCollectionByID(ctx, id)
	if err != nil {
		return err
	}
	if col == nil || col.UserID != userID {
		return errors.New("collection not found")
	}
	return s.collectionRepo.AddPlace(ctx, id, userID, placeID)
}

// RemovePlace removes a place from a collection owned by the user.
func (s *CollectionService) RemovePlace(ctx context.Context, userID primitive.ObjectID, collectionID string, placeID string) error {
	id, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return errors.New("collection not found")
	}
	col, err := s.collectionRepo.GetCollectionByID(ctx, id)
	if err != nil {
		return err
	}
	if col == nil || col.UserID != userID {
		return errors.New("collection not found")
	}
	return s.collectionRepo.RemovePlace(ctx, id, userID, placeID)
}

// GetCollectionPlaces returns the saved place details for all place IDs in a collection.
func (s *CollectionService) GetCollectionPlaces(ctx context.Context, userID primitive.ObjectID, collectionID string) ([]model.PlaceResult, error) {
	id, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, errors.New("collection not found")
	}
	col, err := s.collectionRepo.GetCollectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if col == nil || col.UserID != userID {
		return nil, errors.New("collection not found")
	}
	if len(col.PlaceIDs) == 0 {
		return []model.PlaceResult{}, nil
	}
	savedMap, err := s.savedPlaceRepo.GetSavedPlacesMap(ctx, userID)
	if err != nil {
		return nil, err
	}
	places := make([]model.PlaceResult, 0, len(col.PlaceIDs))
	for _, pid := range col.PlaceIDs {
		if place, ok := savedMap[pid]; ok {
			places = append(places, place)
		}
	}
	return places, nil
}

// DeleteCollection deletes a collection owned by the user.
func (s *CollectionService) DeleteCollection(ctx context.Context, userID primitive.ObjectID, collectionID string) error {
	id, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return errors.New("collection not found")
	}
	col, err := s.collectionRepo.GetCollectionByID(ctx, id)
	if err != nil {
		return err
	}
	if col == nil || col.UserID != userID {
		return errors.New("collection not found")
	}
	return s.collectionRepo.DeleteCollection(ctx, id, userID)
}

func toCollectionResponse(col model.Collection) model.CollectionResponse {
	placeIDs := col.PlaceIDs
	if placeIDs == nil {
		placeIDs = []string{}
	}
	return model.CollectionResponse{
		ID:        col.ID.Hex(),
		Name:      col.Name,
		PlaceIDs:  placeIDs,
		CreatedAt: col.CreatedAt,
		UpdatedAt: col.UpdatedAt,
	}
}
