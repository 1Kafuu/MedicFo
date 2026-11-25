package route

import (
	"github.com/gofiber/fiber/v2"

	"medicfo/app/service"
	"medicfo/helper"
	"medicfo/middleware"
)

func RegisterNotification(api fiber.Router, deps Dependencies) {
	notifications := api.Group("/notifications",
		middleware.RequireAuth(deps.JWT),
		middleware.RequirePermission(deps.Permissions, "notification:read"),
	)
	notifications.Get("/mine", listNotifications(deps.NotificationService))
}

func listNotifications(svc *service.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.ErrUnauthorized
		}
		items, err := svc.ListForUser(c.Context(), user.UserID)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "ok", items)
	}
}
