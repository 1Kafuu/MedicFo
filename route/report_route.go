package route

import (
	"github.com/gofiber/fiber/v2"

	"medicfo/app/service"
	"medicfo/helper"
	"medicfo/middleware"
)

func RegisterReport(api fiber.Router, deps Dependencies) {
	reports := api.Group("/reports", middleware.RequireAuth(deps.JWT))
	reports.Get("/appointments",
		middleware.RequirePermission(deps.Permissions, "report:read"),
		listAppointmentReport(deps.ReportService),
	)
}

func listAppointmentReport(svc *service.ReportService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Format dipilih lebih dulu: format yang tidak didukung tidak perlu
		// membebani database dengan query yang akan dibuang.
		format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
		if err != nil {
			return err
		}
		items, err := svc.AppointmentsByDate(c.Context(), c.Query("date"))
		if err != nil {
			return err
		}
		if format == helper.FormatCSV {
			return helper.WriteAppointmentsCSV(c, items)
		}
		return respond(c, fiber.StatusOK, "ok", items)
	}
}
