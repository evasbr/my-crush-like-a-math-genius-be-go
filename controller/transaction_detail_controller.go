package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
)

type TransactionDetailController struct {
	service.TransactionDetailService
	configuration.Config
	Redis *redis.Client
	log *logrus.Entry
}

func NewTransactionDetailController(transactionDetailService *service.TransactionDetailService, config configuration.Config, redis *redis.Client) *TransactionDetailController {
	return &TransactionDetailController{
		TransactionDetailService: *transactionDetailService,
		Config:                   config,
		Redis:                    redis,
		log:                      common.Log.WithField("scope", "TransactionDetailController"),
	}
}

func (controller TransactionDetailController) Route(router fiber.Router) {
	detail := router.Group("/transaction-detail")
	detail.Get("/:id", middleware.RequireAuth([]string{"ROLE_USER", "read:transaction"}, controller.Config, controller.Redis), controller.FindById)
}

// FindById func gets one exists transaction detail.
// @Description Get one exists transaction detail.
// @Summary get one exists transaction detail
// @Tags Transaction Detail
// @Accept json
// @Produce json
// @Param id path string true "Transaction Detail Id"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/transaction-detail/{id} [get]
func (controller TransactionDetailController) FindById(c *fiber.Ctx) error {
	id := c.Params("id")

	result := controller.TransactionDetailService.FindById(c.UserContext(), id)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}
