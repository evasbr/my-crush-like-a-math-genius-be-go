package controller

import (
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

type AttemptController struct {
	AttemptService service.AttemptService
	Config         configuration.Config
	Redis          *redis.Client
	log            *logrus.Entry
}

func NewAttemptController(attemptService *service.AttemptService, config configuration.Config, redis *redis.Client) *AttemptController {
	return &AttemptController{
		AttemptService: *attemptService,
		Config:         config,
		Redis:          redis,
		log:            common.Log.WithField("scope", "AttemptController"),
	}
}

func (controller *AttemptController) Route(router fiber.Router) {
	attempts := router.Group("/attempts")
	requireAuth := middleware.RequireAuth([]string{}, controller.Config, controller.Redis)

	// List history (supports query filter topicId)
	attempts.Get("/", requireAuth, controller.FindAll)
	// Start attempt
	attempts.Post("/start", requireAuth, controller.StartAttempt)
	// Submit answer (linear step)
	attempts.Post("/submit-answer", requireAuth, controller.SubmitAnswer)
	// Get active/next question
	attempts.Get("/:id/next-question", requireAuth, controller.GetNextQuestion)
	// FindOne attempt details
	// attempts.Get("/:id", requireAuth, controller.FindByID)
	// Get attempt details by attempt ID
	attempts.Get("/:id/details", requireAuth, controller.GetAttemptDetails)

	// Topic attempts level info
	router.Get("/topics/:topicId/attempts/info", requireAuth, controller.GetTopicAttemptInfo)
}

func (controller *AttemptController) getUserId(c *fiber.Ctx) (string, error) {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Missing or invalid JWT token")
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fiber.NewError(fiber.StatusForbidden, "Invalid token claims")
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", fiber.NewError(fiber.StatusForbidden, "Missing user_id claim in JWT")
	}
	return userIDStr, nil
}

// GetTopicAttemptInfo func gets level settings stats and attempts info for a topic.
// @Description gets attempts info & level question counts for a specific topic.
// @Summary topic attempts info
// @Tags Attempt
// @Accept json
// @Produce json
// @Param topicId path string true "Topic ID"
// @Success 200 {object} model.GeneralResponse{data=model.TopicAttemptInfoResponse}
// @Security JWT
// @Router /api/v1/topics/{topicId}/attempts/info [get]
func (controller *AttemptController) GetTopicAttemptInfo(c *fiber.Ctx) error {
	topicId := c.Params("topicId")
	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	response, err := controller.AttemptService.GetTopicAttemptInfo(c.UserContext(), topicId, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// StartAttempt func starts a quiz attempt.
// @Description starts a new attempt session.
// @Summary start attempt
// @Tags Attempt
// @Accept json
// @Produce json
// @Param request body model.StartAttemptRequest true "Start Attempt Request Body"
// @Success 200 {object} model.GeneralResponse{data=model.AttemptSessionResponse}
// @Security JWT
// @Router /api/v1/attempts/start [post]
func (controller *AttemptController) StartAttempt(c *fiber.Ctx) error {
	var request model.StartAttemptRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	response, err := controller.AttemptService.StartAttempt(c.UserContext(), request, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// GetNextQuestion func gets the active/next question.
// @Description gets the next unanswered question in the attempt session.
// @Summary get active question
// @Tags Attempt
// @Accept json
// @Produce json
// @Param id path string true "Attempt Session ID"
// @Success 200 {object} model.GeneralResponse{data=model.ActiveQuestionResponse}
// @Security JWT
// @Router /api/v1/attempts/{id}/next-question [get]
func (controller *AttemptController) GetNextQuestion(c *fiber.Ctx) error {
	id := c.Params("id")
	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	response, err := controller.AttemptService.GetNextQuestion(c.UserContext(), id, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// SubmitAnswer func submits an answer to a question.
// @Description submits a single answer option and returns correctness/session status.
// @Summary submit answer
// @Tags Attempt
// @Accept json
// @Produce json
// @Param request body model.SubmitAnswerRequest true "Submit Answer Request Body"
// @Success 200 {object} model.GeneralResponse{data=model.SubmitAnswerResponse}
// @Security JWT
// @Router /api/v1/attempts/submit-answer [post]
func (controller *AttemptController) SubmitAnswer(c *fiber.Ctx) error {
	var request model.SubmitAnswerRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	response, err := controller.AttemptService.SubmitAnswer(c.UserContext(), request, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// FindByID func gets a specific attempt session history detail.
// @Description gets attempt session result/details by ID.
// @Summary get attempt detail
// @Tags Attempt
// @Accept json
// @Produce json
// @Param id path string true "Attempt Session ID"
// @Success 200 {object} model.GeneralResponse{data=model.AttemptSessionResponse}
// @Security JWT
// @Router /api/v1/attempts/{id} [get]
func (controller *AttemptController) FindByID(c *fiber.Ctx) error {
	id := c.Params("id")
	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	response, err := controller.AttemptService.FindByID(c.UserContext(), id, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// GetAttemptDetails func gets attempt details by attempt ID.
// @Description gets list of attempt details (questions and answers) by attempt session ID.
// @Summary get attempt details
// @Tags Attempt
// @Accept json
// @Produce json
// @Param id path string true "Attempt Session ID"
// @Success 200 {object} model.GeneralResponse{data=[]model.AttemptDetailDto}
// @Security JWT
// @Router /api/v1/attempts/{id}/details [get]
func (controller *AttemptController) GetAttemptDetails(c *fiber.Ctx) error {
	id := c.Params("id")
	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	response, err := controller.AttemptService.GetAttemptDetails(c.UserContext(), id, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}


// FindAll func retrieves attempt history list.
// @Description gets list of attempt sessions for the user.
// @Summary attempt history list
// @Tags Attempt
// @Accept json
// @Produce json
// @Param topicId query string false "Filter by Topic ID"
// @Param page query int false "Page number"
// @Param limit query int false "Limit number"
// @Success 200 {object} model.GeneralResponse{data=[]model.AttemptSessionResponse}
// @Security JWT
// @Router /api/v1/attempts [get]
func (controller *AttemptController) FindAll(c *fiber.Ctx) error {
	var filter model.AttemptFilter
	if err := c.QueryParser(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	userId, err := controller.getUserId(c)
	if err != nil {
		return err
	}

	sessions, _, err := controller.AttemptService.FindAll(c.UserContext(), filter, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    sessions,
	})
}
