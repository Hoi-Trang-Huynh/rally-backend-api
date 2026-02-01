package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RallyRepository interface {
	CreateRally(ctx context.Context, rally *model.Rally) error
	GetRallyByID(ctx context.Context, rallyID string) (*model.Rally, error)
	UpdateRally(ctx context.Context, rallyID string, updates *model.UpdateRallyRequest) (*model.Rally, error)
}

type rallyRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewRallyRepository(db *mongo.Database) RallyRepository {
	return &rallyRepository{
		db:         db,
		collection: db.Collection("rallies"),
	}
}

func (r *rallyRepository) CreateRally(ctx context.Context, rally *model.Rally) error {
	if rally.ID.IsZero() {
		rally.ID = primitive.NewObjectID()
	}

	now := time.Now()
	if rally.CreatedAt.IsZero() {
		rally.CreatedAt = now
	}
	if rally.UpdatedAt.IsZero() {
		rally.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, rally)
	return err
}

func (r *rallyRepository) GetRallyByID(ctx context.Context, rallyID string) (*model.Rally, error) {
	objectID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, err
	}

	var rally model.Rally
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&rally)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &rally, nil
}

func (r *rallyRepository) UpdateRally(ctx context.Context, rallyID string, updates *model.UpdateRallyRequest) (*model.Rally, error) {
	objectID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, err
	}

	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	if updates.Name != nil {
		updateDoc["name"] = *updates.Name
	}
	if updates.Description != nil {
		updateDoc["description"] = *updates.Description
	}
	if updates.CoverImageUrl != nil {
		updateDoc["cover_image_url"] = *updates.CoverImageUrl
	}
	if updates.Status != nil {
		updateDoc["status"] = *updates.Status
	}
	if updates.StartDate != nil {
		updateDoc["start_date"] = *updates.StartDate
	}
	if updates.EndDate != nil {
		updateDoc["end_date"] = *updates.EndDate
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	return r.GetRallyByID(ctx, rallyID)
}
