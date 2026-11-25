package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hibiken/asynq"

	"medicfo/app/model"
)

// RedisOpt dipakai client (relay), server (worker), dan inspector.
func RedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:         GetENV("REDIS_ADDR", "localhost:6379"),
		Password:     GetENV("REDIS_PASSWORD", ""),
		DB:           GetENVInt("REDIS_DB", 0),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

// NewInspector hanya MEMBACA keadaan queue. Ia tidak menghubungi Redis saat
// dibuat, sehingga API tetap bisa menyala walaupun Redis sedang mati.
func NewInspector() *asynq.Inspector {
	return asynq.NewInspector(RedisOpt())
}

// NewWorkerServer menyiapkan server Asynq yang hanya memproses queue notifikasi.
func NewWorkerServer(logger *slog.Logger) *asynq.Server {
	return asynq.NewServer(RedisOpt(), asynq.Config{
		Concurrency: GetENVInt("WORKER_CONCURRENCY", 5),
		Queues:      map[string]int{model.QueueNotifications: 1},
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			return time.Duration(n+1) * 2 * time.Second
		},
		ShutdownTimeout: 10 * time.Second,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retry, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			logger.Warn("task_gagal",
				slog.String("type", task.Type()),
				slog.String("attempt", fmt.Sprintf("%d/%d", retry+1, maxRetry+1)),
				slog.String("error", err.Error()),
			)
		}),
		Logger:   newAsynqLogger(logger),
		LogLevel: asynq.WarnLevel,
	})
}

type asynqLogger struct{ logger *slog.Logger }

func newAsynqLogger(logger *slog.Logger) asynqLogger {
	return asynqLogger{logger: logger.With(slog.String("component", "asynq"))}
}

func (l asynqLogger) Debug(args ...any) { l.logger.Debug(fmt.Sprint(args...)) }
func (l asynqLogger) Info(args ...any)  { l.logger.Info(fmt.Sprint(args...)) }
func (l asynqLogger) Warn(args ...any)  { l.logger.Warn(fmt.Sprint(args...)) }
func (l asynqLogger) Error(args ...any) { l.logger.Error(fmt.Sprint(args...)) }
func (l asynqLogger) Fatal(args ...any) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
