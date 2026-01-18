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

// FollowUserItem represents a user in followers/following list
type FollowUserItem struct {
	ID        string `json:"id" example:"507f1f77bcf86cd799439011"`
	Username  string `json:"username" example:"johndoe"`
	FirstName string `json:"firstName,omitempty" example:"John"`
	LastName  string `json:"lastName,omitempty" example:"Doe"`
	AvatarUrl string `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
} //@name FollowUserItem

// FollowListResponse represents a paginated list of followers or following
type FollowListResponse struct {
	Users      []FollowUserItem `json:"users"`
	Total      int64            `json:"total" example:"100"`
	Page       int              `json:"page" example:"1"`
	PageSize   int              `json:"pageSize" example:"20"`
	TotalPages int              `json:"totalPages" example:"5"`
} //@name FollowListResponse
