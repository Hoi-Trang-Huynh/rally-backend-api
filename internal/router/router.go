package router

import (
	"context"
	"time"

	fb "firebase.google.com/go/v4"
	_ "github.com/Hoi-Trang-Huynh/rally-backend-api/api/docs"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/handler"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/database"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/infrastructure/firebase"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/middleware"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/service"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func Setup(cfg *config.Config) *fiber.App {
	db := database.GetDB()
	internalDB := database.GetInternalDB()

	// Auth relies on the unique firebase_uid / username indexes — fail fast if
	// they cannot be created.
	idxCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := repository.EnsureUserIndexes(idxCtx, db); err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepository(db)
	followRepo := repository.NewFollowRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(internalDB)
	rallyRepo := repository.NewRallyRepository(db)
	eventRepo := repository.NewEventRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	participantRepo := repository.NewRallyParticipantRepository(db)
	inviteLinkRepo := repository.NewInviteLinkRepository(db)

	fbApp := firebase.GetClient()

	cld, err := utils.NewCloudinaryUploader(cfg.Cloudinary.URL)
	if err != nil {
		panic(err)
	}

	app, err := SetupWithDeps(userRepo, followRepo, feedbackRepo, rallyRepo, eventRepo, activityRepo, participantRepo, inviteLinkRepo, fbApp, cld, cfg.Server.AllowedOrigins)
	if err != nil {
		panic(err)
	}

	return app
}

func SetupWithDeps(
	userRepo repository.UserRepository,
	followRepo repository.FollowRepository,
	feedbackRepo repository.FeedbackRepository,
	rallyRepo repository.RallyRepository,
	eventRepo repository.EventRepository,
	activityRepo repository.ActivityRepository,
	participantRepo repository.RallyParticipantRepository,
	inviteLinkRepo repository.InviteLinkRepository,
	fbApp *fb.App,
	cld *utils.CloudinaryUploader,
	allowedOrigins string,
) (*fiber.App, error) {

	// Create Firebase auth client once — token verification happens solely in
	// the auth middleware.
	firebaseAuth, err := fbApp.Auth(context.Background())
	if err != nil {
		return nil, err
	}

	app := fiber.New()

	app.Use(middleware.Logger())
	app.Use(middleware.CORS(allowedOrigins))

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	followService := service.NewFollowService(followRepo, userRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	rallyService := service.NewRallyService(database.GetDB(), rallyRepo, participantRepo, userRepo)
	eventService := service.NewEventService(eventRepo, rallyRepo, participantRepo, userRepo)
	activityService := service.NewActivityService(activityRepo, eventRepo, participantRepo, userRepo)
	participantService := service.NewRallyParticipantService(participantRepo, rallyRepo, userRepo, followRepo)
	inviteLinkService := service.NewInviteLinkService(inviteLinkRepo, participantRepo, rallyRepo, userRepo, eventRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	mediaHandler := handler.NewMediaHandler(cld, userService)
	followHandler := handler.NewFollowHandler(followService)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService)
	rallyHandler := handler.NewRallyHandler(rallyService)
	eventHandler := handler.NewEventHandler(eventService)
	activityHandler := handler.NewActivityHandler(activityService)
	participantHandler := handler.NewRallyParticipantHandler(participantService)
	inviteLinkHandler := handler.NewInviteLinkHandler(inviteLinkService)

	// Auth middleware chain: verify the Firebase ID token once, then resolve
	// (JIT-provisioning) the MongoDB user from the verified claims.
	auth := middleware.AuthRequired(firebaseAuth)
	resolveUser := middleware.ResolveFirebaseUser(userRepo)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/health", handler.HealthCheck)
	v1.Get("/version", handler.VersionCheck)

	// Throttle the public auth endpoints (per-IP) against abuse/enumeration
	authLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
	})

	authRoutes := v1.Group("/auth", authLimiter)
	authRoutes.Post("/register", auth, resolveUser, authHandler.Register)
	authRoutes.Get("/check-email", authHandler.CheckEmailAvailability)
	authRoutes.Get("/check-username", authHandler.CheckUsernameAvailability)

	users := v1.Group("/user")
	users.Get("/me/profile", auth, resolveUser, userHandler.GetMyProfile)
	users.Get("/me/profile/details", auth, resolveUser, userHandler.GetMyProfileDetails)
	users.Get("/me/invitations", auth, resolveUser, participantHandler.GetPendingInvitations) // TODO: temporary until realtime notifications
	users.Get("/search", userHandler.SearchUsers)
	users.Get("/:id/profile", followHandler.GetUserPublicProfile)
	users.Put("/:id/profile", auth, resolveUser, userHandler.UpdateProfile)
	users.Post("/:id/follow", auth, resolveUser, followHandler.FollowUser)
	users.Delete("/:id/follow", auth, resolveUser, followHandler.UnfollowUser)
	users.Get("/:id/follow/status", auth, resolveUser, followHandler.GetFollowStatus)
	users.Get("/:id/followers", followHandler.GetFollowersList)
	users.Get("/:id/following", followHandler.GetFollowingList)
	users.Get("/:id/friends", followHandler.GetFriendsList)
	users.Get("/:id/rallies", auth, rallyHandler.GetRalliesList)

	media := v1.Group("/media")
	media.Post("/sign", mediaHandler.GetUploadSignature)
	media.Post("/verify-avatar", auth, resolveUser, mediaHandler.VerifyAvatar)

	feedback := v1.Group("/feedback")
	feedback.Post("/", feedbackHandler.CreateFeedback)
	feedback.Get("/", feedbackHandler.GetFeedbackList)
	feedback.Patch("/:id/resolve", feedbackHandler.UpdateFeedbackStatus)

	// Convenience aliases for rally access middleware
	loadParticipant := middleware.LoadRallyParticipant(participantRepo)
	joined := middleware.RequireJoined()
	ownerOrEditor := middleware.RequireRole("owner", "editor")

	// Rally routes (all require auth + resolved user)
	rallies := v1.Group("/rallies", auth, resolveUser)
	rallies.Post("/join-via-link", inviteLinkHandler.JoinViaLink)                                                              // No rally ID — manual validation
	rallies.Get("/invite-links/:token/preview", inviteLinkHandler.PreviewInviteLink)                                           // Preview an invite link
	rallies.Post("/", rallyHandler.CreateRally)                                                                                // No rally ID yet
	rallies.Get("/:id", loadParticipant, rallyHandler.GetRally)                                                                // Allows invited — handler checks status
	rallies.Put("/:id", loadParticipant, joined, ownerOrEditor, rallyHandler.UpdateRally)                                      // Owner/Editor + joined
	rallies.Post("/:id/events", loadParticipant, joined, ownerOrEditor, eventHandler.CreateEvent)                              // Owner/Editor + joined
	rallies.Get("/:id/participants", loadParticipant, joined, participantHandler.GetParticipantsList)                          // Any joined participant
	rallies.Get("/:id/invitable-friends", loadParticipant, joined, participantHandler.GetInvitableFriends)                     // Any joined participant
	rallies.Post("/:id/participants", loadParticipant, joined, ownerOrEditor, participantHandler.InviteParticipant)            // Owner/Editor + joined
	rallies.Put("/:id/participants/:participantId", loadParticipant, participantHandler.UpdateParticipant)                     // Conditional — service handles self vs. others
	rallies.Post("/:id/invite-links", loadParticipant, joined, ownerOrEditor, inviteLinkHandler.CreateInviteLink)              // Owner/Editor + joined (extra owner check for elevated roles in service)
	rallies.Get("/:id/invite-links", loadParticipant, joined, ownerOrEditor, inviteLinkHandler.GetActiveInviteLinks)           // Owner/Editor + joined
	rallies.Delete("/:id/invite-links/:token", loadParticipant, joined, ownerOrEditor, inviteLinkHandler.DeactivateInviteLink) // Owner/Editor + joined

	// Event routes (auth + resolved user, rally access checked in service via event lookup)
	events := v1.Group("/events", auth, resolveUser)
	events.Put("/:id", eventHandler.UpdateEvent)
	events.Post("/:id/activities", activityHandler.CreateActivity)

	// Activity routes (auth + resolved user, rally access checked in service via activity lookup)
	activities := v1.Group("/activities", auth, resolveUser)
	activities.Put("/:id", activityHandler.UpdateActivity)

	return app, nil
}
