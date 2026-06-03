package app

import (
	"evasbr/mclamg/client/restclient"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/controller"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	repository "evasbr/mclamg/repository/impl"
	service "evasbr/mclamg/service/impl"

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
	productRepository := repository.NewProductRepositoryImpl(database)
	transactionRepository := repository.NewTransactionRepositoryImpl(database)
	transactionDetailRepository := repository.NewTransactionDetailRepositoryImpl(database)
	userRepository := repository.NewUserRepositoryImpl(database)
	authRepository := repository.NewAuthRepositoryImpl(database)

	//rest client
	httpBinRestClient := restclient.NewHttpBinRestClient()

	//service
	productService := service.NewProductServiceImpl(&productRepository, redis)
	transactionService := service.NewTransactionServiceImpl(&transactionRepository)
	transactionDetailService := service.NewTransactionDetailServiceImpl(&transactionDetailRepository)
	userService := service.NewUserServiceImpl(&userRepository)
	authService := service.NewAuthServiceImpl(&userRepository, &authRepository, redis, config)
	httpBinService := service.NewHttpBinServiceImpl(&httpBinRestClient)

	//controller
	productController := controller.NewProductController(&productService, config, redis)
	transactionController := controller.NewTransactionController(&transactionService, config, redis)
	transactionDetailController := controller.NewTransactionDetailController(&transactionDetailService, config, redis)
	userController := controller.NewUserController(&userService, config, redis)
	authController := controller.NewAuthController(&authService, config, redis)
	httpBinController := controller.NewHttpBinController(&httpBinService)

	//setup fiber
	app := fiber.New(configuration.NewFiberConfiguration())
	api := app.Group("/api/v1")
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(requestid.New())
	app.Use(middleware.RequestID()) // Mount Request ID middleware and UserContext propagator

	//routing
	productController.Route(api)
	transactionController.Route(api)
	transactionDetailController.Route(api)
	userController.Route(api)
	authController.Route(api)
	httpBinController.Route(api)

	//swagger
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
