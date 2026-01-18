package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	FirebaseUID     string             `json:"firebaseUid" bson:"firebase_uid"`
	Email           string             `json:"email" bson:"email"`
	Username        string             `json:"username" bson:"username"`
	FirstName       string             `json:"firstName" bson:"first_name"`
	LastName        string             `json:"lastName" bson:"last_name"`
	AvatarUrl       string             `json:"avatarUrl" bson:"avatar_url"`
	BioText         string             `json:"bioText" bson:"bio_text"`
	PhoneNumber     string             `json:"phoneNumber" bson:"phone_number"`
	CreatedAt       time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updated_at"`
	IsActive        bool               `json:"isActive" bson:"is_active"`
	IsEmailVerified bool               `json:"isEmailVerified" bson:"is_email_verified"`
	IsOnboarding    bool               `json:"isOnboarding" bson:"is_onboarding"`
	FollowersCount  int                `json:"followersCount" bson:"followers_count"`
	FollowingCount  int                `json:"followingCount" bson:"following_count"`
}

// ProfileUpdateRequest represents the request payload for updating user profile
type ProfileUpdateRequest struct {
	Username        *string `json:"username,omitempty"`
	FirstName       *string `json:"firstName,omitempty"`
	LastName        *string `json:"lastName,omitempty"`
	AvatarUrl       *string `json:"avatarUrl,omitempty"`
	BioText         *string `json:"bioText,omitempty"`
	PhoneNumber     *string `json:"phoneNumber,omitempty"`
	IsActive        *bool   `json:"isActive,omitempty"`
	IsEmailVerified *bool   `json:"isEmailVerified,omitempty"`
	IsOnboarding    *bool   `json:"isOnboarding,omitempty"`
} //@name ProfileUpdateRequest

// ProfileResponse represents the user profile response (for syncing)
type ProfileResponse struct {
	ID              string    `json:"id" example:"507f1f77bcf86cd799439011"`
	Email           string    `json:"email" example:"john@example.com"`
	Username        string    `json:"username,omitempty" example:"johndoe"`
	FirstName       string    `json:"firstName,omitempty" example:"John"`
	LastName        string    `json:"lastName,omitempty" example:"Doe"`
	AvatarUrl       string    `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
	CreatedAt       time.Time `json:"createdAt" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       time.Time `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
	IsActive        bool      `json:"isActive" example:"true"`
	IsEmailVerified bool      `json:"isEmailVerified" example:"true"`
	IsOnboarding    bool      `json:"isOnboarding" example:"false"`
} //@name ProfileResponse

// ProfileDetailsResponse represents the profile details for profile page view
type ProfileDetailsResponse struct {
	ID             string `json:"id" example:"507f1f77bcf86cd799439011"`
	BioText        string `json:"bioText,omitempty" example:"Software developer passionate about rally racing"`
	FollowersCount int    `json:"followersCount" example:"150"`
	FollowingCount int    `json:"followingCount" example:"75"`
} //@name ProfileDetailsResponse

// UserPublicProfileResponse represents the public profile for viewing other users
type UserPublicProfileResponse struct {
	ID             string `json:"id" example:"507f1f77bcf86cd799439011"`
	Username       string `json:"username" example:"johndoe"`
	FirstName      string `json:"firstName,omitempty" example:"John"`
	LastName       string `json:"lastName,omitempty" example:"Doe"`
	AvatarUrl      string `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
	BioText        string `json:"bioText,omitempty" example:"Software developer passionate about rally racing"`
	FollowersCount int    `json:"followersCount" example:"150"`
	FollowingCount int    `json:"followingCount" example:"75"`
} //@name UserPublicProfileResponse

// UserSearchResult represents a single user in search results
type UserSearchResult struct {
	ID        string `json:"id" example:"507f1f77bcf86cd799439011"`
	Username  string `json:"username" example:"johndoe"`
	FirstName string `json:"firstName,omitempty" example:"John"`
	LastName  string `json:"lastName,omitempty" example:"Doe"`
	AvatarUrl string `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
} //@name UserSearchResult

// UserSearchResponse represents the paginated search response
type UserSearchResponse struct {
	Users      []UserSearchResult `json:"users"`
	Total      int64              `json:"total" example:"100"`
	Page       int                `json:"page" example:"1"`
	PageSize   int                `json:"pageSize" example:"20"`
	TotalPages int                `json:"totalPages" example:"5"`
} //@name UserSearchResponse
