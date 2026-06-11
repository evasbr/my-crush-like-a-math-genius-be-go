package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	UserService service.UserService
	Config      configuration.Config
	Redis       *redis.Client
	log         *logrus.Entry
}

func NewUserController(userService *service.UserService, config configuration.Config, redis *redis.Client) *UserController {
	return &UserController{
		UserService: *userService,
		Config:      config,
		Redis:       redis,
		log:         common.Log.WithField("scope", "UserController"),
	}
}

func (controller *UserController) Route(router fiber.Router) {
	users := router.Group("/users")
	users.Get("/me", middleware.RequireAuth([]string{}, controller.Config, controller.Redis), controller.GetMyProfile)
	users.Patch("/me/profile-picture", middleware.RequireAuth([]string{}, controller.Config, controller.Redis), controller.UpdateProfilePicture)
}

// GetMyProfile func gets current user profile.
// @Description gets current user profile based on JWT claim.
// @Summary get profile
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse{data=model.UserProfileResponse}
// @Security JWT
// @Router /api/v1/users/me [get]
func (controller *UserController) GetMyProfile(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
			Code:    401,
			Message: "Unauthorized",
			Data:    "Missing or invalid JWT token in context",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
			Code:    403,
			Message: "Forbidden",
			Data:    "Invalid token claims",
		})
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
			Code:    403,
			Message: "Forbidden",
			Data:    "Missing user_id claim in JWT",
		})
	}

	profile, err := controller.UserService.FindByID(c.UserContext(), userIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.GeneralResponse{
			Code:    404,
			Message: "Not Found",
			Data:    err.Error(),
		})
	}

	var lastName string
	if profile.LastName != nil {
		lastName = *profile.LastName
	}
	var gender string
	if profile.Gender != nil {
		gender = *profile.Gender
	}
	var profilePictureURL string
	if profile.ProfilePictureURL != nil {
		profilePictureURL = *profile.ProfilePictureURL
	}

	var roles []string
	for _, userRole := range profile.UserRoles {
		roles = append(roles, userRole.Role.Name)
	}

	var permissionsResponse map[string]interface{}
	if controller.Config.Get("AUTH_MODE") != "RBAC" {
		var permissionsList []map[string]interface{}
		for _, userRole := range profile.UserRoles {
			permissionsList = append(permissionsList, userRole.Role.Permissions)
		}
		permissionsResponse = common.MergePermissions(permissionsList)
	}

	username := profile.Email
	for _, a := range profile.Authentications {
		if a.Method == string(entity.MethodLocalUsername) {
			username = a.ProviderUserID
			break
		}
	}

	response := model.UserProfileResponse{
		Id:                profile.ID.String(),
		Email:             profile.Email,
		Username:          username,
		FirstName:         profile.FirstName,
		LastName:          lastName,
		Gender:            gender,
		ProfilePictureURL: profilePictureURL,
		Status:            profile.Status,
		Roles:             roles,
		Permissions:       permissionsResponse,
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// UpdateProfilePicture func updates current user profile picture.
// @Description updates current user profile picture.
// @Summary update profile picture
// @Tags User
// @Accept multipart/form-data
// @Produce json
// @Param profile_picture formData file true "Profile picture file (JPEG, PNG, WebP, max 1MB)"
// @Success 200 {object} model.GeneralResponse{data=model.UserProfileResponse}
// @Security JWT
// @Router /api/v1/users/me/profile-picture [patch]
func (controller *UserController) UpdateProfilePicture(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
			Code:    401,
			Message: "Unauthorized",
			Data:    "Missing or invalid JWT token in context",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
			Code:    403,
			Message: "Forbidden",
			Data:    "Invalid token claims",
		})
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
			Code:    403,
			Message: "Forbidden",
			Data:    "Missing user_id claim in JWT",
		})
	}

	fileHeader, err := c.FormFile("profile_picture")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    "profile_picture field is required",
		})
	}

	updatedUser, err := controller.UserService.UpdateProfilePicture(c.UserContext(), userIDStr, fileHeader)
	if err != nil {
		return err
	}

	var profilePictureURL string
	if updatedUser.ProfilePictureURL != nil {
		profilePictureURL = *updatedUser.ProfilePictureURL
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data: fiber.Map{
			"profile_picture_url": profilePictureURL,
		},
	})
}



