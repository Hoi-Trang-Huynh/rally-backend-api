package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents an event/stop within a rally
type Event struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	RallyID       primitive.ObjectID `json:"rallyId" bson:"rally_id"`
	GooglePlaceID string             `json:"googlePlaceId" bson:"google_place_id"`
	Name          string             `json:"name" bson:"name"`
	Lat           float64            `json:"lat" bson:"lat"`
	Lng           float64            `json:"lng" bson:"lng"`
	StartTime     *time.Time         `json:"startTime" bson:"start_time"`
	EndTime       *time.Time         `json:"endTime" bson:"end_time"`
	Notes         string             `json:"notes" bson:"notes"`
	VisitOrder    int                `json:"visitOrder" bson:"visit_order"`
	CreatedAt     time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateEventRequest represents the request payload for creating an event
type CreateEventRequest struct {
	GooglePlaceID string     `json:"googlePlaceId,omitempty"`
	Name          string     `json:"name"`
	Lat           float64    `json:"lat,omitempty"`
	Lng           float64    `json:"lng,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	VisitOrder    int        `json:"visitOrder,omitempty"`
} //@name CreateEventRequest

// UpdateEventRequest represents the request payload for updating an event
type UpdateEventRequest struct {
	GooglePlaceID *string    `json:"googlePlaceId,omitempty"`
	Name          *string    `json:"name,omitempty"`
	Lat           *float64   `json:"lat,omitempty"`
	Lng           *float64   `json:"lng,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	VisitOrder    *int       `json:"visitOrder,omitempty"`
} //@name UpdateEventRequest

// EventResponse represents the API response for an event
type EventResponse struct {
	ID            string     `json:"id" example:"507f1f77bcf86cd799439011"`
	RallyID       string     `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	GooglePlaceID string     `json:"googlePlaceId,omitempty" example:"ChIJN1t_tDeuEmsRUsoyG83frY4"`
	Name          string     `json:"name" example:"Golden Gate Bridge"`
	Lat           float64    `json:"lat" example:"37.8199"`
	Lng           float64    `json:"lng" example:"-122.4783"`
	StartTime     *time.Time `json:"startTime,omitempty" example:"2025-07-01T09:00:00Z"`
	EndTime       *time.Time `json:"endTime,omitempty" example:"2025-07-01T12:00:00Z"`
	Notes         string     `json:"notes,omitempty" example:"Arrive early for parking"`
	VisitOrder    int        `json:"visitOrder" example:"1"`
	CreatedAt     time.Time  `json:"createdAt" example:"2025-01-15T10:30:00Z"`
	UpdatedAt     time.Time  `json:"updatedAt" example:"2025-01-15T10:30:00Z"`
} //@name EventResponse
