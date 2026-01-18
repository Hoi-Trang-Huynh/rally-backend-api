package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Follow represents a follow relationship between two users
type Follow struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	FollowerID  primitive.ObjectID `json:"followerId" bson:"follower_id"`
	FollowingID primitive.ObjectID `json:"followingId" bson:"following_id"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// FollowResponse represents the API response for follow operations
type FollowResponse struct {
	Success     bool   `json:"success" example:"true"`
	Message     string `json:"message" example:"Successfully followed user"`
	IsFollowing bool   `json:"isFollowing" example:"true"`
} //@name FollowResponse

// FollowStatusResponse represents the follow status check response
type FollowStatusResponse struct {
	IsFollowing bool `json:"isFollowing" example:"true"`
} //@name FollowStatusResponse
