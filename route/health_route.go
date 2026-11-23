package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"medicfo/helper"
)

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ErrServiceUnavailable
		}
		body, status := helper.Ok(
			fiber.StatusOK,
			"server dan PostgreSQL berjalan",
			nil,
			nil,
		)

		return c.Status(status).Type("json").SendString(body)
	}
}
