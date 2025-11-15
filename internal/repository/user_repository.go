package repository

import (
	"context"
	"errors"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
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

// GetUserByID finds a user by UserID
func (r *userRepository) GetUserByID(ctx context.Context, firebaseUID string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"firebaseuid": firebaseUID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
