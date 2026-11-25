package route

import (
	"github.com/gofiber/fiber/v2"

	"medicfo/app/service"
	"medicfo/middleware"
)

func RegisterAdmin(api fiber.Router, deps Dependencies) {
	admin := api.Group("/admin", middleware.RequireAuth(deps.JWT))
	admin.Get("/queue",
		middleware.RequirePermission(deps.Permissions, "admin:queue"),
		monitorQueue(deps.NotificationService),
	)
}

func monitorQueue(svc *service.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		stats, err := svc.Monitor(c.Context())
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "keadaan queue berhasil diambil", stats)
	}
}
