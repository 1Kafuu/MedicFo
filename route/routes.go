package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"medicfo/app/service"
	"medicfo/helper"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	AuthService *service.AuthService
}

// Register memasang seluruh API di bawah /api/v1.
func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool))
	RegisterAuth(api, deps)
	// FT-03: RegisterDoctor(api, deps.DoctorService)  ← 1 baris + file sendiri
	// FT-05: RegisterAppointment(api, deps.AppointmentService)
}

// ErrorHandler memetakan SEMUA error menjadi envelope di satu tempat.
func ErrorHandler(c *fiber.Ctx, err error) error {
	body, status := helper.FromError(err)
	return c.Status(status).SendString(body)
}
