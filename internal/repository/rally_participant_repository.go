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

type RallyParticipantRepository interface {
	CreateParticipant(ctx context.Context, participant *model.RallyParticipant) error
	GetParticipant(ctx context.Context, participantID string) (*model.RallyParticipant, error)
	GetParticipantByRallyAndUser(ctx context.Context, rallyID, userID primitive.ObjectID) (*model.RallyParticipant, error)
	UpdateParticipant(ctx context.Context, participantID string, updates *model.UpdateParticipantRequest) (*model.RallyParticipant, error)
}

type rallyParticipantRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewRallyParticipantRepository(db *mongo.Database) RallyParticipantRepository {
	return &rallyParticipantRepository{
		db:         db,
		collection: db.Collection("rally_participants"),
	}
}

func (r *rallyParticipantRepository) CreateParticipant(ctx context.Context, participant *model.RallyParticipant) error {
	if participant.ID.IsZero() {
		participant.ID = primitive.NewObjectID()
	}

	if participant.InvitedAt.IsZero() {
		participant.InvitedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, participant)
	return err
}

func (r *rallyParticipantRepository) GetParticipant(ctx context.Context, participantID string) (*model.RallyParticipant, error) {
	objectID, err := primitive.ObjectIDFromHex(participantID)
	if err != nil {
		return nil, err
	}

	var participant model.RallyParticipant
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&participant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &participant, nil
}

func (r *rallyParticipantRepository) GetParticipantByRallyAndUser(ctx context.Context, rallyID, userID primitive.ObjectID) (*model.RallyParticipant, error) {
	var participant model.RallyParticipant
	err := r.collection.FindOne(ctx, bson.M{
		"rally_id": rallyID,
		"user_id":  userID,
	}).Decode(&participant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &participant, nil
}

func (r *rallyParticipantRepository) UpdateParticipant(ctx context.Context, participantID string, updates *model.UpdateParticipantRequest) (*model.RallyParticipant, error) {
	objectID, err := primitive.ObjectIDFromHex(participantID)
	if err != nil {
		return nil, err
	}

	updateDoc := bson.M{}

	if updates.Role != nil {
		updateDoc["role"] = *updates.Role
	}
	if updates.Status != nil {
		updateDoc["status"] = *updates.Status
		if *updates.Status == "joined" {
			now := time.Now()
			updateDoc["joined_at"] = now
		}
	}

	if len(updateDoc) == 0 {
		return r.GetParticipant(ctx, participantID)
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	return r.GetParticipant(ctx, participantID)
}
