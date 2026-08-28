package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/config"
	appfavorite "Lullify_Backend/internal/application/favorite"
	apphistory "Lullify_Backend/internal/application/history"
	appstream "Lullify_Backend/internal/application/stream"
	apptrack "Lullify_Backend/internal/application/track"
	appuser "Lullify_Backend/internal/application/user"
	httphandler "Lullify_Backend/internal/infrastructure/http"
	"Lullify_Backend/internal/infrastructure/observability"
	"Lullify_Backend/internal/infrastructure/postgres"
	infraredis "Lullify_Backend/internal/infrastructure/redis"
	"Lullify_Backend/internal/infrastructure/storage"
	infrastream "Lullify_Backend/internal/infrastructure/stream"
	"Lullify_Backend/internal/infrastructure/token"
)

func main() {
	cfg := config.Load()

	// ── Logger ─────────────────────────────────────────
	observability.InitLogger(cfg.OTELServiceName)

	// ── OTEL Tracer ────────────────────────────────────
	ctx := context.Background()
	shutdownTracer, err := observability.InitTracer(ctx, cfg.OTELServiceName, cfg.OTELEndpoint)
	if err != nil {
		observability.Logger.Fatal().Err(err).Msg("cannot initialize tracer")
	}
	defer func() {
		if shutdownErr := shutdownTracer(ctx); shutdownErr != nil {
			observability.Logger.Error().Err(shutdownErr).Msg("tracer shutdown error")
		}
	}()

	// ── PostgreSQL ─────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		observability.Logger.Fatal().Err(err).Msg("cannot connect to database")
	}
	defer pool.Close()

	// ── Redis ──────────────────────────────────────────
	redisClient, err := infraredis.NewClient(cfg.RedisURL)
	if err != nil {
		observability.Logger.Fatal().Err(err).Msg("cannot connect to redis")
	}
	defer func() { _ = redisClient.Close() }()

	// ── Storage ────────────────────────────────────────
	objectStorage, err := storage.New(storage.Options{
		Provider:  cfg.StorageProvider,
		LocalPath: cfg.StoragePath,
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		Bucket:    cfg.MinIOBucket,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		observability.Logger.Fatal().Err(err).Msg("cannot initialize object storage")
	}

	// ── Repositories ───────────────────────────────────
	userRepo := postgres.NewUserRepository(pool)
	streamRepo := postgres.NewStreamRepository(pool)
	playlistRepo := postgres.NewPlaylistRepository(pool)
	trackRepo := postgres.NewTrackRepository(pool)
	historyRepo := postgres.NewHistoryRepository(pool)
	favoriteRepo := postgres.NewFavoriteRepository(pool)

	// ── Services ───────────────────────────────────────
	jwtService := token.NewJWTService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	streamEngine := infrastream.NewStreamEngine()

	// ── Use cases ──────────────────────────────────────
	registerUC := appuser.NewRegisterUseCase(userRepo)
	loginUC := appuser.NewLoginUseCase(userRepo)
	createStreamUC := appstream.NewCreateUseCase(streamRepo)
	startStreamUC := appstream.NewStartUseCase(streamRepo, streamEngine, redisClient)
	stopStreamUC := appstream.NewStopUseCase(streamRepo, streamEngine)
	uploadTrackUC := apptrack.NewUploadUseCase(trackRepo, playlistRepo, objectStorage, cfg.MaxUploadSizeBytes)
	recordHistoryUC := apphistory.NewRecordUseCase(historyRepo)
	listHistoryUC := apphistory.NewListUseCase(historyRepo)
	addFavoriteUC := appfavorite.NewAddUseCase(favoriteRepo)
	removeFavoriteUC := appfavorite.NewRemoveUseCase(favoriteRepo)
	listFavoritesUC := appfavorite.NewListUseCase(favoriteRepo)

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
	playlistHandler := httphandler.NewPlaylistHandler(playlistRepo, trackRepo, jwtService)
	trackHandler := httphandler.NewTrackHandler(uploadTrackUC, jwtService, cfg.MaxUploadSizeBytes)
	historyHandler := httphandler.NewHistoryHandler(recordHistoryUC, listHistoryUC, jwtService)
	favoriteHandler := httphandler.NewFavoriteHandler(addFavoriteUC, removeFavoriteUC, listFavoritesUC, jwtService)
	adminHandler := httphandler.NewAdminHandler(userRepo, jwtService)
	healthHandler := httphandler.NewHealthHandler(pool, redisClient)

	// ── Router ─────────────────────────────────────────
	router := httphandler.NewRouter(
		authHandler,
		streamHandler,
		playlistHandler,
		trackHandler,
		historyHandler,
		favoriteHandler,
		adminHandler,
		healthHandler,
	)

	// ── Start server ───────────────────────────────────
	observability.Logger.Info().Str("port", cfg.Port).Msg("Lullify listening")
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		observability.Logger.Fatal().Err(err).Msg("server error")
	}
}
