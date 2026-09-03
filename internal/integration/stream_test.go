package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	appstream "Lullify_Backend/internal/application/stream"
	appuser "Lullify_Backend/internal/application/user"
	"Lullify_Backend/internal/infrastructure/observability"
	"Lullify_Backend/internal/infrastructure/postgres"
	infrastream "Lullify_Backend/internal/infrastructure/stream"
)

func init() {
	observability.InitLogger("integration-test")
}

func TestStreamFlow_CreateAndStart(t *testing.T) {
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB)
	streamRepo := postgres.NewStreamRepository(testDB)
	trackRepo := postgres.NewTrackRepository(testDB)
	engine := infrastream.NewStreamEngine(testRedis)

	registerUC := appuser.NewRegisterUseCase(userRepo)
	createStreamUC := appstream.NewCreateUseCase(streamRepo)
	startStreamUC := appstream.NewStartUseCase(streamRepo, engine, testRedis, trackRepo, "")
	stopStreamUC := appstream.NewStopUseCase(streamRepo, engine)

	// Crée un utilisateur broadcaster
	email := "broadcaster_" + uuid.New().String()[:8] + "@test.com"
	u, err := registerUC.Execute(ctx, appuser.RegisterInput{
		Username: "broadcaster_int",
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	defer userRepo.DeleteByID(ctx, u.ID) //nolint:errcheck

	// Crée un stream
	s, err := createStreamUC.Execute(ctx, appstream.CreateInput{
		OwnerID:    u.ID,
		Title:      "Integration Test Stream",
		MountPoint: "int-test-" + uuid.New().String()[:8],
	})
	if err != nil {
		t.Fatalf("Create stream error: %v", err)
	}
	defer streamRepo.Delete(ctx, s.ID) //nolint:errcheck

	// Démarre le stream
	err = startStreamUC.Execute(ctx, appstream.StartInput{
		StreamID: s.ID,
		OwnerID:  u.ID,
	})
	if err != nil {
		t.Fatalf("Start stream error: %v", err)
	}

	if !engine.IsRunning(s.ID) {
		t.Error("expected stream to be running")
	}

	// Arrête le stream
	err = stopStreamUC.Execute(ctx, appstream.StopInput{
		StreamID: s.ID,
		OwnerID:  u.ID,
	})
	if err != nil {
		t.Fatalf("Stop stream error: %v", err)
	}
}

func TestStreamFlow_Create_DuplicateMountPoint(t *testing.T) {
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB)
	streamRepo := postgres.NewStreamRepository(testDB)

	registerUC := appuser.NewRegisterUseCase(userRepo)
	createStreamUC := appstream.NewCreateUseCase(streamRepo)

	email := "broadcaster2_" + uuid.New().String()[:8] + "@test.com"
	u, err := registerUC.Execute(ctx, appuser.RegisterInput{
		Username: "broadcaster_int2",
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	defer userRepo.DeleteByID(ctx, u.ID) //nolint:errcheck

	mountPoint := "dup-mount-" + uuid.New().String()[:8]

	s, err := createStreamUC.Execute(ctx, appstream.CreateInput{
		OwnerID:    u.ID,
		Title:      "Stream 1",
		MountPoint: mountPoint,
	})
	if err != nil {
		t.Fatalf("Create stream 1 error: %v", err)
	}
	defer streamRepo.Delete(ctx, s.ID) //nolint:errcheck

	_, err = createStreamUC.Execute(ctx, appstream.CreateInput{
		OwnerID:    u.ID,
		Title:      "Stream 2",
		MountPoint: mountPoint,
	})
	if err == nil {
		t.Fatal("expected error for duplicate mount point, got nil")
	}
}
