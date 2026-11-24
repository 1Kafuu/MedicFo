package route

import (
	"github.com/gofiber/fiber/v2"

	"medicfo/app/model"
	"medicfo/app/service"
	"medicfo/helper"
	"medicfo/middleware"
)

func RegisterDoctor(api fiber.Router, deps Dependencies) {
	doctors := api.Group("/doctors", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)
	doctors.Post("/", middleware.RequirePermission(deps.Permissions, "doctor:create"), createDoctor(deps.DoctorService))
	doctors.Patch("/:id", middleware.RequirePermission(deps.Permissions, "doctor:update"), updateDoctor(deps.DoctorService))
	doctors.Get("/", middleware.RequirePermission(deps.Permissions, "doctor:list"), listDoctors(deps.DoctorService))
	doctors.Get("/:id", middleware.RequirePermission(deps.Permissions, "doctor:list"), getDoctor(deps.DoctorService))
}

func createDoctor(svc *service.DoctorService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.CreateDoctorRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrBadRequest
		}

		doctor, err := svc.Create(c.Context(), req)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusCreated, "doctor created", doctor)
	}
}

func updateDoctor(svc *service.DoctorService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return helper.ErrBadRequest
		}

		var req model.UpdateDoctorRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrBadRequest
		}

		doctor, err := svc.Update(c.Context(), int64(id), req)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "doctor updated", doctor)
	}
}

func listDoctors(svc *service.DoctorService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, err := svc.List(c.Context(), c.QueryInt("limit", 0), c.Query("cursor"))
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "ok", page)
	}
}

func getDoctor(svc *service.DoctorService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return helper.ErrBadRequest
		}

		doctor, err := svc.Get(c.Context(), int64(id))
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "ok", doctor)
	}
}
