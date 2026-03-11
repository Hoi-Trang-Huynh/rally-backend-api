package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Feedback struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username"`
	AvatarUrl      string             `bson:"avatar_url" json:"avatarUrl"`
	AttachmentURLs []string           `bson:"attachment_urls" json:"attachmentUrls"`
	Comment        string             `bson:"comment" json:"comment"`
	Categories     []string           `bson:"categories" json:"categories"`
	Resolved       bool               `bson:"resolved" json:"resolved"`
	CreatedAt      time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updatedAt"`
}

type CreateFeedbackRequest struct {
	Username       string   `json:"username" validate:"required"`
	AvatarUrl      string   `json:"avatarUrl"`
	AttachmentURLs []string `json:"attachmentUrls"`
	Comment        string   `json:"comment" validate:"required"`
	Categories     []string `json:"categories"`
}

type FeedbackListResponse struct {
	Feedbacks  []Feedback `json:"feedbacks"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

type UpdateFeedbackStatusRequest struct {
	Resolved bool `json:"resolved"`
}
