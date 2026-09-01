package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	infraredis "Lullify_Backend/internal/infrastructure/redis"
)

var (
	testDB    *pgxpool.Pool
	testRedis *infraredis.Client
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://lullify:lullify_dev@localhost:5432/lullify?sslmode=disable"
	}

	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	ctx := context.Background()

	var err error
	testDB, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		// Skip all integration tests if DB not available
		os.Exit(0)
	}

	if err := testDB.Ping(ctx); err != nil {
		os.Exit(0)
	}

	testRedis, err = infraredis.NewClient(redisURL)
	if err != nil {
		os.Exit(0)
	}

	code := m.Run()

	testDB.Close()
	_ = testRedis.Close()
	os.Exit(code)
}
