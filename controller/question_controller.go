package controller

import (
	"strings"

	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type QuestionController struct {
	QuestionService service.QuestionService
	Config          configuration.Config
	Redis           *redis.Client
	log             *logrus.Entry
}

func NewQuestionController(questionService *service.QuestionService, config configuration.Config, redis *redis.Client) *QuestionController {
	return &QuestionController{
		QuestionService: *questionService,
		Config:          config,
		Redis:           redis,
		log:             common.Log.WithField("scope", "QuestionController"),
	}
}

func (controller *QuestionController) Route(router fiber.Router) {
	questions := router.Group("/questions")

	// Optional authentication routes
	optionalAuth := middleware.OptionalAuth(controller.Config, controller.Redis)
	questions.Get("/", optionalAuth, controller.FindAll)
	questions.Get("/:id", optionalAuth, controller.FindByID)

	// Protected routes (Super Admin only)
	requireSuperAdmin := middleware.RequireAuth([]string{"super_admin"}, controller.Config, controller.Redis)
	questions.Post("/batch", requireSuperAdmin, controller.CreateBatch)
	questions.Put("/:id", requireSuperAdmin, controller.Update)
	questions.Delete("/:id", requireSuperAdmin, controller.Delete)
}

func (controller *QuestionController) checkIsSuperAdmin(c *fiber.Ctx) bool {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return false
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	rolesRaw, exists := claims["roles"]
	if !exists {
		return false
	}
	rolesSlice, ok := rolesRaw.([]interface{})
	if !ok {
		return false
	}
	for _, r := range rolesSlice {
		if rStr, ok := r.(string); ok && strings.EqualFold(rStr, "super_admin") {
			return true
		}
	}
	return false
}

// FindAll func lists all questions.
// @Description list all questions.
// @Summary list questions
// @Tags Question
// @Accept json
// @Produce json
// @Param topicId query string true "Filter by topic ID"
// @Param page query int false "Page number"
// @Param limit query int false "Limit number"
// @Success 200 {object} model.GeneralResponse{data=[]model.QuestionResponse}
// @Router /api/v1/questions [get]
func (controller *QuestionController) FindAll(c *fiber.Ctx) error {
	var filter model.QuestionFilter
	if err := c.QueryParser(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	common.Validate(filter)

	includeIsCorrect := controller.checkIsSuperAdmin(c)

	response, err := controller.QuestionService.FindAll(c.UserContext(), filter, includeIsCorrect)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// FindByID func gets a specific question.
// @Description get a specific question by UUID.
// @Summary get question
// @Tags Question
// @Accept json
// @Produce json
// @Param id path string true "Question ID"
// @Success 200 {object} model.GeneralResponse{data=model.QuestionResponse}
// @Router /api/v1/questions/{id} [get]
func (controller *QuestionController) FindByID(c *fiber.Ctx) error {
	id := c.Params("id")
	includeIsCorrect := controller.checkIsSuperAdmin(c)

	response, err := controller.QuestionService.FindByID(c.UserContext(), id, includeIsCorrect)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// CreateBatch func creates questions in batch.
// @Description create questions in batch for a topic.
// @Summary create questions batch
// @Tags Question
// @Accept json
// @Produce json
// @Param request body model.CreateQuestionBatchRequest true "Create Questions Batch Request Body"
// @Success 201 {object} model.GeneralResponse{data=[]model.QuestionResponse}
// @Security JWT
// @Router /api/v1/questions/batch [post]
func (controller *QuestionController) CreateBatch(c *fiber.Ctx) error {
	var request model.CreateQuestionBatchRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	response, err := controller.QuestionService.CreateBatch(c.UserContext(), request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(model.GeneralResponse{
		Code:    201,
		Message: "Success",
		Data:    response,
	})
}

// Update func updates an existing question and its answers.
// @Description update an existing question and sync its answers.
// @Summary update question
// @Tags Question
// @Accept json
// @Produce json
// @Param id path string true "Question ID"
// @Param request body model.UpdateQuestionRequest true "Update Question Request Body"
// @Success 200 {object} model.GeneralResponse{data=model.QuestionResponse}
// @Security JWT
// @Router /api/v1/questions/{id} [put]
func (controller *QuestionController) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var request model.UpdateQuestionRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	response, err := controller.QuestionService.Update(c.UserContext(), request, id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// Delete func deletes a question.
// @Description delete a question by UUID.
// @Summary delete question
// @Tags Question
// @Accept json
// @Produce json
// @Param id path string true "Question ID"
// @Success 200 {object} model.GeneralResponse{data=string}
// @Security JWT
// @Router /api/v1/questions/{id} [delete]
func (controller *QuestionController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := controller.QuestionService.Delete(c.UserContext(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    "Question deleted successfully",
	})
}
