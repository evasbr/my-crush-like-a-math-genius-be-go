package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type TopicController struct {
	TopicService service.TopicService
	Config       configuration.Config
	Redis        *redis.Client
	log          *logrus.Entry
}

func NewTopicController(topicService *service.TopicService, config configuration.Config, redis *redis.Client) *TopicController {
	return &TopicController{
		TopicService: *topicService,
		Config:       config,
		Redis:        redis,
		log:          common.Log.WithField("scope", "TopicController"),
	}
}

func (controller *TopicController) Route(router fiber.Router) {
	topics := router.Group("/topics")

	// Public routes
	topics.Get("/", controller.FindAll)
	topics.Get("/:id", controller.FindByID)

	// Protected routes (Super Admin only)
	requireSuperAdmin := middleware.RequireAuth([]string{"SUPER_ADMIN"}, controller.Config, controller.Redis)
	topics.Post("/", requireSuperAdmin, controller.Create)
	topics.Put("/:id", requireSuperAdmin, controller.Update)
	topics.Delete("/:id", requireSuperAdmin, controller.Delete)
}

// FindAll func lists all topics.
// @Description list all topics.
// @Summary list topics
// @Tags Topic
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit number"
// @Success 200 {object} model.GeneralResponse{data=[]model.TopicResponse}
// @Router /api/v1/topics [get]
func (controller *TopicController) FindAll(c *fiber.Ctx) error {
	var filter model.TopicFilter
	if err := c.QueryParser(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	response, err := controller.TopicService.FindAll(c.UserContext(), filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// FindByID func gets a specific topic.
// @Description get a specific topic by UUID.
// @Summary get topic
// @Tags Topic
// @Accept json
// @Produce json
// @Param id path string true "Topic ID"
// @Success 200 {object} model.GeneralResponse{data=model.TopicResponse}
// @Router /api/v1/topics/{id} [get]
func (controller *TopicController) FindByID(c *fiber.Ctx) error {
	id := c.Params("id")
	response, err := controller.TopicService.FindByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// Create func creates a new topic.
// @Description create a new topic.
// @Summary create topic
// @Tags Topic
// @Accept json
// @Produce json
// @Param request body model.CreateTopicRequest true "Create Topic Request Body"
// @Success 201 {object} model.GeneralResponse{data=model.TopicResponse}
// @Security JWT
// @Router /api/v1/topics [post]
func (controller *TopicController) Create(c *fiber.Ctx) error {
	var request model.CreateTopicRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	response, err := controller.TopicService.Create(c.UserContext(), request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(model.GeneralResponse{
		Code:    201,
		Message: "Success",
		Data:    response,
	})
}

// Update func updates an existing topic.
// @Description update an existing topic.
// @Summary update topic
// @Tags Topic
// @Accept json
// @Produce json
// @Param id path string true "Topic ID"
// @Param request body model.UpdateTopicRequest true "Update Topic Request Body"
// @Success 200 {object} model.GeneralResponse{data=model.TopicResponse}
// @Security JWT
// @Router /api/v1/topics/{id} [put]
func (controller *TopicController) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var request model.UpdateTopicRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	response, err := controller.TopicService.Update(c.UserContext(), request, id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// Delete func deletes a topic.
// @Description delete a topic by UUID.
// @Summary delete topic
// @Tags Topic
// @Accept json
// @Produce json
// @Param id path string true "Topic ID"
// @Success 200 {object} model.GeneralResponse{data=string}
// @Security JWT
// @Router /api/v1/topics/{id} [delete]
func (controller *TopicController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := controller.TopicService.Delete(c.UserContext(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    "Topic deleted successfully",
	})
}
