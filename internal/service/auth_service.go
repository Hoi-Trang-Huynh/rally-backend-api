package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// CompleteRegistration applies the optional initial profile fields to a
// freshly resolved user. The user itself is provisioned by the auth
// middleware from the verified Firebase token, so this is idempotent: calling
// it again for an existing user without profile fields is a no-op login.
func (s *AuthService) CompleteRegistration(ctx context.Context, user *model.User, req *model.RegisterRequest) (*model.User, error) {
	if req == nil || (req.Username == nil && req.FirstName == nil && req.LastName == nil) {
		return user, nil
	}

	updates := &model.ProfileUpdateRequest{
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updated, err := s.userRepo.UpdateUserProfile(ctx, user.ID.Hex(), updates)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("username is already taken")
		}
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return updated, nil
}

// CheckEmailAvailability checks if an email is available for registration
func (s *AuthService) CheckEmailAvailability(ctx context.Context, email string) (bool, error) {
	exists, err := s.userRepo.ExistsEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email availability: %w", err)
	}
	return !exists, nil
}

// CheckUsernameAvailability checks if a username is available
func (s *AuthService) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	exists, err := s.userRepo.ExistsUsername(ctx, username)
	if err != nil {
		return false, fmt.Errorf("failed to check username availability: %w", err)
	}
	return !exists, nil
}
