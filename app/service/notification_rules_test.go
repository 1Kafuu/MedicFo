package service

import (
	"errors"
	"strings"
	"testing"

	"medicfo/app/model"
	"medicfo/helper"
)

func TestBuildMessage(t *testing.T) {
	event := model.AppointmentCreatedEvent{
		UserID: 21, AppointmentID: 98, AppointmentDate: "2026-11-26",
		QueueNumber: 7, PatientName: "Budi", DoctorName: "dr. Sinta",
	}
	user := model.User{ID: 21, Email: "budi@mail.com"}

	msg, err := BuildMessage(model.EventAppointmentCreated, event, user)
	if err != nil {
		t.Fatalf("BuildMessage: %v", err)
	}
	if msg.To != user.Email {
		t.Errorf("To = %q, ingin %q", msg.To, user.Email)
	}
	if !strings.Contains(msg.Subject, event.AppointmentDate) {
		t.Errorf("Subject tidak memuat tanggal: %q", msg.Subject)
	}
	if !strings.Contains(msg.Body, "dr. Sinta") || !strings.Contains(msg.Body, "7") {
		t.Errorf("Body tidak memuat dokter/nomor antrean: %q", msg.Body)
	}
}

func TestBuildMessageUnknownEvent(t *testing.T) {
	if _, err := BuildMessage("unknown.event", model.AppointmentCreatedEvent{}, model.User{Email: "a@b.c"}); err == nil {
		t.Fatal("event tidak dikenal seharusnya menghasilkan error")
	}
}

func TestBuildMessageMissingEmailIsPermanent(t *testing.T) {
	_, err := BuildMessage(model.EventAppointmentCreated, model.AppointmentCreatedEvent{}, model.User{ID: 1})
	if !errors.Is(err, helper.ErrPermanentDelivery) {
		t.Fatalf("err = %v, ingin ErrPermanentDelivery", err)
	}
}
