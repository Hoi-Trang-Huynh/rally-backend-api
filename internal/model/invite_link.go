package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InviteLink represents a stateful link/QR token that can be used to join a rally
type InviteLink struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	RallyID     primitive.ObjectID `json:"rallyId" bson:"rally_id"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"created_by"`
	Token       string             `json:"token" bson:"token"` // Unique UUID or secure random string
	RoleToGrant ParticipantRole    `json:"roleToGrant" bson:"role_to_grant"`
	ExpiresAt   *time.Time         `json:"expiresAt,omitempty" bson:"expires_at,omitempty"`
	MaxUses     int                `json:"maxUses" bson:"max_uses"`         // 0 means unlimited
	CurrentUses int                `json:"currentUses" bson:"current_uses"` // How many times it has been used
	IsActive    bool               `json:"isActive" bson:"is_active"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
}

// CreateInviteLinkRequest represents the request to create a new invite link/QR
type CreateInviteLinkRequest struct {
	Role          ParticipantRole `json:"role" validate:"omitempty,oneof=owner editor participant" example:"participant"`
	ExpiresInDays *int            `json:"expiresInDays,omitempty" validate:"omitempty,min=1,max=365" example:"7"`
	MaxUses       int             `json:"maxUses" validate:"omitempty,min=0" example:"10"` // 0 for unlimited
} //@name CreateInviteLinkRequest

// InviteLinkResponse represents the API response for an invite link
type InviteLinkResponse struct {
	ID          string          `json:"id" example:"507f1f77bcf86cd799439011"`
	RallyID     string          `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	CreatedBy   string          `json:"createdBy" example:"507f1f77bcf86cd799439013"`
	Token       string          `json:"token" example:"e7b8e9d0-f1a2-4b3c-9d8e-f7a6b5c4d3e2"`
	RoleToGrant ParticipantRole `json:"roleToGrant" example:"participant"`
	ExpiresAt   *time.Time      `json:"expiresAt,omitempty" example:"2025-01-15T10:30:00Z"`
	MaxUses     int             `json:"maxUses" example:"10"`
	CurrentUses int             `json:"currentUses" example:"2"`
	IsActive    bool            `json:"isActive" example:"true"`
	CreatedAt   time.Time       `json:"createdAt" example:"2025-01-01T10:30:00Z"`
} //@name InviteLinkResponse

// JoinViaLinkRequest represents the request to join a rally via an invite link/QR
type JoinViaLinkRequest struct {
	Token string `json:"token" validate:"required" example:"e7b8e9d0-f1a2-4b3c-9d8e-f7a6b5c4d3e2"`
} //@name JoinViaLinkRequest

// JoinViaLinkResponse represents the API response after successfully joining via a link
type JoinViaLinkResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Successfully joined the rally"`
	RallyID string `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	Role    string `json:"role" example:"participant"`
	Status  string `json:"status" example:"invited"`
} //@name JoinViaLinkResponse

// InviteLinkPreviewOwner contains the rally owner's public info for the preview card
type InviteLinkPreviewOwner struct {
	Username  string `json:"username" example:"johndoe"`
	FirstName string `json:"firstName,omitempty" example:"John"`
	LastName  string `json:"lastName,omitempty" example:"Doe"`
	AvatarUrl string `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
} //@name InviteLinkPreviewOwner

// InviteLinkPreviewResponse represents the preview card data for an invitation
type InviteLinkPreviewResponse struct {
	RallyID       string                 `json:"rallyId" example:"507f1f77bcf86cd799439012"`
	RallyName     string                 `json:"rallyName" example:"Summer Road Trip"`
	Description   interface{}            `json:"description,omitempty"`
	CoverImageUrl string                 `json:"coverImageUrl,omitempty" example:"https://example.com/cover.jpg"`
	Status        RallyStatus            `json:"status" example:"active"`
	StartDate     *time.Time             `json:"startDate,omitempty" example:"2025-07-01T00:00:00Z"`
	EndDate       *time.Time             `json:"endDate,omitempty" example:"2025-07-15T00:00:00Z"`
	Owner         InviteLinkPreviewOwner `json:"owner"`
	RoleOffered   ParticipantRole        `json:"roleOffered" example:"participant"`
	MemberCount   int64                  `json:"memberCount" example:"5"`
	EventCount    int64                  `json:"eventCount" example:"3"`
	ParticipantID string                 `json:"participantId" example:"507f1f77bcf86cd799439015"`
	InvitedAt     time.Time              `json:"invitedAt" example:"2025-01-15T10:30:00Z"`
} //@name InviteLinkPreviewResponse
