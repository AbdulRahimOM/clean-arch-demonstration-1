// Package httputil provides HTTP test helpers shared across handler and integration tests.
package httputil

import (
	"github.com/gofiber/fiber/v2"
)

// UserIDMiddleware returns a Fiber middleware that sets "user_id" in Locals.
// Use in tests so handlers that read c.Locals("user_id") do not panic.
func UserIDMiddleware(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	}
}
