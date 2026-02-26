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

type EventRepository interface {
	CreateEvent(ctx context.Context, event *model.Event) error
	GetEventByID(ctx context.Context, eventID string) (*model.Event, error)
	UpdateEvent(ctx context.Context, eventID string, updates *model.UpdateEventRequest) (*model.Event, error)
	CountEventsByRally(ctx context.Context, rallyID primitive.ObjectID) (int64, error)
}

type eventRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewEventRepository(db *mongo.Database) EventRepository {
	return &eventRepository{
		db:         db,
		collection: db.Collection("events"),
	}
}

func (r *eventRepository) CreateEvent(ctx context.Context, event *model.Event) error {
	if event.ID.IsZero() {
		event.ID = primitive.NewObjectID()
	}

	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *eventRepository) GetEventByID(ctx context.Context, eventID string) (*model.Event, error) {
	objectID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, err
	}

	var event model.Event
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&event)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) UpdateEvent(ctx context.Context, eventID string, updates *model.UpdateEventRequest) (*model.Event, error) {
	objectID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, err
	}

	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	if updates.GooglePlaceID != nil {
		updateDoc["google_place_id"] = *updates.GooglePlaceID
	}
	if updates.Name != nil {
		updateDoc["name"] = *updates.Name
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
	if updates.VisitOrder != nil {
		updateDoc["visit_order"] = *updates.VisitOrder
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	return r.GetEventByID(ctx, eventID)
}

func (r *eventRepository) CountEventsByRally(ctx context.Context, rallyID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"rally_id": rallyID})
}
