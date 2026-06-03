package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/repository/impl"
	impl2 "evasbr/mclamg/service/impl"
	"io"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"golang.org/x/crypto/bcrypt"
)

func createTestApp() *fiber.App {
	//setup fiber
	app := fiber.New(configuration.NewFiberConfiguration())
	app.Use(recover.New())
	app.Use(cors.New())

	//routing
	productController.Route(app)
	transactionController.Route(app)
	transactionDetailController.Route(app)
	userController.Route(app)
	authController.Route(app)

	// Route under v1/api prefix group to support prefix tests
	v1 := app.Group("/v1/api")
	productController.Route(v1)
	transactionController.Route(v1)
	transactionDetailController.Route(v1)
	userController.Route(v1)
	authController.Route(v1)

	// Demo admin dashboard endpoint (RBAC / PBAC config)
	app.Get("/admin/dashboard",
		middleware.RequireAuth([]string{"admin", "superadmin", "read:dashboard"}, config, redisClient),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
				Code:    200,
				Message: "Success",
				Data:    "Welcome to Admin Dashboard!",
			})
		},
	)

	return app
}

// setup configuration
var config = configuration.New("../.env.test")
var database = configuration.NewDatabase(config)
var redisClient = configuration.NewRedis(config)

// repository
var productRepository = impl.NewProductRepositoryImpl(database)
var transactionRepository = impl.NewTransactionRepositoryImpl(database)
var transactionDetailRepository = impl.NewTransactionDetailRepositoryImpl(database)
var userRepository = impl.NewUserRepositoryImpl(database)
var authRepository = impl.NewAuthRepositoryImpl(database)

// service
var productService = impl2.NewProductServiceImpl(&productRepository, redisClient)
var transactionService = impl2.NewTransactionServiceImpl(&transactionRepository)
var transactionDetailService = impl2.NewTransactionDetailServiceImpl(&transactionDetailRepository)
var userService = impl2.NewUserServiceImpl(&userRepository)
var authService = impl2.NewAuthServiceImpl(&userRepository, &authRepository, redisClient, config)

// controller
var productController = NewProductController(&productService, config, redisClient)
var transactionController = NewTransactionController(&transactionService, config, redisClient)
var transactionDetailController = NewTransactionDetailController(&transactionDetailService, config, redisClient)
var userController = NewUserController(&userService, config, redisClient)
var authController = NewAuthController(&authService, config, redisClient)

var appTest = createTestApp()

func deleteAllUsers() {
	database.Exec("DELETE FROM authentications")
	database.Exec("DELETE FROM user_roles")
	database.Exec("DELETE FROM users")
	database.Exec("DELETE FROM roles")

	database.Exec("INSERT INTO roles (id, name, permissions) VALUES ('11b38d48-8605-4e1f-8630-2c2120fbd682', 'SUPER_ADMIN', '{\"fullaccess\": true}')")
	database.Exec("INSERT INTO roles (id, name, permissions) VALUES ('d4df0794-22e8-4a30-9039-bbcd76447b56', 'SUB_ADMIN', '{\"product\": [\"read:product\", \"write:product\"]}')")
	database.Exec("INSERT INTO roles (id, name, permissions) VALUES ('22b38d48-8605-4e1f-8630-2c2120fbd682', 'ROLE_ADMIN', '{\"profile\": [\"read:profile\"]}')")
	database.Exec("INSERT INTO roles (id, name, permissions) VALUES ('32b38d48-8605-4e1f-8630-2c2120fbd682', 'ROLE_USER', '{\"profile\": [\"read:profile\"]}')")
}

func createTestUser(username, email, gender, firstName, lastName, password string, roles []string) {
	payload := repository.RegisterUserPayload{
		Username:  &username,
		Email:     email,
		FirstName: firstName,
		LastName:  &lastName,
		Gender:    &gender,
		Password:  password,
		RoleIDs:   roles,
	}
	_, err := authRepository.Register(context.Background(), payload)
	exception.PanicLogging(err)
}

func authenticationCreate() map[string]interface{} {
	deleteAllUsers()

	password, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	exception.PanicLogging(err)
	roles := []string{"22b38d48-8605-4e1f-8630-2c2120fbd682", "32b38d48-8605-4e1f-8630-2c2120fbd682"}
	createTestUser("admin", "admin@example.com", "male", "admin", "admin", string(password), roles)

	userModel := model.LoginRequest{
		Identifier: "admin",
		Password:   "admin",
	}

	userRequestBody, _ := json.Marshal(userModel)

	userRequest := httptest.NewRequest("POST", "/authentication/login", bytes.NewBuffer(userRequestBody))
	userRequest.Header.Set("Content-Type", "application/json")
	userRequest.Header.Set("Accept", "application/json")

	userResponse, _ := appTest.Test(userRequest)

	userResponseBody, _ := io.ReadAll(userResponse.Body)
	println("DEBUG AUTHENTICATION BODY: ", string(userResponseBody))
	userWebResponse := model.GeneralResponse{}
	_ = json.Unmarshal(userResponseBody, &userWebResponse)

	userJsonData, _ := json.Marshal(userWebResponse.Data)

	tokenResponse := map[string]interface{}{}
	_ = json.Unmarshal(userJsonData, &tokenResponse)

	return tokenResponse
}
