package service

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FollowService struct {
	firebaseAuth *auth.Client
	followRepo   repository.FollowRepository
	userRepo     repository.UserRepository
}

func NewFollowService(firebaseApp *firebase.App, followRepo repository.FollowRepository, userRepo repository.UserRepository) (*FollowService, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &FollowService{
		firebaseAuth: authClient,
		followRepo:   followRepo,
		userRepo:     userRepo,
	}, nil
}

// FollowUser creates a follow relationship between the authenticated user and target user
func (s *FollowService) FollowUser(ctx context.Context, idToken string, targetUserID string) (*model.FollowResponse, error) {
	// Verify Firebase token to get current user
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Get current user by Firebase UID
	currentUser, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	if currentUser == nil {
		return nil, errors.New("current user not found")
	}

	// Convert target user ID to ObjectID
	targetObjID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return nil, errors.New("invalid target user ID")
	}

	// Prevent self-follow
	if currentUser.ID == targetObjID {
		return nil, errors.New("cannot follow yourself")
	}

	// Check if target user exists
	targetUser, err := s.userRepo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}
	if targetUser == nil {
		return nil, errors.New("target user not found")
	}

	// Check if already following
	existingFollow, err := s.followRepo.GetFollow(ctx, currentUser.ID, targetObjID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing follow: %w", err)
	}
	if existingFollow != nil {
		return &model.FollowResponse{
			Success:     true,
			Message:     "Already following this user",
			IsFollowing: true,
		}, nil
	}

	// Create follow relationship
	_, err = s.followRepo.CreateFollow(ctx, currentUser.ID, targetObjID)
	if err != nil {
		return nil, fmt.Errorf("failed to create follow: %w", err)
	}

	// Increment target user's followers count
	if err := s.userRepo.IncrementFollowersCount(ctx, targetUserID); err != nil {
		return nil, fmt.Errorf("failed to increment followers count: %w", err)
	}

	// Increment current user's following count
	if err := s.userRepo.IncrementFollowingCount(ctx, currentUser.ID.Hex()); err != nil {
		return nil, fmt.Errorf("failed to increment following count: %w", err)
	}

	return &model.FollowResponse{
		Success:     true,
		Message:     "Successfully followed user",
		IsFollowing: true,
	}, nil
}

// UnfollowUser removes a follow relationship between the authenticated user and target user
func (s *FollowService) UnfollowUser(ctx context.Context, idToken string, targetUserID string) (*model.FollowResponse, error) {
	// Verify Firebase token to get current user
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Get current user by Firebase UID
	currentUser, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	if currentUser == nil {
		return nil, errors.New("current user not found")
	}

	// Convert target user ID to ObjectID
	targetObjID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return nil, errors.New("invalid target user ID")
	}

	// Check if follow relationship exists
	existingFollow, err := s.followRepo.GetFollow(ctx, currentUser.ID, targetObjID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing follow: %w", err)
	}
	if existingFollow == nil {
		return &model.FollowResponse{
			Success:     true,
			Message:     "Not following this user",
			IsFollowing: false,
		}, nil
	}

	// Delete follow relationship
	if err := s.followRepo.DeleteFollow(ctx, currentUser.ID, targetObjID); err != nil {
		return nil, fmt.Errorf("failed to delete follow: %w", err)
	}

	// Decrement target user's followers count
	if err := s.userRepo.DecrementFollowersCount(ctx, targetUserID); err != nil {
		return nil, fmt.Errorf("failed to decrement followers count: %w", err)
	}

	// Decrement current user's following count
	if err := s.userRepo.DecrementFollowingCount(ctx, currentUser.ID.Hex()); err != nil {
		return nil, fmt.Errorf("failed to decrement following count: %w", err)
	}

	return &model.FollowResponse{
		Success:     true,
		Message:     "Successfully unfollowed user",
		IsFollowing: false,
	}, nil
}

// IsFollowing checks if the authenticated user follows the target user
func (s *FollowService) IsFollowing(ctx context.Context, idToken string, targetUserID string) (*model.FollowStatusResponse, error) {
	// Verify Firebase token to get current user
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Get current user by Firebase UID
	currentUser, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	if currentUser == nil {
		return nil, errors.New("current user not found")
	}

	// Convert target user ID to ObjectID
	targetObjID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return nil, errors.New("invalid target user ID")
	}

	// Check if follow relationship exists
	existingFollow, err := s.followRepo.GetFollow(ctx, currentUser.ID, targetObjID)
	if err != nil {
		return nil, fmt.Errorf("failed to check follow status: %w", err)
	}

	return &model.FollowStatusResponse{
		IsFollowing: existingFollow != nil,
	}, nil
}

// GetUserPublicProfile retrieves the public profile of a user including follow counts
func (s *FollowService) GetUserPublicProfile(ctx context.Context, userID string) (*model.UserPublicProfileResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &model.UserPublicProfileResponse{
		ID:             user.ID.Hex(),
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		AvatarUrl:      user.AvatarUrl,
		BioText:        user.BioText,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
	}, nil
}
