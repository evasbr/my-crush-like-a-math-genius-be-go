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

type LeaderboardController struct {
	ClassroomService service.ClassroomService
	Config           configuration.Config
	Redis            *redis.Client
	log              *logrus.Entry
}

func NewLeaderboardController(classroomService *service.ClassroomService, config configuration.Config, redis *redis.Client) *LeaderboardController {
	return &LeaderboardController{
		ClassroomService: *classroomService,
		Config:           config,
		Redis:            redis,
		log:              common.Log.WithField("scope", "LeaderboardController"),
	}
}

func (controller *LeaderboardController) Route(router fiber.Router) {
	requireAuth := middleware.RequireAuth([]string{}, controller.Config, controller.Redis)
	router.Get("/leaderboard", requireAuth, controller.GetLeaderboard)
}

func (controller *LeaderboardController) getRequestUserInfo(c *fiber.Ctx) (userID string, isSuperAdmin bool, err error) {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", false, fiber.NewError(fiber.StatusUnauthorized, "Missing or invalid JWT token in context")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", false, fiber.NewError(fiber.StatusForbidden, "Invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", false, fiber.NewError(fiber.StatusForbidden, "Missing user_id claim in JWT")
	}

	var roles []interface{}
	rolesVal := claims["roles"]
	if rolesVal != nil {
		roles, _ = rolesVal.([]interface{})
	}

	for _, role := range roles {
		if rStr, ok := role.(string); ok && rStr == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	return userIDStr, isSuperAdmin, nil
}

// GetLeaderboard func gets classroom leaderboard.
// @Description get classroom leaderboard.
// @Summary get classroom leaderboard
// @Tags Leaderboard
// @Accept json
// @Produce json
// @Param classroomId query string true "Classroom ID"
// @Param topicId query string false "Filter by Topic ID"
// @Success 200 {object} model.GeneralResponse{data=[]model.LeaderboardEntry}
// @Security JWT
// @Router /api/v1/leaderboard [get]
func (controller *LeaderboardController) GetLeaderboard(c *fiber.Ctx) error {
	classroomId := c.Query("classroomId")
	if classroomId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    "classroomId query parameter is required",
		})
	}
	topicID := c.Query("topicId")
	userIDStr, isSuperAdmin, err := controller.getRequestUserInfo(c)
	if err != nil {
		return err
	}

	response, err := controller.ClassroomService.GetLeaderboard(c.UserContext(), classroomId, topicID, userIDStr, isSuperAdmin)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}
