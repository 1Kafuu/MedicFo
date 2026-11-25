package helper

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"medicfo/app/model"
)

func TestWriteAppointmentsCSV(t *testing.T) {
	rows := []model.AppointmentReport{{
		QueueNumber: 7,
		PatientName: "Budi, Santoso",
		DoctorName:  "dr. \"Sinta\"",
		Status:      "completed",
		CheckedInAt: "2026-11-26T08:14:00Z",
	}}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return WriteAppointmentsCSV(c, rows)
	})
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", resp.StatusCode)
	}
	if got := resp.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(got, FormatCSV) {
		t.Errorf("Content-Type = %q, ingin awalan %q", got, FormatCSV)
	}
	if got := resp.Header.Get(fiber.HeaderContentDisposition); !strings.Contains(got, "attachment") || !strings.Contains(got, "appointments.csv") {
		t.Errorf("Content-Disposition = %q", got)
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV tidak valid: %v", err)
	}
	want := [][]string{
		{"queue_number", "patient_name", "doctor_name", "status", "checked_in_at"},
		{"7", "Budi, Santoso", `dr. "Sinta"`, "completed", "2026-11-26T08:14:00Z"},
	}
	if len(records) != len(want) {
		t.Fatalf("jumlah baris = %d, ingin %d", len(records), len(want))
	}
	for i := range want {
		for j := range want[i] {
			if records[i][j] != want[i][j] {
				t.Errorf("baris %d kolom %d = %q, ingin %q", i, j, records[i][j], want[i][j])
			}
		}
	}
}
