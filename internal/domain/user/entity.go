package user

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser        Role = "user"
	RoleBroadcaster Role = "broadcaster"
	RoleAdmin       Role = "admin"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
