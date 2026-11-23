package middleware

import (
	"mime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"medicfo/helper"
)

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func CORS(allowedOrigins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	})
}

func RequireJSON(c *fiber.Ctx) error {
	if !methodsWithBody[c.Method()] {
		return c.Next()
	}

	contentType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || contentType != fiber.MIMEApplicationJSON {
		return helper.ErrUnsupportedMediaType
	}
	return c.Next()
}
