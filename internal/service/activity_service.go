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

type ActivityService struct {
	firebaseAuth    *auth.Client
	activityRepo    repository.ActivityRepository
	eventRepo       repository.EventRepository
	participantRepo repository.RallyParticipantRepository
	userRepo        repository.UserRepository
}

func NewActivityService(
	firebaseApp *firebase.App,
	activityRepo repository.ActivityRepository,
	eventRepo repository.EventRepository,
	participantRepo repository.RallyParticipantRepository,
	userRepo repository.UserRepository,
) (*ActivityService, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &ActivityService{
		firebaseAuth:    authClient,
		activityRepo:    activityRepo,
		eventRepo:       eventRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}, nil
}

func (s *ActivityService) authenticateUser(ctx context.Context, idToken string) (*model.User, error) {
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

func (s *ActivityService) validateRallyAccessViaEvent(ctx context.Context, userID primitive.ObjectID, eventID string, requiredRoles []string) (*model.Event, error) {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return nil, errors.New("event not found")
	}

	rallyObjID := event.RallyID
	participant, err := s.participantRepo.GetParticipantByRallyAndUser(ctx, rallyObjID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if participant == nil {
		return nil, errors.New("unauthorized: not a participant of this rally")
	}

	for _, role := range requiredRoles {
		if string(participant.Role) == role {
			return event, nil
		}
	}

	return nil, errors.New("unauthorized: insufficient permissions")
}

// CreateActivity creates a new activity within an event (requires owner or editor role in the event's rally)
func (s *ActivityService) CreateActivity(ctx context.Context, idToken string, eventID string, req *model.CreateActivityRequest) (*model.ActivityResponse, error) {
	user, err := s.authenticateUser(ctx, idToken)
	if err != nil {
		return nil, err
	}

	_, err = s.validateRallyAccessViaEvent(ctx, user.ID, eventID, []string{"owner", "editor"})
	if err != nil {
		return nil, err
	}

	eventObjID, _ := primitive.ObjectIDFromHex(eventID)
	activity := &model.Activity{
		ID:            primitive.NewObjectID(),
		EventID:       eventObjID,
		Name:          req.Name,
		Description:   req.Description,
		Status:        "planned",
		GooglePlaceID: req.GooglePlaceID,
		Lat:           req.Lat,
		Lng:           req.Lng,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Notes:         req.Notes,
		ActivityOrder: req.ActivityOrder,
	}

	if err := s.activityRepo.CreateActivity(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	return s.ConvertToActivityResponse(activity), nil
}

// UpdateActivity updates an existing activity (requires owner or editor role in the activity's rally)
func (s *ActivityService) UpdateActivity(ctx context.Context, idToken string, activityID string, req *model.UpdateActivityRequest) (*model.ActivityResponse, error) {
	user, err := s.authenticateUser(ctx, idToken)
	if err != nil {
		return nil, err
	}

	activity, err := s.activityRepo.GetActivityByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}
	if activity == nil {
		return nil, errors.New("activity not found")
	}

	_, err = s.validateRallyAccessViaEvent(ctx, user.ID, activity.EventID.Hex(), []string{"owner", "editor"})
	if err != nil {
		return nil, err
	}

	updated, err := s.activityRepo.UpdateActivity(ctx, activityID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	return s.ConvertToActivityResponse(updated), nil
}

// ConvertToActivityResponse converts an Activity model to ActivityResponse
func (s *ActivityService) ConvertToActivityResponse(activity *model.Activity) *model.ActivityResponse {
	return &model.ActivityResponse{
		ID:            activity.ID.Hex(),
		EventID:       activity.EventID.Hex(),
		Name:          activity.Name,
		Description:   activity.Description,
		Status:        activity.Status,
		GooglePlaceID: activity.GooglePlaceID,
		Lat:           activity.Lat,
		Lng:           activity.Lng,
		StartTime:     activity.StartTime,
		EndTime:       activity.EndTime,
		Notes:         activity.Notes,
		ActivityOrder: activity.ActivityOrder,
		CreatedAt:     activity.CreatedAt,
		UpdatedAt:     activity.UpdatedAt,
	}
}
