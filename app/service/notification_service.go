package service

import (
	"context"

	"medicfo/app/model"
	"medicfo/app/repository"
)

const notificationListLimit = 50

type NotificationService struct {
	notifications repository.NotificationRepository
}

func NewNotificationService(notifications repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

// ListForUser mengembalikan notifikasi milik user. Ownership dipaksa dengan
// memakai userID dari token, bukan dari query string.
func (s *NotificationService) ListForUser(ctx context.Context, userID int64) ([]model.Notification, error) {
	return s.notifications.FindByUser(ctx, userID, notificationListLimit)
}
