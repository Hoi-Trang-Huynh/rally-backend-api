package service

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
)

type AuthService struct {
	firebaseAuth *auth.Client
	userRepo     repository.UserRepository
}

func NewAuthService(firebaseApp *firebase.App, userRepo repository.UserRepository) (*AuthService, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &AuthService{
		firebaseAuth: authClient,
		userRepo:     userRepo,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, idToken string) (*model.User, error) {
	// Verify Firebase token
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Fetch user from DB
	user, err := s.userRepo.GetUserByID(ctx, token.UID)
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}
