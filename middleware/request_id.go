package middleware

import (
	"context"
	"evasbr/mclamg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// RequestID is a middleware that generates a unique ID for each request
// and propagates it into the Go context.Context (UserContext) so it is
// accessible to services and repositories.
func RequestID() fiber.Handler {
	// First initialize standard requestid middleware
	ridMiddleware := requestid.New()

	return func(c *fiber.Ctx) error {
		// 1. Run standard requestid middleware to generate requestid local
		if err := ridMiddleware(c); err != nil {
			return err
		}

		// 2. Fetch the request ID from Locals
		reqID := c.Locals("requestid")
		if reqID != nil {
			if strID, ok := reqID.(string); ok && strID != "" {
				// 3. Inject it into the UserContext using our custom key type and string key
				ctx := c.UserContext()
				ctx = context.WithValue(ctx, common.RequestIDKey, strID)
				ctx = context.WithValue(ctx, "requestid", strID)
				c.SetUserContext(ctx)
			}
		}

		return c.Next()
	}
}
