package middleware

import (
	"strings"

	"evasbr/mclamg/configuration"
	"evasbr/mclamg/model"
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v9"
	"github.com/golang-jwt/jwt/v4"
)

// RequireAuth is a middleware that verifies JWT and checks roles or permissions depending on AUTH_MODE.
func RequireAuth(allowed []string, config configuration.Config, redisClient *redis.Client) fiber.Handler {
	jwtSecret := config.Get("JWT_SECRET_KEY")

	// Validate configuration at startup / route registration time (acting as a build/initialization error)
	if allowed == nil {
		panic("RequireAuth initialization error: allowed list cannot be nil")
	}

	return func(c *fiber.Ctx) error {
		// Read AUTH_MODE at request time to ensure it respects dynamic environment values if modified
		authModeReq := config.Get("AUTH_MODE")

		// 1. Extract Access Token from Cookie or Authorization Header
		tokenStr := c.Cookies("access_token")
		if tokenStr == "" {
			// Fallback to Authorization Header
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenStr = parts[1]
				} else if len(parts) == 1 {
					tokenStr = parts[0]
				}
			}
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
				Code:    401,
				Message: "Unauthorized",
				Data:    "Missing Access Token",
			})
		}

		// 2. Validate token (check expiration and signature)
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			// Check if the error is due to token expiration
			if ve, ok := err.(*jwt.ValidationError); ok && (ve.Errors&jwt.ValidationErrorExpired) != 0 {
				return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
					Code:    401,
					Message: "Access token expired",
					Data:    "EXPIRED_ACCESS_TOKEN",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
				Code:    401,
				Message: "Unauthorized",
				Data:    "Invalid Access Token",
			})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
				Code:    401,
				Message: "Unauthorized",
				Data:    "Invalid Access Token",
			})
		}

		// 3. Check if token exists in Redis whitelist
		exists, rErr := redisClient.Exists(c.Context(), "whitelist:token:"+tokenStr).Result()
		if rErr != nil || exists == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(model.GeneralResponse{
				Code:    401,
				Message: "Unauthorized",
				Data:    "Access Token is revoked or not whitelisted",
			})
		}

		// 4. Save payload to Locals
		c.Locals("user", token)

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
				Code:    403,
				Message: "Forbidden",
				Data:    "Invalid token claims",
			})
		}

		// 5. Check auth mode
		if len(allowed) == 0 {
			return c.Next()
		}

		switch authModeReq {
		case "RBAC":
			tokenRolesRaw, exists := claims["roles"]
			if !exists {
				return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
					Code:    403,
					Message: "Forbidden",
					Data:    "Access Denied (Roles claim missing)",
				})
			}

			tokenRolesSlice, ok := tokenRolesRaw.([]interface{})
			if !ok {
				return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
					Code:    403,
					Message: "Forbidden",
					Data:    "Access Denied (Invalid roles claims)",
				})
			}

			var tokenRoles []string
			for _, r := range tokenRolesSlice {
				if str, ok := r.(string); ok {
					tokenRoles = append(tokenRoles, str)
				}
			}

			match := false
			for _, role := range tokenRoles {
				for _, allowedRole := range allowed {
					if role == allowedRole {
						match = true
						break
					}
				}
				if match {
					break
				}
			}

			if !match {
				return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
					Code:    403,
					Message: "Forbidden",
					Data:    "Access Denied (Insufficient Role)",
				})
			}

		case "PBAC":
			tokenPermissionsRaw, exists := claims["permissions"]
			if !exists {
				return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
					Code:    403,
					Message: "Forbidden",
					Data:    "Access Denied (Permissions claim missing)",
				})
			}

			tokenPermissionsMap, ok := tokenPermissionsRaw.(map[string]interface{})
			if !ok {
				return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
					Code:    403,
					Message: "Forbidden",
					Data:    "Access Denied (Invalid permissions claims format)",
				})
			}

			if !hasPermission(tokenPermissionsMap, allowed) {
				return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
					Code:    403,
					Message: "Forbidden",
					Data:    "Access Denied (Insufficient Permission)",
				})
			}

		default:
			// Fallback: block access if authMode is empty or invalid
			return c.Status(fiber.StatusForbidden).JSON(model.GeneralResponse{
				Code:    403,
				Message: "Forbidden",
				Data:    "Access Denied (Invalid Auth Mode)",
			})
		}

		return c.Next()
	}
}

// hasPermission checks if user permissions contain FULLACCESS: true or at least one of the allowedPermissions.
func hasPermission(userPerms map[string]interface{}, allowedPermissions []string) bool {
	// Check FULLACCESS
	if fa, ok := userPerms["FULLACCESS"]; ok {
		if faBool, ok := fa.(bool); ok && faBool {
			return true
		}
	}

	// Check each allowed permission
	for _, allowed := range allowedPermissions {
		for _, val := range userPerms {
			if list, ok := val.([]interface{}); ok {
				for _, p := range list {
					if pStr, ok := p.(string); ok && pStr == allowed {
						return true
					}
				}
			} else if listStr, ok := val.([]string); ok {
				for _, p := range listStr {
					if p == allowed {
						return true
					}
				}
			}
		}
	}
	return false
}
