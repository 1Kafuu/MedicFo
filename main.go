package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"medicfo/app/repository"
	"medicfo/app/service"
	"medicfo/config"
	"medicfo/database"
	"medicfo/helper"
	"medicfo/route"
)

const jwtSecretMinLength = 32

func main() {
	config.LoadENV()
	logger := config.NewLogger()

	// 1. Secret JWT wajib ada dan cukup panjang — fail fast.
	jwtSecret := config.GetENV("JWT_SECRET", "")
	if len(jwtSecret) < jwtSecretMinLength {
		logger.Error("JWT_SECRET tidak ada atau terlalu pendek",
			slog.Int("minLength", jwtSecretMinLength))
		os.Exit(1)
	}

	// 2. Koneksi database — fail fast juga.
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Susun dependensi: helper → repository → service → route.
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetENV("JWT_ISSUER", "medicfo"),
		time.Duration(config.GetENVInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	refreshTTL := time.Duration(config.GetENVInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour

	authSvc := service.NewAuthService(
		repository.NewUserRepository(pool),
		repository.NewTokenRepository(pool),
		jwtManager,
		refreshTTL,
	)

	// 4. Fiber + error handler terpusat.
	app := fiber.New(fiber.Config{ErrorHandler: route.ErrorHandler})
	route.Register(app, route.Dependencies{JWT: jwtManager, AuthService: authSvc})

	// 5. Jalankan server + graceful shutdown.
	port := config.GetENV("APP_PORT", "3000")
	logger.Info("server berjalan", slog.String("port", port))

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("sinyal berhenti diterima, menutup server...")
	_ = app.Shutdown()
}
