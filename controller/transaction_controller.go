package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
)

type TransactionController struct {
	service.TransactionService
	configuration.Config
	Redis *redis.Client
	log *logrus.Entry
}

func NewTransactionController(transactionService *service.TransactionService, config configuration.Config, redis *redis.Client) *TransactionController {
	return &TransactionController{
		TransactionService: *transactionService,
		Config:             config,
		Redis:              redis,
		log:                common.Log.WithField("scope", "TransactionController"),
	}
}

func (controller TransactionController) Route(router fiber.Router) {
	transaction := router.Group("/transaction")
	transaction.Post("/", middleware.RequireAuth([]string{"ROLE_USER", "write:transaction"}, controller.Config, controller.Redis), controller.Create)
	transaction.Delete("/:id", middleware.RequireAuth([]string{"ROLE_USER", "write:transaction"}, controller.Config, controller.Redis), controller.Delete)
	transaction.Get("/:id", middleware.RequireAuth([]string{"ROLE_USER", "read:transaction"}, controller.Config, controller.Redis), controller.FindById)
	transaction.Get("/", middleware.RequireAuth([]string{"ROLE_USER", "read:transaction"}, controller.Config, controller.Redis), controller.FindAll)
}

// Create func create transaction.
// @Description create transaction.
// @Summary create transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param request body model.TransactionCreateUpdateModel true "Request Body"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/transaction [post]
func (controller TransactionController) Create(c *fiber.Ctx) error {
	var request model.TransactionCreateUpdateModel
	err := c.BodyParser(&request)
	exception.PanicLogging(err)

	response := controller.TransactionService.Create(c.UserContext(), request)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// Delete func delete one exists transaction.
// @Description delete one exists transaction.
// @Summary delete one exists transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path string true "Transaction Id"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/transaction/{id} [delete]
func (controller TransactionController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	controller.TransactionService.Delete(c.UserContext(), id)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
	})
}

// FindById func gets one exists transaction.
// @Description Get one exists transaction.
// @Summary get one exists transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path string true "Transaction Id"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/transaction/{id} [get]
func (controller TransactionController) FindById(c *fiber.Ctx) error {
	id := c.Params("id")

	result := controller.TransactionService.FindById(c.UserContext(), id)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}

// FindAll func gets all exists transaction.
// @Description Get all exists transaction.
// @Summary get all exists transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/transaction [get]
func (controller TransactionController) FindAll(c *fiber.Ctx) error {
	result := controller.TransactionService.FindAll(c.UserContext())
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}
