package user_test

import (
	"testing"

	"Lullify_Backend/internal/domain/user"
	"github.com/google/uuid"
)

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

func TestUser_Role(t *testing.T) {
	u := &user.User{
		ID:   uuid.New(),
		Role: user.RoleAdmin,
	}
	if u.Role != user.RoleAdmin {
		t.Errorf("expected role admin, got %s", u.Role)
	}
}
