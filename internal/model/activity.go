package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Activity represents an activity within an event
type Activity struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	EventID       primitive.ObjectID `json:"eventId" bson:"event_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	Status        string             `json:"status" bson:"status"`
	GooglePlaceID string             `json:"googlePlaceId" bson:"google_place_id"`
	Lat           float64            `json:"lat" bson:"lat"`
	Lng           float64            `json:"lng" bson:"lng"`
	StartTime     *time.Time         `json:"startTime" bson:"start_time"`
	EndTime       *time.Time         `json:"endTime" bson:"end_time"`
	Notes         string             `json:"notes" bson:"notes"`
	ActivityOrder int                `json:"activityOrder" bson:"activity_order"`
	CreatedAt     time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateActivityRequest represents the request payload for creating an activity
type CreateActivityRequest struct {
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	GooglePlaceID string     `json:"googlePlaceId,omitempty"`
	Lat           float64    `json:"lat,omitempty"`
	Lng           float64    `json:"lng,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	ActivityOrder int        `json:"activityOrder,omitempty"`
} //@name CreateActivityRequest

// UpdateActivityRequest represents the request payload for updating an activity
type UpdateActivityRequest struct {
	Name          *string    `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Status        *string    `json:"status,omitempty"`
	GooglePlaceID *string    `json:"googlePlaceId,omitempty"`
	Lat           *float64   `json:"lat,omitempty"`
	Lng           *float64   `json:"lng,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	ActivityOrder *int       `json:"activityOrder,omitempty"`
} //@name UpdateActivityRequest

// ActivityResponse represents the API response for an activity
type ActivityResponse struct {
	ID            string     `json:"id" example:"507f1f77bcf86cd799439011"`
	EventID       string     `json:"eventId" example:"507f1f77bcf86cd799439012"`
	Name          string     `json:"name" example:"Bike across the bridge"`
	Description   string     `json:"description,omitempty" example:"Rent bikes and ride across"`
	Status        string     `json:"status" example:"planned"`
	GooglePlaceID string     `json:"googlePlaceId,omitempty" example:"ChIJN1t_tDeuEmsRUsoyG83frY4"`
	Lat           float64    `json:"lat,omitempty" example:"37.8199"`
	Lng           float64    `json:"lng,omitempty" example:"-122.4783"`
	StartTime     *time.Time `json:"startTime,omitempty" example:"2025-07-01T09:00:00Z"`
	EndTime       *time.Time `json:"endTime,omitempty" example:"2025-07-01T10:00:00Z"`
	Notes         string     `json:"notes,omitempty" example:"Bring sunscreen"`
	ActivityOrder int        `json:"activityOrder" example:"1"`
	CreatedAt     time.Time  `json:"createdAt" example:"2025-01-15T10:30:00Z"`
	UpdatedAt     time.Time  `json:"updatedAt" example:"2025-01-15T10:30:00Z"`
} //@name ActivityResponse
