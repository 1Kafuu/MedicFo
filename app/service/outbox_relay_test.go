package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
	"medicfo/app/repository"
)

type fakeOutboxRepo struct {
	events       []model.OutboxEvent
	fetchErr     error
	marked       []int64
	withTxCalled bool
}

func (f *fakeOutboxRepo) WithTx(repository.DBTX) repository.OutboxRepository {
	f.withTxCalled = true
	return f
}

func (f *fakeOutboxRepo) FetchUnpublished(context.Context, int) ([]model.OutboxEvent, error) {
	return f.events, f.fetchErr
}

func (f *fakeOutboxRepo) MarkPublished(_ context.Context, ids []int64) error {
	f.marked = append([]int64(nil), ids...)
	return nil
}

type fakeTxOnce struct{ err error }

func (f fakeTxOnce) WithTxOnce(_ context.Context, fn func(repository.DBTX) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(nil)
}

type fakeEnqueuer struct {
	enqueue func(*asynq.Task) error
	calls   int
}

func (f *fakeEnqueuer) EnqueueContext(_ context.Context, task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	f.calls++
	if f.enqueue != nil {
		return nil, f.enqueue(task)
	}
	return &asynq.TaskInfo{}, nil
}

func newTestRelay(outbox repository.OutboxRepository, tx TxOnceRunner, queue TaskEnqueuer) *OutboxRelay {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewOutboxRelay(outbox, tx, queue, logger, time.Second, 10)
}

func TestRelayOnceMarksOnlyPublishedEvents(t *testing.T) {
	outbox := &fakeOutboxRepo{events: []model.OutboxEvent{{ID: 1}, {ID: 2}, {ID: 3}}}
	queueCalls := 0
	queue := &fakeEnqueuer{enqueue: func(*asynq.Task) error {
		if queueCalls == 1 {
			return errors.New("redis mati")
		}
		queueCalls++
		return nil
	}}
	// Event kedua gagal: hanya event pertama yang boleh ditandai published.

	relay := newTestRelay(outbox, fakeTxOnce{}, queue)
	published, err := relay.RelayOnce(context.Background())
	if err == nil {
		t.Fatal("error enqueue seharusnya dikembalikan")
	}
	if published != 1 {
		t.Errorf("published = %d, ingin 1", published)
	}
	if !reflect.DeepEqual(outbox.marked, []int64{1}) {
		t.Errorf("marked = %v, ingin [1]", outbox.marked)
	}
	if !outbox.withTxCalled {
		t.Error("relay harus memakai transaction")
	}
}

func TestRelayOnceTaskConflictCountsAsPublished(t *testing.T) {
	outbox := &fakeOutboxRepo{events: []model.OutboxEvent{{ID: 5}, {ID: 6}}}
	queue := &fakeEnqueuer{enqueue: func(*asynq.Task) error {
		return asynq.ErrTaskIDConflict
	}}

	relay := newTestRelay(outbox, fakeTxOnce{}, queue)
	published, err := relay.RelayOnce(context.Background())
	if err != nil {
		t.Fatalf("TaskID conflict harus dianggap sukses: %v", err)
	}
	if published != 2 {
		t.Errorf("published = %d, ingin 2", published)
	}
	if !reflect.DeepEqual(outbox.marked, []int64{5, 6}) {
		t.Errorf("marked = %v, ingin [5 6]", outbox.marked)
	}
}

func TestRelayOnceWithoutEvents(t *testing.T) {
	outbox := &fakeOutboxRepo{}
	queue := &fakeEnqueuer{}
	relay := newTestRelay(outbox, fakeTxOnce{}, queue)

	published, err := relay.RelayOnce(context.Background())
	if err != nil || published != 0 {
		t.Fatalf("published=%d err=%v, ingin 0 dan nil", published, err)
	}
	if queue.calls != 0 || len(outbox.marked) != 0 {
		t.Error("tanpa event tidak boleh ada enqueue atau mark")
	}
}
