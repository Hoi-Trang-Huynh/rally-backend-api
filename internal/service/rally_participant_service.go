package service

import (
	"context"
	"errors"
	"fmt"

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
	firebaseAuth *auth.Client,
	participantRepo repository.RallyParticipantRepository,
	rallyRepo repository.RallyRepository,
	userRepo repository.UserRepository,
) *RallyParticipantService {
	return &RallyParticipantService{
		firebaseAuth:    firebaseAuth,
		participantRepo: participantRepo,
		rallyRepo:       rallyRepo,
		userRepo:        userRepo,
	}
}

// InviteParticipant invites a user to a rally (requires owner or editor role)
func (s *RallyParticipantService) InviteParticipant(ctx context.Context, idToken string, rallyID string, req *model.InviteParticipantRequest) (*model.RallyParticipantResponse, error) {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
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
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
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
		if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner"}); err != nil {
			return nil, err
		}
	}

	// Status changes: allowed for the participant themselves or owner/editor
	if req.Status != nil {
		isSelf := user.ID == participant.UserID
		if !isSelf {
			if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
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

// GetParticipantsList retrieves a paginated list of participants for a given rally
func (s *RallyParticipantService) GetParticipantsList(ctx context.Context, idToken string, rallyID string, role string, page, pageSize int) (*model.ParticipantListResponse, error) {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}

	participant, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, rallyObjID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check participation: %w", err)
	}

	if participant == nil {
		return nil, errors.New("unauthorized: not a participant of this rally")
	}

	if participant.Status != model.ParticipationStatusJoined {
		return nil, errors.New("unauthorized: you have been invited but not yet joined this rally")
	}

	participants, total, err := s.participantRepo.GetParticipantsList(ctx, rallyObjID, role, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants list: %w", err)
	}

	// Calculate pagination using utils (assuming utils is available, adjust if not, or just handle manually)
	// If utils is not imported, let's do it manually:
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return &model.ParticipantListResponse{
		Participants: participants,
		Total:        int(total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
		Pagination: model.PaginationMetadata{
			HasNextPage:     page < totalPages,
			HasPreviousPage: page > 1,
		},
	}, nil
}
