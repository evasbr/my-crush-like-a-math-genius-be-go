package app

import (
	"evasbr/mclamg/client/restclient"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/controller"
	"evasbr/mclamg/docs"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	repository "evasbr/mclamg/repository/impl"
	service "evasbr/mclamg/service/impl"
	"evasbr/mclamg/common"
	"evasbr/mclamg/exception"
	"os"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

// BuildApp initializes all repositories, services, controllers, middlewares, and routes.
// It returns a configured *fiber.App instance.
func BuildApp(config configuration.Config, database *gorm.DB, redis *redis.Client) *fiber.App {
	//repository
	userRepository := repository.NewUserRepositoryImpl(database)
	authRepository := repository.NewAuthRepositoryImpl(database)
	topicRepository := repository.NewTopicRepositoryImpl(database)
	questionRepository := repository.NewQuestionRepositoryImpl(database)
	attemptRepository := repository.NewAttemptRepositoryImpl(database)
	classroomRepository := repository.NewClassroomRepositoryImpl(database)

	//rest client
	httpBinRestClient := restclient.NewHttpBinRestClient()

	//file storage
	fileStorage, err := common.NewFileStorage(config.Get("CLOUDINARY_URL"))
	exception.PanicLogging(err)

	//service
	userService := service.NewUserServiceImpl(&userRepository, fileStorage)
	authService := service.NewAuthServiceImpl(&userRepository, &authRepository, redis, config)
	httpBinService := service.NewHttpBinServiceImpl(&httpBinRestClient)
	topicService := service.NewTopicServiceImpl(&topicRepository, fileStorage)
	questionService := service.NewQuestionServiceImpl(&questionRepository)
	attemptService := service.NewAttemptServiceImpl(&attemptRepository, &topicRepository)
	classroomService := service.NewClassroomServiceImpl(&classroomRepository, fileStorage)

	//controller
	userController := controller.NewUserController(&userService, config, redis)
	authController := controller.NewAuthController(&authService, config, redis)
	httpBinController := controller.NewHttpBinController(&httpBinService)
	topicController := controller.NewTopicController(&topicService, config, redis)
	questionController := controller.NewQuestionController(&questionService, config, redis)
	attemptController := controller.NewAttemptController(&attemptService, config, redis)
	classroomController := controller.NewClassroomController(&classroomService, config, redis)

	//setup fiber
	app := fiber.New(configuration.NewFiberConfiguration())
	api := app.Group("/api/v1")
	app.Use(recover.New())
	allowedOrigins := config.Get("CORS_ALLOWED_ORIGINS")
	allowCredentials := true
	if allowedOrigins == "" || allowedOrigins == "*" {
		allowedOrigins = "*"
		allowCredentials = false
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: allowCredentials,
	}))
	app.Use(requestid.New())
	app.Use(middleware.RequestID()) // Mount Request ID middleware and UserContext propagator

	//routing
	userController.Route(api)
	authController.Route(api)
	httpBinController.Route(api)
	topicController.Route(api)
	questionController.Route(api)
	attemptController.Route(api)
	classroomController.Route(api)

	//swagger
	docs.SwaggerInfo.Host = ""
	if os.Getenv("VERCEL") == "1" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
			Code:    200,
			Message: "Hello world",
		})
	})

	return app
}
