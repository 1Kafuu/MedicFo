package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"medicfo/helper"
)

// RequireAuth memeriksa access token pada header Authorization.
func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.ErrUnauthorized
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			// "kedaluwarsa" vs "tidak valid" dibedakan: client perlu tahu
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.ErrTokenExpired
			}
			return helper.ErrUnauthorized
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}
	return token, nil
}
