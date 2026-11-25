package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
	"medicfo/app/repository"
)

const notificationListLimit = 50

// QueueInspector adalah kemampuan membaca keadaan queue. *asynq.Inspector
// memenuhi interface ini. Interface menjaga service tetap bisa diuji tanpa
// Redis dan tidak bergantung pada implementasi Asynq secara langsung.
type QueueInspector interface {
	GetQueueInfo(queue string) (*asynq.QueueInfo, error)
}

type NotificationService struct {
	notifications repository.NotificationRepository
	outbox        repository.OutboxRepository
	inspector     QueueInspector
}

func NewNotificationService(
	notifications repository.NotificationRepository,
	outbox repository.OutboxRepository,
	inspector QueueInspector,
) *NotificationService {
	return &NotificationService{
		notifications: notifications,
		outbox:        outbox,
		inspector:     inspector,
	}
}

// ListForUser mengembalikan notifikasi milik user. Ownership dipaksa dengan
// memakai userID dari token, bukan dari query string.
func (s *NotificationService) ListForUser(ctx context.Context, userID int64) ([]model.Notification, error) {
	return s.notifications.FindByUser(ctx, userID, notificationListLimit)
}

// Monitor mengumpulkan statistik outbox dan queue. Kegagalan membaca Redis
// tidak menggagalkan respons:admin justru membutuhkan angka outbox saat queue
// sedang bermasalah. Detail error Redis sengaja tidak diteruskan ke client.
func (s *NotificationService) Monitor(ctx context.Context) (model.QueueStats, error) {
	pending, published, oldest, err := s.outbox.Stats(ctx)
	if err != nil {
		return model.QueueStats{}, fmt.Errorf("statistik outbox: %w", err)
	}
	sent, err := s.notifications.Count(ctx)
	if err != nil {
		return model.QueueStats{}, fmt.Errorf("hitung notifikasi: %w", err)
	}

	stats := model.QueueStats{
		OutboxPending:     pending,
		OutboxPublished:   published,
		OldestPendingAge:  oldest.Round(time.Millisecond).String(),
		NotificationsSent: sent,
	}

	info, err := s.inspector.GetQueueInfo(model.QueueNotifications)
	if err != nil || info == nil {
		stats.QueueError = "queue tidak dapat dibaca; Redis mungkin tidak tersedia atau queue belum pernah dipakai"
		return stats, nil
	}
	stats.Queue = &model.QueueInfo{
		Name:      info.Queue,
		Pending:   info.Pending,
		Active:    info.Active,
		Scheduled: info.Scheduled,
		Retry:     info.Retry,
		Archived:  info.Archived,
		Completed: info.Completed,
		Processed: info.Processed,
		Failed:    info.Failed,
	}
	return stats, nil
}
