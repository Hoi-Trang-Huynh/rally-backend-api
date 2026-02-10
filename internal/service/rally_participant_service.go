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

type RallyParticipantService struct {
	firebaseAuth    *auth.Client
	participantRepo repository.RallyParticipantRepository
	rallyRepo       repository.RallyRepository
	userRepo        repository.UserRepository
}

func NewRallyParticipantService(
	firebaseApp *firebase.App,
	participantRepo repository.RallyParticipantRepository,
	rallyRepo repository.RallyRepository,
	userRepo repository.UserRepository,
) (*RallyParticipantService, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &RallyParticipantService{
		firebaseAuth:    authClient,
		participantRepo: participantRepo,
		rallyRepo:       rallyRepo,
		userRepo:        userRepo,
	}, nil
}

func (s *RallyParticipantService) authenticateUser(ctx context.Context, idToken string) (*model.User, error) {
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

func (s *RallyParticipantService) validateRallyAccess(ctx context.Context, userID primitive.ObjectID, rallyID string, requiredRoles []string) error {
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
		if string(participant.Role) == role {
			return nil
		}
	}

	return errors.New("unauthorized: insufficient permissions")
}

// InviteParticipant invites a user to a rally (requires owner or editor role)
func (s *RallyParticipantService) InviteParticipant(ctx context.Context, idToken string, rallyID string, req *model.InviteParticipantRequest) (*model.RallyParticipantResponse, error) {
	user, err := s.authenticateUser(ctx, idToken)
	if err != nil {
		return nil, err
	}

	if err := s.validateRallyAccess(ctx, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
		return nil, err
	}

	// Verify rally exists
	rally, err := s.rallyRepo.GetRallyByID(ctx, rallyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rally: %w", err)
	}
	if rally == nil {
		return nil, errors.New("rally not found")
	}

	// Verify target user exists
	targetUser, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}
	if targetUser == nil {
		return nil, errors.New("target user not found")
	}

	// Check for duplicate participant
	rallyObjID, _ := primitive.ObjectIDFromHex(rallyID)
	existing, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, rallyObjID, targetUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing participant: %w", err)
	}
	if existing != nil {
		return nil, errors.New("user is already a participant")
	}

	// Default role to "participant" if not specified
	var role model.ParticipantRole
	if req.Role != nil && *req.Role != "" {
		role = *req.Role
	} else {
		role = model.ParticipantRoleParticipant
	}

	participant := &model.RallyParticipant{
		ID:        primitive.NewObjectID(),
		RallyID:   rallyObjID,
		UserID:    targetUser.ID,
		Role:      role,
		Status:    model.ParticipationStatusInvited,
		InvitedBy: &user.ID,
	}

	if err := s.participantRepo.CreateParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to create participant: %w", err)
	}

	return s.ConvertToParticipantResponse(participant), nil
}

// UpdateParticipant updates a participant's role or status
func (s *RallyParticipantService) UpdateParticipant(ctx context.Context, idToken string, rallyID string, participantID string, req *model.UpdateParticipantRequest) (*model.RallyParticipantResponse, error) {
	user, err := s.authenticateUser(ctx, idToken)
	if err != nil {
		return nil, err
	}

	// Get the participant being updated
	participant, err := s.participantRepo.GetParticipant(ctx, participantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil {
		return nil, errors.New("participant not found")
	}

	// Role changes require owner permission
	if req.Role != nil {
		if err := s.validateRallyAccess(ctx, user.ID, rallyID, []string{"owner"}); err != nil {
			return nil, err
		}
	}

	// Status changes: allowed for the participant themselves or owner/editor
	if req.Status != nil {
		isSelf := user.ID == participant.UserID
		if !isSelf {
			if err := s.validateRallyAccess(ctx, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
				return nil, err
			}
		}
	}

	updated, err := s.participantRepo.UpdateParticipant(ctx, participantID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update participant: %w", err)
	}

	return s.ConvertToParticipantResponse(updated), nil
}

// ConvertToParticipantResponse converts a RallyParticipant model to RallyParticipantResponse
func (s *RallyParticipantService) ConvertToParticipantResponse(p *model.RallyParticipant) *model.RallyParticipantResponse {
	invitedBy := ""
	if p.InvitedBy != nil {
		invitedBy = p.InvitedBy.Hex()
	}

	return &model.RallyParticipantResponse{
		ID:        p.ID.Hex(),
		RallyID:   p.RallyID.Hex(),
		UserID:    p.UserID.Hex(),
		Role:      p.Role,
		Status:    p.Status,
		InvitedBy: invitedBy,
		JoinedAt:  p.JoinedAt,
		InvitedAt: p.InvitedAt,
	}
}
