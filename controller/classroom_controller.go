package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type ClassroomController struct {
	ClassroomService service.ClassroomService
	Config           configuration.Config
	Redis            *redis.Client
	log              *logrus.Entry
}

func NewClassroomController(classroomService *service.ClassroomService, config configuration.Config, redis *redis.Client) *ClassroomController {
	return &ClassroomController{
		ClassroomService: *classroomService,
		Config:           config,
		Redis:            redis,
		log:              common.Log.WithField("scope", "ClassroomController"),
	}
}

func (controller *ClassroomController) Route(router fiber.Router) {
	classrooms := router.Group("/classrooms")

	// Protected routes (Super Admin only)
	requireSuperAdmin := middleware.RequireAuth([]string{"SUPER_ADMIN"}, controller.Config, controller.Redis)
	classrooms.Post("/", requireSuperAdmin, controller.Create)
	classrooms.Get("/", requireSuperAdmin, controller.FindAll)
	classrooms.Put("/:id", requireSuperAdmin, controller.Update)
	classrooms.Delete("/:id", requireSuperAdmin, controller.Delete)

	// Protected routes (Any authenticated user)
	requireAuth := middleware.RequireAuth([]string{}, controller.Config, controller.Redis)
	classrooms.Get("/me", requireAuth, controller.FindMyClassrooms)
	classrooms.Get("/:id", requireAuth, controller.FindByID)
	classrooms.Post("/join", requireAuth, controller.JoinByCode)
	classrooms.Get("/:id/members", requireAuth, controller.ListMembers)
}

func (controller *ClassroomController) getRequestUserInfo(c *fiber.Ctx) (userID string, isSuperAdmin bool, err error) {
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

	tokenRolesRaw, exists := claims["roles"]
	if exists {
		if tokenRolesSlice, ok := tokenRolesRaw.([]interface{}); ok {
			for _, r := range tokenRolesSlice {
				if str, ok := r.(string); ok && str == "SUPER_ADMIN" {
					isSuperAdmin = true
					break
				}
			}
		}
	}

	return userIDStr, isSuperAdmin, nil
}

// Create func creates a new classroom.
// @Description create a new classroom.
// @Summary create classroom
// @Tags Classroom
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Classroom Name"
// @Param description formData string false "Description"
// @Param is_external_invite_enable formData boolean false "Enable External Invite"
// @Param cover_img formData file false "Cover image file (JPEG, PNG, WebP, max 1MB)"
// @Param wallpaper_img formData file false "Wallpaper image file (JPEG, PNG, WebP, max 1MB)"
// @Success 201 {object} model.GeneralResponse{data=model.ClassroomResponse}
// @Security JWT
// @Router /api/v1/classrooms [post]
func (controller *ClassroomController) Create(c *fiber.Ctx) error {
	userIDStr, _, err := controller.getRequestUserInfo(c)
	if err != nil {
		return err
	}

	var request model.CreateClassroomRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	coverHeader, err := c.FormFile("cover_img")
	if err == nil && coverHeader != nil {
		if err := common.ValidateImageFile(coverHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid cover_img: %v", err),
			})
		}
	}

	wallpaperHeader, err := c.FormFile("wallpaper_img")
	if err == nil && wallpaperHeader != nil {
		if err := common.ValidateImageFile(wallpaperHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid wallpaper_img: %v", err),
			})
		}
	}

	response, err := controller.ClassroomService.Create(c.UserContext(), request, coverHeader, wallpaperHeader, userIDStr)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(model.GeneralResponse{
		Code:    201,
		Message: "Success",
		Data:    response,
	})
}

// Update func updates an existing classroom.
// @Description update an existing classroom.
// @Summary update classroom
// @Tags Classroom
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Classroom ID"
// @Param name formData string false "Classroom Name"
// @Param description formData string false "Description"
// @Param is_external_invite_enable formData boolean false "Enable External Invite"
// @Param status formData string false "Status"
// @Param cover_img formData file false "Cover image file (JPEG, PNG, WebP, max 1MB)"
// @Param wallpaper_img formData file false "Wallpaper image file (JPEG, PNG, WebP, max 1MB)"
// @Success 200 {object} model.GeneralResponse{data=model.ClassroomResponse}
// @Security JWT
// @Router /api/v1/classrooms/{id} [put]
func (controller *ClassroomController) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var request model.UpdateClassroomRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	coverHeader, err := c.FormFile("cover_img")
	if err == nil && coverHeader != nil {
		if err := common.ValidateImageFile(coverHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid cover_img: %v", err),
			})
		}
	}

	wallpaperHeader, err := c.FormFile("wallpaper_img")
	if err == nil && wallpaperHeader != nil {
		if err := common.ValidateImageFile(wallpaperHeader); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
				Code:    400,
				Message: "Bad Request",
				Data:    fmt.Sprintf("invalid wallpaper_img: %v", err),
			})
		}
	}

	response, err := controller.ClassroomService.Update(c.UserContext(), request, coverHeader, wallpaperHeader, id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// Delete func deletes a classroom.
// @Description delete a classroom by UUID.
// @Summary delete classroom
// @Tags Classroom
// @Accept json
// @Produce json
// @Param id path string true "Classroom ID"
// @Success 200 {object} model.GeneralResponse{data=string}
// @Security JWT
// @Router /api/v1/classrooms/{id} [delete]
func (controller *ClassroomController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := controller.ClassroomService.Delete(c.UserContext(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    "Classroom deleted successfully",
	})
}

// FindAll func lists all classrooms.
// @Description list all classrooms.
// @Summary list classrooms
// @Tags Classroom
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse{data=[]model.ClassroomResponse}
// @Security JWT
// @Router /api/v1/classrooms [get]
func (controller *ClassroomController) FindAll(c *fiber.Ctx) error {
	response, err := controller.ClassroomService.FindAll(c.UserContext())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// FindMyClassrooms func lists classrooms joined/created by the current user.
// @Description list user's classrooms.
// @Summary list my classrooms
// @Tags Classroom
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse{data=[]model.ClassroomResponse}
// @Security JWT
// @Router /api/v1/classrooms/me [get]
func (controller *ClassroomController) FindMyClassrooms(c *fiber.Ctx) error {
	userIDStr, isSuperAdmin, err := controller.getRequestUserInfo(c)
	if err != nil {
		return err
	}

	response, err := controller.ClassroomService.FindMyClassrooms(c.UserContext(), userIDStr, isSuperAdmin)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// FindByID func gets a specific classroom.
// @Description get a specific classroom by UUID (user must be a member or admin).
// @Summary get classroom
// @Tags Classroom
// @Accept json
// @Produce json
// @Param id path string true "Classroom ID"
// @Success 200 {object} model.GeneralResponse{data=model.ClassroomResponse}
// @Security JWT
// @Router /api/v1/classrooms/{id} [get]
func (controller *ClassroomController) FindByID(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDStr, isSuperAdmin, err := controller.getRequestUserInfo(c)
	if err != nil {
		return err
	}

	response, err := controller.ClassroomService.FindByID(c.UserContext(), id, userIDStr, isSuperAdmin)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// JoinByCode func joins a classroom using an invite code.
// @Description join classroom by code.
// @Summary join classroom
// @Tags Classroom
// @Accept json
// @Produce json
// @Param request body model.JoinClassroomRequest true "Join Classroom Request Body"
// @Success 200 {object} model.GeneralResponse{data=model.ClassroomResponse}
// @Security JWT
// @Router /api/v1/classrooms/join [post]
func (controller *ClassroomController) JoinByCode(c *fiber.Ctx) error {
	userIDStr, _, err := controller.getRequestUserInfo(c)
	if err != nil {
		return err
	}

	var request model.JoinClassroomRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	response, err := controller.ClassroomService.JoinByCode(c.UserContext(), request, userIDStr)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// ListMembers func lists all members in a classroom.
// @Description list classroom members.
// @Summary list classroom members
// @Tags Classroom
// @Accept json
// @Produce json
// @Param id path string true "Classroom ID"
// @Success 200 {object} model.GeneralResponse{data=[]model.ClassroomMemberResponse}
// @Security JWT
// @Router /api/v1/classrooms/{id}/members [get]
func (controller *ClassroomController) ListMembers(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDStr, isSuperAdmin, err := controller.getRequestUserInfo(c)
	if err != nil {
		return err
	}

	response, err := controller.ClassroomService.ListMembers(c.UserContext(), id, userIDStr, isSuperAdmin)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}
