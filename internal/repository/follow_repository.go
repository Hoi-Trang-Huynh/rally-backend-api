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

type FollowRepository interface {
	CreateFollow(ctx context.Context, followerID, followingID primitive.ObjectID) (*model.Follow, error)
	DeleteFollow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	GetFollow(ctx context.Context, followerID, followingID primitive.ObjectID) (*model.Follow, error)
	GetFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*model.Follow, error)
	GetFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*model.Follow, error)
}

type followRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewFollowRepository initializes a MongoDB-backed FollowRepository
func NewFollowRepository(db *mongo.Database) FollowRepository {
	return &followRepository{
		db:         db,
		collection: db.Collection("follows"),
	}
}

// CreateFollow creates a new follow relationship
func (r *followRepository) CreateFollow(ctx context.Context, followerID, followingID primitive.ObjectID) (*model.Follow, error) {
	now := time.Now()
	follow := &model.Follow{
		ID:          primitive.NewObjectID(),
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := r.collection.InsertOne(ctx, follow)
	if err != nil {
		return nil, err
	}

	return follow, nil
}

// DeleteFollow removes a follow relationship
func (r *followRepository) DeleteFollow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"follower_id":  followerID,
		"following_id": followingID,
	})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("follow relationship not found")
	}

	return nil
}

// GetFollow retrieves a follow relationship if it exists
func (r *followRepository) GetFollow(ctx context.Context, followerID, followingID primitive.ObjectID) (*model.Follow, error) {
	var follow model.Follow
	err := r.collection.FindOne(ctx, bson.M{
		"follower_id":  followerID,
		"following_id": followingID,
	}).Decode(&follow)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &follow, nil
}

// GetFollowers retrieves users who follow the given user
func (r *followRepository) GetFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*model.Follow, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"following_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []*model.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, err
	}

	return follows, nil
}

// GetFollowing retrieves users that the given user follows
func (r *followRepository) GetFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*model.Follow, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"follower_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []*model.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, err
	}

	return follows, nil
}
