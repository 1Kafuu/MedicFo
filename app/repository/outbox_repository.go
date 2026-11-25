package repository

import (
	"context"
	"fmt"

	"medicfo/app/model"
)

type OutboxRepository interface {
	WithTx(tx DBTX) OutboxRepository
	// FetchUnpublished mengunci baris yang diambil sampai transaction selesai.
	// Pemanggilnya wajib berada di dalam transaction.
	FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkPublished(ctx context.Context, ids []int64) error
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
