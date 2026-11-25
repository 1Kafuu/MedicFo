package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
	"medicfo/app/repository"
	"medicfo/helper"
)

// NotificationWorker mengerjakan task notification:send.
type NotificationWorker struct {
	users         repository.UserRepository
	notifications repository.NotificationRepository
	tx            TxOnceRunner
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	notifications repository.NotificationRepository,
	tx TxOnceRunner,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		users: users, notifications: notifications,
		tx: tx, mailer: mailer, logger: logger,
	}
}

// HandleSend idempoten: UNIQUE (outbox_id) pada tabel notifications
// memastikan task yang dikirim ulang tidak mengirim pesan kedua kalinya.
func (w *NotificationWorker) HandleSend(ctx context.Context, task *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(task.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak", asynq.SkipRetry)
	}
	if job.OutboxID <= 0 {
		return fmt.Errorf("%w: outbox_id tidak valid", asynq.SkipRetry)
	}

	var event model.AppointmentCreatedEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak", asynq.SkipRetry, job.OutboxID)
	}
	if job.EventType != model.EventAppointmentCreated {
		return fmt.Errorf("%w: event %q tidak dikenal", asynq.SkipRetry, job.EventType)
	}

	user, err := w.users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err // gangguan database: layak dicoba lagi
	}

	msg, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID,
			UserID:   user.ID,
			Channel:  "email",
			Subject:  msg.Subject,
			Body:     msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		// Pengiriman dan claim berada dalam satu transaction: gagal kirim
		// membatalkan claim, sukses kirim mengunci claim.
		return w.mailer.Send(ctx, msg)
	})
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notifikasi_diproses",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("sudah_terkirim", alreadySent),
	)
	return nil
}
