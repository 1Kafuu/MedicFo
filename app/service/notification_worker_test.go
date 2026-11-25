package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
	"medicfo/app/repository"
	"medicfo/helper"
)

type fakeUserRepo struct {
	user     model.User
	notFound bool
}

func (f *fakeUserRepo) WithTx(repository.DBTX) repository.UserRepository { return f }
func (f *fakeUserRepo) Create(context.Context, model.User) (model.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) FindByEmail(context.Context, string) (model.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) FindByID(context.Context, int64) (model.User, error) {
	if f.notFound {
		return model.User{}, repository.ErrNotFound
	}
	return f.user, nil
}

type fakeNotificationRepo struct {
	claimed  bool
	calls    int
	total    int
	countErr error
}

func (f *fakeNotificationRepo) WithTx(repository.DBTX) repository.NotificationRepository { return f }
func (f *fakeNotificationRepo) Claim(_ context.Context, _ model.Notification) (bool, error) {
	f.calls++
	return f.claimed, nil
}
func (f *fakeNotificationRepo) FindByUser(context.Context, int64, int) ([]model.Notification, error) {
	return nil, nil
}

func (f *fakeNotificationRepo) Count(context.Context) (int, error) {
	return f.total, f.countErr
}

type fakeMailer struct {
	calls int
	err   error
}

func (f *fakeMailer) Send(context.Context, model.Message) error {
	f.calls++
	return f.err
}

func newNotificationTask(t *testing.T, outboxID int64, eventType string, event model.AppointmentCreatedEvent) *asynq.Task {
	t.Helper()
	rawEvent, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(model.NotificationJob{
		OutboxID: outboxID, EventType: eventType, Payload: rawEvent,
	})
	if err != nil {
		t.Fatal(err)
	}
	return asynq.NewTask(model.TaskSendNotification, payload)
}

func newTestWorker(users repository.UserRepository, notifications repository.NotificationRepository, mailer helper.Mailer) *NotificationWorker {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewNotificationWorker(users, notifications, fakeTxOnce{}, mailer, logger)
}

func TestHandleSendSendsOnce(t *testing.T) {
	event := model.AppointmentCreatedEvent{UserID: 21, AppointmentDate: "2026-11-26"}
	users := &fakeUserRepo{user: model.User{ID: 21, Email: "budi@mail.com"}}
	notifications := &fakeNotificationRepo{claimed: true}
	mailer := &fakeMailer{}
	worker := newTestWorker(users, notifications, mailer)

	if err := worker.HandleSend(context.Background(), newNotificationTask(t, 9, model.EventAppointmentCreated, event)); err != nil {
		t.Fatalf("HandleSend: %v", err)
	}
	if mailer.calls != 1 || notifications.calls != 1 {
		t.Errorf("mailer=%d claim=%d, ingin 1 dan 1", mailer.calls, notifications.calls)
	}
}

func TestHandleSendSkipsAlreadyClaimed(t *testing.T) {
	event := model.AppointmentCreatedEvent{UserID: 21, AppointmentDate: "2026-11-26"}
	users := &fakeUserRepo{user: model.User{ID: 21, Email: "budi@mail.com"}}
	notifications := &fakeNotificationRepo{claimed: false}
	mailer := &fakeMailer{}
	worker := newTestWorker(users, notifications, mailer)

	if err := worker.HandleSend(context.Background(), newNotificationTask(t, 9, model.EventAppointmentCreated, event)); err != nil {
		t.Fatalf("HandleSend: %v", err)
	}
	if mailer.calls != 0 {
		t.Errorf("mailer dipanggil %d kali pada task duplikat", mailer.calls)
	}
}

func TestHandleSendPermanentFailureSkipsRetry(t *testing.T) {
	event := model.AppointmentCreatedEvent{UserID: 21, AppointmentDate: "2026-11-26"}
	users := &fakeUserRepo{user: model.User{ID: 21, Email: "budi@mail.com"}}
	notifications := &fakeNotificationRepo{claimed: true}
	mailer := &fakeMailer{err: helper.ErrPermanentDelivery}
	worker := newTestWorker(users, notifications, mailer)

	err := worker.HandleSend(context.Background(), newNotificationTask(t, 9, model.EventAppointmentCreated, event))
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("err = %v, ingin SkipRetry", err)
	}
}

func TestHandleSendUnknownUserSkipsRetry(t *testing.T) {
	event := model.AppointmentCreatedEvent{UserID: 21, AppointmentDate: "2026-11-26"}
	users := &fakeUserRepo{notFound: true}
	notifications := &fakeNotificationRepo{}
	mailer := &fakeMailer{}
	worker := newTestWorker(users, notifications, mailer)

	err := worker.HandleSend(context.Background(), newNotificationTask(t, 9, model.EventAppointmentCreated, event))
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("err = %v, ingin SkipRetry", err)
	}
	if mailer.calls != 0 || notifications.calls != 0 {
		t.Error("user tidak ditemukan tidak boleh mengirim atau mengklaim")
	}
}
