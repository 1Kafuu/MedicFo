package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"medicfo/app/model"
	"medicfo/helper"
)

type fakeReportRepo struct {
	rows    []model.AppointmentReport
	err     error
	gotDate string
	called  bool
}

func (f *fakeReportRepo) ListAppointmentsByDate(_ context.Context, date string) ([]model.AppointmentReport, error) {
	f.called = true
	f.gotDate = date
	return f.rows, f.err
}

func TestAppointmentsByDateValid(t *testing.T) {
	repo := &fakeReportRepo{rows: []model.AppointmentReport{{QueueNumber: 7}}}
	svc := NewReportService(repo)

	items, err := svc.AppointmentsByDate(context.Background(), "2026-11-26")
	if err != nil {
		t.Fatalf("AppointmentsByDate: %v", err)
	}
	if len(items) != 1 || items[0].QueueNumber != 7 {
		t.Errorf("items = %+v, ingin satu baris", items)
	}
	if !repo.called || repo.gotDate != "2026-11-26" {
		t.Errorf("repo dipanggil=%v dengan date=%q", repo.called, repo.gotDate)
	}
}

func TestAppointmentsByDateRejectsInvalidDate(t *testing.T) {
	for _, date := range []string{"", "26-11-2026", "2026-13-01"} {
		repo := &fakeReportRepo{}
		svc := NewReportService(repo)

		_, err := svc.AppointmentsByDate(context.Background(), date)
		if err == nil {
			t.Fatalf("date %q seharusnya ditolak", date)
		}
		ae := helper.AsAppError(err)
		if ae.Status != http.StatusBadRequest {
			t.Errorf("date %q status = %d, ingin 400", date, ae.Status)
		}
		if repo.called {
			t.Errorf("date %q tidak boleh menjalankan query", date)
		}
	}
}

func TestAppointmentsByDatePropagatesRepoError(t *testing.T) {
	svc := NewReportService(&fakeReportRepo{err: errors.New("db down")})
	if _, err := svc.AppointmentsByDate(context.Background(), "2026-11-26"); err == nil {
		t.Fatal("error repository harus diteruskan")
	}
}
