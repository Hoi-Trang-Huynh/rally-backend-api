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
	eventRepo       repository.EventRepository
}

// NewInviteLinkService initializes a new InviteLinkService
func NewInviteLinkService(
	firebaseAuth *auth.Client,
	inviteLinkRepo repository.InviteLinkRepository,
	participantRepo repository.RallyParticipantRepository,
	rallyRepo repository.RallyRepository,
	userRepo repository.UserRepository,
	eventRepo repository.EventRepository,
) *InviteLinkService {
	return &InviteLinkService{
		firebaseAuth:    firebaseAuth,
		inviteLinkRepo:  inviteLinkRepo,
		participantRepo: participantRepo,
		rallyRepo:       rallyRepo,
		userRepo:        userRepo,
		eventRepo:       eventRepo,
	}
}

// CreateInviteLink generates a new invite link token for a rally (middleware ensures owner/editor)
func (s *InviteLinkService) CreateInviteLink(ctx context.Context, user *model.User, callerParticipant *model.RallyParticipant, rallyID string, req *model.CreateInviteLinkRequest) (*model.InviteLinkResponse, error) {
	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, errors.New("invalid rally ID")
	}

	// Only owners can create owner/editor links
	if req.Role == model.ParticipantRoleOwner || req.Role == model.ParticipantRoleEditor {
		if callerParticipant.Role != model.ParticipantRoleOwner {
			return nil, errors.New("only owners can create links for owner/editor roles")
		}
	}

	// Set defaults
	roleToGrant := model.ParticipantRoleParticipant
	if req.Role != "" {
		roleToGrant = req.Role
	}

	token := uuid.New().String()

	// Default to 7 days expiration
	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		exp := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &exp
	} else {
		exp := time.Now().AddDate(0, 0, 7)
		expiresAt = &exp
	}

	// Default to 5 max uses
	maxUses := 5
	if req.MaxUses > 0 {
		maxUses = req.MaxUses
	}

	link := &model.InviteLink{
		ID:          primitive.NewObjectID(),
		RallyID:     rallyObjID,
		CreatedBy:   user.ID,
		Token:       token,
		RoleToGrant: roleToGrant,
		ExpiresAt:   expiresAt,
		MaxUses:     maxUses,
		CurrentUses: 0,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := s.inviteLinkRepo.CreateInviteLink(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	return convertToInviteLinkResponse(link), nil
}

// GetActiveInviteLinks retrieves all active invite links for a rally (middleware ensures owner/editor)
func (s *InviteLinkService) GetActiveInviteLinks(ctx context.Context, rallyID string) ([]*model.InviteLinkResponse, error) {

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

// DeactivateInviteLink revokes an existing invite link (middleware ensures owner/editor)
func (s *InviteLinkService) DeactivateInviteLink(ctx context.Context, rallyID string, token string) error {

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

// PreviewInviteLink gets details about an invitation link for a preview card
func (s *InviteLinkService) PreviewInviteLink(ctx context.Context, idToken string, token string) (*model.InviteLinkPreviewResponse, error) {
	user, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	link, err := s.inviteLinkRepo.GetInviteLinkByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil || !link.IsActive {
		return nil, errors.New("link is invalid or inactive")
	}

	// We will validate expiration and uses ONLY IF the user is not already joined or invited.
	// So we need to check participation status first.

	isExpired := link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now())
	isMaxedOut := link.MaxUses > 0 && link.CurrentUses >= link.MaxUses

	// Fetch rally details
	rally, err := s.rallyRepo.GetRallyByID(ctx, link.RallyID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get rally: %w", err)
	}
	if rally == nil {
		return nil, errors.New("rally not found")
	}

	// Check if already a participant (status=joined)
	existing, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, link.RallyID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing participation: %w", err)
	}

	// If they are not already joined/invited, enforce expiration/max uses
	if existing == nil || (existing.Status != model.ParticipationStatusJoined && existing.Status != model.ParticipationStatusInvited) {
		if isExpired {
			return nil, errors.New("link is expired")
		}
		if isMaxedOut {
			return nil, errors.New("link has reached its maximum uses")
		}
	}
	// Note: already-joined users can still preview (frontend will redirect them)

	// Get owner info
	owner, err := s.userRepo.GetUserByID(ctx, rally.OwnerID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get rally owner: %w", err)
	}
	if owner == nil {
		return nil, errors.New("rally owner not found")
	}

	// Get member and event counts
	memberCount, err := s.participantRepo.CountJoinedParticipants(ctx, rally.ID)
	if err != nil {
		memberCount = 1 // Default to 1 (the owner) if count fails
	}

	eventCount, err := s.eventRepo.CountEventsByRally(ctx, rally.ID)
	if err != nil {
		eventCount = 0
	}

	participantID := ""
	participantStatus := ""
	if existing != nil {
		participantID = existing.ID.Hex()
		participantStatus = string(existing.Status)
	}

	return &model.InviteLinkPreviewResponse{
		RallyID:       rally.ID.Hex(),
		RallyName:     rally.Name,
		Description:   rally.Description,
		CoverImageUrl: rally.CoverImageUrl,
		Status:        rally.Status,
		StartDate:     rally.StartDate,
		EndDate:       rally.EndDate,
		Owner: model.InviteLinkPreviewOwner{
			Username:  owner.Username,
			FirstName: owner.FirstName,
			LastName:  owner.LastName,
			AvatarUrl: owner.AvatarUrl,
		},
		RoleOffered:       link.RoleToGrant,
		MemberCount:       memberCount,
		EventCount:        eventCount,
		ParticipantID:     participantID,
		ParticipantStatus: participantStatus,
		InvitedAt:         link.CreatedAt,
	}, nil
}

// JoinViaLink allows a user to accept an invite link and join a rally directly.
// This is called when the user taps "Accept" on the invite preview screen.
// It creates or updates the participant record with "joined" status and increments link usage.
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
		_ = s.inviteLinkRepo.DeactivateInviteLink(ctx, token)
		return nil, errors.New("invite link has expired")
	}

	// Validate Max Uses
	if link.MaxUses > 0 && link.CurrentUses >= link.MaxUses {
		_ = s.inviteLinkRepo.DeactivateInviteLink(ctx, token)
		return nil, errors.New("invite link has reached its maximum number of uses")
	}

	// Check if already a participant
	existing, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, link.RallyID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing participant status: %w", err)
	}

	now := time.Now()
	joinedStatus := model.ParticipationStatusJoined

	if existing != nil {
		if existing.Status == model.ParticipationStatusJoined {
			// Already joined — no-op, no usage increment
			return &model.JoinViaLinkResponse{
				Success: true,
				Message: "You are already a participant of this rally",
				RallyID: link.RallyID.Hex(),
				Role:    string(existing.Role),
				Status:  string(model.ParticipationStatusJoined),
			}, nil
		}

		// Determine the role to apply
		var finalRole model.ParticipantRole
		if existing.Status == model.ParticipationStatusInvited {
			// Take the higher of existing invite role vs link role
			finalRole = higherRole(existing.Role, link.RoleToGrant)
		} else {
			// Declined or left — fresh start, use the link's role
			finalRole = link.RoleToGrant
		}

		// Update to joined with the resolved role
		updates := &model.UpdateParticipantRequest{
			Status: &joinedStatus,
			Role:   &finalRole,
		}
		if _, err := s.participantRepo.UpdateParticipant(ctx, existing.ID.Hex(), updates); err != nil {
			return nil, fmt.Errorf("failed to join rally: %w", err)
		}

		// Increment link usage on successful join
		if err := s.inviteLinkRepo.IncrementLinkUsage(ctx, token); err != nil {
			return nil, fmt.Errorf("failed to process invite link usage: %w", err)
		}

		return &model.JoinViaLinkResponse{
			Success: true,
			Message: "Successfully joined the rally",
			RallyID: link.RallyID.Hex(),
			Role:    string(finalRole),
			Status:  string(model.ParticipationStatusJoined),
		}, nil
	}

	// No existing record — create new participant with "joined" status
	participant := &model.RallyParticipant{
		ID:        primitive.NewObjectID(),
		RallyID:   link.RallyID,
		UserID:    user.ID,
		Role:      link.RoleToGrant,
		Status:    model.ParticipationStatusJoined,
		InvitedBy: &link.CreatedBy,
		JoinedAt:  &now,
		InvitedAt: now,
	}

	if err := s.participantRepo.CreateParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to join rally: %w", err)
	}

	// Increment link usage on successful join
	if err := s.inviteLinkRepo.IncrementLinkUsage(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to process invite link usage: %w", err)
	}

	return &model.JoinViaLinkResponse{
		Success: true,
		Message: "Successfully joined the rally",
		RallyID: link.RallyID.Hex(),
		Role:    string(link.RoleToGrant),
		Status:  string(model.ParticipationStatusJoined),
	}, nil
}

// higherRole returns the more privileged of two roles.
// Precedence: owner > editor > participant.
func higherRole(a, b model.ParticipantRole) model.ParticipantRole {
	rank := map[model.ParticipantRole]int{
		model.ParticipantRoleParticipant: 0,
		model.ParticipantRoleEditor:      1,
		model.ParticipantRoleOwner:       2,
	}
	if rank[b] > rank[a] {
		return b
	}
	return a
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
