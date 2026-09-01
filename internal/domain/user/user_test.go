package user_test

import (
	"testing"

	"Lullify_Backend/internal/domain/user"
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
