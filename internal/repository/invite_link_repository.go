package repository

import (
	"context"
	"errors"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type InviteLinkRepository interface {
	CreateInviteLink(ctx context.Context, link *model.InviteLink) error
	GetInviteLinkByToken(ctx context.Context, token string) (*model.InviteLink, error)
	GetActiveInviteLinksByRally(ctx context.Context, rallyID primitive.ObjectID) ([]*model.InviteLink, error)
	DeactivateInviteLink(ctx context.Context, token string) error
	IncrementLinkUsage(ctx context.Context, token string) error
}

type inviteLinkRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewInviteLinkRepository initializes a MongoDB-backed InviteLinkRepository
func NewInviteLinkRepository(db *mongo.Database) InviteLinkRepository {
	return &inviteLinkRepository{
		db:         db,
		collection: db.Collection("invite_links"),
	}
}

// CreateInviteLink inserts a new invite link into the database
func (r *inviteLinkRepository) CreateInviteLink(ctx context.Context, link *model.InviteLink) error {
	_, err := r.collection.InsertOne(ctx, link)
	return err
}

// GetInviteLinkByToken finds an invite link by its unique token
func (r *inviteLinkRepository) GetInviteLinkByToken(ctx context.Context, token string) (*model.InviteLink, error) {
	var link model.InviteLink
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&link)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &link, nil
}

// GetActiveInviteLinksByRally retrieves all active invite links for a specific rally
func (r *inviteLinkRepository) GetActiveInviteLinksByRally(ctx context.Context, rallyID primitive.ObjectID) ([]*model.InviteLink, error) {
	filter := bson.M{
		"rally_id":  rallyID,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var links []*model.InviteLink
	if err := cursor.All(ctx, &links); err != nil {
		return nil, err
	}
	if links == nil {
		links = []*model.InviteLink{}
	}

	return links, nil
}

// DeactivateInviteLink marks an invite link as inactive (revoked)
func (r *inviteLinkRepository) DeactivateInviteLink(ctx context.Context, token string) error {
	filter := bson.M{"token": token}
	update := bson.M{"$set": bson.M{"is_active": false}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("invite link not found or already inactive")
	}

	return nil
}

// IncrementLinkUsage increments the current_uses counter of an invite link
func (r *inviteLinkRepository) IncrementLinkUsage(ctx context.Context, token string) error {
	filter := bson.M{"token": token}
	update := bson.M{"$inc": bson.M{"current_uses": 1}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("invite link not found")
	}

	return nil
}
