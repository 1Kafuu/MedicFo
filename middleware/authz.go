package middleware

import (
	"medicfo/helper"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="api"`)
			return helper.ErrUnauthorized
		}
		if !perms.Can(user.Role, permission) {
			return helper.ErrForbidden
		}
		return c.Next()
	}
}
