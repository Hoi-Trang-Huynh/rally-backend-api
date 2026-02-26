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

// ResolveFirebaseUser verifies the Firebase ID token from c.Locals("idToken")
// and loads the corresponding user from the database.
// On success, stores the *model.User in c.Locals("user").
func ResolveFirebaseUser(firebaseAuth *auth.Client, userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idToken, ok := c.Locals("idToken").(string)
		if !ok || idToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Authorization token is required",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := firebaseAuth.VerifyIDToken(ctx, idToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "Invalid or expired token",
			})
		}

		user, err := userRepo.GetUserByFirebaseUID(ctx, token.UID)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Message: "User not found",
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
