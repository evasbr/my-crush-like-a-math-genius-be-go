package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type TransactionDetailController struct {
	service.TransactionDetailService
	configuration.Config
	log *logrus.Entry
}

func NewTransactionDetailController(transactionDetailService *service.TransactionDetailService, config configuration.Config) *TransactionDetailController {
	return &TransactionDetailController{
		TransactionDetailService: *transactionDetailService,
		Config:                   config,
		log:                      common.Log.WithField("scope", "TransactionDetailController"),
	}
}

func (controller TransactionDetailController) Route(app *fiber.App) {
	app.Get("/v1/api/transaction-detail/:id", middleware.AuthenticateJWT("ROLE_USER", controller.Config), controller.FindById)
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
