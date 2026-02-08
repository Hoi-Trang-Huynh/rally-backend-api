package router

import (
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
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func Setup(cfg *config.Config) *fiber.App {
	db := database.GetDB()
	internalDB := database.GetInternalDB()

	userRepo := repository.NewUserRepository(db)
	followRepo := repository.NewFollowRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(internalDB)
	rallyRepo := repository.NewRallyRepository(db)
	eventRepo := repository.NewEventRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	participantRepo := repository.NewRallyParticipantRepository(db)

	fbApp := firebase.GetClient()

	cld, err := utils.NewCloudinaryUploader(cfg.Cloudinary.URL)
	if err != nil {
		panic(err)
	}

	app, err := SetupWithDeps(userRepo, followRepo, feedbackRepo, rallyRepo, eventRepo, activityRepo, participantRepo, fbApp, cld)
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
	fbApp *fb.App,
	cld *utils.CloudinaryUploader,
) (*fiber.App, error) {

	app := fiber.New()

	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	authService, err := service.NewAuthService(fbApp, userRepo)
	if err != nil {
		return nil, err
	}

	userService, err := service.NewUserService(fbApp, userRepo)
	if err != nil {
		return nil, err
	}

	followService, err := service.NewFollowService(fbApp, followRepo, userRepo)
	if err != nil {
		return nil, err
	}

	feedbackService := service.NewFeedbackService(feedbackRepo)

	rallyService, err := service.NewRallyService(fbApp, rallyRepo, participantRepo, userRepo)
	if err != nil {
		return nil, err
	}

	eventService, err := service.NewEventService(fbApp, eventRepo, rallyRepo, participantRepo, userRepo)
	if err != nil {
		return nil, err
	}

	activityService, err := service.NewActivityService(fbApp, activityRepo, eventRepo, participantRepo, userRepo)
	if err != nil {
		return nil, err
	}

	participantService, err := service.NewRallyParticipantService(fbApp, participantRepo, rallyRepo, userRepo)
	if err != nil {
		return nil, err
	}

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	mediaHandler := handler.NewMediaHandler(cld, userService)
	followHandler := handler.NewFollowHandler(followService)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService)
	rallyHandler := handler.NewRallyHandler(rallyService)
	eventHandler := handler.NewEventHandler(eventService)
	activityHandler := handler.NewActivityHandler(activityService)
	participantHandler := handler.NewRallyParticipantHandler(participantService)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/health", handler.HealthCheck)
	v1.Get("/version", handler.VersionCheck)

	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.RegisterOrLogin)
	auth.Post("/login", authHandler.Login)
	auth.Get("/check-email", authHandler.CheckEmailAvailability)
	auth.Get("/check-username", authHandler.CheckUsernameAvailability)

	users := v1.Group("/user")
	users.Get("/me/profile", userHandler.GetMyProfile)
	users.Get("/me/profile/details", userHandler.GetMyProfileDetails)
	users.Get("/search", userHandler.SearchUsers)
	users.Get("/:id/profile", followHandler.GetUserPublicProfile)
	users.Put("/:id/profile", userHandler.UpdateProfile)
	users.Post("/:id/follow", followHandler.FollowUser)
	users.Delete("/:id/follow", followHandler.UnfollowUser)
	users.Get("/:id/follow/status", followHandler.GetFollowStatus)
	users.Get("/:id/followers", followHandler.GetFollowersList)
	users.Get("/:id/following", followHandler.GetFollowingList)
	users.Get("/:id/friends", followHandler.GetFriendsList)

	media := v1.Group("/media")
	media.Post("/sign", mediaHandler.GetUploadSignature)
	media.Post("/verify-avatar", mediaHandler.VerifyAvatar)

	feedback := v1.Group("/feedback")
	feedback.Post("/", feedbackHandler.CreateFeedback)
	feedback.Get("/", feedbackHandler.GetFeedbackList)
	feedback.Patch("/:id/resolve", feedbackHandler.UpdateFeedbackStatus)

	// Rally routes
	rallies := v1.Group("/rallies")
	rallies.Post("/", rallyHandler.CreateRally)
	rallies.Put("/:id", rallyHandler.UpdateRally)
	rallies.Post("/:id/events", eventHandler.CreateEvent)
	rallies.Post("/:id/participants", participantHandler.InviteParticipant)
	rallies.Put("/:id/participants/:participantId", participantHandler.UpdateParticipant)

	// Event routes
	events := v1.Group("/events")
	events.Put("/:id", eventHandler.UpdateEvent)
	events.Post("/:id/activities", activityHandler.CreateActivity)

	// Activity routes
	activities := v1.Group("/activities")
	activities.Put("/:id", activityHandler.UpdateActivity)

	return app, nil
}
