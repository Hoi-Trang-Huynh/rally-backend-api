package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Rally represents a rally/trip document in MongoDB
type Rally struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	OwnerID       primitive.ObjectID `json:"ownerId" bson:"owner_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	CoverImageUrl string             `json:"coverImageUrl" bson:"cover_image_url"`
	Status        string             `json:"status" bson:"status"`
	StartDate     *time.Time         `json:"startDate" bson:"start_date"`
	EndDate       *time.Time         `json:"endDate" bson:"end_date"`
	CreatedAt     time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateRallyRequest represents the request payload for creating a rally
type CreateRallyRequest struct {
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	CoverImageUrl string     `json:"coverImageUrl,omitempty"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"`
} //@name CreateRallyRequest

// UpdateRallyRequest represents the request payload for updating a rally
type UpdateRallyRequest struct {
	Name          *string    `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CoverImageUrl *string    `json:"coverImageUrl,omitempty"`
	Status        *string    `json:"status,omitempty"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"`
} //@name UpdateRallyRequest

// RallyResponse represents the API response for a rally
type RallyResponse struct {
	ID            string     `json:"id" example:"507f1f77bcf86cd799439011"`
	OwnerID       string     `json:"ownerId" example:"507f1f77bcf86cd799439012"`
	Name          string     `json:"name" example:"Summer Road Trip"`
	Description   string     `json:"description,omitempty" example:"A road trip through California"`
	CoverImageUrl string     `json:"coverImageUrl,omitempty" example:"https://example.com/cover.jpg"`
	Status        string     `json:"status" example:"draft"`
	StartDate     *time.Time `json:"startDate,omitempty" example:"2025-07-01T00:00:00Z"`
	EndDate       *time.Time `json:"endDate,omitempty" example:"2025-07-15T00:00:00Z"`
	CreatedAt     time.Time  `json:"createdAt" example:"2025-01-15T10:30:00Z"`
	UpdatedAt     time.Time  `json:"updatedAt" example:"2025-01-15T10:30:00Z"`
} //@name RallyResponse
