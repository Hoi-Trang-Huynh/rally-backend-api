package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RallyStatus represents the status of a rally
type RallyStatus string

const (
	RallyStatusDraft     RallyStatus = "draft"
	RallyStatusActive    RallyStatus = "active"
	RallyStatusInactive  RallyStatus = "inactive"
	RallyStatusCompleted RallyStatus = "completed"
	RallyStatusArchived  RallyStatus = "archived"
)

// Rally represents a rally/trip document in MongoDB
type Rally struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	OwnerID       primitive.ObjectID `json:"ownerId" bson:"owner_id"`
	Name          string             `json:"name" bson:"name"`
	Description   interface{}        `json:"description" bson:"description"`
	CoverImageUrl string             `json:"coverImageUrl" bson:"cover_image_url"`
	Status        RallyStatus        `json:"status" bson:"status"`
	StartDate     *time.Time         `json:"startDate" bson:"start_date"`
	EndDate       *time.Time         `json:"endDate" bson:"end_date"`
	CreatedAt     time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateRallyRequest represents the request payload for creating a rally
type CreateRallyRequest struct {
	Name          string                     `json:"name"`
	Description   interface{}                `json:"description,omitempty"`
	CoverImageUrl string                     `json:"coverImageUrl,omitempty"`
	StartDate     *time.Time                 `json:"startDate,omitempty"`
	EndDate       *time.Time                 `json:"endDate,omitempty"`
	Participants  []InviteParticipantRequest `json:"participants,omitempty"`
} //@name CreateRallyRequest

// UpdateRallyRequest represents the request payload for updating a rally
type UpdateRallyRequest struct {
	Name          *string      `json:"name,omitempty"`
	Description   *interface{} `json:"description,omitempty"`
	CoverImageUrl *string      `json:"coverImageUrl,omitempty"`
	Status        *RallyStatus `json:"status,omitempty"`
	StartDate     *time.Time   `json:"startDate,omitempty"`
	EndDate       *time.Time   `json:"endDate,omitempty"`
} //@name UpdateRallyRequest

// RallyResponse represents the API response for a rally
type RallyResponse struct {
	ID            string      `json:"id" example:"507f1f77bcf86cd799439011"`
	OwnerID       string      `json:"ownerId" example:"507f1f77bcf86cd799439012"`
	Name          string      `json:"name" example:"Summer Road Trip"`
	Description   interface{} `json:"description,omitempty"`
	CoverImageUrl string      `json:"coverImageUrl,omitempty" example:"https://example.com/cover.jpg"`
	Status        RallyStatus `json:"status" example:"draft"`
	StartDate     *time.Time  `json:"startDate,omitempty" example:"2025-07-01T00:00:00Z"`
	EndDate       *time.Time  `json:"endDate,omitempty" example:"2025-07-15T00:00:00Z"`
	CreatedAt     time.Time   `json:"createdAt" example:"2025-01-15T10:30:00Z"`
	UpdatedAt     time.Time   `json:"updatedAt" example:"2025-01-15T10:30:00Z"`
} //@name RallyResponse

// RallyJoinResponse represents the API response for GetRally, including user's role and status
type RallyJoinResponse struct {
	*RallyResponse
	CurrentUserRole   ParticipantRole     `json:"currentUserRole" example:"owner"`
	CurrentUserStatus ParticipationStatus `json:"currentUserStatus" example:"joined"`
} //@name RallyJoinResponse

// RallyListItem represents a simplified rally item for list views
type RallyListItem struct {
	ID        string      `json:"id" example:"507f1f77bcf86cd799439011"`
	Name      string      `json:"name" example:"Summer Road Trip"`
	Status    RallyStatus `json:"status" example:"active"`
	StartDate *time.Time  `json:"startDate,omitempty" example:"2025-07-01T00:00:00Z"`
	EndDate   *time.Time  `json:"endDate,omitempty" example:"2025-07-15T00:00:00Z"`
	UpdatedAt time.Time   `json:"updatedAt" example:"2025-01-15T10:30:00Z"`
} //@name RallyListItem

// RalliesListResponse represents the API response for rallies list
type RalliesListResponse struct {
	Rallies    []RallyListItem    `json:"rallies"`
	Total      int                `json:"total" example:"100"`
	Page       int                `json:"page" example:"1"`
	PageSize   int                `json:"pageSize" example:"20"`
	TotalPages int                `json:"totalPages" example:"5"`
	Pagination PaginationMetadata `json:"pagination"`
} //@name RalliesListResponse

// PaginationMetadata provides pagination information
type PaginationMetadata struct {
	HasNextPage     bool `json:"hasNextPage" example:"true"`
	HasPreviousPage bool `json:"hasPreviousPage" example:"false"`
} //@name PaginationMetadata
