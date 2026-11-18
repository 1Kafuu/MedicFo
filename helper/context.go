package helper

import (
	"github.com/gofiber/fiber/v2"

	"medicfo/app/model"
)

// LocalsAuthUser adalah kunci penyimpanan identitas di Locals.
const LocalsAuthUser = "authUser"

// CurrentUser membaca identitas yang dipasang middleware RequireAuth.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
