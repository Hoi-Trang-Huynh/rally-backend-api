package service

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
)

type UserService struct {
	firebaseAuth *auth.Client
	userRepo     repository.UserRepository
}

func NewUserService(firebaseApp *firebase.App, userRepo repository.UserRepository) (*UserService, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &UserService{
		firebaseAuth: authClient,
		userRepo:     userRepo,
	}, nil
}

// GetUserProfile retrieves user profile by user ID (MongoDB ObjectID string)
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserProfileByToken retrieves user profile by Firebase ID token
func (s *UserService) GetUserProfileByToken(ctx context.Context, idToken string) (*model.User, error) {
	// Verify Firebase token
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Get user by Firebase UID
	user, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUserProfile updates user profile information
func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, updates *model.ProfileUpdateRequest) (*model.User, error) {
	// Check if user exists
	existingUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Update the profile
	updatedUser, err := s.userRepo.UpdateUserProfile(ctx, userID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return updatedUser, nil
}

// ValidateUserOwnership validates that the Firebase token belongs to the user being modified
func (s *UserService) ValidateUserOwnership(ctx context.Context, idToken, userID string) error {
	// Verify Firebase token
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// Get user by Firebase UID
	user, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if the user ID matches (convert ObjectID to string for comparison)
	if user.ID.Hex() != userID {
		return errors.New("unauthorized: cannot modify another user's profile")
	}

	return nil
}

// ConvertToProfileResponse converts User model to ProfileResponse (for syncing)
func (s *UserService) ConvertToProfileResponse(user *model.User) *model.ProfileResponse {
	return &model.ProfileResponse{
		ID:              user.ID.Hex(), // Convert ObjectID to string
		Email:           user.Email,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		AvatarUrl:       user.AvatarUrl,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		IsOnboarding:    user.IsOnboarding,
	}
}

// ConvertToProfileDetailsResponse converts User model to ProfileDetailsResponse (for profile page)
func (s *UserService) ConvertToProfileDetailsResponse(user *model.User) *model.ProfileDetailsResponse {
	return &model.ProfileDetailsResponse{
		ID:             user.ID.Hex(),
		BioText:        user.BioText,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
	}
}

// SearchUsers searches for users by query string with pagination
func (s *UserService) SearchUsers(ctx context.Context, query string, page, pageSize int) (*model.UserSearchResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50 // Max page size to prevent abuse
	}

	users, total, err := s.userRepo.SearchUsers(ctx, query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert users to search results
	results := make([]model.UserSearchResult, len(users))
	for i, user := range users {
		results[i] = model.UserSearchResult{
			ID:        user.ID.Hex(),
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			AvatarUrl: user.AvatarUrl,
		}
	}

	// Calculate total pages
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.UserSearchResponse{
		Users:      results,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
