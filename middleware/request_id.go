package middleware

import (
	"context"
	"evasbr/mclamg/common"
	"github.com/gofiber/fiber/v2"
)

// RequestID is a middleware that propagates the unique request ID (set by the
// standard requestid middleware) into the Go context.Context (UserContext)
// so it is accessible to services and repositories.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Fetch the request ID from Locals (which is populated by the standard requestid middleware)
		reqID := c.Locals("requestid")
		if reqID != nil {
			if strID, ok := reqID.(string); ok && strID != "" {
				// Inject it into the UserContext using our custom key type and string key
				ctx := c.UserContext()
				ctx = context.WithValue(ctx, common.RequestIDKey, strID)
				ctx = context.WithValue(ctx, "requestid", strID)
				c.SetUserContext(ctx)
			}
		}

		return c.Next()
	}
}
