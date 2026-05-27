package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionRepository defines the persistence contract for place collections.
type CollectionRepository interface {
	GetUserCollections(ctx context.Context, userID primitive.ObjectID) ([]model.Collection, error)
	GetCollectionByID(ctx context.Context, id primitive.ObjectID) (*model.Collection, error)
	CreateCollection(ctx context.Context, col *model.Collection) error
	AddPlace(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, placeID string) error
	RemovePlace(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, placeID string) error
	DeleteCollection(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type collectionRepository struct {
	collection *mongo.Collection
}

// NewCollectionRepository returns a MongoDB-backed CollectionRepository.
func NewCollectionRepository(db *mongo.Database) CollectionRepository {
	return &collectionRepository{
		collection: db.Collection("collections"),
	}
}

// GetUserCollections returns all collections owned by the user, newest first.
func (r *collectionRepository) GetUserCollections(ctx context.Context, userID primitive.ObjectID) ([]model.Collection, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cols []model.Collection
	if err := cursor.All(ctx, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

// GetCollectionByID returns a single collection. Returns nil, nil if not found.
func (r *collectionRepository) GetCollectionByID(ctx context.Context, id primitive.ObjectID) (*model.Collection, error) {
	var col model.Collection
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&col)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &col, nil
}

// CreateCollection inserts a new collection document.
func (r *collectionRepository) CreateCollection(ctx context.Context, col *model.Collection) error {
	now := time.Now()
	col.ID = primitive.NewObjectID()
	col.CreatedAt = now
	col.UpdatedAt = now
	if col.PlaceIDs == nil {
		col.PlaceIDs = []string{}
	}
	_, err := r.collection.InsertOne(ctx, col)
	return err
}

// AddPlace appends a placeID to the collection (idempotent via $addToSet).
func (r *collectionRepository) AddPlace(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, placeID string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "user_id": userID},
		bson.M{
			"$addToSet": bson.M{"place_ids": placeID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// RemovePlace removes a placeID from the collection.
func (r *collectionRepository) RemovePlace(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, placeID string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "user_id": userID},
		bson.M{
			"$pull": bson.M{"place_ids": placeID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// DeleteCollection removes the collection document.
func (r *collectionRepository) DeleteCollection(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	return err
}
