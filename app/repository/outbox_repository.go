package repository

import (
	"context"
	"fmt"
	"time"

	"medicfo/app/model"
)

type OutboxRepository interface {
	WithTx(tx DBTX) OutboxRepository
	// FetchUnpublished mengunci baris yang diambil sampai transaction selesai.
	// Pemanggilnya wajib berada di dalam transaction.
	FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkPublished(ctx context.Context, ids []int64) error
	// Stats menghitung event pending/published dan usia event pending tertua.
	Stats(ctx context.Context) (pending, published int, oldest time.Duration, err error)
}

type outboxPostgresRepository struct{ db DBTX }

func NewOutboxRepository(db DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: db}
}

func (r *outboxPostgresRepository) WithTx(tx DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: tx}
}

// FetchUnpublished memakai FOR UPDATE SKIP LOCKED agar dua relay yang jalan
// bersamaan membagi baris, bukan saling menunggu atau mengambil baris yang sama.
func (r *outboxPostgresRepository) FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_type, payload, created_at
		FROM outbox
		WHERE published_at IS NULL
		ORDER BY id
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, fmt.Errorf("mengambil outbox: %w", err)
	}
	defer rows.Close()

	events := []model.OutboxEvent{}
	for rows.Next() {
		var e model.OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca row outbox: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca row outbox: %w", err)
	}
	return events, nil
}

func (r *outboxPostgresRepository) MarkPublished(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Exec(ctx,
		`UPDATE outbox SET published_at = NOW() WHERE id = ANY($1)`, ids)
	if err != nil {
		return fmt.Errorf("menandai outbox: %w", err)
	}
	return nil
}

// Stats memakai satu query agregat. EXTRACT(EPOCH ...) dikonversi ke
// time.Duration di Go agar tidak bergantung pada pemetaan tipe interval.
func (r *outboxPostgresRepository) Stats(ctx context.Context) (int, int, time.Duration, error) {
	var pending, published int
	var seconds float64
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE published_at IS NULL),
			COUNT(*) FILTER (WHERE published_at IS NOT NULL),
			COALESCE(EXTRACT(EPOCH FROM (
				NOW() - MIN(CASE WHEN published_at IS NULL THEN created_at END)
			))::double precision, 0)
		FROM outbox`,
	).Scan(&pending, &published, &seconds)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("mengambil statistik outbox: %w", err)
	}
	return pending, published, time.Duration(seconds * float64(time.Second)), nil
}
