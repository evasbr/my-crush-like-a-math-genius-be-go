package controller

import (
	"encoding/json"
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"fmt"

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
// @Param classroomId query string false "Filter by Classroom ID"
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
// @Accept multipart/form-data
// @Produce json
// @Param classroom_id formData string true "Classroom ID"
// @Param name formData string true "Topic Name"
// @Param description formData string false "Description"
// @Param female_normal_dialog formData string true "Female Normal Dialog"
// @Param male_normal_dialog formData string true "Male Normal Dialog"
// @Param female_dating_dialog formData string true "Female Dating Dialog"
// @Param male_dating_dialog formData string true "Male Dating Dialog"
// @Param status formData string true "Status"
// @Param level_settings formData string true "Level Settings JSON Array"
// @Param max_attempts formData integer true "Max Attempts"
// @Param female_normal_img formData file true "Female Normal image file (JPEG, PNG, WebP, max 1MB)"
// @Param male_normal_img formData file true "Male Normal image file (JPEG, PNG, WebP, max 1MB)"
// @Param female_dating_img formData file true "Female Dating image file (JPEG, PNG, WebP, max 1MB)"
// @Param male_dating_img formData file true "Male Dating image file (JPEG, PNG, WebP, max 1MB)"
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

	if len(request.LevelSettings) == 0 && c.FormValue("level_settings") != "" {
		var settings []model.LevelSettingDto
		if err := json.Unmarshal([]byte(c.FormValue("level_settings")), &settings); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid level_settings JSON format: %v", err),
			})
		}
		request.LevelSettings = settings
	}

	femaleNormalHeader, err := c.FormFile("female_normal_img")
	if err == nil && femaleNormalHeader != nil {
		if err := common.ValidateImageFile(femaleNormalHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid female_normal_img: %v", err),
			})
		}
	}

	maleNormalHeader, err := c.FormFile("male_normal_img")
	if err == nil && maleNormalHeader != nil {
		if err := common.ValidateImageFile(maleNormalHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid male_normal_img: %v", err),
			})
		}
	}

	femaleDatingHeader, err := c.FormFile("female_dating_img")
	if err == nil && femaleDatingHeader != nil {
		if err := common.ValidateImageFile(femaleDatingHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid female_dating_img: %v", err),
			})
		}
	}

	maleDatingHeader, err := c.FormFile("male_dating_img")
	if err == nil && maleDatingHeader != nil {
		if err := common.ValidateImageFile(maleDatingHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid male_dating_img: %v", err),
			})
		}
	}

	response, err := controller.TopicService.Create(c.UserContext(), request, femaleNormalHeader, maleNormalHeader, femaleDatingHeader, maleDatingHeader)
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
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Topic ID"
// @Param classroom_id formData string false "Classroom ID"
// @Param name formData string false "Topic Name"
// @Param description formData string false "Description"
// @Param female_normal_dialog formData string false "Female Normal Dialog"
// @Param male_normal_dialog formData string false "Male Normal Dialog"
// @Param female_dating_dialog formData string false "Female Dating Dialog"
// @Param male_dating_dialog formData string false "Male Dating Dialog"
// @Param status formData string false "Status"
// @Param level_settings formData string false "Level Settings JSON Array"
// @Param max_attempts formData integer false "Max Attempts"
// @Param female_normal_img formData file false "Female Normal image file (JPEG, PNG, WebP, max 1MB)"
// @Param male_normal_img formData file false "Male Normal image file (JPEG, PNG, WebP, max 1MB)"
// @Param female_dating_img formData file false "Female Dating image file (JPEG, PNG, WebP, max 1MB)"
// @Param male_dating_img formData file false "Male Dating image file (JPEG, PNG, WebP, max 1MB)"
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

	if len(request.LevelSettings) == 0 && c.FormValue("level_settings") != "" {
		var settings []model.LevelSettingDto
		if err := json.Unmarshal([]byte(c.FormValue("level_settings")), &settings); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid level_settings JSON format: %v", err),
			})
		}
		request.LevelSettings = settings
	}

	femaleNormalHeader, err := c.FormFile("female_normal_img")
	if err == nil && femaleNormalHeader != nil {
		if err := common.ValidateImageFile(femaleNormalHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid female_normal_img: %v", err),
			})
		}
	}

	maleNormalHeader, err := c.FormFile("male_normal_img")
	if err == nil && maleNormalHeader != nil {
		if err := common.ValidateImageFile(maleNormalHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid male_normal_img: %v", err),
			})
		}
	}

	femaleDatingHeader, err := c.FormFile("female_dating_img")
	if err == nil && femaleDatingHeader != nil {
		if err := common.ValidateImageFile(femaleDatingHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid female_dating_img: %v", err),
			})
		}
	}

	maleDatingHeader, err := c.FormFile("male_dating_img")
	if err == nil && maleDatingHeader != nil {
		if err := common.ValidateImageFile(maleDatingHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid male_dating_img: %v", err),
			})
		}
	}

	response, err := controller.TopicService.Update(c.UserContext(), request, femaleNormalHeader, maleNormalHeader, femaleDatingHeader, maleDatingHeader, id)
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
