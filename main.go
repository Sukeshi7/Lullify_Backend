package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/config"
	appstream "Lullify_Backend/internal/application/stream"
	apptrack "Lullify_Backend/internal/application/track"
	appuser "Lullify_Backend/internal/application/user"
	httphandler "Lullify_Backend/internal/infrastructure/http"
	"Lullify_Backend/internal/infrastructure/postgres"
	"Lullify_Backend/internal/infrastructure/storage"
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
		log.Fatalf("cannot initialize object storage: %v", err)
	}

	userRepo := postgres.NewUserRepository(pool)
	streamRepo := postgres.NewStreamRepository(pool)
	playlistRepo := postgres.NewPlaylistRepository(pool)
	trackRepo := postgres.NewTrackRepository(pool)

	jwtService := token.NewJWTService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	streamEngine := infrastream.NewStreamEngine()

	registerUC := appuser.NewRegisterUseCase(userRepo)
	loginUC := appuser.NewLoginUseCase(userRepo)
	createStreamUC := appstream.NewCreateUseCase(streamRepo)
	startStreamUC := appstream.NewStartUseCase(streamRepo, streamEngine)
	stopStreamUC := appstream.NewStopUseCase(streamRepo, streamEngine)
	uploadTrackUC := apptrack.NewUploadUseCase(trackRepo, playlistRepo, objectStorage, cfg.MaxUploadSizeBytes)

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

	router := httphandler.NewRouter(authHandler, streamHandler, playlistHandler, trackHandler)

	log.Printf("Lullify listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
