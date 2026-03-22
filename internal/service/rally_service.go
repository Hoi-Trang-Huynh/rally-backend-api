package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RallyService struct {
	db              *mongo.Database
	firebaseAuth    *auth.Client
	rallyRepo       repository.RallyRepository
	participantRepo repository.RallyParticipantRepository
	userRepo        repository.UserRepository
}

func NewRallyService(
	db *mongo.Database,
	firebaseAuth *auth.Client,
	rallyRepo repository.RallyRepository,
	participantRepo repository.RallyParticipantRepository,
	userRepo repository.UserRepository,
) *RallyService {
	return &RallyService{
		db:              db,
		firebaseAuth:    firebaseAuth,
		rallyRepo:       rallyRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}
}

// CreateRally creates a new rally, auto-adds the creator as owner, and invites participants
func (s *RallyService) CreateRally(ctx context.Context, user *model.User, req *model.CreateRallyRequest) (*model.RallyResponse, error) {

	session, err := s.db.Client().StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	var rallyResponse *model.RallyResponse

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		rally := &model.Rally{
			ID:            primitive.NewObjectID(),
			OwnerID:       user.ID,
			Name:          req.Name,
			Description:   req.Description,
			CoverImageUrl: req.CoverImageUrl,
			Status:        model.RallyStatusDraft,
			StartDate:     req.StartDate,
			EndDate:       req.EndDate,
		}

		if err := s.rallyRepo.CreateRally(sessCtx, rally); err != nil {
			return nil, fmt.Errorf("failed to create rally: %w", err)
		}

		// Auto-add creator as owner participant
		now := time.Now()
		ownerParticipant := &model.RallyParticipant{
			ID:        primitive.NewObjectID(),
			RallyID:   rally.ID,
			UserID:    user.ID,
			Role:      model.ParticipantRoleOwner,
			Status:    model.ParticipationStatusJoined,
			JoinedAt:  &now,
			InvitedAt: now,
		}

		if err := s.participantRepo.CreateParticipant(sessCtx, ownerParticipant); err != nil {
			return nil, fmt.Errorf("failed to create owner participant: %w", err)
		}

		// Process invited participants
		for _, p := range req.Participants {
			// Skip if user tries to invite themselves (owner already added)
			if p.UserID == user.ID.Hex() {
				continue
			}

			targetUser, err := s.userRepo.GetUserByID(sessCtx, p.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get target user %s: %w", p.UserID, err)
			}
			if targetUser == nil {
				return nil, fmt.Errorf("target user %s not found", p.UserID)
			}

			// Check for duplicate participant
			existing, err := s.participantRepo.GetParticipantByRallyAndUser(sessCtx, rally.ID, targetUser.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing participant: %w", err)
			}
			if existing != nil {
				continue // Duplicate, skip
			}

			role := model.ParticipantRoleParticipant
			if p.Role != nil {
				role = *p.Role
			}

			participant := &model.RallyParticipant{
				ID:        primitive.NewObjectID(),
				RallyID:   rally.ID,
				UserID:    targetUser.ID,
				Role:      role,
				Status:    model.ParticipationStatusInvited,
				InvitedBy: &user.ID,
				InvitedAt: now,
			}

			if err := s.participantRepo.CreateParticipant(sessCtx, participant); err != nil {
				return nil, fmt.Errorf("failed to invite participant %s: %w", p.UserID, err)
			}
		}

		rallyResponse = s.ConvertToRallyResponse(rally)
		return rallyResponse, nil
	})

	if err != nil {
		return nil, err
	}

	return rallyResponse, nil
}

// UpdateRally updates an existing rally (middleware ensures owner or editor role)
func (s *RallyService) UpdateRally(ctx context.Context, rallyID string, req *model.UpdateRallyRequest) (*model.RallyResponse, error) {

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

// GetRalliesList retrieves a filtered and sorted list of rallies for a specific user with pagination
func (s *RallyService) GetRalliesList(ctx context.Context, idToken string, userID string, nameFilter string, statusFilter string, sortOrder string, page int, pageSize int) (*model.RalliesListResponse, error) {
	// Authenticate the requesting user
	_, err := authenticateUser(ctx, s.firebaseAuth, s.userRepo, idToken)
	if err != nil {
		return nil, err
	}

	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Verify the target user exists
	targetUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}
	if targetUser == nil {
		return nil, errors.New("user not found")
	}

	rallies, total, err := s.rallyRepo.GetRalliesList(ctx, userObjectID, nameFilter, statusFilter, sortOrder, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get rallies list: %w", err)
	}

	// Convert to list items
	rallyItems := make([]model.RallyListItem, len(rallies))
	for i, rally := range rallies {
		rallyItems[i] = model.RallyListItem{
			ID:        rally.ID.Hex(),
			Name:      rally.Name,
			Status:    rally.Status,
			StartDate: rally.StartDate,
			EndDate:   rally.EndDate,
			UpdatedAt: rally.UpdatedAt,
		}
	}

	// Calculate pagination metadata
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &model.RalliesListResponse{
		Rallies:    rallyItems,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Pagination: model.PaginationMetadata{
			HasNextPage:     page < totalPages,
			HasPreviousPage: page > 1,
		},
	}, nil
}

// GetRally retrieves a specific rally by ID (middleware ensures participant exists)
func (s *RallyService) GetRally(ctx context.Context, participant *model.RallyParticipant, rallyID string) (*model.RallyJoinResponse, error) {
	// Middleware already loaded participant; check for allowed statuses
	if participant.Status != model.ParticipationStatusJoined && participant.Status != model.ParticipationStatusInvited {
		return nil, errors.New("unauthorized: you must be joined or invited to view this rally")
	}

	// Fetch rally details
	rally, err := s.rallyRepo.GetRallyByID(ctx, rallyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rally: %w", err)
	}
	if rally == nil {
		return nil, errors.New("rally not found")
	}

	return &model.RallyJoinResponse{
		RallyResponse:     s.ConvertToRallyResponse(rally),
		CurrentUserRole:   participant.Role,
		CurrentUserStatus: participant.Status,
	}, nil
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
