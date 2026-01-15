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

type UserRepository interface {
	GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUserProfile(ctx context.Context, userID string, updates *model.ProfileUpdateRequest) (*model.User, error)
	ExistsEmail(ctx context.Context, email string) (bool, error)
	ExistsUsername(ctx context.Context, username string) (bool, error)
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
