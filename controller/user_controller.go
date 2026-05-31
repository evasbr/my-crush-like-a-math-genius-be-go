package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func NewUserController(userService *service.UserService, config configuration.Config) *UserController {
	return &UserController{
		UserService: *userService,
		Config:      config,
		log:         common.Log.WithField("scope", "UserController"),
	}
}

type UserController struct {
	service.UserService
	configuration.Config
	log *logrus.Entry
}

func (controller UserController) Route(app *fiber.App) {
	app.Post("/v1/api/authentication", controller.Authentication)
}

// Authentication func Authenticate user.
// @Description authenticate user.
// @Summary authenticate user
// @Tags Authenticate user
// @Accept json
// @Produce json
// @Param request body model.UserModel true "Request Body"
// @Success 200 {object} model.GeneralResponse
// @Router /v1/api/authentication [post]
func (controller UserController) Authentication(c *fiber.Ctx) error {
	var request model.UserModel
	err := c.BodyParser(&request)
	exception.PanicLogging(err)

	result := controller.UserService.Authentication(c.UserContext(), request)
	var userRoles []map[string]interface{}
	for _, userRole := range result.UserRoles {
		userRoles = append(userRoles, map[string]interface{}{
			"role": userRole.Role,
		})
	}
	tokenJwtResult := common.GenerateToken(result.Username, userRoles, controller.Config)
	resultWithToken := map[string]interface{}{
		"token":    tokenJwtResult,
		"username": result.Username,
		"role":     userRoles,
	}
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    resultWithToken,
	})
}
