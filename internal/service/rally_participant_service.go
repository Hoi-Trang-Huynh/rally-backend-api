package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RallyParticipantService struct {
	participantRepo repository.RallyParticipantRepository
	rallyRepo       repository.RallyRepository
	userRepo        repository.UserRepository
	followRepo      repository.FollowRepository
}

func NewRallyParticipantService(
	participantRepo repository.RallyParticipantRepository,
	rallyRepo repository.RallyRepository,
	userRepo repository.UserRepository,
	followRepo repository.FollowRepository,
) *RallyParticipantService {
	return &RallyParticipantService{
		participantRepo: participantRepo,
		rallyRepo:       rallyRepo,
		userRepo:        userRepo,
		followRepo:      followRepo,
	}
}

// InviteParticipant invites a user to a rally (middleware ensures owner or editor role)
func (s *RallyParticipantService) InviteParticipant(ctx context.Context, user *model.User, rallyID string, req *model.InviteParticipantRequest) (*model.RallyParticipantResponse, error) {

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
	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}
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
func (s *RallyParticipantService) UpdateParticipant(ctx context.Context, user *model.User, callerParticipant *model.RallyParticipant, rallyID string, participantID string, req *model.UpdateParticipantRequest) (*model.RallyParticipantResponse, error) {
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
		if callerParticipant.Status != model.ParticipationStatusJoined || callerParticipant.Role != model.ParticipantRoleOwner {
			return nil, errors.New("unauthorized: only owners can change roles")
		}
	}

	// Status changes: allowed for the participant themselves or owner/editor
	if req.Status != nil {
		isSelf := user.ID == participant.UserID
		if !isSelf {
			if callerParticipant.Status != model.ParticipationStatusJoined {
				return nil, errors.New("unauthorized: participant status is not active")
			}
			if callerParticipant.Role != model.ParticipantRoleOwner && callerParticipant.Role != model.ParticipantRoleEditor {
				return nil, errors.New("unauthorized: insufficient permissions")
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

// GetParticipantsList retrieves a paginated list of participants for a given rally (middleware ensures joined participant)
func (s *RallyParticipantService) GetParticipantsList(ctx context.Context, rallyID string, role string, page, pageSize int) (*model.ParticipantListResponse, error) {
	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}

	participants, total, err := s.participantRepo.GetParticipantsList(ctx, rallyObjID, role, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants list: %w", err)
	}

	totalPages := utils.CalcTotalPages(total, pageSize)

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

// GetPendingInvitations retrieves all pending ("invited" status) invitations for the current user.
// TODO: This is a temporary endpoint until realtime notifications are implemented.
func (s *RallyParticipantService) GetPendingInvitations(ctx context.Context, user *model.User) (*model.PendingInvitationsResponse, error) {
	items, err := s.participantRepo.GetPendingInvitations(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}

	return &model.PendingInvitationsResponse{
		Invitations: items,
		Total:       len(items),
	}, nil
}

// GetInvitableFriends retrieves friends who can be invited to a rally (middleware ensures joined participant)
func (s *RallyParticipantService) GetInvitableFriends(ctx context.Context, user *model.User, rallyID string, query string, page, pageSize int) (*model.FriendListResponse, error) {
	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}

	// Verify rally exists
	rally, err := s.rallyRepo.GetRallyByID(ctx, rallyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rally: %w", err)
	}
	if rally == nil {
		return nil, errors.New("rally not found")
	}

	users, total, err := s.followRepo.GetInvitableFriends(ctx, user.ID, rallyObjID, query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitable friends: %w", err)
	}

	totalPages := utils.CalcTotalPages(total, pageSize)

	return &model.FriendListResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
