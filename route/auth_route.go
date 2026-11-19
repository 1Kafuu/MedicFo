package route

import (
	"github.com/gofiber/fiber/v2"

	"medicfo/app/model"
	"medicfo/app/service"
	"medicfo/helper"
	"medicfo/middleware"
)

// RegisterAuth memasang semua endpoint /auth.
// Menerima fiber.Router (grup /api/v1 dari routes.go) — tidak tahu cara grup dibuat.
func RegisterAuth(api fiber.Router, deps Dependencies) {
	auth := api.Group("/auth")
	auth.Post("/register", register(deps.AuthService))
	auth.Post("/login", login(deps.AuthService))
	auth.Post("/refresh", refresh(deps.AuthService))
	auth.Post("/logout", logout(deps.AuthService))
	auth.Get("/me", middleware.RequireAuth(deps.JWT), me(deps.AuthService))
}

// respond: pembungkus kecil agar handler konsisten.
func respond(c *fiber.Ctx, status int, message string, data any) error {
	body, s := helper.Ok(status, message, data, nil)
	return c.Status(s).SendString(body)
}

// register: POST /api/v1/auth/register (201)
func register(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrBadRequest
		}

		user, err := svc.Register(c.Context(), req)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusCreated, "register success", user)
	}
}

// login: POST /api/v1/auth/login (200, pasangan token)
func login(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrBadRequest
		}

		pair, err := svc.Login(c.Context(), req)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "login success", pair)
	}
}

// refresh: POST /api/v1/auth/refresh (200, pasangan token baru)
func refresh(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.RefreshRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrBadRequest
		}

		pair, err := svc.Refresh(c.Context(), req)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "token refreshed", pair)
	}
}

// logout: POST /api/v1/auth/logout (200, selalu sukses)
func logout(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.RefreshRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.ErrBadRequest
		}

		if err := svc.Logout(c.Context(), req); err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "logout success", nil)
	}
}

// me: GET /api/v1/auth/me (200, profil user dari access token)
func me(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authUser, ok := helper.CurrentUser(c)
		if !ok {
			return helper.ErrUnauthorized
		}

		user, err := svc.Me(c.Context(), authUser)
		if err != nil {
			return err
		}
		return respond(c, fiber.StatusOK, "profile fetched", user)
	}
}
