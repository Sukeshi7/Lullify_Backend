package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"Lullify_Backend/config"
	appuser "Lullify_Backend/internal/application/user"
	httphandler "Lullify_Backend/internal/infrastructure/http"
	"Lullify_Backend/internal/infrastructure/postgres"
	"Lullify_Backend/internal/infrastructure/token"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	jwtService := token.NewJWTService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	registerUC := appuser.NewRegisterUseCase(userRepo)
	loginUC := appuser.NewLoginUseCase(userRepo)

	authHandler := httphandler.NewAuthHandler(registerUC, loginUC, userRepo, jwtService)

	router := httphandler.NewRouter(authHandler)

	log.Printf("Lullify listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
