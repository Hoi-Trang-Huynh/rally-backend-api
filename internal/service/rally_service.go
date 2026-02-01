package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RallyService struct {
	firebaseAuth    *auth.Client
	rallyRepo       repository.RallyRepository
	participantRepo repository.RallyParticipantRepository
	userRepo        repository.UserRepository
}

func NewRallyService(
	firebaseApp *firebase.App,
	rallyRepo repository.RallyRepository,
	participantRepo repository.RallyParticipantRepository,
	userRepo repository.UserRepository,
) (*RallyService, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &RallyService{
		firebaseAuth:    authClient,
		rallyRepo:       rallyRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}, nil
}

func (s *RallyService) authenticateUser(ctx context.Context, idToken string) (*model.User, error) {
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	user, err := s.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *RallyService) ValidateRallyAccess(ctx context.Context, userID primitive.ObjectID, rallyID string, requiredRoles []string) error {
	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return errors.New("invalid rally ID")
	}

	participant, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, rallyObjID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if participant == nil {
		return errors.New("unauthorized: not a participant of this rally")
	}

	for _, role := range requiredRoles {
		if participant.Role == role {
			return nil
		}
	}

	return errors.New("unauthorized: insufficient permissions")
}

// CreateRally creates a new rally and auto-adds the creator as owner participant
func (s *RallyService) CreateRally(ctx context.Context, idToken string, req *model.CreateRallyRequest) (*model.RallyResponse, error) {
	user, err := s.authenticateUser(ctx, idToken)
	if err != nil {
		return nil, err
	}

	rally := &model.Rally{
		ID:            primitive.NewObjectID(),
		OwnerID:       user.ID,
		Name:          req.Name,
		Description:   req.Description,
		CoverImageUrl: req.CoverImageUrl,
		Status:        "draft",
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
	}

	if err := s.rallyRepo.CreateRally(ctx, rally); err != nil {
		return nil, fmt.Errorf("failed to create rally: %w", err)
	}

	// Auto-add creator as owner participant
	now := time.Now()
	participant := &model.RallyParticipant{
		ID:        primitive.NewObjectID(),
		RallyID:   rally.ID,
		UserID:    user.ID,
		Role:      "owner",
		Status:    "joined",
		JoinedAt:  &now,
		InvitedAt: now,
	}

	if err := s.participantRepo.CreateParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to create owner participant: %w", err)
	}

	return s.ConvertToRallyResponse(rally), nil
}

// UpdateRally updates an existing rally (requires owner or editor role)
func (s *RallyService) UpdateRally(ctx context.Context, idToken string, rallyID string, req *model.UpdateRallyRequest) (*model.RallyResponse, error) {
	user, err := s.authenticateUser(ctx, idToken)
	if err != nil {
		return nil, err
	}

	if err := s.ValidateRallyAccess(ctx, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
		return nil, err
	}

	existing, err := s.rallyRepo.GetRallyByID(ctx, rallyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rally: %w", err)
	}
	if existing == nil {
		return nil, errors.New("rally not found")
	}

	updated, err := s.rallyRepo.UpdateRally(ctx, rallyID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update rally: %w", err)
	}

	return s.ConvertToRallyResponse(updated), nil
}

// ConvertToRallyResponse converts a Rally model to RallyResponse
func (s *RallyService) ConvertToRallyResponse(rally *model.Rally) *model.RallyResponse {
	return &model.RallyResponse{
		ID:            rally.ID.Hex(),
		OwnerID:       rally.OwnerID.Hex(),
		Name:          rally.Name,
		Description:   rally.Description,
		CoverImageUrl: rally.CoverImageUrl,
		Status:        rally.Status,
		StartDate:     rally.StartDate,
		EndDate:       rally.EndDate,
		CreatedAt:     rally.CreatedAt,
		UpdatedAt:     rally.UpdatedAt,
	}
}
