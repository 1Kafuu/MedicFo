package service

import (
	"errors"
	"fmt"

	"medicfo/app/model"
	"medicfo/helper"
)

var errUnknownEvent = errors.New("unknown event type")

// BuildMessage menyusun pesan dari event. Fungsi murni: tidak menyentuh
// database, Redis, atau mailer sehingga mudah diuji.
func BuildMessage(eventType string, e model.AppointmentCreatedEvent, user model.User) (model.Message, error) {
	if eventType != model.EventAppointmentCreated {
		return model.Message{}, fmt.Errorf("%w: %q", errUnknownEvent, eventType)
	}
	if user.Email == "" {
		return model.Message{}, fmt.Errorf("%w: user %d tidak punya email", helper.ErrPermanentDelivery, user.ID)
	}

	return model.Message{
		To:      user.Email,
		Subject: fmt.Sprintf("Pengingat antrean %s", e.AppointmentDate),
		Body: fmt.Sprintf(
			"Halo %s, Anda memiliki antrean dengan %s pada tanggal %s. Nomor antrean Anda adalah %d.",
			e.PatientName, e.DoctorName, e.AppointmentDate, e.QueueNumber,
		),
	}, nil
}
