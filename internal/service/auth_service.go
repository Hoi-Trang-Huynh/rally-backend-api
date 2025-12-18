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

func (s *AuthService) RegisterOrLogin(ctx context.Context, idToken string) (*model.User, error) {
	// Verify the Firebase ID token
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired Firebase token")
	}

	// Extract email from token
	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email not found in token claims")
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err == nil && existingUser != nil {
		// User exists, return it
		return existingUser, nil
	}

	// User doesn't exist, create new user
	newUser := &model.User{
		ID:              primitive.NewObjectID(), // MongoDB will generate this automatically in CreateUser if zero
		FirebaseUID:     token.UID,
		Email:           email,
		IsActive:        true,
		IsEmailVerified: false,
		IsOnboarding:    true,
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

func (s *AuthService) Login(ctx context.Context, idToken string) (*model.User, error) {
	// Verify Firebase token
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Fetch user from DB
	user, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}
