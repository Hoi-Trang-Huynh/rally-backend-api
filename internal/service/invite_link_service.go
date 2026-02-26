package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InviteLinkService struct {
	firebaseAuth    *auth.Client
	inviteLinkRepo  repository.InviteLinkRepository
	participantRepo repository.RallyParticipantRepository
	rallyRepo       repository.RallyRepository
	userRepo        repository.UserRepository
}

// NewInviteLinkService initializes a new InviteLinkService
func NewInviteLinkService(
	firebaseAuth *auth.Client,
	inviteLinkRepo repository.InviteLinkRepository,
	participantRepo repository.RallyParticipantRepository,
	rallyRepo repository.RallyRepository,
	userRepo repository.UserRepository,
) *InviteLinkService {
	return &InviteLinkService{
		firebaseAuth:    firebaseAuth,
		inviteLinkRepo:  inviteLinkRepo,
		participantRepo: participantRepo,
		rallyRepo:       rallyRepo,
		userRepo:        userRepo,
	}
}

// CreateInviteLink generates a new invite link token for a rally
func (s *InviteLinkService) CreateInviteLink(ctx context.Context, idToken string, rallyID string, req *model.CreateInviteLinkRequest) (*model.InviteLinkResponse, error) {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	// Verify user has owner/editor role in the rally
	if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
		return nil, err
	}

	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}

	// Only owners can create owner/editor links
	if req.Role == model.ParticipantRoleOwner || req.Role == model.ParticipantRoleEditor {
		if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner"}); err != nil {
			return nil, errors.New("only owners can create links for owner/editor roles")
		}
	}

	// Set defaults
	roleToGrant := model.ParticipantRoleParticipant
	if req.Role != "" {
		roleToGrant = req.Role
	}

	token := uuid.New().String()

	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		exp := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &exp
	}

	link := &model.InviteLink{
		ID:          primitive.NewObjectID(),
		RallyID:     rallyObjID,
		CreatedBy:   user.ID,
		Token:       token,
		RoleToGrant: roleToGrant,
		ExpiresAt:   expiresAt,
		MaxUses:     req.MaxUses,
		CurrentUses: 0,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := s.inviteLinkRepo.CreateInviteLink(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	return convertToInviteLinkResponse(link), nil
}

// GetActiveInviteLinks retrieves all active invite links for a rally
func (s *InviteLinkService) GetActiveInviteLinks(ctx context.Context, idToken string, rallyID string) ([]*model.InviteLinkResponse, error) {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	// Require owner/editor to list links
	if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
		return nil, err
	}

	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}

	links, err := s.inviteLinkRepo.GetActiveInviteLinksByRally(ctx, rallyObjID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active invite links: %w", err)
	}

	responses := make([]*model.InviteLinkResponse, len(links))
	for i, link := range links {
		responses[i] = convertToInviteLinkResponse(link)
	}

	return responses, nil
}

// DeactivateInviteLink revokes an existing invite link
func (s *InviteLinkService) DeactivateInviteLink(ctx context.Context, idToken string, rallyID string, token string) error {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return err
	}

	// Require owner/editor to revoke links
	if err := validateRallyAccess(ctx, s.participantRepo, user.ID, rallyID, []string{"owner", "editor"}); err != nil {
		return err
	}

	// Optionally check if the link belongs to this rally
	link, err := s.inviteLinkRepo.GetInviteLinkByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil || !link.IsActive {
		return errors.New("link not found or already inactive")
	}
	if link.RallyID.Hex() != rallyID {
		return errors.New("link does not belong to this rally")
	}

	// Owners can revoke any link, editors can only revoke links they created or lower tier links
	// Assuming simple validation for now: owner/editor can revoke. Add strict checks if necessary.

	if err := s.inviteLinkRepo.DeactivateInviteLink(ctx, token); err != nil {
		return fmt.Errorf("failed to deactivate link: %w", err)
	}

	return nil
}

// JoinViaLink allows a user to join a rally using a valid invite link token
func (s *InviteLinkService) JoinViaLink(ctx context.Context, idToken string, token string) (*model.JoinViaLinkResponse, error) {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	// Find the link
	link, err := s.inviteLinkRepo.GetInviteLinkByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to check invite link: %w", err)
	}
	if link == nil || !link.IsActive {
		return nil, errors.New("invalid or expired invite link")
	}

	// Validate Expiration
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		// Attempt to deactivate it since it's expired
		_ = s.inviteLinkRepo.DeactivateInviteLink(ctx, token)
		return nil, errors.New("invite link has expired")
	}

	// Validate Max Uses
	if link.MaxUses > 0 && link.CurrentUses >= link.MaxUses {
		_ = s.inviteLinkRepo.DeactivateInviteLink(ctx, token)
		return nil, errors.New("invite link has reached its maximum number of uses")
	}

	// Increment uses now (to prevent race conditions overriding max limit, but a distributed lock would be better)
	if err := s.inviteLinkRepo.IncrementLinkUsage(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to process invite link usage: %w", err)
	}

	// Check if already a participant
	existing, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, link.RallyID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing participant status: %w", err)
	}

	if existing != nil {
		if existing.Status == model.ParticipationStatusJoined {
			// Already joined, let it succeed but state so
			return &model.JoinViaLinkResponse{
				Success: true,
				Message: "You are already a participant of this rally",
				RallyID: link.RallyID.Hex(),
				Role:    string(existing.Role),
			}, nil
		}

		if existing.Status == model.ParticipationStatusInvited {
			return &model.JoinViaLinkResponse{
				Success: true,
				Message: "You have successfully received an invitation to this rally. Please confirm to join.",
				RallyID: link.RallyID.Hex(),
				Role:    string(existing.Role), // keep existing role
			}, nil
		}

		// Update existing participant status to Invited (if they had Left or Declined previously)
		status := model.ParticipationStatusInvited
		updates := &model.UpdateParticipantRequest{
			Status: &status,
			Role:   &link.RoleToGrant, // Inherit role from the link
		}

		_, err := s.participantRepo.UpdateParticipant(ctx, existing.ID.Hex(), updates)
		if err != nil {
			return nil, fmt.Errorf("failed to process invite: %w", err)
		}
	} else {
		// Make a new participant record with INVITED status
		participant := &model.RallyParticipant{
			ID:        primitive.NewObjectID(),
			RallyID:   link.RallyID,
			UserID:    user.ID,
			Role:      link.RoleToGrant,
			Status:    model.ParticipationStatusInvited,
			InvitedBy: &link.CreatedBy,
		}

		if err := s.participantRepo.CreateParticipant(ctx, participant); err != nil {
			// Revert the link usage? Or just accept some minor inaccuracy
			return nil, fmt.Errorf("failed to process invite: %w", err)
		}
	}

	return &model.JoinViaLinkResponse{
		Success: true,
		Message: "Successfully received invitation. Please confirm to join.",
		RallyID: link.RallyID.Hex(),
		Role:    string(link.RoleToGrant),
	}, nil
}

func convertToInviteLinkResponse(link *model.InviteLink) *model.InviteLinkResponse {
	return &model.InviteLinkResponse{
		ID:          link.ID.Hex(),
		RallyID:     link.RallyID.Hex(),
		CreatedBy:   link.CreatedBy.Hex(),
		Token:       link.Token,
		RoleToGrant: link.RoleToGrant,
		ExpiresAt:   link.ExpiresAt,
		MaxUses:     link.MaxUses,
		CurrentUses: link.CurrentUses,
		IsActive:    link.IsActive,
		CreatedAt:   link.CreatedAt,
	}
}
