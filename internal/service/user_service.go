package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
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
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("username is already taken")
		}
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return updatedUser, nil
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
	page, pageSize = utils.ClampPagination(page, pageSize, 50)

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

	totalPages := utils.CalcTotalPages(total, pageSize)

	return &model.UserSearchResponse{
		Users:      results,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
