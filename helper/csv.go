package helper

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"medicfo/app/model"
)

// WriteAppointmentsCSV menuliskan laporan antrean sebagai berkas CSV.
// encoding/csv dipakai agar nilai dengan koma, tanda kutip, atau baris baru
// tetap menjadi satu kolom yang valid (RFC 4180).
func WriteAppointmentsCSV(c *fiber.Ctx, rows []model.AppointmentReport) error {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	if err := writer.Write([]string{
		"queue_number", "patient_name", "doctor_name", "status", "checked_in_at",
	}); err != nil {
		return ErrInternal
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.Itoa(row.QueueNumber), row.PatientName, row.DoctorName,
			row.Status, row.CheckedInAt,
		}); err != nil {
			return ErrInternal
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return ErrInternal
	}

	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="appointments.csv"`)
	return c.SendString(buffer.String())
}
