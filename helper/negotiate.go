package helper

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

// Negotiate memilih format response dari header Accept.
//
// Content-Type menjelaskan format body yang dikirim; Accept menjelaskan
// format yang diminta client sebagai balasan. Accept kosong atau */*
// dianggap "terserah server" dan memilih format pertama yang ditawarkan.
func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	if len(offered) == 0 {
		return "", ErrInternal
	}
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))
	if accept == "" {
		return offered[0], nil
	}

	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", &AppError{
			Code:   "NOT_ACCEPTABLE",
			Status: http.StatusNotAcceptable,
			Detail: "format yang diminta tidak tersedia, pilih: " + strings.Join(offered, ", "),
		}
	}
	return chosen, nil
}
