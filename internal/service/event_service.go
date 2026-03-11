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

type EventService struct {
	firebaseAuth    *auth.Client
	eventRepo       repository.EventRepository
	rallyRepo       repository.RallyRepository
	participantRepo repository.RallyParticipantRepository
	userRepo        repository.UserRepository
}

func NewEventService(
	firebaseAuth *auth.Client,
	eventRepo repository.EventRepository,
	rallyRepo repository.RallyRepository,
	participantRepo repository.RallyParticipantRepository,
	userRepo repository.UserRepository,
) *EventService {
	return &EventService{
		firebaseAuth:    firebaseAuth,
		eventRepo:       eventRepo,
		rallyRepo:       rallyRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}
}

// CreateEvent creates a new event within a rally (middleware ensures owner or editor role)
func (s *EventService) CreateEvent(ctx context.Context, user *model.User, rallyID string, req *model.CreateEventRequest) (*model.EventResponse, error) {

	rally, err := s.rallyRepo.GetRallyByID(ctx, rallyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rally: %w", err)
	}
	if rally == nil {
		return nil, errors.New("rally not found")
	}

	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, fmt.Errorf("invalid rally ID: %w", err)
	}
	event := &model.Event{
		ID:            primitive.NewObjectID(),
		RallyID:       rallyObjID,
		GooglePlaceID: req.GooglePlaceID,
		Name:          req.Name,
		Lat:           req.Lat,
		Lng:           req.Lng,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Notes:         req.Notes,
		VisitOrder:    req.VisitOrder,
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return s.ConvertToEventResponse(event), nil
}

// UpdateEvent updates an existing event (requires owner or editor role in the event's rally)
func (s *EventService) UpdateEvent(ctx context.Context, user *model.User, eventID string, req *model.UpdateEventRequest) (*model.EventResponse, error) {

	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return nil, errors.New("event not found")
	}

	if err := validateRallyAccess(ctx, s.participantRepo, user.ID, event.RallyID.Hex(), []string{"owner", "editor"}); err != nil {
		return nil, err
	}

	updated, err := s.eventRepo.UpdateEvent(ctx, eventID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return s.ConvertToEventResponse(updated), nil
}

// ConvertToEventResponse converts an Event model to EventResponse
func (s *EventService) ConvertToEventResponse(event *model.Event) *model.EventResponse {
	return &model.EventResponse{
		ID:            event.ID.Hex(),
		RallyID:       event.RallyID.Hex(),
		GooglePlaceID: event.GooglePlaceID,
		Name:          event.Name,
		Lat:           event.Lat,
		Lng:           event.Lng,
		StartTime:     event.StartTime,
		EndTime:       event.EndTime,
		Notes:         event.Notes,
		VisitOrder:    event.VisitOrder,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	}
}
