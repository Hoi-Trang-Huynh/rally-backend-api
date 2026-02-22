package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParticipantRole represents the role of a participant in a rally
type ParticipantRole string

const (
	ParticipantRoleOwner       ParticipantRole = "owner"
	ParticipantRoleEditor      ParticipantRole = "editor"
	ParticipantRoleParticipant ParticipantRole = "participant"
)

// ParticipationStatus represents the status of a participant in a rally
type ParticipationStatus string

const (
	ParticipationStatusInvited  ParticipationStatus = "invited"
	ParticipationStatusJoined   ParticipationStatus = "joined"
	ParticipationStatusDeclined ParticipationStatus = "declined"
	ParticipationStatusLeft     ParticipationStatus = "left"
)

// RallyParticipant represents a participant entry in a rally
type RallyParticipant struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id"`
	RallyID   primitive.ObjectID  `json:"rallyId" bson:"rally_id"`
	UserID    primitive.ObjectID  `json:"userId" bson:"user_id"`
	Role      ParticipantRole     `json:"role" bson:"role"`
	Status    ParticipationStatus `json:"status" bson:"status"`
	InvitedBy *primitive.ObjectID `json:"invitedBy" bson:"invited_by"`
	JoinedAt  *time.Time          `json:"joinedAt" bson:"joined_at"`
	InvitedAt time.Time           `json:"invitedAt" bson:"invited_at"`
}

// InviteParticipantRequest represents the request to invite a user to a rally
type InviteParticipantRequest struct {
	UserID string           `json:"userId"`
	Role   *ParticipantRole `json:"role,omitempty"`
} //@name InviteParticipantRequest

// UpdateParticipantRequest represents the request to update a participant's role or status
type UpdateParticipantRequest struct {
	Role   *ParticipantRole     `json:"role,omitempty"`
	Status *ParticipationStatus `json:"status,omitempty"`
} //@name UpdateParticipantRequest

// RallyParticipantResponse represents the API response for a rally participant
type RallyParticipantResponse struct {
	ID        string              `json:"id" example:"507f1f77bcf86cd799439011"`
	RallyID   string              `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	UserID    string              `json:"userId" example:"507f1f77bcf86cd799439013"`
	Role      ParticipantRole     `json:"role" example:"participant"`
	Status    ParticipationStatus `json:"status" example:"invited"`
	InvitedBy string              `json:"invitedBy,omitempty" example:"507f1f77bcf86cd799439014"`
	JoinedAt  *time.Time          `json:"joinedAt,omitempty" example:"2025-01-15T10:30:00Z"`
	InvitedAt time.Time           `json:"invitedAt" example:"2025-01-15T10:30:00Z"`
} //@name RallyParticipantResponse

// ParticipantUserInfo contains basic info about a user in a participant list
type ParticipantUserInfo struct {
	ID        string `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011"`
	Username  string `json:"username" bson:"username" example:"johndoe"`
	FirstName string `json:"firstName,omitempty" bson:"first_name" example:"John"`
	LastName  string `json:"lastName,omitempty" bson:"last_name" example:"Doe"`
	AvatarUrl string `json:"avatarUrl,omitempty" bson:"avatar_url" example:"https://example.com/avatar.jpg"`
} //@name ParticipantUserInfo

// RallyParticipantDetailResponse represents a detailed participant entry including user info
type RallyParticipantDetailResponse struct {
	ID        string               `json:"id" example:"507f1f77bcf86cd799439011"`
	RallyID   string               `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	Role      ParticipantRole      `json:"role" example:"participant"`
	Status    ParticipationStatus  `json:"status" example:"joined"`
	JoinedAt  *time.Time           `json:"joinedAt,omitempty" example:"2025-01-15T10:30:00Z"`
	InvitedAt time.Time            `json:"invitedAt" example:"2025-01-15T10:30:00Z"`
	User      ParticipantUserInfo  `json:"user"`
	InvitedBy *ParticipantUserInfo `json:"invitedBy,omitempty"`
} //@name RallyParticipantDetailResponse

// ParticipantListResponse represents the API response for participants list
type ParticipantListResponse struct {
	Participants []RallyParticipantDetailResponse `json:"participants"`
	Total        int                              `json:"total" example:"100"`
	Page         int                              `json:"page" example:"1"`
	PageSize     int                              `json:"pageSize" example:"20"`
	TotalPages   int                              `json:"totalPages" example:"5"`
	Pagination   PaginationMetadata               `json:"pagination"`
} //@name ParticipantListResponse
