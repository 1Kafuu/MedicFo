package repository

import (
	"context"
	"fmt"

	"medicfo/app/model"
)

type ReportRepository interface {
	ListAppointmentsByDate(ctx context.Context, date string) ([]model.AppointmentReport, error)
}

type reportPostgresRepository struct{ db DBTX }

func NewReportRepository(db DBTX) ReportRepository {
	return &reportPostgresRepository{db: db}
}

// ListAppointmentsByDate mengambil seluruh antrean pada satu tanggal.
//
// checked_in_at sengaja dibaca lewat to_jsonb: API.md meminta kolom ini,
// tetapi skema appointments di DATABASE.md belum mencantumkannya. Dengan
// begini query tidak gagal bila kolom belum tersedia; nilainya kosong sampai
// migrasi booking menyediakannya.
func (r *reportPostgresRepository) ListAppointmentsByDate(ctx context.Context, date string) ([]model.AppointmentReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			a.queue_number,
			u.username,
			d.name,
			a.status,
			COALESCE(to_jsonb(a)->>'checked_in_at', '')
		FROM appointments a
		JOIN users u ON u.id = a.patient_id
		JOIN doctor_schedules ds ON ds.id = a.schedule_id
		JOIN doctors d ON d.id = ds.doctor_id
		WHERE a.appointment_date = $1::date
		ORDER BY a.queue_number`, date)
	if err != nil {
		return nil, fmt.Errorf("mengambil laporan appointments: %w", err)
	}
	defer rows.Close()

	result := []model.AppointmentReport{}
	for rows.Next() {
		var item model.AppointmentReport
		if err := rows.Scan(&item.QueueNumber, &item.PatientName, &item.DoctorName,
			&item.Status, &item.CheckedInAt); err != nil {
			return nil, fmt.Errorf("membaca row laporan appointments: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca row laporan appointments: %w", err)
	}
	return result, nil
}
