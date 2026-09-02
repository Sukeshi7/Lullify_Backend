package token_test

import (
	"strings"
	"testing"
	"time"

	"Lullify_Backend/internal/domain/user"
	"Lullify_Backend/internal/infrastructure/token"

	"github.com/google/uuid"
)

func newService() *token.JWTService {
	return token.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		7*24*time.Hour,
	)
}

func newUser() *user.User {
	return &user.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     user.RoleUser,
	}
}

func TestGenerateAndParseAccess(t *testing.T) {
	svc := newService()
	u := newUser()

	access, _, err := svc.GenerateTokens(u)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	claims, err := svc.ParseAccess(access)
	if err != nil {
		t.Fatalf("ParseAccess error: %v", err)
	}

	if claims.UserID != u.ID {
		t.Errorf("expected UserID %s, got %s", u.ID, claims.UserID)
	}
	if claims.Role != u.Role {
		t.Errorf("expected role %s, got %s", u.Role, claims.Role)
	}
}

func TestGenerateAndParseRefresh(t *testing.T) {
	svc := newService()
	u := newUser()

	_, refresh, err := svc.GenerateTokens(u)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	claims, err := svc.ParseRefresh(refresh)
	if err != nil {
		t.Fatalf("ParseRefresh error: %v", err)
	}

	if claims.UserID != u.ID {
		t.Errorf("expected UserID %s, got %s", u.ID, claims.UserID)
	}
}

func TestParseAccess_InvalidToken(t *testing.T) {
	svc := newService()
	_, err := svc.ParseAccess("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseRefresh_InvalidToken(t *testing.T) {
	svc := newService()
	_, err := svc.ParseRefresh("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseAccess_WithRefreshToken(t *testing.T) {
	svc := newService()
	u := newUser()

	_, refresh, err := svc.GenerateTokens(u)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	_, err = svc.ParseAccess(refresh)
	if err == nil {
		t.Fatal("expected error when using refresh token as access token, got nil")
	}
}

func TestParseRefresh_WithAccessToken(t *testing.T) {
	svc := newService()
	u := newUser()

	access, _, err := svc.GenerateTokens(u)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	_, err = svc.ParseRefresh(access)
	if err == nil {
		t.Fatal("expected error when using access token as refresh token, got nil")
	}
}

func TestGenerateTokens_BroadcasterRole(t *testing.T) {
	svc := newService()
	u := &user.User{
		ID:       uuid.New(),
		Username: "broadcaster",
		Email:    "broadcaster@example.com",
		Role:     user.RoleBroadcaster,
	}

	access, _, err := svc.GenerateTokens(u)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	claims, err := svc.ParseAccess(access)
	if err != nil {
		t.Fatalf("ParseAccess error: %v", err)
	}

	if claims.Role != user.RoleBroadcaster {
		t.Errorf("expected role broadcaster, got %s", claims.Role)
	}
}

func TestGenerateTokens_AdminRole(t *testing.T) {
	svc := newService()
	u := &user.User{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@example.com",
		Role:     user.RoleAdmin,
	}

	access, _, err := svc.GenerateTokens(u)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	claims, err := svc.ParseAccess(access)
	if err != nil {
		t.Fatalf("ParseAccess error: %v", err)
	}

	if claims.Role != user.RoleAdmin {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestParseAccess_EmptyString(t *testing.T) {
	svc := newService()
	_, err := svc.ParseAccess("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestParseRefresh_EmptyString(t *testing.T) {
	svc := newService()
	_, err := svc.ParseRefresh("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestGenerateTokens_MultipleUsers(t *testing.T) {
	svc := newService()

	users := []*user.User{
		{ID: uuid.New(), Username: "u1", Email: "u1@test.com", Role: user.RoleUser},
		{ID: uuid.New(), Username: "u2", Email: "u2@test.com", Role: user.RoleBroadcaster},
		{ID: uuid.New(), Username: "u3", Email: "u3@test.com", Role: user.RoleAdmin},
	}

	for _, u := range users {
		access, refresh, err := svc.GenerateTokens(u)
		if err != nil {
			t.Fatalf("GenerateTokens error for %s: %v", u.Username, err)
		}

		claims, err := svc.ParseAccess(access)
		if err != nil {
			t.Fatalf("ParseAccess error for %s: %v", u.Username, err)
		}
		if claims.UserID != u.ID {
			t.Errorf("user %s: expected ID %s, got %s", u.Username, u.ID, claims.UserID)
		}

		refreshClaims, err := svc.ParseRefresh(refresh)
		if err != nil {
			t.Fatalf("ParseRefresh error for %s: %v", u.Username, err)
		}
		if refreshClaims.UserID != u.ID {
			t.Errorf("user %s: refresh token ID mismatch", u.Username)
		}
	}
}

func TestNewJWTService(t *testing.T) {
	svc := token.NewJWTService("access", "refresh", 15*time.Minute, 7*24*time.Hour)
	if svc == nil {
		t.Fatal("expected non-nil JWTService")
	}
}

func TestParseAccess_WrongSecret(t *testing.T) {
	other := token.NewJWTService("wrong", "wrong", time.Minute, time.Hour)
	u := &user.User{ID: uuid.New(), Role: user.RoleUser}
	access, _, _ := other.GenerateTokens(u)

	svc := newService()
	_, err := svc.ParseAccess(access)
	if err == nil {
		t.Fatal("expected error for token signed with wrong secret")
	}
}

func TestParseAccess_Malformed(t *testing.T) {
	svc := newService()
	cases := []string{
		"not.a.jwt",
		"",
		strings.Repeat("a", 500),
	}
	for _, c := range cases {
		_, err := svc.ParseAccess(c)
		if err == nil {
			t.Errorf("expected error for malformed token, got nil")
		}
	}
}
