package main

import (
	"context"
	"evasbr/mclamg/client/restclient"
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/controller"
	"evasbr/mclamg/db/seed"
	_ "evasbr/mclamg/docs"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	repository "evasbr/mclamg/repository/impl"
	service "evasbr/mclamg/service/impl"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
)

// @title Go Fiber Clean Architecture
// @version 1.0.0
// @description Baseline project using Go Fiber
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9999
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
// @description Authorization For JWT
func main() {
	// Parse CLI flags
	runSeed := flag.Bool("seed", false, "run database seeders")
	runRollback := flag.Bool("seed-rollback", false, "rollback database seeders")
	createSeed := flag.String("seed-create", "", "generate a new empty seeder file with timestamp prefix")
	flag.Parse()

	if *createSeed != "" {
		fmt.Printf("Generating new seeder: %s...\n", *createSeed)
		if err := seed.Create(*createSeed); err != nil {
			log.Fatalf("Failed to create seeder: %v", err)
		}
		return
	}

	//setup configuration
	config := configuration.New()
	database := configuration.NewDatabase(config)

	if *runSeed {
		fmt.Println("Running database seeders...")
		if err := seed.Run(database); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Println("Seeding completed successfully.")
		return
	}

	if *runRollback {
		fmt.Println("Rolling back database seeders...")
		if err := seed.Rollback(database, 1); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("Rollback completed successfully.")
		return
	}

	redis := configuration.NewRedis(config)

	//repository
	productRepository := repository.NewProductRepositoryImpl(database)
	transactionRepository := repository.NewTransactionRepositoryImpl(database)
	transactionDetailRepository := repository.NewTransactionDetailRepositoryImpl(database)
	userRepository := repository.NewUserRepositoryImpl(database)

	//rest client
	httpBinRestClient := restclient.NewHttpBinRestClient()

	//service
	productService := service.NewProductServiceImpl(&productRepository, redis)
	transactionService := service.NewTransactionServiceImpl(&transactionRepository)
	transactionDetailService := service.NewTransactionDetailServiceImpl(&transactionDetailRepository)
	userService := service.NewUserServiceImpl(&userRepository)
	httpBinService := service.NewHttpBinServiceImpl(&httpBinRestClient)

	//controller
	productController := controller.NewProductController(&productService, config)
	transactionController := controller.NewTransactionController(&transactionService, config)
	transactionDetailController := controller.NewTransactionDetailController(&transactionDetailService, config)
	userController := controller.NewUserController(&userService, config)
	httpBinController := controller.NewHttpBinController(&httpBinService)

	//setup fiber
	app := fiber.New(configuration.NewFiberConfiguration())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(requestid.New())
	app.Use(middleware.RequestID()) // Mount Request ID middleware and UserContext propagator

	//routing
	productController.Route(app)
	transactionController.Route(app)
	transactionDetailController.Route(app)
	userController.Route(app)
	httpBinController.Route(app)

	//swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check
	app.Get("/", HealthCheck)

	//start app in a goroutine
	go func() {
		err := app.Listen(config.Get("SERVER.PORT"))
		if err != nil {
			common.Logger(context.Background(), "Main").Infof("Server startup error: %v", err)
		}
	}()

	// Listening for OS termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	common.Logger(context.Background(), "Main").Info("Shutting down server gracefully...")

	// Attempt graceful shutdown
	if err := app.Shutdown(); err != nil {
		common.Logger(context.Background(), "Main").Errorf("Server forced to shutdown: %v", err)
	}

	common.Logger(context.Background(), "Main").Info("Server exited")
}

// HealthCheck func returns application health status.
// @Description health check endpoint.
// @Summary health check
// @Tags Health
// @Produce json
// @Success 200 {object} model.GeneralResponse
// @Router / [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Hello world",
	})
}
