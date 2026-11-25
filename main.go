package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"medicfo/app/repository"
	"medicfo/app/service"
	"medicfo/config"
	"medicfo/database"
	"medicfo/helper"
	"medicfo/middleware"
	"medicfo/route"
)

const jwtSecretMinLength = 32

func main() {
	config.LoadENV()
	logger := config.NewLogger()

	// Satu binary, tiga peran: api, worker, atau keduanya.
	role := flag.String("role", config.GetENV("APP_ROLE", "api"), "api | worker | all")
	flag.Parse()
	runAPI := *role == "api" || *role == "all"
	runWorker := *role == "worker" || *role == "all"
	if !runAPI && !runWorker {
		logger.Error("role tidak dikenal", slog.String("role", *role))
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if runWorker {
		stopWorker, err := startWorker(ctx, logger, pool)
		if err != nil {
			logger.Error("gagal menyalakan worker", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer stopWorker()
	}

	var app *fiber.App
	if runAPI {
		jwtSecret := config.GetENV("JWT_SECRET", "")
		if len(jwtSecret) < jwtSecretMinLength {
			logger.Error("JWT_SECRET tidak ada atau terlalu pendek",
				slog.Int("minLength", jwtSecretMinLength))
			os.Exit(1)
		}

		permissionRepo := repository.NewPermissionRepository(pool)
		rawPermissions, err := permissionRepo.ListRolePermissions(context.Background())
		if err != nil {
			logger.Error("gagal memuat permissions", slog.String("error", err.Error()))
			os.Exit(1)
		}
		permissionSet := helper.NewPermissionSet(rawPermissions)

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
			repository.NewTxManager(pool),
			refreshTTL,
		)
		userRepo := repository.NewUserRepository(pool)
		doctorSvc := service.NewDoctorService(repository.NewDoctorRepository(pool), userRepo)
		notificationSvc := service.NewNotificationService(repository.NewNotificationRepository(pool))

		app = fiber.New(fiber.Config{ErrorHandler: route.ErrorHandler})
		allowedOrigins := strings.TrimSpace(config.GetENV("ALLOWED_ORIGINS", "http://localhost:5173"))
		if allowedOrigins == "" || strings.Contains(allowedOrigins, "*") {
			logger.Error("ALLOWED_ORIGINS kosong atau wildcard")
			os.Exit(1)
		}
		middleware.RegisterHTTP(app, logger, allowedOrigins)
		route.Register(app, route.Dependencies{
			Pool:                pool,
			JWT:                 jwtManager,
			AuthService:         authSvc,
			DoctorService:       doctorSvc,
			NotificationService: notificationSvc,
			Permissions:         permissionSet,
		})

		port := config.GetENV("APP_PORT", "3000")
		logger.Info("server berjalan", slog.String("port", port))
		go func() {
			if err := app.Listen(":" + port); err != nil {
				logger.Error("server berhenti", slog.String("error", err.Error()))
				os.Exit(1)
			}
		}()
	}

	<-ctx.Done()
	logger.Info("sinyal berhenti diterima, menutup server...")
	if app != nil {
		_ = app.Shutdown()
	}
}
