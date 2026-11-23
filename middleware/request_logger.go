package middleware

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"medicfo/helper"
)

func RegisterHTTP(app *fiber.App, logger *slog.Logger, allowedOrigins string) {
	app.Use(requestid.New())
	app.Use(helmet.New())
	app.Use(CORS(allowedOrigins))
	app.Use(RequestLogger(logger))
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		started := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			var appErr *helper.AppError
			if errors.As(err, &appErr) {
				status = appErr.Status
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(started)),
			slog.String("ip", c.IP()),
		}
		if requestID, ok := c.Locals("requestid").(string); ok && requestID != "" {
			attrs = append(attrs, slog.String("request_id", requestID))
		}
		if user, ok := helper.CurrentUser(c); ok {
			attrs = append(attrs, slog.Int64("user_id", user.UserID), slog.String("role", user.Role))
		}
		logger.Info("http_request", attrs...)
		return err
	}
}
