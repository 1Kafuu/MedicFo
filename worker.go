package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"medicfo/app/model"
	"medicfo/app/repository"
	"medicfo/app/service"
	"medicfo/config"
	"medicfo/helper"
)

// startWorker menyalakan relay outbox dan queue worker. Mengembalikan fungsi
// untuk menghentikan keduanya saat context dibatalkan.
func startWorker(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) (func(), error) {
	mailer, err := helper.NewLogMailer(
		"logs/mail.log",
		time.Duration(config.GetENVInt("MAILER_DELAY_MS", 0))*time.Millisecond,
		config.GetENVInt("MAILER_FAIL_EVERY", 0),
	)
	if err != nil {
		return nil, err
	}

	interval := time.Duration(config.GetENVInt("RELAY_INTERVAL_MS", 1000)) * time.Millisecond
	if interval <= 0 {
		interval = time.Second
	}
	batch := config.GetENVInt("RELAY_BATCH", 100)
	if batch <= 0 {
		batch = 100
	}

	txManager := repository.NewTxManager(pool)
	client := asynq.NewClient(config.RedisOpt())
	relay := service.NewOutboxRelay(
		repository.NewOutboxRepository(pool),
		txManager,
		client,
		logger,
		interval,
		batch,
	)
	worker := service.NewNotificationWorker(
		repository.NewUserRepository(pool),
		repository.NewNotificationRepository(pool),
		txManager,
		mailer,
		logger,
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(model.TaskSendNotification, worker.HandleSend)

	server := config.NewWorkerServer(logger)
	if err := server.Start(mux); err != nil {
		_ = client.Close()
		return nil, err
	}

	relayDone := make(chan struct{})
	go func() {
		relay.Run(ctx)
		close(relayDone)
	}()

	logger.Info("worker berjalan", slog.String("queue", model.QueueNotifications))
	return func() {
		<-relayDone
		server.Shutdown()
		_ = client.Close()
	}, nil
}
