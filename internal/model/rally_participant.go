package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RallyParticipant represents a participant entry in a rally
type RallyParticipant struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id"`
	RallyID   primitive.ObjectID  `json:"rallyId" bson:"rally_id"`
	UserID    primitive.ObjectID  `json:"userId" bson:"user_id"`
	Role      string              `json:"role" bson:"role"`
	Status    string              `json:"status" bson:"status"`
	InvitedBy *primitive.ObjectID `json:"invitedBy" bson:"invited_by"`
	JoinedAt  *time.Time          `json:"joinedAt" bson:"joined_at"`
	InvitedAt time.Time           `json:"invitedAt" bson:"invited_at"`
}

// InviteParticipantRequest represents the request to invite a user to a rally
type InviteParticipantRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role,omitempty"`
} //@name InviteParticipantRequest

// UpdateParticipantRequest represents the request to update a participant's role or status
type UpdateParticipantRequest struct {
	Role   *string `json:"role,omitempty"`
	Status *string `json:"status,omitempty"`
} //@name UpdateParticipantRequest

// RallyParticipantResponse represents the API response for a rally participant
type RallyParticipantResponse struct {
	ID        string     `json:"id" example:"507f1f77bcf86cd799439011"`
	RallyID   string     `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	UserID    string     `json:"userId" example:"507f1f77bcf86cd799439013"`
	Role      string     `json:"role" example:"participant"`
	Status    string     `json:"status" example:"invited"`
	InvitedBy string     `json:"invitedBy,omitempty" example:"507f1f77bcf86cd799439014"`
	JoinedAt  *time.Time `json:"joinedAt,omitempty" example:"2025-01-15T10:30:00Z"`
	InvitedAt time.Time  `json:"invitedAt" example:"2025-01-15T10:30:00Z"`
} //@name RallyParticipantResponse
