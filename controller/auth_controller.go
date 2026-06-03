package controller

import (
	"strconv"
	"strings"
	"time"

	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	AuthService service.AuthService
	Config      configuration.Config
	Redis       *redis.Client
	log         *logrus.Entry
}

func NewAuthController(authService *service.AuthService, config configuration.Config, redis *redis.Client) *AuthController {
	return &AuthController{
		AuthService: *authService,
		Config:      config,
		Redis:       redis,
		log:         common.Log.WithField("scope", "AuthController"),
	}
}

func (controller *AuthController) Route(router fiber.Router) {
	authentication := router.Group("/authentication")
	authentication.Post("/login", controller.Login)
	authentication.Post("/register", controller.Register)
	authentication.Post("/refresh", controller.Refresh)
	authentication.Post("/logout", controller.Logout)
}

// Register func handles user registration.
// @Description register a new user.
// @Summary register user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register Request Body"
// @Success 201 {object} model.GeneralResponse{data=model.RegisterResponse}
// @Router /api/v1/authentication/register [post]
func (controller *AuthController) Register(c *fiber.Ctx) error {
	var request model.RegisterRequest
	err := c.BodyParser(&request)
	exception.PanicLogging(err)

	newUser, err := controller.AuthService.Register(c.UserContext(), request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.GeneralResponse{
			Code:    400,
			Message: "Bad Request",
			Data:    err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.GeneralResponse{
		Code:    201,
		Message: "Success",
		Data:    newUser,
	})
}

// Login func handles user authentication.
// @Description login a user.
// @Summary login user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login Request Body"
// @Success 200 {object} model.GeneralResponse{data=model.LoginResponse}
// @Router /api/v1/authentication/login [post]
func (controller *AuthController) Login(c *fiber.Ctx) error {
	var request model.LoginRequest
	err := c.BodyParser(&request)
	exception.PanicLogging(err)

	result, err := controller.AuthService.Login(c.UserContext(), request)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
			Code:    401,
			Message: "Unauthorized",
			Data:    err.Error(),
		})
	}

	// Fetch expiration settings to set matching cookie expirations
	jwtExpiredMinutes := 15
	if expStr := controller.Config.Get("JWT_EXPIRE_MINUTES_COUNT"); expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			jwtExpiredMinutes = val
		}
	}

	refreshExpiredMinutes := 10080
	if refExpStr := controller.Config.Get("JWT_REFRESH_EXPIRE_MINUTES_COUNT"); refExpStr != "" {
		if val, err := strconv.Atoi(refExpStr); err == nil {
			refreshExpiredMinutes = val
		}
	}

	// Set httpOnly cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Expires:  time.Now().Add(time.Minute * time.Duration(jwtExpiredMinutes)),
		HTTPOnly: true,
		Secure:   false, // Change to true if HTTPS in production
		SameSite: "Lax",
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Expires:  time.Now().Add(time.Minute * time.Duration(refreshExpiredMinutes)),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}

// Refresh func handles token rotation.
// @Description refresh access token using refresh token.
// @Summary refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse{data=model.RefreshTokenResponse}
// @Security JWT
// @Router /api/v1/authentication/refresh [post]
func (controller *AuthController) Refresh(c *fiber.Ctx) error {
	// Get Refresh Token from Cookie or Fallback Header
	refreshTokenStr := c.Cookies("refresh_token")
	if refreshTokenStr == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				refreshTokenStr = parts[1]
			} else if len(parts) == 1 {
				refreshTokenStr = parts[0]
			}
		}
	}

	if refreshTokenStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
			Code:    401,
			Message: "Refresh token missing",
			Data:    "MISSING_REFRESH_TOKEN",
		})
	}

	res, err := controller.AuthService.RefreshToken(c.UserContext(), refreshTokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
			Code:    401,
			Message: "Unauthorized",
			Data:    err.Error(),
		})
	}

	jwtExpiredMinutes := 15
	if expStr := controller.Config.Get("JWT_EXPIRE_MINUTES_COUNT"); expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			jwtExpiredMinutes = val
		}
	}

	// If refresh token rotated, set both cookies, otherwise just access token
	if res.Rotated {
		refreshExpiredMinutes := 10080
		if refExpStr := controller.Config.Get("JWT_REFRESH_EXPIRE_MINUTES_COUNT"); refExpStr != "" {
			if val, err := strconv.Atoi(refExpStr); err == nil {
				refreshExpiredMinutes = val
			}
		}

		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    res.AccessToken,
			Expires:  time.Now().Add(time.Minute * time.Duration(jwtExpiredMinutes)),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
		})

		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    res.RefreshToken,
			Expires:  time.Now().Add(time.Minute * time.Duration(refreshExpiredMinutes)),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
		})
	} else {
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    res.AccessToken,
			Expires:  time.Now().Add(time.Minute * time.Duration(jwtExpiredMinutes)),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    res,
	})
}

// Logout func logs out user.
// @Description log out user and clear tokens/cookies.
// @Summary logout user
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse{data=string}
// @Security JWT
// @Router /api/v1/authentication/logout [post]
func (controller *AuthController) Logout(c *fiber.Ctx) error {
	accessTokenStr := c.Cookies("access_token")
	refreshTokenStr := c.Cookies("refresh_token")

	// Fallback to headers
	if accessTokenStr == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				accessTokenStr = parts[1]
			} else if len(parts) == 1 {
				accessTokenStr = parts[0]
			}
		}
	}

	err := controller.AuthService.Logout(c.UserContext(), accessTokenStr, refreshTokenStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.GeneralResponse{
			Code:    500,
			Message: "Error logging out",
			Data:    err.Error(),
		})
	}

	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})

	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    "Logged out successfully",
	})
}
