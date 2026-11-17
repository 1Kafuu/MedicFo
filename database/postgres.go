package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"medicfo/config"
)

// NewPool membuat connection pool ke PostgreSQL.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.GetENV("DB_USER", "postgres"),
		os.Getenv("DB_PASSWORD"),
		config.GetENV("DB_HOST", "localhost"),
		config.GetENVInt("DB_PORT", 5432),
		config.GetENV("DB_NAME", "praktikum_backend"),
		config.GetENV("DB_SSLMODE", "disable"),
	)

	pgCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi database tidak valid: %w", err)
	}

	pgCfg.MaxConns = int32(config.GetENVInt("DB_MAX_CONNS", 10))
	pgCfg.MinConns = 2
	pgCfg.MaxConnLifetime = time.Hour
	pgCfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %w", err)
	}

	// Ping memastikan kredensial benar dan server memang dapat dihubungi.
	// Tanpa ini, kesalahan baru ketahuan saat permintaan pertama masuk.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	return pool, nil
}
