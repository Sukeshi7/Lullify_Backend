package user_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/user"
)

func TestUser_Fields(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	u := &user.User{
		ID:        id,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      user.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if u.ID != id {
		t.Errorf("expected ID %s, got %s", id, u.ID)
	}
	if u.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", u.Username)
	}
	if u.Role != user.RoleUser {
		t.Errorf("expected role user, got %s", u.Role)
	}
}

func TestRole_Constants(t *testing.T) {
	if user.RoleUser != "user" {
		t.Errorf("expected RoleUser to be 'user', got %s", user.RoleUser)
	}
	if user.RoleBroadcaster != "broadcaster" {
		t.Errorf("expected RoleBroadcaster to be 'broadcaster', got %s", user.RoleBroadcaster)
	}
	if user.RoleAdmin != "admin" {
		t.Errorf("expected RoleAdmin to be 'admin', got %s", user.RoleAdmin)
	}
}

func TestUser_Roles(t *testing.T) {
	cases := []struct {
		role     user.Role
		expected string
	}{
		{user.RoleUser, "user"},
		{user.RoleBroadcaster, "broadcaster"},
		{user.RoleAdmin, "admin"},
	}

	for _, c := range cases {
		if string(c.role) != c.expected {
			t.Errorf("expected role %s, got %s", c.expected, c.role)
		}
	}
}
