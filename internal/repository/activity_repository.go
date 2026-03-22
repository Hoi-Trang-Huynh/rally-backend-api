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

type ActivityRepository interface {
	CreateActivity(ctx context.Context, activity *model.Activity) error
	GetActivityByID(ctx context.Context, activityID string) (*model.Activity, error)
	UpdateActivity(ctx context.Context, activityID string, updates *model.UpdateActivityRequest) (*model.Activity, error)
}

type activityRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewActivityRepository(db *mongo.Database) ActivityRepository {
	return &activityRepository{
		db:         db,
		collection: db.Collection("activities"),
	}
}

func (r *activityRepository) CreateActivity(ctx context.Context, activity *model.Activity) error {
	if activity.ID.IsZero() {
		activity.ID = primitive.NewObjectID()
	}

	now := time.Now()
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = now
	}
	if activity.UpdatedAt.IsZero() {
		activity.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, activity)
	return err
}

func (r *activityRepository) GetActivityByID(ctx context.Context, activityID string) (*model.Activity, error) {
	objectID, err := primitive.ObjectIDFromHex(activityID)
	if err != nil {
		return nil, err
	}

	var activity model.Activity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&activity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &activity, nil
}

func (r *activityRepository) UpdateActivity(ctx context.Context, activityID string, updates *model.UpdateActivityRequest) (*model.Activity, error) {
	objectID, err := primitive.ObjectIDFromHex(activityID)
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
	if updates.Status != nil {
		updateDoc["status"] = *updates.Status
	}
	if updates.GooglePlaceID != nil {
		updateDoc["google_place_id"] = *updates.GooglePlaceID
	}
	if updates.Lat != nil {
		updateDoc["lat"] = *updates.Lat
	}
	if updates.Lng != nil {
		updateDoc["lng"] = *updates.Lng
	}
	if updates.StartTime != nil {
		updateDoc["start_time"] = *updates.StartTime
	}
	if updates.EndTime != nil {
		updateDoc["end_time"] = *updates.EndTime
	}
	if updates.Notes != nil {
		updateDoc["notes"] = *updates.Notes
	}
	if updates.ActivityOrder != nil {
		updateDoc["activity_order"] = *updates.ActivityOrder
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	return r.GetActivityByID(ctx, activityID)
}
