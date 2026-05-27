package repository

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SavedPlaceRepository defines the persistence contract for saved places.
type SavedPlaceRepository interface {
	GetSavedPlaces(ctx context.Context, userID primitive.ObjectID) ([]model.PlaceResult, error)
	GetSavedPlacesMap(ctx context.Context, userID primitive.ObjectID) (map[string]model.PlaceResult, error)
	SavePlace(ctx context.Context, userID primitive.ObjectID, place model.PlaceResult) error
	RemovePlace(ctx context.Context, userID primitive.ObjectID, placeID string) error
}

type savedPlaceRepository struct {
	collection *mongo.Collection
}

// NewSavedPlaceRepository returns a MongoDB-backed SavedPlaceRepository.
func NewSavedPlaceRepository(db *mongo.Database) SavedPlaceRepository {
	return &savedPlaceRepository{
		collection: db.Collection("saved_places"),
	}
}

// GetSavedPlaces returns all places saved by the user, newest first.
func (r *savedPlaceRepository) GetSavedPlaces(ctx context.Context, userID primitive.ObjectID) ([]model.PlaceResult, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"saved_at": -1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []model.SavedPlace
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	places := make([]model.PlaceResult, len(docs))
	for i, d := range docs {
		places[i] = d.PlaceData
	}
	return places, nil
}

// GetSavedPlacesMap returns a map of placeID → PlaceResult for all saved places of the user.
func (r *savedPlaceRepository) GetSavedPlacesMap(ctx context.Context, userID primitive.ObjectID) (map[string]model.PlaceResult, error) {
	places, err := r.GetSavedPlaces(ctx, userID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]model.PlaceResult, len(places))
	for _, p := range places {
		m[p.ID] = p
	}
	return m, nil
}

// SavePlace upserts a place snapshot for the user.
func (r *savedPlaceRepository) SavePlace(ctx context.Context, userID primitive.ObjectID, place model.PlaceResult) error {
	filter := bson.M{"user_id": userID, "place_data.id": place.ID}
	update := bson.M{
		"$set": bson.M{
			"user_id":    userID,
			"place_data": place,
			"saved_at":   time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

// RemovePlace deletes a user's saved place by Google Place ID.
func (r *savedPlaceRepository) RemovePlace(ctx context.Context, userID primitive.ObjectID, placeID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID, "place_data.id": placeID})
	return err
}
