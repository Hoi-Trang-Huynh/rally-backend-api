package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	FirebaseUID string             `json:"firebaseUid" bson:"firebase_uid"`
	Email       string             `json:"email" bson:"email"`
	DisplayName string             `json:"displayName" bson:"display_name"`
	FirstName   string             `json:"firstName" bson:"first_name"`
	LastName    string             `json:"lastName" bson:"last_name"`
	ProfilePic  string             `json:"profilePic" bson:"profile_pic"`
	Bio         string             `json:"bio" bson:"bio"`
	Phone       string             `json:"phone" bson:"phone"`
	DateOfBirth *time.Time         `json:"dateOfBirth" bson:"date_of_birth"`
	Location    string             `json:"location" bson:"location"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// ProfileUpdateRequest represents the request payload for updating user profile
type ProfileUpdateRequest struct {
	DisplayName *string    `json:"displayName,omitempty"`
	FirstName   *string    `json:"firstName,omitempty"`
	LastName    *string    `json:"lastName,omitempty"`
	ProfilePic  *string    `json:"profilePic,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	Location    *string    `json:"location,omitempty"`
} //@name ProfileUpdateRequest

// ProfileResponse represents the user profile response
type ProfileResponse struct {
	ID          string     `json:"id" example:"507f1f77bcf86cd799439011"`
	Email       string     `json:"email" example:"john@example.com"`
	DisplayName string     `json:"displayName,omitempty" example:"John Doe"`
	FirstName   string     `json:"firstName,omitempty" example:"John"`
	LastName    string     `json:"lastName,omitempty" example:"Doe"`
	ProfilePic  string     `json:"profilePic,omitempty" example:"https://example.com/profile.jpg"`
	Bio         string     `json:"bio,omitempty" example:"Software developer passionate about rally racing"`
	Phone       string     `json:"phone,omitempty" example:"+1234567890"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty" example:"1990-01-15T00:00:00Z"`
	Location    string     `json:"location,omitempty" example:"San Francisco, CA"`
	CreatedAt   time.Time  `json:"createdAt" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time  `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
} //@name ProfileResponse
