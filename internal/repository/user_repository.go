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

type UserRepository interface {
	GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUserProfile(ctx context.Context, userID string, updates *model.ProfileUpdateRequest) (*model.User, error)
	ExistsEmail(ctx context.Context, email string) (bool, error)
	ExistsUsername(ctx context.Context, username string) (bool, error)
	IncrementFollowersCount(ctx context.Context, userID string) error
	DecrementFollowersCount(ctx context.Context, userID string) error
	IncrementFollowingCount(ctx context.Context, userID string) error
	DecrementFollowingCount(ctx context.Context, userID string) error
	SearchUsers(ctx context.Context, query string, page, pageSize int) ([]*model.User, int64, error)
}

type userRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewUserRepository initializes a MongoDB-backed UserRepository
func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		db:         db,
		collection: db.Collection("users"),
	}
}

// GetUserByFirebaseUID finds a user by Firebase UID
func (r *userRepository) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"firebase_uid": firebaseUID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by MongoDB ObjectID
func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	// Convert string ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser inserts a new user document
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	// Generate new MongoDB ObjectID if not set
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	// Set timestamps if not already set
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// UpdateUserProfile updates user profile fields
func (r *userRepository) UpdateUserProfile(ctx context.Context, userID string, updates *model.ProfileUpdateRequest) (*model.User, error) {
	// Convert string ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Build update document
	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	// Only update fields that are provided (not nil)
	if updates.Username != nil {
		updateDoc["username"] = *updates.Username
	}
	if updates.FirstName != nil {
		updateDoc["first_name"] = *updates.FirstName
	}
	if updates.LastName != nil {
		updateDoc["last_name"] = *updates.LastName
	}
	if updates.AvatarUrl != nil {
		updateDoc["avatar_url"] = *updates.AvatarUrl
	}
	if updates.BioText != nil {
		updateDoc["bio_text"] = *updates.BioText
	}
	if updates.PhoneNumber != nil {
		updateDoc["phone_number"] = *updates.PhoneNumber
	}
	if updates.IsActive != nil {
		updateDoc["is_active"] = *updates.IsActive
	}
	if updates.IsEmailVerified != nil {
		updateDoc["is_email_verified"] = *updates.IsEmailVerified
	}
	if updates.IsOnboarding != nil {
		updateDoc["is_onboarding"] = *updates.IsOnboarding
	}

	// Perform the update
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	// Return updated user
	return r.GetUserByID(ctx, userID)
}

// ExistsEmail checks if an email already exists in the database
func (r *userRepository) ExistsEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsUsername checks if a username already exists in the database
func (r *userRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IncrementFollowersCount increments the followers_count for a user
func (r *userRepository) IncrementFollowersCount(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"followers_count": 1}},
	)
	return err
}

// DecrementFollowersCount decrements the followers_count for a user
func (r *userRepository) DecrementFollowersCount(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"followers_count": -1}},
	)
	return err
}

// IncrementFollowingCount increments the following_count for a user
func (r *userRepository) IncrementFollowingCount(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"following_count": 1}},
	)
	return err
}

// DecrementFollowingCount decrements the following_count for a user
func (r *userRepository) DecrementFollowingCount(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"following_count": -1}},
	)
	return err
}

// SearchUsers searches for users by username, first name, or last name with pagination
func (r *userRepository) SearchUsers(ctx context.Context, query string, page, pageSize int) ([]*model.User, int64, error) {
	// Create case-insensitive regex pattern for flexible matching
	regexPattern := primitive.Regex{Pattern: query, Options: "i"}

	// Build search filter - matches username, first_name, or last_name
	filter := bson.M{
		"$or": []bson.M{
			{"username": regexPattern},
			{"first_name": regexPattern},
			{"last_name": regexPattern},
		},
		"is_active": true, // Only search active users
	}

	// Get total count for pagination
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip value for pagination
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// Find users with pagination
	cursor, err := r.collection.Find(ctx, filter,
		&options.FindOptions{
			Skip:  &skip,
			Limit: &limit,
			Sort:  bson.M{"username": 1}, // Sort by username alphabetically
		},
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
