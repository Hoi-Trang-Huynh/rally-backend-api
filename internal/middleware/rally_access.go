package middleware

import (
	"context"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResolveFirebaseUser loads the user matching the verified Firebase token in
// c.Locals("authToken"). The user is provisioned on first sight (JIT) and the
// email/email_verified claims are kept in sync with MongoDB.
// On success, stores the *model.User in c.Locals("user").
// Must be chained AFTER AuthRequired.
func ResolveFirebaseUser(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := c.Locals("authToken").(*auth.Token)
		if !ok || token == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Authorization token is required",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := userRepo.EnsureUser(ctx, firebaseUserInfoFromToken(token))
		if err != nil || user == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to resolve user",
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}

// LoadRallyParticipant loads the RallyParticipant record for the current user
// and the rally identified by c.Params("id").
// On success, stores the *model.RallyParticipant in c.Locals("rallyParticipant").
// Returns 403 if the user has no participant record at all.
func LoadRallyParticipant(participantRepo repository.RallyParticipantRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*model.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "User not resolved",
			})
		}

		rallyID := c.Params("id")
		if rallyID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: "Rally ID is required",
			})
		}

		rallyObjID, err := primitive.ObjectIDFromHex(rallyID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Message: "Invalid rally ID",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		participant, err := participantRepo.GetParticipantByRallyAndUser(ctx, rallyObjID, user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Message: "Failed to check rally access",
			})
		}
		if participant == nil {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: "Not a participant of this rally",
			})
		}

		c.Locals("rallyParticipant", participant)
		return c.Next()
	}
}

// RequireJoined checks that the loaded rally participant has status "joined".
// Must be chained AFTER LoadRallyParticipant.
func RequireJoined() fiber.Handler {
	return func(c *fiber.Ctx) error {
		participant, ok := c.Locals("rallyParticipant").(*model.RallyParticipant)
		if !ok || participant == nil {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: "Rally participant not loaded",
			})
		}

		if participant.Status != model.ParticipationStatusJoined {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: "Participant status is not active (must be joined)",
			})
		}

		return c.Next()
	}
}

// RequireRole checks that the loaded rally participant has one of the specified roles.
// Must be chained AFTER LoadRallyParticipant (and typically after RequireJoined).
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		participant, ok := c.Locals("rallyParticipant").(*model.RallyParticipant)
		if !ok || participant == nil {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Message: "Rally participant not loaded",
			})
		}

		for _, role := range roles {
			if string(participant.Role) == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: "Insufficient permissions",
		})
	}
}
