package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
	"medicfo/app/repository"
)

// TxOnceRunner menjalankan satu transaction tanpa retry. Dipakai relay dan
// worker karena keduanya melakukan aksi di luar database di dalam callback;
// mengulang callback akan mengulang aksi tersebut.
type TxOnceRunner interface {
	WithTxOnce(ctx context.Context, fn func(repository.DBTX) error) error
}

// TaskEnqueuer adalah kemampuan mengirim task ke queue. *asynq.Client
// memenuhi interface ini.
type TaskEnqueuer interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// OutboxRelay memindahkan event dari PostgreSQL ke Redis. API tidak pernah
// berbicara langsung dengan Redis, sehingga event tetap tersimpan bila Redis
// sedang mati.
type OutboxRelay struct {
	outbox   repository.OutboxRepository
	tx       TxOnceRunner
	queue    TaskEnqueuer
	logger   *slog.Logger
	interval time.Duration
	batch    int
}

func NewOutboxRelay(
	outbox repository.OutboxRepository,
	tx TxOnceRunner,
	queue TaskEnqueuer,
	logger *slog.Logger,
	interval time.Duration,
	batch int,
) *OutboxRelay {
	return &OutboxRelay{
		outbox: outbox, tx: tx, queue: queue, logger: logger,
		interval: interval, batch: batch,
	}
}

// Run berjalan sampai ctx dibatalkan.
func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		n, err := r.RelayOnce(ctx)
		if err != nil {
			r.logger.Warn("relay_gagal", slog.Int("published", n), slog.String("error", err.Error()))
		} else if n > 0 {
			r.logger.Info("relay_selesai", slog.Int("published", n))
		}
		if err == nil && n == r.batch && ctx.Err() == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// RelayOnce mem-publish paling banyak satu batch dan mengembalikan jumlah
// event yang berhasil dikirim ke Redis.
func (r *OutboxRelay) RelayOnce(ctx context.Context) (int, error) {
	var published int
	var enqueueErr error

	err := r.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		repo := r.outbox.WithTx(tx)
		events, err := repo.FetchUnpublished(ctx, r.batch)
		if err != nil {
			return err
		}

		ids := make([]int64, 0, len(events))
		for _, event := range events {
			// Hanya event yang BERHASIL di-enqueue yang boleh ditandai.
			if err := r.publish(ctx, event); err != nil {
				enqueueErr = err
				break
			}
			ids = append(ids, event.ID)
		}
		if err := repo.MarkPublished(ctx, ids); err != nil {
			return err
		}
		published = len(ids)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return published, enqueueErr
}

func (r *OutboxRelay) publish(ctx context.Context, event model.OutboxEvent) error {
	job := model.NotificationJob{
		OutboxID: event.ID, EventType: event.EventType, Payload: event.Payload,
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("menyusun task outbox %d: %w", event.ID, err)
	}

	opts := []asynq.Option{
		// TaskID dari id outbox menutup celah relay mati setelah enqueue
		// tetapi sebelum commit. Task duplikat ditolak Redis.
		asynq.TaskID(fmt.Sprintf("outbox-%d", event.ID)),
		asynq.Queue(model.QueueNotifications),
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
		asynq.Retention(24 * time.Hour),
	}
	if delay, err := reminderDelay(event, time.Now()); err == nil && delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	task := asynq.NewTask(model.TaskSendNotification, payload, opts...)
	_, err = r.queue.EnqueueContext(ctx, task)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil // sudah pernah di-enqueue pada putaran sebelumnya
	}
	if err != nil {
		return fmt.Errorf("enqueue outbox %d: %w", event.ID, err)
	}
	return nil
}

// reminderDelay menghitung jeda sampai H-1 dari tanggal appointment.
// Tanggal dengan format selain YYYY-MM-DD diabaikan (task dikirim langsung).
func reminderDelay(event model.OutboxEvent, now time.Time) (time.Duration, error) {
	if event.EventType != model.EventAppointmentCreated {
		return 0, errors.New("bukan event appointment")
	}
	var payload model.AppointmentCreatedEvent
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return 0, err
	}
	date, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(payload.AppointmentDate), time.UTC)
	if err != nil {
		return 0, err
	}
	return date.Add(-24 * time.Hour).Sub(now), nil
}
