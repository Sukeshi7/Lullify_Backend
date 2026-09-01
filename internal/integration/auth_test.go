package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	appuser "Lullify_Backend/internal/application/user"
	"Lullify_Backend/internal/infrastructure/postgres"
	"Lullify_Backend/internal/infrastructure/token"
)

func TestAuthFlow_RegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB)
	jwtSvc := token.NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)

	registerUC := appuser.NewRegisterUseCase(userRepo)
	loginUC := appuser.NewLoginUseCase(userRepo)

	email := "integration_" + uuid.New().String()[:8] + "@test.com"

	// Register
	u, err := registerUC.Execute(ctx, appuser.RegisterInput{
		Username: "intuser",
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}

	// Login
	logged, err := loginUC.Execute(ctx, appuser.LoginInput{
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if logged.ID != u.ID {
		t.Errorf("expected same user ID after login")
	}

	// Generate tokens
	access, refresh, err := jwtSvc.GenerateTokens(logged)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	// Parse access token
	claims, err := jwtSvc.ParseAccess(access)
	if err != nil {
		t.Fatalf("ParseAccess error: %v", err)
	}
	if claims.UserID != u.ID {
		t.Errorf("expected UserID %s, got %s", u.ID, claims.UserID)
	}

	// Parse refresh token
	refreshClaims, err := jwtSvc.ParseRefresh(refresh)
	if err != nil {
		t.Fatalf("ParseRefresh error: %v", err)
	}
	if refreshClaims.UserID != u.ID {
		t.Errorf("expected UserID %s, got %s", u.ID, refreshClaims.UserID)
	}

	// Cleanup
	_ = userRepo.DeleteByID(ctx, u.ID)
}

func TestAuthFlow_WrongPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB)
	registerUC := appuser.NewRegisterUseCase(userRepo)
	loginUC := appuser.NewLoginUseCase(userRepo)

	email := "integration_wrong_" + uuid.New().String()[:8] + "@test.com"

	u, err := registerUC.Execute(ctx, appuser.RegisterInput{
		Username: "wrongpwduser",
		Email:    email,
		Password: "correctpassword",
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	defer userRepo.DeleteByID(ctx, u.ID) //nolint:errcheck

	_, err = loginUC.Execute(ctx, appuser.LoginInput{
		Email:    email,
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestAuthFlow_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB)
	registerUC := appuser.NewRegisterUseCase(userRepo)

	email := "integration_dup_" + uuid.New().String()[:8] + "@test.com"

	u, err := registerUC.Execute(ctx, appuser.RegisterInput{
		Username: "dupuser1",
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("First register error: %v", err)
	}
	defer userRepo.DeleteByID(ctx, u.ID) //nolint:errcheck

	_, err = registerUC.Execute(ctx, appuser.RegisterInput{
		Username: "dupuser2",
		Email:    email,
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}
