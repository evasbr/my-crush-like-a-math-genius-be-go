package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type HttpBinController struct {
	service.HttpBinService
	log *logrus.Entry
}

func NewHttpBinController(httpBinService *service.HttpBinService) *HttpBinController {
	return &HttpBinController{
		HttpBinService: *httpBinService,
		log:            common.Log.WithField("scope", "HttpBinController"),
	}
}

func (controller HttpBinController) Route(router fiber.Router) {
	httpbin := router.Group("/httpbin")
	httpbin.Get("/", controller.PostHttpBin)
}

func (controller HttpBinController) PostHttpBin(c *fiber.Ctx) error {

	controller.HttpBinService.PostMethod(c.UserContext())
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    nil,
	})
}
