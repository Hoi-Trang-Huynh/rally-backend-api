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

type FollowRepository interface {
	CreateFollow(ctx context.Context, followerID, followingID primitive.ObjectID) (*model.Follow, error)
	DeleteFollow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	GetFollow(ctx context.Context, followerID, followingID primitive.ObjectID) (*model.Follow, error)
	GetFollowers(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Follow, int64, error)
	GetFollowing(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Follow, int64, error)
	GetFriends(ctx context.Context, userID primitive.ObjectID, query string, page, pageSize int) ([]*model.Follow, int64, error)
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

// GetFollowers retrieves users who follow the given user with pagination
func (r *followRepository) GetFollowers(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Follow, int64, error) {
	filter := bson.M{"following_id": userID}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	cursor, err := r.collection.Find(ctx, filter,
		&options.FindOptions{
			Skip:  &skip,
			Limit: &limit,
			Sort:  bson.M{"created_at": -1}, // Most recent first
		},
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var follows []*model.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

// GetFollowing retrieves users that the given user follows with pagination
func (r *followRepository) GetFollowing(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Follow, int64, error) {
	filter := bson.M{"follower_id": userID}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	cursor, err := r.collection.Find(ctx, filter,
		&options.FindOptions{
			Skip:  &skip,
			Limit: &limit,
			Sort:  bson.M{"created_at": -1}, // Most recent first
		},
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var follows []*model.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

// GetFriends retrieves mutual friends (users who follow each other) with optional search and pagination
func (r *followRepository) GetFriends(ctx context.Context, userID primitive.ObjectID, query string, page, pageSize int) ([]*model.Follow, int64, error) {
	// Calculate skip for pagination
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// Build the aggregation pipeline
	matchStage := bson.M{
		"$match": bson.M{
			"follower_id": userID,
		},
	}

	// Lookup to check if the followed user also follows back
	lookupMutualStage := bson.M{
		"$lookup": bson.M{
			"from": "follows",
			"let":  bson.M{"following_id": "$following_id"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{
							"$and": bson.A{
								bson.M{"$eq": bson.A{"$follower_id", "$$following_id"}},
								bson.M{"$eq": bson.A{"$following_id", userID}},
							},
						},
					},
				},
			},
			"as": "mutual",
		},
	}

	// Filter only mutual follows (friends)
	filterMutualStage := bson.M{
		"$match": bson.M{
			"mutual.0": bson.M{"$exists": true},
		},
	}

	// Lookup user details
	lookupUserStage := bson.M{
		"$lookup": bson.M{
			"from":         "users",
			"localField":   "following_id",
			"foreignField": "_id",
			"as":           "user",
		},
	}

	// Unwind user array
	unwindStage := bson.M{
		"$unwind": "$user",
	}

	// Build match filter for user search and active status
	userMatchFilter := bson.M{
		"user.is_active": true,
	}

	// Add search filter if query is provided
	if query != "" {
		regexPattern := primitive.Regex{Pattern: query, Options: "i"}
		userMatchFilter["$or"] = bson.A{
			bson.M{"user.username": regexPattern},
			bson.M{"user.first_name": regexPattern},
			bson.M{"user.last_name": regexPattern},
		}
	}

	userMatchStage := bson.M{
		"$match": userMatchFilter,
	}

	// Sort by username
	sortStage := bson.M{
		"$sort": bson.M{"user.username": 1},
	}

	// Use $facet to get both count and paginated results in one query
	facetStage := bson.M{
		"$facet": bson.M{
			"metadata": bson.A{
				bson.M{"$count": "total"},
			},
			"data": bson.A{
				bson.M{"$skip": skip},
				bson.M{"$limit": limit},
			},
		},
	}

	// Build the complete pipeline
	pipeline := []bson.M{
		matchStage,
		lookupMutualStage,
		filterMutualStage,
		lookupUserStage,
		unwindStage,
		userMatchStage,
		sortStage,
		facetStage,
	}

	// Execute aggregation
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Parse results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	// Extract total count and data
	var total int64 = 0
	var follows []*model.Follow

	if len(results) > 0 {
		result := results[0]

		// Get total count
		if metadata, ok := result["metadata"].(primitive.A); ok && len(metadata) > 0 {
			if metaDoc, ok := metadata[0].(bson.M); ok {
				if totalVal, ok := metaDoc["total"].(int32); ok {
					total = int64(totalVal)
				} else if totalVal, ok := metaDoc["total"].(int64); ok {
					total = totalVal
				}
			}
		}

		// Get data
		if data, ok := result["data"].(primitive.A); ok {
			follows = make([]*model.Follow, 0, len(data))
			for _, item := range data {
				if doc, ok := item.(bson.M); ok {
					follow := &model.Follow{
						ID:          doc["_id"].(primitive.ObjectID),
						FollowerID:  doc["follower_id"].(primitive.ObjectID),
						FollowingID: doc["following_id"].(primitive.ObjectID),
					}
					// Extract timestamps if present
					if createdAt, ok := doc["created_at"].(primitive.DateTime); ok {
						follow.CreatedAt = createdAt.Time()
					}
					if updatedAt, ok := doc["updated_at"].(primitive.DateTime); ok {
						follow.UpdatedAt = updatedAt.Time()
					}
					follows = append(follows, follow)
				}
			}
		}
	}

	return follows, total, nil
}
