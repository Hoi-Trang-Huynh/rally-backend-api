package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// validateRallyAccess checks that a user is a participant of the given rally
// and has one of the required roles.
func validateRallyAccess(ctx context.Context, participantRepo repository.RallyParticipantRepository, userID primitive.ObjectID, rallyID string, requiredRoles []string) error {
	rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return errors.New("invalid rally ID")
	}

	participant, err := participantRepo.GetParticipantByRallyAndUser(ctx, rallyObjID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if participant == nil {
		return errors.New("unauthorized: not a participant of this rally")
	}
	if participant.Status != model.ParticipationStatusJoined {
		return errors.New("unauthorized: participant status is not active (must be joined)")
	}

	for _, role := range requiredRoles {
		if string(participant.Role) == role {
			return nil
		}
	}

	return errors.New("unauthorized: insufficient permissions")
}
