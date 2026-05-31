package middleware

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/model"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

// AuthenticateJWT is a route protection middleware (Guard) that verifies the JSON Web Token (JWT)
// from the request header 'Authorization: Bearer <token>' and performs role-based authorization.
//
// Key Features:
// 1. Checks the presence & validity of the JWT token using 'JWT_SECRET_KEY'.
// 2. Extracts the role claims ('roles') from the token payload and validates them against the required role.
// 3. Injects the verified user token into Fiber's local context (accessible via c.Locals("user")).
//
// Route registration example (in Controller):
//
//	func (controller *ProductController) Route(app *fiber.App) {
//	    // Only users with 'ROLE_ADMIN' are allowed to access this endpoint
//	    app.Post("/v1/api/product", middleware.AuthenticateJWT("ROLE_ADMIN", controller.Config), controller.Create)
//	}
//
// Extracting User Payload in Controller example:
//
//	func (controller *ProductController) Create(c *fiber.Ctx) error {
//	    // 1. Retrieve the jwt.Token from Locals
//	    userToken := c.Locals("user").(*jwt.Token)
//
//	    // 2. Extract claims (JWT payload)
//	    claims := userToken.Claims.(jwt.MapClaims)
//	    username := claims["username"].(string)
//	    roles := claims["roles"].([]interface{})
//
//	    // ... use username/roles for audit logging or database records ...
//	}
func AuthenticateJWT(role string, config configuration.Config) func(*fiber.Ctx) error {
	jwtSecret := config.Get("JWT_SECRET_KEY")
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(jwtSecret),
		SuccessHandler: func(ctx *fiber.Ctx) error {
			user := ctx.Locals("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			roles := claims["roles"].([]interface{})

			common.Logger(ctx.UserContext(), "JWT").Info("role function ", role, " role user ", roles)
			for _, roleInterface := range roles {
				roleMap := roleInterface.(map[string]interface{})
				if roleMap["role"] == role {
					return ctx.Next()
				}
			}

			return ctx.
				Status(fiber.StatusUnauthorized).
				JSON(model.GeneralResponse{
					Code:    401,
					Message: "Unauthorized",
					Data:    "Invalid Role",
				})
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return c.
					Status(fiber.StatusBadRequest).
					JSON(model.GeneralResponse{
						Code:    400,
						Message: "Bad Request",
						Data:    "Missing or malformed JWT",
					})
			} else {
				return c.
					Status(fiber.StatusUnauthorized).
					JSON(model.GeneralResponse{
						Code:    401,
						Message: "Unauthorized",
						Data:    "Invalid or expired JWT",
					})
			}
		},
	})
}
