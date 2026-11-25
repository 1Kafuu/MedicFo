package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
)

type fakeInspector struct {
	info      *asynq.QueueInfo
	err       error
	queueName string
}

func (f *fakeInspector) GetQueueInfo(queue string) (*asynq.QueueInfo, error) {
	f.queueName = queue
	return f.info, f.err
}

func TestMonitorMapsOutboxAndQueueStats(t *testing.T) {
	outbox := &fakeOutboxRepo{pending: 3, published: 7, oldest: 1500 * time.Millisecond}
	notifications := &fakeNotificationRepo{total: 11}
	inspector := &fakeInspector{info: &asynq.QueueInfo{
		Queue: "notifications", Pending: 2, Active: 1, Scheduled: 4,
		Retry: 5, Archived: 6, Completed: 8, Processed: 9, Failed: 10,
	}}
	svc := NewNotificationService(notifications, outbox, inspector)

	stats, err := svc.Monitor(context.Background())
	if err != nil {
		t.Fatalf("Monitor: %v", err)
	}
	if stats.OutboxPending != 3 || stats.OutboxPublished != 7 {
		t.Errorf("outbox pending/published = %d/%d, ingin 3/7", stats.OutboxPending, stats.OutboxPublished)
	}
	if stats.OldestPendingAge != "1.5s" {
		t.Errorf("OldestPendingAge = %q, ingin 1.5s", stats.OldestPendingAge)
	}
	if stats.NotificationsSent != 11 {
		t.Errorf("NotificationsSent = %d, ingin 11", stats.NotificationsSent)
	}
	if stats.QueueError != "" {
		t.Errorf("QueueError = %q, ingin kosong", stats.QueueError)
	}
	if stats.Queue == nil {
		t.Fatal("Queue tidak boleh nil saat Redis terbaca")
	}
	want := model.QueueInfo{
		Name: "notifications", Pending: 2, Active: 1, Scheduled: 4,
		Retry: 5, Archived: 6, Completed: 8, Processed: 9, Failed: 10,
	}
	if *stats.Queue != want {
		t.Errorf("Queue = %+v, ingin %+v", *stats.Queue, want)
	}
	if inspector.queueName != model.QueueNotifications {
		t.Errorf("queue yang dibaca = %q, ingin %q", inspector.queueName, model.QueueNotifications)
	}
}

func TestMonitorRedisFailureKeepsOutboxStats(t *testing.T) {
	outbox := &fakeOutboxRepo{pending: 3, oldest: 2 * time.Second}
	notifications := &fakeNotificationRepo{total: 1}
	inspector := &fakeInspector{err: errors.New("redis dial tcp: internal host")}
	svc := NewNotificationService(notifications, outbox, inspector)

	stats, err := svc.Monitor(context.Background())
	if err != nil {
		t.Fatalf("Redis mati tidak boleh menggagalkan Monitor: %v", err)
	}
	if stats.OutboxPending != 3 || stats.NotificationsSent != 1 {
		t.Errorf("stats = %+v, ingin data outbox tetap ada", stats)
	}
	if stats.Queue != nil {
		t.Error("Queue harus nil saat Redis gagal")
	}
	if stats.QueueError == "" {
		t.Error("QueueError harus diisi saat Redis gagal")
	}
	if strings.Contains(stats.QueueError, "redis dial tcp") {
		t.Errorf("detail internal Redis bocor: %q", stats.QueueError)
	}
}

func TestMonitorNilQueueInfoTreatedAsUnavailable(t *testing.T) {
	svc := NewNotificationService(&fakeNotificationRepo{}, &fakeOutboxRepo{}, &fakeInspector{})
	stats, err := svc.Monitor(context.Background())
	if err != nil {
		t.Fatalf("Monitor: %v", err)
	}
	if stats.Queue != nil || stats.QueueError == "" {
		t.Errorf("stats = %+v, ingin QueueError tanpa Queue", stats)
	}
}

func TestMonitorOutboxErrorPropagated(t *testing.T) {
	svc := NewNotificationService(
		&fakeNotificationRepo{},
		&fakeOutboxRepo{statsErr: errors.New("db down")},
		&fakeInspector{},
	)
	if _, err := svc.Monitor(context.Background()); err == nil {
		t.Fatal("error statistik outbox harus diteruskan")
	}
}

func TestMonitorCountErrorPropagated(t *testing.T) {
	svc := NewNotificationService(
		&fakeNotificationRepo{countErr: errors.New("db down")},
		&fakeOutboxRepo{},
		&fakeInspector{},
	)
	if _, err := svc.Monitor(context.Background()); err == nil {
		t.Fatal("error hitung notifikasi harus diteruskan")
	}
}
