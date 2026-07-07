package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}
