package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/config"
	appstream "Lullify_Backend/internal/application/stream"
	appuser "Lullify_Backend/internal/application/user"
	httphandler "Lullify_Backend/internal/infrastructure/http"
	"Lullify_Backend/internal/infrastructure/postgres"
	infrastream "Lullify_Backend/internal/infrastructure/stream"
	"Lullify_Backend/internal/infrastructure/token"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer pool.Close()

	// ── Repositories ───────────────────────────────────
	userRepo := postgres.NewUserRepository(pool)
	streamRepo := postgres.NewStreamRepository(pool)

	// ── Services ───────────────────────────────────────
	jwtService := token.NewJWTService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	streamEngine := infrastream.NewStreamEngine()

	// ── Use cases ──────────────────────────────────────
	registerUC := appuser.NewRegisterUseCase(userRepo)
	loginUC := appuser.NewLoginUseCase(userRepo)
	createStreamUC := appstream.NewCreateUseCase(streamRepo)
	startStreamUC := appstream.NewStartUseCase(streamRepo, streamEngine)
	stopStreamUC := appstream.NewStopUseCase(streamRepo, streamEngine)

	// ── Handlers ───────────────────────────────────────
	authHandler := httphandler.NewAuthHandler(registerUC, loginUC, userRepo, jwtService)
	streamHandler := httphandler.NewStreamHandler(
		createStreamUC,
		startStreamUC,
		stopStreamUC,
		streamRepo,
		jwtService,
		streamEngine,
	)

	// ── Router ─────────────────────────────────────────
	router := httphandler.NewRouter(authHandler, streamHandler)

	log.Printf("Lullify listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
