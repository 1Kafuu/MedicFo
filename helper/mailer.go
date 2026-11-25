package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"medicfo/app/model"
)

// ErrPermanentDelivery menandakan kegagalan yang tidak perlu di-retry.
var ErrPermanentDelivery = errors.New("permanent delivery failure")

// Mailer adalah pengirim pesan. Implementasi tiruan menulis ke file log.
type Mailer interface {
	Send(ctx context.Context, msg model.Message) error
}

type logMailer struct {
	path      string
	delay     time.Duration
	failEvery int
	attempts  atomic.Int64
	mu        sync.Mutex
}

// NewLogMailer menyiapkan mailer tiruan. failEvery = 0 berarti tidak pernah
// gagal; N berarti setiap attempt ke-N gagal sementara.
func NewLogMailer(path string, delay time.Duration, failEvery int) (Mailer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("menyiapkan folder mailer: %w", err)
	}
	return &logMailer{path: path, delay: delay, failEvery: failEvery}, nil
}

func (m *logMailer) Send(ctx context.Context, msg model.Message) error {
	if m.delay > 0 {
		timer := time.NewTimer(m.delay)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	attempt := m.attempts.Add(1)
	if m.failEvery > 0 && attempt%int64(m.failEvery) == 0 {
		return fmt.Errorf("simulasi kegagalan pengiriman pada attempt %d", attempt)
	}

	line, err := json.Marshal(struct {
		Time    string `json:"time"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}{
		Time: time.Now().UTC().Format(time.RFC3339),
		To:   msg.To, Subject: msg.Subject, Body: msg.Body,
	})
	if err != nil {
		return fmt.Errorf("menyusun log mailer: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	f, err := os.OpenFile(m.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("membuka log mailer: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, string(line)); err != nil {
		return fmt.Errorf("menulis log mailer: %w", err)
	}
	return nil
}
