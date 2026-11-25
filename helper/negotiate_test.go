package helper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func negotiateFormat(t *testing.T, accept string) (string, error) {
	t.Helper()
	app := fiber.New()
	var format string
	var negotiateErr error
	app.Get("/", func(c *fiber.Ctx) error {
		format, negotiateErr = Negotiate(c, FormatJSON, FormatCSV)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if accept != "" {
		req.Header.Set(fiber.HeaderAccept, accept)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", resp.StatusCode)
	}
	return format, negotiateErr
}

func TestNegotiateDefaultsToJSON(t *testing.T) {
	for _, accept := range []string{"", "*/*"} {
		format, err := negotiateFormat(t, accept)
		if err != nil {
			t.Fatalf("Accept %q: %v", accept, err)
		}
		if format != FormatJSON {
			t.Errorf("Accept %q menghasilkan %q, ingin %q", accept, format, FormatJSON)
		}
	}
}

func TestNegotiateCSV(t *testing.T) {
	format, err := negotiateFormat(t, "text/csv")
	if err != nil {
		t.Fatalf("Negotiate: %v", err)
	}
	if format != FormatCSV {
		t.Errorf("format = %q, ingin %q", format, FormatCSV)
	}
}

func TestNegotiateUnsupportedFormatIsNotAcceptable(t *testing.T) {
	_, err := negotiateFormat(t, "application/xml")
	if err == nil {
		t.Fatal("format tidak didukung harus menghasilkan error")
	}
	ae := AsAppError(err)
	if ae.Code != "NOT_ACCEPTABLE" || ae.Status != http.StatusNotAcceptable {
		t.Fatalf("error = %+v, ingin NOT_ACCEPTABLE/406", ae)
	}
}
